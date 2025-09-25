package api

import (
	"context"
	"fmt"
	"net/url"
	"places-informer/models"
)

func (c *Client) SearchLocations(ctx context.Context, query string) ([]models.Location, error) {
	encodedQuery := url.QueryEscape(query)
	apiURL := fmt.Sprintf("https://graphhopper.com/api/1/geocode?q=%s&locale=ru&key=%s", encodedQuery, c.graphHopperAPIKey)
	var response models.GraphHopperResponse
	err := c.makeRequest(ctx, apiURL, &response)
	if err != nil {
		return nil, fmt.Errorf("graphhopper request failed: %w", err)
	}
	locations := make([]models.Location, 0, len(response.Hits))
	for _, hit := range response.Hits {
		address := hit.Name
		if hit.Street != "" {
			address = hit.Street
			if hit.Housenumber != "" {
				address += ", " + hit.Housenumber
			}
		}
		locations = append(locations, models.Location{
			Name:      hit.Name,
			Latitude:  hit.Point.Lat,
			Longitude: hit.Point.Lng,
			City:      hit.City,
			Country:   hit.Country,
			Address:   address,
		})
	}
	return locations, nil
}
