package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	hue "github.com/ezebunandu/gohue"
)

type lightManager struct {
	cfg       *config
	isOn      bool
	bridge    *hue.Bridge
	lights    map[string]*hue.Light // Map of light names to light objects
	powerChan chan bool             // true for on, false for off
}

func newLightManager(cfg *config) (*lightManager, error) {
	bridge, err := hue.NewBridge(cfg.HueIPAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create bridge: %w", err)
	}

	if err := bridge.Login(cfg.HueID); err != nil {
		return nil, fmt.Errorf("failed to login to bridge: %w", err)
	}

	// Initialize lights map
	lights := make(map[string]*hue.Light)
	for _, lightConfig := range cfg.Lights {
		light, err := bridge.GetLightByName(lightConfig.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get light %s: %w", lightConfig.Name, err)
		}
		lights[lightConfig.Name] = &light
	}

	return &lightManager{
		cfg:       cfg,
		bridge:    bridge,
		lights:    lights,
		powerChan: make(chan bool, 2),
	}, nil
}

func (lm *lightManager) setState(on bool) {
	for name, light := range lm.lights {
		var err error
		if on {
			err = light.On()
		} else {
			err = light.Off()
		}

		if err != nil {
			log.Printf("Failed to set light %s state to %v: %v", name, on, err)
			continue
		}
	}

	lm.isOn = on
}

func (lm *lightManager) scheduleStateChanges() {
	for {
		now := time.Now()
		nextChange := lm.calculateNextChangeTime(now)
		wait := nextChange.Sub(now)

		log.Printf("Next state change scheduled for %v", nextChange.Format("15:04"))

		time.Sleep(wait)

		shouldBeOn := !isNight(lm.cfg.NightStart.t, lm.cfg.NightEnd.t)
		lm.setState(shouldBeOn)
	}
}

func (lm *lightManager) calculateNextChangeTime(now time.Time) time.Time {
	today := now.Truncate(24 * time.Hour)

	start := time.Date(today.Year(), today.Month(), today.Day(),
		lm.cfg.NightStart.t.Hour(), lm.cfg.NightStart.t.Minute(), 0, 0, today.Location())
	end := time.Date(today.Year(), today.Month(), today.Day(),
		lm.cfg.NightEnd.t.Hour(), lm.cfg.NightEnd.t.Minute(), 0, 0, today.Location())

	if end.Before(start) {
		end = end.Add(24 * time.Hour)
	}

	if now.Before(start) {
		return start
	} else if now.Before(end) {
		return end
	} else {
		// Next day's start time
		return start.Add(24 * time.Hour)
	}
}

func isNight(start, end time.Time) bool {
	now := time.Now()

	// Create today's date with the start and end times
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), start.Hour(), start.Minute(), 0, 0, now.Location())
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), end.Hour(), end.Minute(), 0, 0, now.Location())

	// Handle overnight periods (when end time is before start time)
	if end.Before(start) {
		if now.After(todayStart) {
			// We're after start time today, so end time should be tomorrow
			todayEnd = todayEnd.Add(24 * time.Hour)
		} else {
			// We're before end time today, so start time should be from yesterday
			todayStart = todayStart.Add(-24 * time.Hour)
		}
	}

	return now.After(todayStart) && now.Before(todayEnd)
}

func (lm *lightManager) run() {
	// First turn all lights off to ensure known state
	lm.setState(false)
	lm.isOn = false

	// Calculate initial state
	shouldBeOn := !isNight(lm.cfg.NightStart.t, lm.cfg.NightEnd.t)
	lm.setState(shouldBeOn)

	// Start scheduling goroutine
	go lm.scheduleStateChanges()

	// Handle manual override requests
	for newState := range lm.powerChan {
		lm.setState(newState)
	}
}

func newMux(cfg *config) (http.Handler, error) {
	mux := http.NewServeMux()

	lm, err := newLightManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create light manager: %w", err)
	}

	go lm.run()

	mux.HandleFunc("/turnOn", func(w http.ResponseWriter, _ *http.Request) {
		log.Println("INFO: Received request to turn on light")
		lm.powerChan <- true
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Turn on request accepted"))
	})

	mux.HandleFunc("/turnOff", func(w http.ResponseWriter, _ *http.Request) {
		log.Println("INFO: Received request to turn off light")
		lm.powerChan <- false
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Turn off request accepted"))
	})

	return mux, nil
}

func main() {
	c := flag.String("c", "config.yml", "Config file")
	flag.Parse()

	cfg, err := newConfig(*c)
	if err != nil {
		log.Println("ERROR:", err)
		os.Exit(1)
	}

	handler, err := newMux(cfg)
	if err != nil {
		log.Println("ERROR:", err)
		os.Exit(1)
	}

	s := &http.Server{
		Addr:         ":8100",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Println("ERROR:", err)
		os.Exit(1)
	}
}
