package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"sort"
    "strconv"
    "gopkg.in/yaml.v3"

	hue "github.com/collinux/gohue"
)

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
    LongitudeStr string `yaml:"longitude"`
    LatitudeStr string `yaml:"latitude"`
    Longitude float64 `yaml:"-"`
    Latitude float64 `yaml:"-"`
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

    cfg.Longitude, err = strconv.ParseFloat(cfg.LongitudeStr, 64)
    if err != nil {
        return nil, fmt.Errorf("invalid longitude value: %v", err)
    }

    cfg.Latitude, err = strconv.ParseFloat(cfg.LatitudeStr, 64)
    if err != nil {
        return nil, fmt.Errorf("invalid latitude value: %v", err)
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