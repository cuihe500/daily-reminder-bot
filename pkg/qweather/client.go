package qweather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Client is a QWeather API client
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClient creates a new QWeather API client
func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// GetLocationID retrieves the location ID for a city name
func (c *Client) GetLocationID(city string) (string, error) {
	params := url.Values{}
	params.Add("location", city)
	params.Add("key", c.apiKey)

	url := fmt.Sprintf("%s/geo/v2/city/lookup?%s", c.baseURL, params.Encode())

	resp, err := c.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get location: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var geoResp GeoLocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return "", fmt.Errorf("failed to decode location response: %w", err)
	}

	if geoResp.Code != "200" || len(geoResp.Location) == 0 {
		return "", fmt.Errorf("location not found for city: %s", city)
	}

	return geoResp.Location[0].ID, nil
}

// GetCurrentWeather retrieves current weather for a location
func (c *Client) GetCurrentWeather(locationID string) (*CurrentWeather, error) {
	params := url.Values{}
	params.Add("location", locationID)
	params.Add("key", c.apiKey)

	url := fmt.Sprintf("%s/v7/weather/now?%s", c.baseURL, params.Encode())

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get weather: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	if weatherResp.Code != "200" {
		return nil, fmt.Errorf("weather API returned code: %s", weatherResp.Code)
	}

	return &weatherResp.Now, nil
}

// GetLifeIndices retrieves life indices (clothing, UV, sports, etc.) for a location
func (c *Client) GetLifeIndices(locationID string) ([]LifeIndex, error) {
	params := url.Values{}
	params.Add("location", locationID)
	params.Add("key", c.apiKey)
	params.Add("type", "0") // 0 = all indices

	url := fmt.Sprintf("%s/v7/indices/1d?%s", c.baseURL, params.Encode())

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get life indices: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var indicesResp LifeIndicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&indicesResp); err != nil {
		return nil, fmt.Errorf("failed to decode life indices response: %w", err)
	}

	if indicesResp.Code != "200" {
		return nil, fmt.Errorf("life indices API returned code: %s", indicesResp.Code)
	}

	return indicesResp.Daily, nil
}
