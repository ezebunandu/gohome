package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	hue "github.com/ezebunandu/gohue"
)

var HueID = os.Getenv("HUE_ID")
var HueIPAddress = os.Getenv("HUE_IP_ADDRESS")


type Lights struct {
	Lights []string `json:"lights"`
}

var lightMappings = map[string]string{
	"lamp_stand_1": "Lamp Stand 1",
	"lamp_stand_2": "Lamp Stand 2",
    "tv_strip_light": "TV Strip Light",
}

func startColorloop(l string){
	log.Println("INFO: starting colorloop")
	bridge, err := hue.NewBridge(HueIPAddress)
	if err != nil {
		log.Printf("ERROR: invalid hue ip address: %s\n", HueIPAddress)
	}
	if err = bridge.Login(HueID); err != nil {
		log.Printf("ERROR: invalid hue ip address: %s\n", HueID)
	}
	light, err := bridge.GetLightByName(l)
	if err != nil {
		log.Printf("could not connect to light: %s", l)
	}
	if err = light.ColorLoop(true); err != nil {
		log.Printf("could not activate colorloop")
	}
}

func newMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("PATCH /colorlooper/{light_name}", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received post request")
		light := r.PathValue("light_name")

		l, ok := lightMappings[light]
		if !ok {
			w.Write([]byte(fmt.Sprintf("%s is not a valid light name", light)))
			log.Printf("ERROR: invalid light name: %s", light)
			return
		} else {
			w.Write([]byte(fmt.Sprintf("Received request to enable colorlooper for %s\n", light)))
			startColorloop(l)
		}
	})
	return mux
}

func main() {
	if HueID == "" || HueIPAddress == ""{
		log.Fatal("must supply hue details")
		os.Exit(1)
	}
	s := &http.Server{
		Addr:         ":3040",
		Handler:      newMux(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Println("ERROR:", err)
		os.Exit(1)
	}
}
