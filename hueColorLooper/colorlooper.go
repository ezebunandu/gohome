package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	hue "github.com/ezebunandu/gohue"
)

var HueID string
var HueIPAddress string

type Lights struct {
	Lights []string `json:"lights"`
}

// map the url-friendly name for lights to their actual names on the hue bridge
var lightMappings = map[string]string{
	"lamp_stand_1":   "Lamp Stand 1",
	"lamp_stand_2":   "Lamp Stand 2",
	"tv_strip_light": "TV Strip Light",
}

func startColorloop(l string) {
	log.Printf("INFO: starting colorloop for : %s", l)
	bridge, err := hue.NewBridge(HueIPAddress)
	if err != nil {
		log.Printf("ERROR: invalid hue ip address: %s\n", HueIPAddress)
	}
	if err = bridge.Login(HueID); err != nil {
		log.Printf("ERROR: invalid hue ip address: %s\n", HueID)
	}
	light, err := bridge.GetLightByName(l)
	if err != nil {
		log.Printf("ERROR: could not connect to light: %s\n", l)
	}
	if err = light.ColorLoop(true); err != nil {
		log.Printf("ERROR: could not activate colorloop: %#v\n", err)
	}
}

func newMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /colorloop/{light_name}", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received post request")
		light := r.PathValue("light_name")
		w.Header().Set("Content-Type", "application/json")
		l, ok := lightMappings[light]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("'%s' is not a valid light name\n", light)))
			log.Printf("ERROR: invalid light name '%s' is invalid\n", light)
			return
		} else {
			w.WriteHeader(http.StatusAccepted)
			log.Printf("INFO: starting colorloop for %s\n", light)
			w.Write([]byte(fmt.Sprintf("Received request to enable colorlooper for '%s'\n", light)))
			startColorloop(l)
		}
	})
	mux.HandleFunc("POST /colorloop/all", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received  request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Received request to enable colorlooper for all lights"))
		log.Println("INFO: starting colorloop for all lights")
		for _, l := range lightMappings {
			if l == "TV Strip Light" {
				continue
			}
			startColorloop(l)
		}
	})
	return mux
}

func main() {
	var ok bool
	HueID, ok = os.LookupEnv("HUE_ID")
	if !ok {
		log.Println("HUE_ID not set")
	}
	HueIPAddress, ok = os.LookupEnv("HUE_IP_ADDRESS")
	if !ok {
		log.Println("HUE_IP_ADDRESS not set")
	}
	if HueID == "" || HueIPAddress == "" {
		log.Fatal("HUE_ID and HUE_IP_ADDRESS environment variables must be set")
	}
	s := &http.Server{
		Addr:         ":3005",
		Handler:      newMux(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Println("ERROR:", err)
		os.Exit(1)
	}
}
