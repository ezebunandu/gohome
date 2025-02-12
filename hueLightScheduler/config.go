package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

type light struct {
	Name string
}

type config struct {
	HueIPAddress string `yaml:"hue_ip_address"`
	HueID        string `yaml:"hue_id"`
	LightName    string `yaml:"light_name"`
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

	return &cfg, nil
}
