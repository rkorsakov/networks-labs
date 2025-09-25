package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	client            HTTPClient
	graphHopperAPIKey string
	openWeatherAPIKey string
	openTripMapAPIKey string
}

func NewAPIClient(graphHopperKey, openWeatherKey, openTripMapKey string) *Client {
	return &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		graphHopperAPIKey: graphHopperKey,
		openWeatherAPIKey: openWeatherKey,
		openTripMapAPIKey: openTripMapKey,
	}
}

func (c *Client) makeRequest(ctx context.Context, url string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decoding JSON: %w", err)
	}

	return nil
}
