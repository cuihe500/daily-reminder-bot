package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Telegram  TelegramConfig  `mapstructure:"telegram"`
	QWeather  QWeatherConfig  `mapstructure:"qweather"`
	OpenAI    OpenAIConfig    `mapstructure:"openai"`
	Holiday   HolidayConfig   `mapstructure:"holiday"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Logger    LoggerConfig    `mapstructure:"logger"`
}

// OpenAIConfig holds OpenAI-compatible API configuration
type OpenAIConfig struct {
	Enabled     bool    `mapstructure:"enabled"`     // Whether to enable AI generation
	APIKey      string  `mapstructure:"api_key"`     // API key
	BaseURL     string  `mapstructure:"base_url"`    // API base URL (supports OpenAI, DeepSeek, etc.)
	Model       string  `mapstructure:"model"`       // Model name (e.g., gpt-4o-mini, deepseek-chat)
	MaxTokens   int     `mapstructure:"max_tokens"`  // Maximum tokens to generate
	Temperature float64 `mapstructure:"temperature"` // Generation temperature (0-2)
	Timeout     int     `mapstructure:"timeout"`     // Request timeout in seconds
	MaxRetries  int     `mapstructure:"max_retries"` // Maximum retry attempts
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
	Type     string `mapstructure:"type"`     // "sqlite" or "mysql"
	Path     string `mapstructure:"path"`     // SQLite database file path
	Host     string `mapstructure:"host"`     // MySQL host
	Port     int    `mapstructure:"port"`     // MySQL port
	User     string `mapstructure:"user"`     // MySQL username
	Password string `mapstructure:"password"` // MySQL password
	DBName   string `mapstructure:"dbname"`   // MySQL database name
	Charset  string `mapstructure:"charset"`  // MySQL charset
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	Timezone string `mapstructure:"timezone"`
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// HolidayConfig holds holiday API configuration
type HolidayConfig struct {
	APIURL   string `mapstructure:"api_url"`   // Holiday API base URL
	CacheTTL int    `mapstructure:"cache_ttl"` // Cache TTL in seconds
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
