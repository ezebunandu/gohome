package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ezebunandu/hue-auto-schedule/pkg/k8s"
	weather "github.com/ezebunandu/hue-auto-schedule/pkg/owm"
)

const BaseURL = "https://api.openweathermap.org"

func main() {
	ns := getenv("NAMESPACE", "default")
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
	sunriseCron := weather.UnixToCron(forecast.Sunrise)
	sunsetCron := weather.UnixToCron(forecast.Sunset)

	fmt.Printf("Sunrise: %s (timestamp: %d)\n", sunriseCron, forecast.Sunrise)
	fmt.Printf("Sunset: %s (timestamp: %d)\n", sunsetCron, forecast.Sunset)

	scheduler, err := k8s.NewScheduler()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Update sunrise CronJob schedule
	fmt.Printf("Updating sunrise CronJob schedule to %s\n", sunriseCron)
	if err := scheduler.ModifyCronJobExecution(ctx, ns, "sunrise", sunriseCron); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to update sunrise CronJob: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully updated sunrise CronJob schedule\n")

	// Update sunset CronJob schedule
	fmt.Printf("Updating sunset CronJob schedule to %s\n", sunsetCron)
	if err := scheduler.ModifyCronJobExecution(ctx, ns, "sunset", sunsetCron); err != nil {
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
