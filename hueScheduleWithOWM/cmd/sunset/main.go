package main

import (
	"fmt"
	"os"

	"github.com/ezebunandu/controller"
)

var Usage = `Usage: sunset <light>`

func main() {
	HueID, HueIPAddress := os.Getenv("HUE_ID"), os.Getenv("HUE_IP_ADDRESS")
	if len(os.Args) < 2 {
		fmt.Println(Usage)
		os.Exit(0)
	}
	light := os.Args[1]
	ctl, err := controller.NewController(HueID, HueIPAddress)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = ctl.EnsureOn(light)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
