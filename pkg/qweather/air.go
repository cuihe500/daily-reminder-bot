package qweather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
)

// GetAirNow retrieves current air quality for a location
func (c *Client) GetAirNow(locationID string) (*AirNow, error) {
	logger.Debug("QWeather.GetAirNow called", zap.String("location_id", locationID))
	start := time.Now()

	params := url.Values{}
	params.Add("location", locationID)
	params.Add("key", c.apiKey)

	requestURL := fmt.Sprintf("%s/v7/air/now?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.doRequest(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get air quality: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	// Read and log response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", zap.Error(err))
		return nil, fmt.Errorf("failed to read air quality response: %w", err)
	}
	logger.Debug("Air quality raw response",
		zap.String("body", string(bodyBytes)))

	var airResp AirNowResponse
	if err := json.Unmarshal(bodyBytes, &airResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err),
			zap.String("body", string(bodyBytes)))
		return nil, fmt.Errorf("failed to decode air quality response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", airResp.Code))

	if airResp.Code != "200" {
		logger.Warn("Air quality data not available",
			zap.String("location_id", locationID),
			zap.String("api_code", airResp.Code))
		return nil, fmt.Errorf("air quality data not available: code %s", airResp.Code)
	}

	logger.Debug("Air quality retrieved",
		zap.String("location_id", locationID),
		zap.String("aqi", airResp.Now.Aqi),
		zap.String("category", airResp.Now.Category),
		zap.Duration("duration", time.Since(start)))
	return &airResp.Now, nil
}

// GetAirDaily retrieves daily air quality forecast for a location
func (c *Client) GetAirDaily(locationID string) ([]AirDaily, error) {
	logger.Debug("QWeather.GetAirDaily called", zap.String("location_id", locationID))
	start := time.Now()

	params := url.Values{}
	params.Add("location", locationID)
	params.Add("key", c.apiKey)

	requestURL := fmt.Sprintf("%s/v7/air/5d?%s", c.baseURL, params.Encode())
	maskedURL := logger.MaskURL(requestURL)

	logger.Debug("Sending HTTP request",
		zap.String("url", maskedURL),
		zap.String("method", "GET"))

	resp, err := c.doRequest(requestURL)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", maskedURL),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get air quality forecast: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", zap.Error(err))
		return nil, fmt.Errorf("failed to read air quality forecast response: %w", err)
	}

	var airResp AirDailyResponse
	if err := json.Unmarshal(bodyBytes, &airResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err),
			zap.String("body", string(bodyBytes)))
		return nil, fmt.Errorf("failed to decode air quality forecast response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", airResp.Code),
		zap.Int("forecast_count", len(airResp.Daily)))

	if airResp.Code != "200" {
		logger.Warn("Air quality forecast not available",
			zap.String("location_id", locationID),
			zap.String("api_code", airResp.Code))
		return nil, fmt.Errorf("air quality forecast not available: code %s", airResp.Code)
	}

	logger.Debug("Air quality forecast retrieved",
		zap.String("location_id", locationID),
		zap.Int("days", len(airResp.Daily)),
		zap.Duration("duration", time.Since(start)))
	return airResp.Daily, nil
}
