package api

import (
	"context"
	"fmt"
	"places-informer/models"
)

func (c *Client) getNearbyPlaces(ctx context.Context, lat, lon float64, radius int) ([]models.PlaceInfo, error) {
	apiURL := fmt.Sprintf("https://api.opentripmap.com/0.1/ru/places/radius?radius=%d&lon=%f&lat=%f&format=geojson&apikey=%s",
		radius, lon, lat, c.openTripMapAPIKey)
	var response models.PlacesResponse
	if err := c.makeRequest(ctx, apiURL, &response); err != nil {
		return nil, fmt.Errorf("opentripmap places request failed: %w", err)
	}
	places := make([]models.PlaceInfo, 0, len(response.Features))
	for _, feature := range response.Features {
		places = append(places, models.PlaceInfo{
			ID:     feature.Properties.Xid,
			Name:   feature.Properties.Name,
			Rating: feature.Properties.Rate,
			Kinds:  feature.Properties.Kinds,
		})
	}

	return places, nil
}

func (c *Client) getPlaceDetails(ctx context.Context, xid string) (*models.PlaceDetailsResponse, error) {
	apiURL := fmt.Sprintf("https://api.opentripmap.com/0.1/ru/places/xid/%s?apikey=%s",
		xid, c.openTripMapAPIKey)
	var details models.PlaceDetailsResponse
	if err := c.makeRequest(ctx, apiURL, &details); err != nil {
		return nil, fmt.Errorf("opentripmap details request failed: %w", err)
	}
	return &details, nil
}

func (c *Client) GetPlacesWithDetails(ctx context.Context, lat, lon float64, radius int) ([]models.PlaceInfo, error) {
	places, err := c.getNearbyPlaces(ctx, lat, lon, radius)
	if err != nil {
		return nil, err
	}
	type result struct {
		index   int
		details *models.PlaceDetailsResponse
		err     error
	}
	results := make(chan result, len(places))
	for i, place := range places {
		go func(idx int, xid string) {
			details, err := c.getPlaceDetails(ctx, xid)
			results <- result{idx, details, err}
		}(i, place.ID)
	}
	placesWithDetails := make([]models.PlaceInfo, len(places))
	for i := 0; i < len(places); i++ {
		res := <-results
		if res.err == nil {
			places[res.index].Details = res.details
		}
		placesWithDetails[res.index] = places[res.index]
	}
	return placesWithDetails, nil
}
