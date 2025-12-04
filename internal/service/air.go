package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"github.com/cuichanghe/daily-reminder-bot/pkg/qweather"
	"go.uber.org/zap"
)

// AirQualityService handles air quality-related business logic
type AirQualityService struct {
	client *qweather.Client
}

// NewAirQualityService creates a new AirQualityService
func NewAirQualityService(client *qweather.Client) *AirQualityService {
	return &AirQualityService{client: client}
}

// GetAirQualityReport generates a formatted air quality report for a city
func (s *AirQualityService) GetAirQualityReport(city string) (string, error) {
	logger.Debug("GetAirQualityReport called", zap.String("city", city))
	start := time.Now()

	// Get location
	logger.Debug("Fetching location", zap.String("city", city))
	location, err := s.client.GetLocation(city)
	if err != nil {
		logger.Error("Failed to get location",
			zap.String("city", city),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return "", fmt.Errorf("failed to get location: %w", err)
	}
	logger.Debug("Location retrieved",
		zap.String("city", city),
		zap.String("location_id", location.ID),
		zap.String("lat", location.Lat),
		zap.String("lon", location.Lon))

	// Get current air quality (v1)
	logger.Debug("Fetching current air quality",
		zap.String("city", city),
		zap.String("lat", location.Lat),
		zap.String("lon", location.Lon))
	airResp, err := s.client.GetAirQualityCurrent(location.Lat, location.Lon)
	if err != nil {
		logger.Error("Failed to get current air quality",
			zap.String("city", city),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return "", fmt.Errorf("failed to get current air quality: %w", err)
	}

	// Find primary index (prefer "qaqi" for China, or "us-epa", or first available)
	var mainIndex qweather.AirQualityIndex
	foundIndex := false
	for _, idx := range airResp.Indexes {
		if idx.Code == "qaqi" {
			mainIndex = idx
			foundIndex = true
			break
		}
	}
	if !foundIndex && len(airResp.Indexes) > 0 {
		mainIndex = airResp.Indexes[0]
		foundIndex = true
	}

	if !foundIndex {
		logger.Warn("No air quality index found", zap.String("city", city))
		return "", fmt.Errorf("no air quality index data available")
	}

	logger.Debug("Current air quality retrieved",
		zap.String("city", city),
		zap.Float64("aqi", mainIndex.Aqi),
		zap.String("category", mainIndex.Category))

	// Get air quality forecast (optional, non-critical)
	// Note: Still using v7 API for forecast as v1 forecast implementation was not requested/planned yet.
	// We use the location ID from GetLocation for this.
	var airForecast []qweather.AirDaily
	logger.Debug("Fetching air quality forecast",
		zap.String("city", city),
		zap.String("location_id", location.ID))
	airForecast, err = s.client.GetAirDaily(location.ID)
	if err != nil {
		logger.Warn("Failed to get air quality forecast",
			zap.String("city", city),
			zap.String("location_id", location.ID),
			zap.Error(err))
		airForecast = nil // Non-critical, continue without forecast
	} else {
		logger.Debug("Air quality forecast retrieved",
			zap.String("city", city),
			zap.Int("days", len(airForecast)))
	}

	// Build report
	var report strings.Builder
	report.WriteString(fmt.Sprintf("📊 %s 空气质量\n\n", city))

	// Current air quality
	report.WriteString("🌫️ 当前状况：\n")
	report.WriteString(fmt.Sprintf("   AQI：%.0f\n", mainIndex.Aqi))
	report.WriteString(fmt.Sprintf("   等级：%s\n", mainIndex.Level))
	report.WriteString(fmt.Sprintf("   类别：%s\n", mainIndex.Category))
	if mainIndex.PrimaryPollutant.Name != "" {
		report.WriteString(fmt.Sprintf("   主要污染物：%s\n", mainIndex.PrimaryPollutant.Name))
	}

	// Pollutant concentrations
	if len(airResp.Pollutants) > 0 {
		report.WriteString("\n💨 污染物浓度：\n")
		for _, p := range airResp.Pollutants {
			if p.Concentration.Value > 0 {
				report.WriteString(fmt.Sprintf("   %s：%.1f %s\n", p.Name, p.Concentration.Value, p.Concentration.Unit))
			}
		}
	}

	// Forecast (if available, show only next 2 days)
	if len(airForecast) > 0 {
		report.WriteString("\n📅 未来预报：\n")
		for i, forecast := range airForecast {
			if i >= 3 { // Show max 3 days
				break
			}
			if i == 0 {
				continue // Skip today, already shown in current status
			}
			dayLabel := "明天"
			if i == 2 {
				dayLabel = "后天"
			}
			report.WriteString(fmt.Sprintf("   %s：AQI %s（%s）\n", dayLabel, forecast.Aqi, forecast.Category))
		}
	}

	logger.Debug("Air quality report generated",
		zap.String("city", city),
		zap.Duration("duration", time.Since(start)))
	return report.String(), nil
}
