package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//go:embed rootPage.html
var rootPageHTML []byte

var errInvalidResponse = errors.New("unexpected response from the server")

type tempValues struct {
    TempC float64 `json:"tempC"`
    TempF float64 `json:"tempF"`
}

func (tv *tempValues) getTempValues(client *http.Client, url string) error {
    response, err := client.Get(url)
    if err != nil {
        log.Println(err)
        return err
    }
    defer response.Body.Close()

    if response.StatusCode != http.StatusOK {
        err := fmt.Errorf(
            "%w: invalid status code: %s",
            errInvalidResponse,
            response.Status,
        )
        log.Println(err)
        return err
    }
    if err := json.NewDecoder(response.Body).Decode(tv); err != nil {
        log.Println(err)
        return err
    }
    return nil
}

type metrics struct {
    results *tempValues
    up float64
    expire time.Time
    sync.RWMutex
}

func (m *metrics) getMetrics(client *http.Client, url string) *metrics {
    m.Lock()
    defer m.Unlock()

    if time.Now().Before(m.expire) {
        return m
    }
    m.up = 1
    if err := m.results.getTempValues(client, url); err != nil {
        m.up = 0
        m.results.TempC = 0
        m.results.TempF = 0
    }

    m.expire = time.Now().Add(2 * time.Second)
    return m
}

func (m *metrics) tempC() float64 {
    m.RLock()
    defer m.RUnlock()

    return m.results.TempC
}

func (m *metrics) tempF() float64 {
    m.RLock()
    defer m.RUnlock()

    return m.results.TempF
}

func (m *metrics) status() float64 {
    m.RLock()
    defer m.RUnlock()

    return m.up
}

func newMux(url string) http.Handler {
    mux := http.NewServeMux()

    client := &http.Client{
        Timeout: 10 * time.Second,
    }
    m := &metrics{
        results: &tempValues{},
    }
    promauto.NewGaugeFunc(
        prometheus.GaugeOpts{
            Name: "pico_temperature",
            Help: "Pico Sensor Temperature",
            ConstLabels: prometheus.Labels{"unit": "celsius"},
        },
        func() float64 {
            return m.getMetrics(client, url).tempC()
        },
    )
    promauto.NewGaugeFunc(
        prometheus.GaugeOpts{
            Name: "pico_temperature",
            Help: "Pico Sensor Temperature",
            ConstLabels: prometheus.Labels{"unit": "fahrenheit"},
        }, 
        func() float64 {
            return m.getMetrics(client, url).tempF()
        },
    )
    promauto.NewGaugeFunc(
        prometheus.GaugeOpts{
            Name: "pico_up",
            Help: "Pico Sensor Server Status.",
        },
        func() float64 {
            return m.getMetrics(client, url).status()
        },
    )
    mux.HandleFunc("/", func (w http.ResponseWriter, _ *http.Request) {
        w.Write(rootPageHTML)
    })
    mux.Handle("/metrics", promhttp.Handler())
    return mux
}

func main() {
    picoURL := os.Getenv("PICO_SERVER_URL")

    s := &http.Server{
        Addr: ":3030",
        Handler: newMux(picoURL),
        ReadTimeout: 10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    if err := s.ListenAndServe(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}