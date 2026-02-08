package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ezebunandu/hue-auto-schedule/pkg/scheduler"
	"github.com/ezebunandu/hue-auto-schedule/pkg/weather"
)

const BaseURL = "https://api.openweathermap.org"

func main() {
	ns := getenv("NAMESPACE", "gohome")
	location := getenv("WEATHER_LOCATION", "Calgary,CA")
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "error: OPENWEATHERMAP_API_KEY environment variable is required\n")
		os.Exit(1)
	}

	// Fetch weather forecast to get sunrise/sunset times
	forecast, err := fetchWeatherForecast(location, apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to fetch weather forecast: %v\n", err)
		os.Exit(1)
	}

	// Convert sunrise and sunset to cron syntax
	sunriseTimeCron := weather.UnixToCron(forecast.Sunrise)
	sunsetTimeCron := weather.UnixToCron(forecast.Sunset)

	fmt.Printf("Sunrise: %s (timestamp: %d)\n", sunriseTimeCron, forecast.Sunrise)
	fmt.Printf("Sunset: %s (timestamp: %d)\n", sunsetTimeCron, forecast.Sunset)

	sched, err := scheduler.NewScheduler(ns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Update sunrise CronJob schedule
	fmt.Printf("Updating sunrise CronJob schedule to %s\n", sunriseTimeCron)
	if err := sched.ModifyCronJobExecution(ctx, "sunrise", sunriseTimeCron); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to update sunrise CronJob: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully updated sunrise CronJob schedule\n")

	// Update sunset CronJob schedule
	fmt.Printf("Updating sunset CronJob schedule to %s\n", sunsetTimeCron)
	if err := sched.ModifyCronJobExecution(ctx, "sunset", sunsetTimeCron); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to update sunset CronJob: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully updated sunset CronJob schedule\n")
}

func fetchWeatherForecast(location, apiKey string) (weather.Forecast, error) {
	url := fmt.Sprintf("%s/data/2.5/weather?q=%s&appid=%s", BaseURL, location, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return weather.Forecast{}, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return weather.Forecast{}, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return weather.Forecast{}, fmt.Errorf("failed to read response body: %w", err)
	}

	forecast, err := weather.ParseResponse(data)
	if err != nil {
		return weather.Forecast{}, fmt.Errorf("failed to parse weather response: %w", err)
	}

	return forecast, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
