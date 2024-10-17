package main

import (
    "fmt"
    "os"

    "github.com/stianeikeland/go-rpio/v4"
)

const pinNumber = 12

type state rpio.State

func (s state) String() string {
    if s == state(rpio.Low) {
        return "Open"
    }

    return "Closed"
}

func main() {
    if err := rpio.Open(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    defer rpio.Close()

    pin := rpio.Pin(pinNumber)
    pin.Input()
    rpio.PullMode(pin, rpio.PullUp)
    door := state(pin.Read())

    fmt.Println("Door is: ", door)
}