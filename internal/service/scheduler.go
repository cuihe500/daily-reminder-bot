package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/internal/repository"
	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"github.com/cuichanghe/daily-reminder-bot/pkg/qweather"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

// SchedulerService handles scheduled tasks
type SchedulerService struct {
	cron        *cron.Cron
	subRepo     *repository.SubscriptionRepository
	weatherSvc  *WeatherService
	todoSvc     *TodoService
	aiSvc       *AIService
	calendarSvc *CalendarService
	bot         *tele.Bot
	timezone    *time.Location
}

// NewSchedulerService creates a new SchedulerService
func NewSchedulerService(
	subRepo *repository.SubscriptionRepository,
	weatherSvc *WeatherService,
	todoSvc *TodoService,
	aiSvc *AIService,
	calendarSvc *CalendarService,
	bot *tele.Bot,
	timezoneStr string,
) (*SchedulerService, error) {
	loc, err := time.LoadLocation(timezoneStr)
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone: %w", err)
	}

	c := cron.New(cron.WithLocation(loc))

	return &SchedulerService{
		cron:        c,
		subRepo:     subRepo,
		weatherSvc:  weatherSvc,
		todoSvc:     todoSvc,
		aiSvc:       aiSvc,
		calendarSvc: calendarSvc,
		bot:         bot,
		timezone:    loc,
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
	logger.Info("Scheduler started")
	return nil
}

// Stop stops the scheduler
func (s *SchedulerService) Stop() {
	s.cron.Stop()
	logger.Info("Scheduler stopped")
}

// checkReminders checks for subscriptions that need reminders at the current time
func (s *SchedulerService) checkReminders() {
	now := time.Now().In(s.timezone)
	currentTime := now.Format("15:04")

	subs, err := s.subRepo.GetByReminderTime(currentTime)
	if err != nil {
		logger.Error("Error getting subscriptions", zap.Error(err))
		return
	}

	for _, sub := range subs {
		go s.sendReminder(sub)
	}
}

// sendReminder sends a daily reminder to a user
func (s *SchedulerService) sendReminder(sub model.Subscription) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	now := time.Now().In(s.timezone)

	// Get location ID and weather data
	locationID, err := s.weatherSvc.Client().GetLocationID(sub.City)
	if err != nil {
		logger.Error("Failed to get location ID", zap.Uint("user_id", sub.UserID), zap.Error(err))
		s.sendFallbackReminder(sub, now, fmt.Sprintf("⚠️ 无法获取 %s 的位置信息", sub.City))
		return
	}

	weather, err := s.weatherSvc.Client().GetCurrentWeather(locationID)
	if err != nil {
		logger.Error("Failed to get weather", zap.Uint("user_id", sub.UserID), zap.Error(err))
		s.sendFallbackReminder(sub, now, fmt.Sprintf("⚠️ 无法获取 %s 的天气信息", sub.City))
		return
	}

	indices, err := s.weatherSvc.Client().GetLifeIndices(locationID)
	if err != nil {
		logger.Warn("Failed to get life indices", zap.Uint("user_id", sub.UserID), zap.Error(err))
		indices = nil
	}

	// Get incomplete todos
	todos, err := s.todoSvc.GetIncompleteTodos(sub.UserID)
	if err != nil {
		logger.Warn("Failed to get todos", zap.Uint("user_id", sub.UserID), zap.Error(err))
		todos = nil
	}

	// Get calendar info
	var calendarInfo string
	if s.calendarSvc != nil {
		calendarInfo = s.calendarSvc.FormatCalendarInfoForAI(now)
	}

	// Try to generate AI reminder
	var message string
	if s.aiSvc != nil && s.aiSvc.IsEnabled() {
		data := ReminderData{
			City:         sub.City,
			Date:         now.Format("2006-01-02"),
			Weather:      weather,
			LifeIndices:  indices,
			Todos:        todos,
			CalendarInfo: calendarInfo,
		}

		aiContent, ok := s.aiSvc.GenerateReminder(ctx, data)
		if ok {
			message = aiContent
		}
	}

	// Fallback to fixed template if AI generation failed or disabled
	if message == "" {
		message = s.buildFallbackMessage(sub.City, weather, indices, todos, now, s.aiSvc != nil && s.aiSvc.IsEnabled())
	}

	// Send message to user
	recipient := &tele.User{ID: sub.User.ChatID}
	_, err = s.bot.Send(recipient, message)
	if err != nil {
		logger.Error("Error sending reminder", zap.Uint("user_id", sub.UserID), zap.Error(err))
	}
}

