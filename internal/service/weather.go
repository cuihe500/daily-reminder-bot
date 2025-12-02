package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"github.com/cuichanghe/daily-reminder-bot/pkg/qweather"
	"go.uber.org/zap"
)

// WeatherService handles weather-related business logic
type WeatherService struct {
	client *qweather.Client // exported via getter for scheduler access
}

// Client returns the underlying QWeather client
func (s *WeatherService) Client() *qweather.Client {
	return s.client
}

// NewWeatherService creates a new WeatherService
func NewWeatherService(client *qweather.Client) *WeatherService {
	return &WeatherService{client: client}
}

// GetWeatherReport generates a formatted weather report for a city
func (s *WeatherService) GetWeatherReport(city string) (string, error) {
	logger.Debug("GetWeatherReport called", zap.String("city", city))
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

	// Get current weather
	logger.Debug("Fetching current weather",
		zap.String("city", city),
		zap.String("location_id", locationID))
	weather, err := s.client.GetCurrentWeather(locationID)
	if err != nil {
		logger.Error("Failed to get current weather",
			zap.String("city", city),
			zap.String("location_id", locationID),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return "", fmt.Errorf("failed to get current weather: %w", err)
	}
	logger.Debug("Current weather retrieved",
		zap.String("city", city),
		zap.String("temp", weather.Temp),
		zap.String("text", weather.Text))

	// Get daily forecast (for max/min temperature)
	logger.Debug("Fetching daily forecast",
		zap.String("city", city),
		zap.String("location_id", locationID))
	forecast, err := s.client.GetDailyForecast(locationID)
	if err != nil {
		logger.Error("Failed to get daily forecast",
			zap.String("city", city),
			zap.String("location_id", locationID),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return "", fmt.Errorf("failed to get daily forecast: %w", err)
	}
	logger.Debug("Daily forecast retrieved",
		zap.String("city", city),
		zap.String("tempMax", forecast.TempMax),
		zap.String("tempMin", forecast.TempMin))

	// Get life indices
	logger.Debug("Fetching life indices",
		zap.String("city", city),
		zap.String("location_id", locationID))
	indices, err := s.client.GetLifeIndices(locationID)
	if err != nil {
		logger.Error("Failed to get life indices",
			zap.String("city", city),
			zap.String("location_id", locationID),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return "", fmt.Errorf("failed to get life indices: %w", err)
	}
	logger.Debug("Life indices retrieved",
		zap.String("city", city),
		zap.Int("indices_count", len(indices)))

	// Format the report
	var report strings.Builder
	report.WriteString(fmt.Sprintf("📍 %s 天气播报\n\n", city))

	// Temperature section
	report.WriteString("🌡️ 温度信息：\n")
	report.WriteString(fmt.Sprintf("   当前温度：%s°C\n", weather.Temp))
	report.WriteString(fmt.Sprintf("   体感温度：%s°C\n", weather.FeelsLike))
	report.WriteString(fmt.Sprintf("   最高温度：%s°C\n", forecast.TempMax))
	report.WriteString(fmt.Sprintf("   最低温度：%s°C\n\n", forecast.TempMin))

	// Weather details
	report.WriteString("☁️ 天气状况：\n")
	report.WriteString(fmt.Sprintf("   当前天气：%s\n", weather.Text))
	report.WriteString(fmt.Sprintf("   白天天气：%s\n", forecast.TextDay))
	report.WriteString(fmt.Sprintf("   夜间天气：%s\n\n", forecast.TextNight))

	// Atmospheric data
	report.WriteString("📊 大气数据：\n")
	report.WriteString(fmt.Sprintf("   相对湿度：%s%%\n", weather.Humidity))
	report.WriteString(fmt.Sprintf("   大气气压：%s hPa\n", forecast.Pressure))
	report.WriteString(fmt.Sprintf("   能见度：%s km\n", forecast.Vis))
	if forecast.Cloud != "" {
		report.WriteString(fmt.Sprintf("   云量：%s%%\n", forecast.Cloud))
	}
	if forecast.Precip != "" && forecast.Precip != "0.0" {
		report.WriteString(fmt.Sprintf("   降水量：%s mm\n", forecast.Precip))
	}
	report.WriteString("\n")

	// Wind information
	report.WriteString("🌬️ 风力信息：\n")
	report.WriteString(fmt.Sprintf("   当前风向：%s %s级（%s km/h）\n", weather.WindDir, weather.WindScale, weather.WindSpeed))
	report.WriteString(fmt.Sprintf("   白天风向：%s %s级\n", forecast.WindDirDay, forecast.WindScaleDay))
	report.WriteString(fmt.Sprintf("   夜间风向：%s %s级\n\n", forecast.WindDirNight, forecast.WindScaleNight))

	// Sun and moon times
	report.WriteString("🌅 日出日落：\n")
	report.WriteString(fmt.Sprintf("   日出时间：%s\n", forecast.Sunrise))
	report.WriteString(fmt.Sprintf("   日落时间：%s\n", forecast.Sunset))
	if forecast.MoonPhase != "" {
		report.WriteString(fmt.Sprintf("   月相：%s\n", forecast.MoonPhase))
	}
	report.WriteString("\n")

	// Add life indices
	report.WriteString("📋 生活指数：\n")
	for _, index := range indices {
		// Filter important indices: dressing (3), UV (5), sports (1)
		if index.Type == "3" || index.Type == "5" || index.Type == "1" {
			emoji := getIndexEmoji(index.Type)
			report.WriteString(fmt.Sprintf("%s %s：%s\n", emoji, index.Name, index.Category))
			if index.Text != "" {
				report.WriteString(fmt.Sprintf("   %s\n", index.Text))
			}
		}
	}

	logger.Info("Weather report generated successfully",
		zap.String("city", city),
		zap.Duration("duration", time.Since(start)))
	return report.String(), nil
}

// getIndexEmoji returns an emoji for a life index type
func getIndexEmoji(indexType string) string {
	switch indexType {
	case "1": // Sports
		return "🏃"
	case "3": // Dressing
		return "👔"
	case "5": // UV
		return "☀️"
	default:
		return "📌"
	}
}
