package main

import (
	"errors"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type yamlHour struct {
    t time.Time
}

func (yh *yamlHour) UnmarshalYAML(v *yaml.Node) error {
    if v.Kind != yaml.ScalarNode {
        return errors.New("value is not scaler")
    }
    var err error
    yh.t, err = time.Parse("3:04pm", v.Value)
    return err
}

type config struct {
    SwitchPinNumber int `yaml:"switch_pin_number"`
    NightStart yamlHour `yaml:"night_start"`
    NightEnd yamlHour `yaml:"night_end"`
}

func newConfig(configFile string) (*config, error) {
    cf, err := os.Open(configFile)
    if err != nil {
        return nil, err
    }
    defer cf.Close()

    var cfg *config
    if err := yaml.NewDecoder(cf).Decode(&cfg); err != nil {
        return nil, err
    }
    if cfg.SwitchPinNumber == 0 {
        log.Println(cfg)
        return nil, errors.New("switch pin needs to be defined")
    }
    if cfg.NightStart.t.IsZero() {
        var err error
        cfg.NightStart.t, err = time.Parse("3:04pm", "9:00pm")
        if err != nil {
            return nil, err
        }
    }
    if cfg.NightEnd.t.IsZero() {
        var err error
        cfg.NightEnd.t, err = time.Parse("3:04pm", "7:00am")
        if err != nil {
            return nil, err
        }
    }
    return cfg, nil
}