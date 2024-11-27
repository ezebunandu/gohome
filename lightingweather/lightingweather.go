package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"time"
    "gopkg.in/yaml.v3"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	owm "github.com/briandowns/openweathermap"
	hue "github.com/collinux/gohue"
)

//go:embed rootPage.html
var rootPageHTML []byte

var errInvalidColor = errors.New("invalid color")

type color struct {
    Color string `yaml:"color"`
    Threshold int `yaml:"threshold"`
}

var colorTranslate = map[string]*[2]float32{
    "blue": hue.BLUE,
    "cyan": hue.CYAN,
    "green": hue.GREEN,
    "orange": hue.ORANGE,
    "pink": hue.PINK,
    "purple": hue.PURPLE,
    "red": hue.RED,
    "white": hue.WHITE,
    "yellow": hue.YELLOW,
}

type config struct {
    Unit string `yaml:"unit"`
    Lang string `yaml:"lang"`
    Location string `yaml:"location"`
    HueID string `yaml:"hue_id"`
    HueIPAddress string `yaml:"hue_ip_address"`
    OWMAPIKey string `yaml:"owm_api_key"`
    LightName string `yaml:"light_name"`
    MaxColor string `yaml:"max_color"`
    Colors []color `yaml:"colors"`
}

func (cfg *config) sortColorRange() *config {
    sort.Slice(cfg.Colors, func(i, j int) bool {
        return cfg.Colors[i].Threshold < cfg.Colors[j].Threshold
    })
    return cfg
}

func newConfig(configFile string) (*config, error) {
    cf, err := os.Open(configFile)
    if err != nil {
        return nil, err
    }
    defer cf.Close()

    var cfg config

    if err := yaml.NewDecoder(cf).Decode(&cfg); err != nil {
        return nil, err
    }

    for _, cl := range cfg.Colors {
        if _, ok := colorTranslate[cl.Color]; !ok {
            return nil, fmt.Errorf("%w:%s", errInvalidColor, cl.Color)
        }
    }

    //Override OWM API Key with env var
    if ownKey, ok := os.LookupEnv("OWM_API_KEY"); ok {
        cfg.OWMAPIKey = ownKey
    }
    //Override HUE ID with env var
    if hueID, ok := os.LookupEnv("HUE_ID"); ok {
        cfg.HueID = hueID
    }

    return cfg.sortColorRange(), nil
}

func pickColor(cfg *config, curTemp int) *[2]float32 {
    for _, cl := range cfg.Colors {
        if curTemp < cl.Threshold {
            return colorTranslate[cl.Color]
        }
    }

    return colorTranslate[cfg.MaxColor]
}

func getCurrentTemperature(cfg *config) (int, error) {
    w, err := owm.NewCurrent(cfg.Unit, cfg.Lang, cfg.OWMAPIKey)

    if err != nil {
        return 0, err
    }

    err = w.CurrentByName(cfg.Location)

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

func lightweather(cfg *config, chRefresh <- chan struct{}) {
    externalWeatherTemp := promauto.NewGauge(prometheus.GaugeOpts{
        Name: "external_weather_temperature",
    })

    run := func ()  {
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
        case <- chRefresh:
            run()
        case <- time.Tick(30 * time.Minute):
            run()
        }
    }
}

func newMux(cfg *config) http.Handler {
    mux := http.NewServeMux()

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        log.Println("INFO: Received request to root")
        w.Write(rootPageHTML)
    })

    chRefresh := make(chan struct{}, 2)

    go lightweather(cfg, chRefresh)

    chRefresh <- struct{}{}

    mux.HandleFunc("POST /refresh", func(w http.ResponseWriter, r *http.Request) {
        log.Println("INFO: Received refresh request")
        chRefresh <- struct{}{}
        w.WriteHeader(http.StatusAccepted)
        w.Write([]byte("Refresh request accepted"))
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
        Addr: ":3040",
        Handler: newMux(cfg),
        ReadTimeout: 10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    if err := s.ListenAndServe(); err != nil {
        log.Println("ERROR:", err)
        os.Exit(1)
    }
}