// buildFallbackMessage builds a fallback message using the fixed template
func (s *SchedulerService) buildFallbackMessage(
	city string,
	weather *qweather.CurrentWeather,
	indices []qweather.LifeIndex,
	todos []model.Todo,
	now time.Time,
	aiWasEnabled bool,
) string {
	var report strings.Builder

	// Date header with calendar info
	report.WriteString("🌅 早安！今日提醒\n")
	if s.calendarSvc != nil {
		dateHeader := s.calendarSvc.FormatDateHeader(now)
		report.WriteString(fmt.Sprintf("📆 %s\n", dateHeader))

		todaySpecial := s.calendarSvc.FormatTodaySpecial(now)
		if todaySpecial != "" {
			report.WriteString(fmt.Sprintf("🎊 %s\n", todaySpecial))
		}
		report.WriteString("\n")

		// Upcoming festivals
		upcomingFestivals := s.calendarSvc.FormatUpcomingFestivals(now, 3)
		if upcomingFestivals != "" {
			report.WriteString(upcomingFestivals)
			report.WriteString("\n")
		}
	} else {
		report.WriteString(fmt.Sprintf("📆 %s\n\n", now.Format("2006-01-02")))
	}

	report.WriteString(fmt.Sprintf("📍 %s 天气播报\n\n", city))
	report.WriteString(fmt.Sprintf("🌡️ 温度：%s°C（体感 %s°C）\n", weather.Temp, weather.FeelsLike))
	report.WriteString(fmt.Sprintf("☁️ 天气：%s\n", weather.Text))
	report.WriteString(fmt.Sprintf("💧 湿度：%s%%\n", weather.Humidity))
	report.WriteString(fmt.Sprintf("🌬️ 风向：%s %s级（%s km/h）\n\n", weather.WindDir, weather.WindScale, weather.WindSpeed))

	// Add life indices
	if len(indices) > 0 {
		report.WriteString("📋 生活指数：\n")
		for _, index := range indices {
			if index.Type == "3" || index.Type == "5" || index.Type == "1" {
				emoji := getIndexEmoji(index.Type)
				report.WriteString(fmt.Sprintf("%s %s：%s\n", emoji, index.Name, index.Category))
				if index.Text != "" {
					report.WriteString(fmt.Sprintf("   %s\n", index.Text))
				}
			}
		}
		report.WriteString("\n")
	}

	// Add todo list
	report.WriteString(s.todoSvc.FormatTodoList(todos))

	// Add AI service unavailable notice
	if aiWasEnabled {
		report.WriteString("\n---\n(AI 服务暂不可用，使用默认模板)")
	}

	return report.String()
}

// sendFallbackReminder sends a simplified fallback reminder when weather data is unavailable
func (s *SchedulerService) sendFallbackReminder(sub model.Subscription, now time.Time, errorMsg string) {
	// Get todos even if weather failed
	todos, _ := s.todoSvc.GetIncompleteTodos(sub.UserID)
	todoReport := s.todoSvc.FormatTodoList(todos)

	var message strings.Builder
	message.WriteString("🌅 早安！今日提醒\n")

	// Add calendar info
	if s.calendarSvc != nil {
		dateHeader := s.calendarSvc.FormatDateHeader(now)
		message.WriteString(fmt.Sprintf("📆 %s\n", dateHeader))

		todaySpecial := s.calendarSvc.FormatTodaySpecial(now)
		if todaySpecial != "" {
			message.WriteString(fmt.Sprintf("🎊 %s\n", todaySpecial))
		}
		message.WriteString("\n")

		upcomingFestivals := s.calendarSvc.FormatUpcomingFestivals(now, 3)
		if upcomingFestivals != "" {
			message.WriteString(upcomingFestivals)
			message.WriteString("\n")
		}
	} else {
		message.WriteString(fmt.Sprintf("📆 %s\n\n", now.Format("2006-01-02")))
	}

	message.WriteString(errorMsg)
	message.WriteString("\n\n")
	message.WriteString(todoReport)

	recipient := &tele.User{ID: sub.User.ChatID}
	_, err := s.bot.Send(recipient, message.String())
	if err != nil {
		logger.Error("Error sending fallback reminder", zap.Uint("user_id", sub.UserID), zap.Error(err))
	}
}
