package api

import (
	"context"
	"fmt"
	"places-informer/models"
)

func (c *Client) GetWeather(ctx context.Context, location *models.Location) (models.Weather, error) {
	apiUrl := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&units=metric&appid=%s", location.Latitude, location.Longitude, c.openWeatherAPIKey)
	var response models.OpenWeatherResponse
	err := c.makeRequest(ctx, apiUrl, &response)
	if err != nil {
		return models.Weather{}, fmt.Errorf("openweather request failed: %w", err)

	}
	return models.Weather{Temp: response.Main.Temp, FeelsLike: response.Main.FeelsLike}, nil
}
