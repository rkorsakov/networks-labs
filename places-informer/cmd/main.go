package main

import (
	"context"
	"fmt"
	"os"
	"places-informer/internal/api"
	"places-informer/internal/ui"
	"time"
)

const (
	timeout = 10 * time.Second
)

func main() {
	graphHopperKey := os.Getenv("GRAPH_HOPPER_KEY")
	openWeatherKey := os.Getenv("OPEN_WEATHER_KEY")
	openTripMapKey := os.Getenv("OPEN_TRIP_KEY")
	if graphHopperKey == "" || openWeatherKey == "" || openTripMapKey == "" {
		fmt.Println("Provide API keys")
		os.Exit(1)
	}
	apiClient := api.NewAPIClient(graphHopperKey, openWeatherKey, openTripMapKey)
	consoleUI := ui.NewConsoleUI()
	for {
		query := consoleUI.GetUserInput("Enter the location to search")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		locations, err := apiClient.SearchLocations(ctx, query)
		if err != nil {
			fmt.Println("Error searching locations:", err)
			continue
		}
		consoleUI.DisplayLocations(locations)
		location := consoleUI.ChooseLocation(locations)
		weather, err := apiClient.GetWeather(ctx, &location)
		if err != nil {
			fmt.Println("Error getting weather:", err)
		}
		consoleUI.DisplayWeather(weather)
		places, err := apiClient.GetPlacesWithDetails(ctx, location.Latitude, location.Longitude, 1000)
		if err != nil {
			fmt.Println("Error getting places:", err)
			continue
		}
		if len(places) == 0 {
			fmt.Println("No places found")
			continue
		}
		consoleUI.DisplayPlaces(places)
	}
}
