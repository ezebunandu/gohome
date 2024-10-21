package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

type state rpio.State

func (s state) String() string {
    if s == state(rpio.Low) {
        return "Open"
    }

    return "Closed"
}

func setupGPIO(pinNumber int) (rpio.Pin, error) {
    if err := rpio.Open(); err != nil {
        log.Println("Error opening GPIO:", err)
        return 0, err
    }

    pin := rpio.Pin(pinNumber)

    pin.Input()
    rpio.PullMode(pin, rpio.PullUp)
    return pin, nil
}

func getDoorState(pin rpio.Pin) state {
    return state(pin.Read())
}

func isNight(start, end time.Time) bool {
    cur := time.Now().Format("15:04")

    now, err := time.Parse("15:04", cur)
    if err != nil {
        log.Println(err)
        return false
    }
    if end.Before(start) {
        end = end.Add(24*time.Hour)
    }
    if now.Before(start) {
        now = now.Add(24 * time.Hour)
    }

    return now.After(start) && now.Before(end)
}

func sendNotification(discordWebhook, message string) {
    u, err := url.Parse(discordWebhook)
    if err != nil {
        log.Println("Invalid Discord webhook URL:", err)
        return
    }

    v := url.Values{}
    v.Set("wait", "true")
    u.RawQuery = v.Encode()

    payload := struct {
        Content string `json:"content"`
    }{
        Content: message,
    }
    
    var body bytes.Buffer
    if err := json.NewEncoder(&body).Encode(payload); err != nil {
        log.Println("Error creating JSON payload:", err)
        return
    }

    request, err := http.NewRequest(http.MethodPost, u.String(), &body)
    if err != nil {
        log.Println("Error creating Discord request:", err)
        return
    }
    request.Header.Add("Content-Type", "application/json")

    client := http.Client{
        Timeout: 10 * time.Second,
    }

    response, err := client.Do(request)
    if err != nil {
        log.Println("Error sending Discord request:", err)
        return
    } else {
        log.Println("Sent request to discord")
    }

    defer response.Body.Close()

    if response.StatusCode != http.StatusOK {
        log.Printf("invalid response from Discord channel: %s", response.Status)
    } else {
        log.Printf("Received a valid response from the Discord channel: %s", response.Status)
    }
}

func checkDoor(pin rpio.Pin, cfg *config, discordWebhookURL string) {
    for range time.Tick(1 * time.Minute) {
        doorState := getDoorState(pin)
        log.Println("Door state:", doorState)

        l, err := time.LoadLocation("America/Edmonton")
        if err != nil {
            log.Println("Error converting location:", err)
        }

        if doorState == state(rpio.Low) {
            if isNight(cfg.NightStart.t, cfg.NightEnd.t) {
                message := fmt.Sprint(
                    "Door open at night:", time.Now().In(l),
                )
                log.Println(message)
                go sendNotification(
                    discordWebhookURL, message,
                )
            } else {
                m := fmt.Sprint("Door is open at regular hour", time.Now().In(l))
                log.Println(m)
            }
        } else if doorState == state(rpio.High) {
            m := fmt.Sprint("Door is closed, no need to notification to discord")
            log.Println(m)
        }
        log.Println("Will check and log door state again in a minute.")
    }
}

func doorStateHandler(pin rpio.Pin) http.HandlerFunc {
    return func(w http.ResponseWriter, _ *http.Request) {
        doorState := getDoorState(pin)
        log.Println("Door state:", doorState)

        response := struct {
            DoorState state `json:"door_state"`
            DoorStateText string `json:"door_state_text"`
        } {
            DoorState: doorState,
            DoorStateText: fmt.Sprint(doorState),
        }
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(response); err != nil {
            log.Println("Error replying door state:", err)
        }
    }
}

func newMux(pin rpio.Pin) http.Handler {
    mux := http.NewServeMux()

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet {
            fmt.Fprintln(w, "Door status API running...")
        } else {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    })
    mux.Handle("/getdoor", doorStateHandler(pin))
    return mux
}

func main() {
    c := flag.String("c", "config.yml", "Config file")
    flag.Parse()

    cfg, err := newConfig(*c)
    if err != nil {
        log.Println("Error opening config file:", err)
        os.Exit(1)
    }

    discordWebhookURL, ok := os.LookupEnv("DISCORD_WEBHOOK_URL")
    if ! ok {
        log.Println("'DISCORD_WEBHOOK_URL' env is required")
        os.Exit(1)
    }

    pin, err := setupGPIO(cfg.SwitchPinNumber)
    if err != nil {
        log.Println("Error opening GPIO:", err)
        os.Exit(1)
    }

    defer rpio.Close()

    go checkDoor(pin, cfg, discordWebhookURL)

    s := &http.Server{
        Addr: ":3060",
        Handler: newMux(pin),
        WriteTimeout: 10 * time.Second,
    }

    log.Println("Starting API server on port: 3060")
    if err := s.ListenAndServe(); err != nil {
        log.Println(err)
        os.Exit(1)
    }
}