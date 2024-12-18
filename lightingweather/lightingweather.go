package main

import (
	_ "embed"
	"flag"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	owm "github.com/briandowns/openweathermap"
	hue "github.com/ezebunandu/gohue"
)

//go:embed rootPage.html
var rootPageHTML []byte

func getCurrentTemperature(cfg *config) (int, error) {
	w, err := owm.NewCurrent(cfg.Unit, cfg.Lang, cfg.OWMAPIKey)

	if err != nil {
		return 0, err
	}

	err = w.CurrentByCoordinates(&owm.Coordinates{
		Longitude: cfg.Longitude,
		Latitude:  cfg.Latitude})

	return int(math.Round(w.Main.Temp)), err
}

func setLight(cfg *config, currentTemp int) error {
	bridge, err := hue.NewBridge(cfg.HueIPAddress)
	if err != nil {
		return err
	}

	if err := bridge.Login(cfg.HueID); err != nil {
		return err
	}

	weatherLight, err := bridge.GetLightByName(cfg.LightName)
	if err != nil {
		return err
	}

	if err := weatherLight.SetColor(
		pickColor(cfg, currentTemp)); err != nil {
		return err
	}
	return weatherLight.On()
}

func lightweather(cfg *config, chRefresh <-chan struct{}) {
	externalWeatherTemp := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "external_weather_temperature",
	})

	run := func() {
		log.Println("INFO: Gettting current temperature")
		currentTemp, err := getCurrentTemperature(cfg)
		if err != nil {
			log.Println("ERROR:", err)
		}

		externalWeatherTemp.Set(float64(currentTemp))

		log.Println("INFO: Setting light")
		if err := setLight(cfg, currentTemp); err != nil {
			log.Println("ERROR:", err)
		}
	}

	for {
		select {
		case <-chRefresh:
			run()
		case <-time.Tick(30 * time.Minute):
			run()
		}
	}
}

func turnOffLight(cfg *config) error {
	bridge, err := hue.NewBridge(cfg.HueIPAddress)
	if err != nil {
		return err
	}

	if err := bridge.Login(cfg.HueID); err != nil {
		return err
	}

	weatherLight, err := bridge.GetLightByName(cfg.LightName)
	if err != nil {
		return err
	}

	return weatherLight.Off()
}

func powerOffLight(cfg *config, chPowerOff <-chan struct{}) {
	run := func() {
		log.Println("INFO: Powering Off the Hue Light")
		err := turnOffLight(cfg)
		if err != nil {
			log.Println("ERROR:", err)
		}
	}

	for {
		select {
		case <-chPowerOff:
			run()
		case <-time.Tick(12 * time.Hour):
			run()
		}
	}
}

func newMux(cfg *config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		log.Println("INFO: Received request to root")
		w.Write(rootPageHTML)
	})

	chRefresh := make(chan struct{}, 2)
	chPowerOff := make(chan struct{}, 2)

	go lightweather(cfg, chRefresh)
	go powerOffLight(cfg, chPowerOff)

	chRefresh <- struct{}{}

	mux.HandleFunc("POST /refresh", func(w http.ResponseWriter, _ *http.Request) {
		log.Println("INFO: Received refresh request")
		chRefresh <- struct{}{}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Refresh request accepted"))
	})

	mux.HandleFunc("POST /powerHueOff", func(w http.ResponseWriter, r *http.Request) {
		log.Println("INFO: Received Poweroff request")
		chPowerOff <- struct{}{}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Poweroff request accepted"))
	})

	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

func main() {
	c := flag.String("c", "config.yml", "Config file")
	flag.Parse()

	cfg, err := newConfig(*c)

	if err != nil {
		log.Println("ERROR:", err)
		os.Exit(1)
	}

	s := &http.Server{
		Addr:         ":3040",
		Handler:      newMux(cfg),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Println("ERROR:", err)
		os.Exit(1)
	}
}
