package qweather

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
)

// GetWarningNow retrieves current weather warnings for a location
func (c *Client) GetWarningNow(locationID string) ([]Warning, error) {
	logger.Debug("QWeather.GetWarningNow called", zap.String("location_id", locationID))
	start := time.Now()

	params := url.Values{}
	params.Add("location", locationID)
	params.Add("key", c.apiKey)

	requestURL := fmt.Sprintf("%s/v7/warning/now?%s", c.baseURL, params.Encode())
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
		return nil, fmt.Errorf("failed to get weather warnings: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var warningResp WarningResponse
	if err := json.NewDecoder(resp.Body).Decode(&warningResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode warning response: %w", err)
	}

	logger.Debug("QWeather API response",
		zap.String("code", warningResp.Code),
		zap.Int("warning_count", len(warningResp.Warning)))

	if warningResp.Code != "200" {
		// Code 204 means no active warnings, which is not an error
		if warningResp.Code == "204" {
			logger.Debug("No active warnings",
				zap.String("location_id", locationID),
				zap.Duration("duration", time.Since(start)))
			return []Warning{}, nil
		}
		logger.Warn("Weather warnings not available",
			zap.String("location_id", locationID),
			zap.String("api_code", warningResp.Code))
		return nil, fmt.Errorf("weather warnings not available: code %s", warningResp.Code)
	}

	logger.Debug("Weather warnings retrieved",
		zap.String("location_id", locationID),
		zap.Int("count", len(warningResp.Warning)),
		zap.Duration("duration", time.Since(start)))
	return warningResp.Warning, nil
}
