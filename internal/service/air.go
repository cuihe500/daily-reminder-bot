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

	// Get location ID
	logger.Debug("Fetching location ID", zap.String("city", city))
	locationID, err := s.client.GetLocationID(city)
	if err != nil {
		logger.Error("Failed to get location ID",
			zap.String("city", city),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return "", fmt.Errorf("failed to get location ID: %w", err)
	}
	logger.Debug("Location ID retrieved",
		zap.String("city", city),
		zap.String("location_id", locationID))

	// Get current air quality
	logger.Debug("Fetching current air quality",
		zap.String("city", city),
		zap.String("location_id", locationID))
	airNow, err := s.client.GetAirNow(locationID)
	if err != nil {
		logger.Error("Failed to get current air quality",
			zap.String("city", city),
			zap.String("location_id", locationID),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return "", fmt.Errorf("failed to get current air quality: %w", err)
	}
	logger.Debug("Current air quality retrieved",
		zap.String("city", city),
		zap.String("aqi", airNow.Aqi),
		zap.String("category", airNow.Category))

	// Get air quality forecast (optional, non-critical)
	var airForecast []qweather.AirDaily
	logger.Debug("Fetching air quality forecast",
		zap.String("city", city),
		zap.String("location_id", locationID))
	airForecast, err = s.client.GetAirDaily(locationID)
	if err != nil {
		logger.Warn("Failed to get air quality forecast",
			zap.String("city", city),
			zap.String("location_id", locationID),
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
	report.WriteString(fmt.Sprintf("   AQI：%s\n", airNow.Aqi))
	report.WriteString(fmt.Sprintf("   等级：%s\n", airNow.Level))
	report.WriteString(fmt.Sprintf("   类别：%s\n", airNow.Category))
	if airNow.Primary != "" && airNow.Primary != "NA" {
		report.WriteString(fmt.Sprintf("   主要污染物：%s\n", airNow.Primary))
	}

	// Pollutant concentrations
	report.WriteString("\n💨 污染物浓度：\n")
	if airNow.Pm2p5 != "" && airNow.Pm2p5 != "0" {
		report.WriteString(fmt.Sprintf("   PM2.5：%s μg/m³\n", airNow.Pm2p5))
	}
	if airNow.Pm10 != "" && airNow.Pm10 != "0" {
		report.WriteString(fmt.Sprintf("   PM10：%s μg/m³\n", airNow.Pm10))
	}
	if airNow.O3 != "" && airNow.O3 != "0" {
		report.WriteString(fmt.Sprintf("   O3：%s μg/m³\n", airNow.O3))
	}
	if airNow.No2 != "" && airNow.No2 != "0" {
		report.WriteString(fmt.Sprintf("   NO2：%s μg/m³\n", airNow.No2))
	}
	if airNow.So2 != "" && airNow.So2 != "0" {
		report.WriteString(fmt.Sprintf("   SO2：%s μg/m³\n", airNow.So2))
	}
	if airNow.Co != "" && airNow.Co != "0" {
		report.WriteString(fmt.Sprintf("   CO：%s mg/m³\n", airNow.Co))
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
