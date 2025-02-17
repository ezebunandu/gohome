package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	hue "github.com/ezebunandu/gohue"
)

func turnOffLight(cfg *config, chPowerOff <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if isNight(cfg.NightStart.t, cfg.NightEnd.t) {
				log.Println("INFO: Turning off light")
				bridge, err := hue.NewBridge(cfg.HueIPAddress)
				if err != nil {
					log.Println("ERROR:", err)
					continue
				}
				if err := bridge.Login(cfg.HueID); err != nil {
					log.Println("ERROR:", err)
					continue
				}
				for _, light := range(cfg.Lights) {
					weatherLight, err := bridge.GetLightByName(light.Name)
					if err != nil {
						log.Println("ERROR:", err)
					}
					if err := weatherLight.Off(); err != nil {
						log.Println("ERROR:", err)
						continue
					}
				}
			}
		case <-chPowerOff:
			log.Println("INFO: Turning off light")
			bridge, err := hue.NewBridge(cfg.HueIPAddress)
			if err != nil {
				log.Println("ERROR:", err)
				continue
			}
			if err := bridge.Login(cfg.HueID); err != nil {
				log.Println("ERROR:", err)
				continue
			}
			for _, light := range(cfg.Lights) {
				weatherLight, err := bridge.GetLightByName(light.Name)
				if err != nil {
					log.Println("ERROR:", err)
				}
				if err := weatherLight.Off(); err != nil {
					log.Println("ERROR:", err)
					continue
				}
			}
		}
	}
}

func turnOnLight(cfg *config, chPowerOn <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !isNight(cfg.NightStart.t, cfg.NightEnd.t) {
				log.Println("INFO: Turning on light")
				bridge, err := hue.NewBridge(cfg.HueIPAddress)
				if err != nil {
					log.Println("ERROR:", err)
					continue
				}
				if err := bridge.Login(cfg.HueID); err != nil {
					log.Println("ERROR:", err)
					continue
				}
				for _, light := range(cfg.Lights) {
					weatherLight, err := bridge.GetLightByName(light.Name)
					if err != nil {
						log.Println("ERROR:", err)
					}
					if err := weatherLight.On(); err != nil {
						log.Println("ERROR:", err)
						continue
					}
				}			}
		case <-chPowerOn:
			bridge, err := hue.NewBridge(cfg.HueIPAddress)
			if err != nil {
				log.Println("ERROR:", err)
				continue
			}
			if err := bridge.Login(cfg.HueID); err != nil {
				log.Println("ERROR:", err)
				continue
			}
			weatherLight, err := bridge.GetLightByName(cfg.LightName)
			if err != nil {
				log.Println("ERROR:", err)
				continue
			}
			if err := weatherLight.On(); err != nil {
				log.Println("ERROR:", err)
			}
		}
	}
}

func isNight(start, end time.Time) bool {
	cur := time.Now().Format("15:04")

	now, err := time.Parse("15:04", cur)
	if err != nil {
		log.Println(err)
		return false
	}
	if end.Before(start) {
		end = end.Add(24 * time.Hour)
	}
	if now.Before(start) {
		now = now.Add(24 * time.Hour)
	}

	return now.After(start) && now.Before(end)
}


func newMux(cfg *config) http.Handler {
	mux := http.NewServeMux()

	chPowerOff := make(chan struct{}, 2)
	chPowerOn := make(chan struct{}, 2)

	go turnOffLight(cfg, chPowerOff)
	go turnOnLight(cfg, chPowerOn)

	mux.HandleFunc("/turnOn", func(w http.ResponseWriter, _ *http.Request) {
		log.Println("INFO: Received request to turn on light")
		chPowerOn <- struct{}{}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Turn on request accepted"))
	})

	mux.HandleFunc("/turnOff", func(w http.ResponseWriter, _ *http.Request) {
		log.Println("INFO: Received request to turn off light")
		chPowerOff <- struct{}{}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Turn off request accepted"))
	})

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
		Addr:         ":8100",
		Handler:      newMux(cfg),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Println("ERROR:", err)
		os.Exit(1)
	}
}
