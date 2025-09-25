package ui

import (
	"bufio"
	"fmt"
	"os"
	"places-informer/models"
	"strconv"
	"strings"
)

type ConsoleUI struct {
	scanner *bufio.Scanner
}

func NewConsoleUI() *ConsoleUI {
	return &ConsoleUI{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (ui *ConsoleUI) GetUserInput(prompt string) string {
	fmt.Println(prompt)
	ui.scanner.Scan()
	return strings.TrimSpace(ui.scanner.Text())
}

func (ui *ConsoleUI) DisplayLocations(locations []models.Location) {
	fmt.Println("\n=== Found locations ===")
	for i, loc := range locations {
		fmt.Printf("%d. %s\n", i+1, loc.Name)
		if loc.Address != "" && loc.Address != loc.Name {
			fmt.Printf("   Address: %s\n", loc.Address)
		}
		if loc.City != "" {
			fmt.Printf("   City: %s\n", loc.City)
		}
		if loc.Country != "" {
			fmt.Printf("   Country: %s\n", loc.Country)
		}
		fmt.Printf("   Coordinates: %.6f, %.6f\n", loc.Latitude, loc.Longitude)
	}
	fmt.Println()
}

func (ui *ConsoleUI) ChooseLocation(locations []models.Location) models.Location {
	fmt.Println("Choose location index:")
	ui.scanner.Scan()
	index, _ := strconv.Atoi(ui.scanner.Text())
	return locations[index-1]
}

func (ui *ConsoleUI) DisplayWeather(weather models.Weather) {
	fmt.Println("Temperature: ", weather.Temp)
	fmt.Println("Feels Like: ", weather.FeelsLike)
}

func (ui *ConsoleUI) DisplayPlaces(places []models.PlaceInfo) {
	fmt.Println("\n=== Points of interest nearby ===")
	for i, place := range places {
		fmt.Printf("\n%d. %s\n", i+1, place.Name)
		fmt.Printf("   Rating: %.1f\n", place.Rating)
		fmt.Printf("   Category: %s\n", place.Kinds)
		if place.Details != nil {
			if place.Details.Info.Descr != "" {
				desc := place.Details.Info.Descr
				if len(desc) > 120 {
					desc = desc[:120] + "..."
				}
				fmt.Printf("   Description: %s\n", desc)
			}
			if place.Details.Wikipedia != "" {
				fmt.Printf("   Wikipedia: %s\n", place.Details.Wikipedia)
			}
		}
	}
}
