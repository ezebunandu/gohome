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
    Longitude float64 `yaml:"longitude"`
    Latitude float64 `yaml:"latitude"`
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

func (cfg *config) UnmarshalYAML(unmarshal func(interface{}) error) error {
    var raw struct {
        LongitudeStr string `yaml:"longitude"`
        LatitudeStr string `yaml:"latitude"`
        Unit string `yaml:"unit"`
        Lang string `yaml:"lang"`
        HueID string `yaml:"hue_id"`
        HueIPAddress string `yaml:"hue_ip_address"`
        OWMAPIKey string `yaml:"owm_api_key"`
        LightName string `yaml:"light_name"`
        MaxColor string `yaml:"max_color"`
        Colors []color `yaml:"colors"`
    }
    if err := unmarshal(&raw); err != nil {
        return err
    }

    longitude, err := strconv.ParseFloat(raw.LongitudeStr, 64)
    if err != nil || longitude < -180 || longitude > 180 {
        return fmt.Errorf("invalid longitude: %v (must be between -180 and 180)", raw.LatitudeStr)
    }

    latitude, err := strconv.ParseFloat(raw.LatitudeStr, 64)
    if err != nil  || latitude < -90 || latitude > 90{
        return fmt.Errorf("invalid latitude: %v (must be between -90 and 90)", raw.LatitudeStr)
    }
    cfg.Unit = raw.Unit
    cfg.Lang = raw.Lang
    cfg.Longitude = longitude
    cfg.Latitude = latitude
    cfg.HueID = raw.HueID
    cfg.HueIPAddress = raw.HueIPAddress
    cfg.OWMAPIKey = raw.OWMAPIKey
    cfg.LightName = raw.LightName
    cfg.MaxColor = raw.MaxColor
    cfg.Colors = raw.Colors

    return nil
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