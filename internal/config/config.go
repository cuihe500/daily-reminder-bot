package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Telegram  TelegramConfig  `mapstructure:"telegram"`
	QWeather  QWeatherConfig  `mapstructure:"qweather"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	Token       string `mapstructure:"token"`
	APIEndpoint string `mapstructure:"api_endpoint"`
}

// QWeatherConfig holds QWeather API configuration
type QWeatherConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	Timezone string `mapstructure:"timezone"`
}

// Load reads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file path
	v.SetConfigFile(configPath)

	// Enable environment variable override
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
