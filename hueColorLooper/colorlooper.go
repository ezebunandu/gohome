package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// webserver listening at a given port
// when it receives a post request with data containing a list of light names
// it searches for the names in a config field
// if it finds at list one matching name, it puts that light into a colorloop
// and responds with the name of the light put into colorloop and a response message
// if it doesn't find at least one matching light, it responds with an error message for the
// list of light names passed with the request

// takes a config struct and a channel
// when a message is received on the channel
// along with a list of light names
// set each light in the list to colorloop mode
// func colorlooper() {}

// request multiplexer than handles a /startloop endpoint
// when a post request is received, along with a list of light names as data with the request
// search the list of lights in the config
// if it finds any that contains the name string
// save the matching name to a slice of names
// call the colorlooper function with that slice of name
// or pass a message to a goroutine in some way that activates the colorlooper

// start a http server in main
// parse for a config file and initializes a struct from the file contents
// listen on port 8888

type Lights struct {
	Lights []string `json:"lights"`
}

func colorlooper(cfg *config, chStartColorLoop <- chan struct{}, lights []string){
}

func newMux() http.Handler {
	mux := http.NewServeMux()

	var l Lights

	mux.HandleFunc("POST /colorlooper", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received post request")
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&l)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(l.Lights) == 0 {
			http.Error(w, "missing or empty 'lights' field in request", http.StatusBadRequest)
			return
		}
		log.Printf("Received request with payload: %+v", l.Lights)
		w.Write([]byte("Received post request to colorlooper"))
	})
	return mux
}

func main() {
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
