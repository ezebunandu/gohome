package main

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type light struct {
	Name string
}

type yamlHour struct {
	t time.Time
}

type config struct {
	HueIPAddress string `yaml:"hue_ip_address"`
	HueID        string `yaml:"hue_id"`
	LightName    string `yaml:"light_name"`
	NightStart yamlHour `yaml:"night_start"`
	NightEnd yamlHour `yaml:"night_end"`
}

func (yh *yamlHour) UnmarshalYAML(v *yaml.Node) error {
	if v.Kind != yaml.ScalarNode {
		return errors.New("not a scaler value")
	}
	var err error
	yh.t, err = time.Parse("4:04pm", v.Value)
	return err
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
	//Override HUE ID with env var
	if hueID, ok := os.LookupEnv("HUE_ID"); ok {
		cfg.HueID = hueID
	}

	if cfg.NightStart.t.IsZero(){
		var err error
		cfg.NightStart.t, err = time.Parse("4:04pm", "10:00pm") // night starts 10pm, if not defined in config file
		if err != nil {
			return nil, err
		}
	}

	if cfg.NightEnd.t.IsZero() {
		var err error
		cfg.NightEnd.t, err = time.Parse("4:04pm", "5:30am") // night ends at 5:30am, if not defined in config
		if err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}
