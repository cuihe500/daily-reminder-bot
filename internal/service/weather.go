package service

import (
	"fmt"
	"strings"

	"github.com/cuichanghe/daily-reminder-bot/pkg/qweather"
)

// WeatherService handles weather-related business logic
type WeatherService struct {
	client *qweather.Client
}

// NewWeatherService creates a new WeatherService
func NewWeatherService(client *qweather.Client) *WeatherService {
	return &WeatherService{client: client}
}

// GetWeatherReport generates a formatted weather report for a city
func (s *WeatherService) GetWeatherReport(city string) (string, error) {
	// Get location ID
	locationID, err := s.client.GetLocationID(city)
	if err != nil {
		return "", fmt.Errorf("failed to get location ID: %w", err)
	}

	// Get current weather
	weather, err := s.client.GetCurrentWeather(locationID)
	if err != nil {
		return "", fmt.Errorf("failed to get current weather: %w", err)
	}

	// Get life indices
	indices, err := s.client.GetLifeIndices(locationID)
	if err != nil {
		return "", fmt.Errorf("failed to get life indices: %w", err)
	}

	// Format the report
	var report strings.Builder
	report.WriteString(fmt.Sprintf("📍 %s 天气播报\n\n", city))
	report.WriteString(fmt.Sprintf("🌡️ 温度：%s°C（体感 %s°C）\n", weather.Temp, weather.FeelsLike))
	report.WriteString(fmt.Sprintf("☁️ 天气：%s\n", weather.Text))
	report.WriteString(fmt.Sprintf("💧 湿度：%s%%\n", weather.Humidity))
	report.WriteString(fmt.Sprintf("🌬️ 风向：%s %s级（%s km/h）\n\n", weather.WindDir, weather.WindScale, weather.WindSpeed))

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
