package service

import (
	"fmt"
	"log"
	"time"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/internal/repository"
	"github.com/robfig/cron/v3"
	tele "gopkg.in/telebot.v3"
)

// SchedulerService handles scheduled tasks
type SchedulerService struct {
	cron       *cron.Cron
	subRepo    *repository.SubscriptionRepository
	weatherSvc *WeatherService
	todoSvc    *TodoService
	bot        *tele.Bot
	timezone   *time.Location
}

// NewSchedulerService creates a new SchedulerService
func NewSchedulerService(
	subRepo *repository.SubscriptionRepository,
	weatherSvc *WeatherService,
	todoSvc *TodoService,
	bot *tele.Bot,
	timezoneStr string,
) (*SchedulerService, error) {
	loc, err := time.LoadLocation(timezoneStr)
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone: %w", err)
	}

	c := cron.New(cron.WithLocation(loc))

	return &SchedulerService{
		cron:       c,
		subRepo:    subRepo,
		weatherSvc: weatherSvc,
		todoSvc:    todoSvc,
		bot:        bot,
		timezone:   loc,
	}, nil
}

// Start starts the scheduler
func (s *SchedulerService) Start() error {
	// Schedule a job every minute to check for reminders
	_, err := s.cron.AddFunc("* * * * *", s.checkReminders)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.cron.Start()
	log.Println("Scheduler started")
	return nil
}

// Stop stops the scheduler
func (s *SchedulerService) Stop() {
	s.cron.Stop()
	log.Println("Scheduler stopped")
}

// checkReminders checks for subscriptions that need reminders at the current time
func (s *SchedulerService) checkReminders() {
	now := time.Now().In(s.timezone)
	currentTime := now.Format("15:04")

	subs, err := s.subRepo.GetByReminderTime(currentTime)
	if err != nil {
		log.Printf("Error getting subscriptions: %v", err)
		return
	}

	for _, sub := range subs {
		go s.sendReminder(sub)
	}
}

// sendReminder sends a daily reminder to a user
func (s *SchedulerService) sendReminder(sub model.Subscription) {
	// Get weather report
	weatherReport, err := s.weatherSvc.GetWeatherReport(sub.City)
	if err != nil {
		log.Printf("Error getting weather for user %d: %v", sub.UserID, err)
		weatherReport = fmt.Sprintf("⚠️ 无法获取 %s 的天气信息", sub.City)
	}

	// Get incomplete todos
	todos, err := s.todoSvc.GetIncompleteTodos(sub.UserID)
	if err != nil {
		log.Printf("Error getting todos for user %d: %v", sub.UserID, err)
	}

	todoReport := s.todoSvc.FormatTodoList(todos)

	// Format the complete reminder message
	message := fmt.Sprintf("🌅 早安！今日提醒 (%s)\n\n%s\n%s",
		time.Now().In(s.timezone).Format("2006-01-02"),
		weatherReport,
		todoReport)

	// Send message to user
	recipient := &tele.User{ID: sub.User.ChatID}
	_, err = s.bot.Send(recipient, message)
	if err != nil {
		log.Printf("Error sending reminder to user %d: %v", sub.UserID, err)
	}
}
