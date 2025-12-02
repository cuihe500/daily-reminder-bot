package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cuichanghe/daily-reminder-bot/internal/bot"
	"github.com/cuichanghe/daily-reminder-bot/internal/config"
	"github.com/cuichanghe/daily-reminder-bot/internal/migration"
	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/internal/repository"
	"github.com/cuichanghe/daily-reminder-bot/internal/service"
	"github.com/cuichanghe/daily-reminder-bot/pkg/holiday"
	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"github.com/cuichanghe/daily-reminder-bot/pkg/openai"
	"github.com/cuichanghe/daily-reminder-bot/pkg/qweather"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "configs/config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize logger
	if err := logger.Init(&cfg.Logger); err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error("Failed to sync logger", zap.Error(err))
		}
	}()

	// Initialize database
	db, err := initDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	subRepo := repository.NewSubscriptionRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	warningRepo := repository.NewWarningLogRepository(db)

	// Initialize QWeather client
	qweatherClient := qweather.NewClient(cfg.QWeather.APIKey, cfg.QWeather.BaseURL)

	// Initialize services
	weatherSvc := service.NewWeatherService(qweatherClient)
	todoSvc := service.NewTodoService(todoRepo)
	airSvc := service.NewAirQualityService(qweatherClient)

	// Initialize AI service
	var aiSvc *service.AIService
	if cfg.OpenAI.Enabled {
		openaiClient := openai.NewClient(
			cfg.OpenAI.APIKey,
			cfg.OpenAI.BaseURL,
			cfg.OpenAI.Model,
			cfg.OpenAI.MaxTokens,
			cfg.OpenAI.Temperature,
			time.Duration(cfg.OpenAI.Timeout)*time.Second,
		)
		aiSvc = service.NewAIService(openaiClient, cfg.OpenAI.MaxRetries, true)
		logger.Info("AI service initialized",
			zap.String("model", cfg.OpenAI.Model),
			zap.String("base_url", cfg.OpenAI.BaseURL))
	} else {
		aiSvc = service.NewAIService(nil, 0, false)
		logger.Info("AI service disabled")
	}

	// Initialize Holiday client and Calendar service
	loc, err := time.LoadLocation(cfg.Scheduler.Timezone)
	if err != nil {
		logger.Fatal("Failed to load timezone", zap.Error(err))
	}

	var holidayClient *holiday.Client
	if cfg.Holiday.APIURL != "" {
		cacheTTL := time.Duration(cfg.Holiday.CacheTTL) * time.Second
		if cacheTTL == 0 {
			cacheTTL = 24 * time.Hour
		}
		holidayClient = holiday.NewClient(cfg.Holiday.APIURL, cacheTTL)
		logger.Info("Holiday API client initialized", zap.String("api_url", cfg.Holiday.APIURL))
	} else {
		logger.Info("Holiday API not configured, using built-in festival data only")
	}

	calendarSvc := service.NewCalendarService(loc, holidayClient)

	// Initialize bot
	teleBot, err := bot.NewBot(cfg.Telegram.Token, cfg.Telegram.APIEndpoint)
	if err != nil {
		logger.Fatal("Failed to create bot", zap.Error(err))
	}

	// Initialize warning service (needs bot for notifications)
	warningSvc := service.NewWarningService(qweatherClient, warningRepo, subRepo, teleBot.Bot)

	// Initialize scheduler
	schedulerSvc, err := service.NewSchedulerService(
		subRepo,
		weatherSvc,
		todoSvc,
		aiSvc,
		calendarSvc,
		warningSvc,
		teleBot.Bot,
		cfg.Scheduler.Timezone,
	)
	if err != nil {
		logger.Fatal("Failed to create scheduler", zap.Error(err))
	}

	// Register handlers
	handlers := bot.NewHandlers(userRepo, subRepo, todoRepo, weatherSvc, todoSvc, airSvc, warningSvc)
	handlers.RegisterHandlers(teleBot.Bot)

	// Start scheduler
	if err := schedulerSvc.Start(); err != nil {
		logger.Fatal("Failed to start scheduler", zap.Error(err))
	}
	defer schedulerSvc.Stop()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		logger.Info("Received shutdown signal")
		schedulerSvc.Stop()
		teleBot.Stop()
		os.Exit(0)
	}()

	// Start bot
	logger.Info("Bot started successfully")
	teleBot.Start()
}

// initDatabase initializes the database and runs migrations
func initDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	gormLogger := logger.NewGormAdapter(logger.Get(), 200*time.Millisecond)

	switch cfg.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: gormLogger})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
		}
		logger.Info("Connected to MySQL database")
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{Logger: gormLogger})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
		}
		logger.Info("Connected to SQLite database")
	default:
		return nil, fmt.Errorf("unsupported database type: %s (must be 'sqlite' or 'mysql')", cfg.Type)
	}

	// Auto migrate models
	if err := db.AutoMigrate(
		&model.User{},
		&model.Subscription{},
		&model.Todo{},
		&model.WarningLog{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Run data migration to multi-subscription model
	if err := migration.MigrateToMultiSubscription(db); err != nil {
		return nil, fmt.Errorf("failed to run data migration: %w", err)
	}

	logger.Info("Database initialized successfully")
	return db, nil
}
