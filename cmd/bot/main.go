package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cuichanghe/daily-reminder-bot/internal/bot"
	"github.com/cuichanghe/daily-reminder-bot/internal/config"
	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/internal/repository"
	"github.com/cuichanghe/daily-reminder-bot/internal/service"
	"github.com/cuichanghe/daily-reminder-bot/pkg/qweather"
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
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := initDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	subRepo := repository.NewSubscriptionRepository(db)
	todoRepo := repository.NewTodoRepository(db)

	// Initialize QWeather client
	qweatherClient := qweather.NewClient(cfg.QWeather.APIKey, cfg.QWeather.BaseURL)

	// Initialize services
	weatherSvc := service.NewWeatherService(qweatherClient)
	todoSvc := service.NewTodoService(todoRepo)

	// Initialize bot
	teleBot, err := bot.NewBot(cfg.Telegram.Token, cfg.Telegram.APIEndpoint)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Initialize scheduler
	schedulerSvc, err := service.NewSchedulerService(
		subRepo,
		weatherSvc,
		todoSvc,
		teleBot.Bot,
		cfg.Scheduler.Timezone,
	)
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	// Register handlers
	handlers := bot.NewHandlers(userRepo, subRepo, todoRepo, weatherSvc, todoSvc)
	handlers.RegisterHandlers(teleBot.Bot)

	// Start scheduler
	if err := schedulerSvc.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	defer schedulerSvc.Stop()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Received shutdown signal")
		schedulerSvc.Stop()
		teleBot.Stop()
		os.Exit(0)
	}()

	// Start bot
	log.Println("Bot started successfully")
	teleBot.Start()
}

// initDatabase initializes the database and runs migrations
func initDatabase(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate models
	if err := db.AutoMigrate(
		&model.User{},
		&model.Subscription{},
		&model.Todo{},
	); err != nil {
		return nil, err
	}

	log.Println("Database initialized successfully")
	return db, nil
}
