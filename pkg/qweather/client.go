package qweather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
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
	logger.Debug("QWeather.GetLocationID called", zap.String("city", city))
	start := time.Now()

	params := url.Values{}
	params.Add("location", city)
	params.Add("key", c.apiKey)

	requestURL := fmt.Sprintf("%s/geo/v2/city/lookup?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.client.Get(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return "", fmt.Errorf("failed to get location: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var geoResp GeoLocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return "", fmt.Errorf("failed to decode location response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", geoResp.Code),
		zap.Int("location_count", len(geoResp.Location)))

	if geoResp.Code != "200" || len(geoResp.Location) == 0 {
		logger.Warn("Location not found",
			zap.String("city", city),
			zap.String("api_code", geoResp.Code))
		return "", fmt.Errorf("location not found for city: %s", city)
	}

	logger.Debug("Location ID retrieved",
		zap.String("city", city),
		zap.String("location_id", geoResp.Location[0].ID),
		zap.Duration("duration", time.Since(start)))
	return geoResp.Location[0].ID, nil
}

// GetCurrentWeather retrieves current weather for a location
func (c *Client) GetCurrentWeather(locationID string) (*CurrentWeather, error) {
	logger.Debug("QWeather.GetCurrentWeather called", zap.String("location_id", locationID))
	start := time.Now()

	params := url.Values{}
	params.Add("location", locationID)
	params.Add("key", c.apiKey)

	requestURL := fmt.Sprintf("%s/v7/weather/now?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.client.Get(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get weather: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", weatherResp.Code))

	if weatherResp.Code != "200" {
		logger.Warn("Weather API error",
			zap.String("location_id", locationID),
			zap.String("api_code", weatherResp.Code))
		return nil, fmt.Errorf("weather API returned code: %s", weatherResp.Code)
	}

	logger.Debug("Current weather retrieved",
		zap.String("location_id", locationID),
		zap.String("temp", weatherResp.Now.Temp),
		zap.String("text", weatherResp.Now.Text),
		zap.Duration("duration", time.Since(start)))
	return &weatherResp.Now, nil
}

// GetLifeIndices retrieves life indices (clothing, UV, sports, etc.) for a location
func (c *Client) GetLifeIndices(locationID string) ([]LifeIndex, error) {
	logger.Debug("QWeather.GetLifeIndices called", zap.String("location_id", locationID))
	start := time.Now()

	params := url.Values{}
	params.Add("location", locationID)
	params.Add("key", c.apiKey)
	params.Add("type", "0") // 0 = all indices

	requestURL := fmt.Sprintf("%s/v7/indices/1d?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.client.Get(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get life indices: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var indicesResp LifeIndicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&indicesResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode life indices response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", indicesResp.Code))

	if indicesResp.Code != "200" {
		logger.Warn("Life indices API error",
			zap.String("location_id", locationID),
			zap.String("api_code", indicesResp.Code))
		return nil, fmt.Errorf("life indices API returned code: %s", indicesResp.Code)
	}

	logger.Debug("Life indices retrieved",
		zap.String("location_id", locationID),
		zap.Int("indices_count", len(indicesResp.Daily)),
		zap.Duration("duration", time.Since(start)))
	return indicesResp.Daily, nil
}

// GetDailyForecast retrieves daily weather forecast for a location
func (c *Client) GetDailyForecast(locationID string) (*DailyForecast, error) {
	logger.Debug("QWeather.GetDailyForecast called", zap.String("location_id", locationID))
	start := time.Now()

	params := url.Values{}
	params.Add("location", locationID)
	params.Add("key", c.apiKey)

	requestURL := fmt.Sprintf("%s/v7/weather/3d?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.client.Get(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get daily forecast: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var forecastResp DailyForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&forecastResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode daily forecast response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", forecastResp.Code))

	if forecastResp.Code != "200" || len(forecastResp.Daily) == 0 {
		logger.Warn("Daily forecast API error",
			zap.String("location_id", locationID),
			zap.String("api_code", forecastResp.Code))
		return nil, fmt.Errorf("daily forecast API returned code: %s", forecastResp.Code)
	}

	logger.Debug("Daily forecast retrieved",
		zap.String("location_id", locationID),
		zap.String("tempMax", forecastResp.Daily[0].TempMax),
		zap.String("tempMin", forecastResp.Daily[0].TempMin),
		zap.Duration("duration", time.Since(start)))
	return &forecastResp.Daily[0], nil
}
