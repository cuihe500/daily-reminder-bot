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
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

// WarningService handles weather warning notifications
type WarningService struct {
	client      *qweather.Client
	warningRepo *repository.WarningLogRepository
	subRepo     *repository.SubscriptionRepository
	bot         *tele.Bot
}

// NewWarningService creates a new WarningService
func NewWarningService(
	client *qweather.Client,
	warningRepo *repository.WarningLogRepository,
	subRepo *repository.SubscriptionRepository,
	bot *tele.Bot,
) *WarningService {
	return &WarningService{
		client:      client,
		warningRepo: warningRepo,
		subRepo:     subRepo,
		bot:         bot,
	}
}

// GetWarnings retrieves weather warnings for a city
func (s *WarningService) GetWarnings(city string) ([]qweather.Warning, error) {
	logger.Debug("GetWarnings called", zap.String("city", city))
	start := time.Now()

	// Get location ID
	locationID, err := s.client.GetLocationID(city)
	if err != nil {
		logger.Error("Failed to get location ID",
			zap.String("city", city),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get location ID: %w", err)
	}

	// Get warnings
	warnings, err := s.client.GetWarningNow(locationID)
	if err != nil {
		logger.Error("Failed to get warnings",
			zap.String("city", city),
			zap.String("location_id", locationID),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to get warnings: %w", err)
	}

	logger.Debug("Warnings retrieved",
		zap.String("city", city),
		zap.Int("count", len(warnings)),
		zap.Duration("duration", time.Since(start)))
	return warnings, nil
}

// GetWarningReport generates a formatted weather warning report
func (s *WarningService) GetWarningReport(city string) (string, error) {
	warnings, err := s.GetWarnings(city)
	if err != nil {
		return "", err
	}

	var report strings.Builder
	report.WriteString(fmt.Sprintf("⚠️ %s 天气预警\n\n", city))

	if len(warnings) == 0 {
		report.WriteString("✅ 当前无生效预警\n")
		return report.String(), nil
	}

	for i, w := range warnings {
		if i > 0 {
			report.WriteString("\n")
		}

		// Warning header with color indicator
		emoji := getWarningEmoji(w.SeverityColor)
		report.WriteString(fmt.Sprintf("%s %s\n", emoji, w.Title))
		report.WriteString(fmt.Sprintf("   发布时间：%s\n", formatTime(w.PubTime)))

		// Time range
		if w.StartTime != "" && w.EndTime != "" {
			report.WriteString(fmt.Sprintf("   生效时间：%s - %s\n",
				formatTime(w.StartTime), formatTime(w.EndTime)))
		}

		// Sender
		if w.Sender != "" {
			report.WriteString(fmt.Sprintf("   发布单位：%s\n", w.Sender))
		}

		// Details
		if w.Text != "" {
			report.WriteString(fmt.Sprintf("\n   详情：\n   %s\n", w.Text))
		}
	}

	return report.String(), nil
}

// CheckAndNotify checks for new warnings and notifies subscribed users
func (s *WarningService) CheckAndNotify(ctx context.Context) error {
	logger.Debug("CheckAndNotify called")
	start := time.Now()

	// Get all active subscriptions with warning enabled, grouped by city
	subs, err := s.subRepo.GetAllActive()
	if err != nil {
		logger.Error("Failed to get subscriptions", zap.Error(err))
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	// Group subscriptions by city to avoid duplicate API calls
	cityMap := make(map[string][]model.Subscription)
	for _, sub := range subs {
		if sub.Active && sub.EnableWarning {
			cityMap[sub.City] = append(cityMap[sub.City], sub)
		}
	}

	logger.Debug("Checking warnings for cities",
		zap.Int("city_count", len(cityMap)))

	// Check warnings for each city
	for city, citySubs := range cityMap {
		if err := s.checkCityWarnings(ctx, city, citySubs); err != nil {
			logger.Warn("Failed to check warnings for city",
				zap.String("city", city),
				zap.Error(err))
			// Continue with other cities even if one fails
		}
	}

	logger.Debug("CheckAndNotify completed",
		zap.Duration("duration", time.Since(start)))
	return nil
}

// checkCityWarnings checks warnings for a specific city and notifies users
func (s *WarningService) checkCityWarnings(ctx context.Context, city string, subs []model.Subscription) error {
	logger.Debug("Checking warnings for city",
		zap.String("city", city),
		zap.Int("subscriber_count", len(subs)))

	// Get location ID
	locationID, err := s.client.GetLocationID(city)
	if err != nil {
		return fmt.Errorf("failed to get location ID for %s: %w", city, err)
	}

	// Get current warnings
	warnings, err := s.client.GetWarningNow(locationID)
	if err != nil {
		return fmt.Errorf("failed to get warnings for %s: %w", city, err)
	}

	if len(warnings) == 0 {
		logger.Debug("No warnings for city", zap.String("city", city))
		return nil
	}

	// Process each warning
	for _, warning := range warnings {
		if err := s.processWarning(ctx, city, locationID, warning, subs); err != nil {
			logger.Warn("Failed to process warning",
				zap.String("warning_id", warning.ID),
				zap.Error(err))
			// Continue with other warnings
		}
	}

	return nil
}

// processWarning processes a single warning and sends notifications if needed
func (s *WarningService) processWarning(
	ctx context.Context,
	city string,
	locationID string,
	warning qweather.Warning,
	subs []model.Subscription,
) error {
	// Check if we've already notified about this warning
	existingLog, err := s.warningRepo.GetByWarningID(warning.ID)
	if err != nil {
		return fmt.Errorf("failed to check warning log: %w", err)
	}

	// If this is a new warning or updated warning, send notification
	shouldNotify := false
	if existingLog == nil {
		// New warning
		shouldNotify = true
		logger.Info("New warning detected",
			zap.String("city", city),
			zap.String("warning_id", warning.ID),
			zap.String("title", warning.Title))
	} else if existingLog.Status != warning.Status {
		// Status changed (e.g., active -> update or cancel)
		shouldNotify = true
		logger.Info("Warning status changed",
			zap.String("city", city),
			zap.String("warning_id", warning.ID),
			zap.String("old_status", existingLog.Status),
			zap.String("new_status", warning.Status))
	}

	if !shouldNotify {
		logger.Debug("Warning already notified, skipping",
			zap.String("warning_id", warning.ID))
		return nil
	}

	// Format notification message
	message := s.formatWarningMessage(city, warning)

	// Send to all subscribers
	successCount := 0
	for _, sub := range subs {
		recipient := &tele.User{ID: sub.User.ChatID}
		if _, err := s.bot.Send(recipient, message); err != nil {
			logger.Warn("Failed to send warning notification",
				zap.Uint("user_id", sub.UserID),
				zap.Int64("chat_id", sub.User.ChatID),
				zap.Error(err))
		} else {
			successCount++
			logger.Debug("Warning notification sent",
				zap.Uint("user_id", sub.UserID))
		}
	}

	logger.Info("Warning notifications sent",
		zap.String("warning_id", warning.ID),
		zap.Int("success_count", successCount),
		zap.Int("total_count", len(subs)))

	// Update or create warning log
	now := time.Now()
	if existingLog == nil {
		// Create new log
		startTime, _ := time.Parse(time.RFC3339, warning.StartTime)
		endTime, _ := time.Parse(time.RFC3339, warning.EndTime)

		newLog := &model.WarningLog{
			WarningID:  warning.ID,
			LocationID: locationID,
			City:       city,
			Type:       warning.Type,
			Level:      warning.Level,
			Title:      warning.Title,
			StartTime:  startTime,
			EndTime:    endTime,
			Status:     warning.Status,
			NotifiedAt: now,
		}
		if err := s.warningRepo.Create(newLog); err != nil {
			return fmt.Errorf("failed to create warning log: %w", err)
		}
	} else {
		// Update existing log
		existingLog.Status = warning.Status
		existingLog.NotifiedAt = now
		if err := s.warningRepo.Update(existingLog); err != nil {
			return fmt.Errorf("failed to update warning log: %w", err)
		}
	}

	return nil
}

// formatWarningMessage formats a warning into a notification message
func (s *WarningService) formatWarningMessage(city string, warning qweather.Warning) string {
	var msg strings.Builder

	emoji := getWarningEmoji(warning.SeverityColor)
	msg.WriteString(fmt.Sprintf("⚠️ %s 天气预警\n\n", city))
	msg.WriteString(fmt.Sprintf("%s %s\n", emoji, warning.Title))
	msg.WriteString(fmt.Sprintf("发布时间：%s\n", formatTime(warning.PubTime)))

	if warning.StartTime != "" && warning.EndTime != "" {
		msg.WriteString(fmt.Sprintf("生效时间：%s - %s\n",
			formatTime(warning.StartTime), formatTime(warning.EndTime)))
	}

	if warning.Sender != "" {
		msg.WriteString(fmt.Sprintf("发布单位：%s\n", warning.Sender))
	}

	if warning.Text != "" {
		msg.WriteString(fmt.Sprintf("\n详情：\n%s\n", warning.Text))
	}

	switch warning.Status {
	case "cancel":
		msg.WriteString("\n✅ 该预警已解除")
	case "update":
		msg.WriteString("\n🔄 该预警已更新")
	}

	return msg.String()
}

// getWarningEmoji returns an emoji based on warning severity color
func getWarningEmoji(severityColor string) string {
	switch severityColor {
	case "Red":
		return "🔴"
	case "Orange":
		return "🟠"
	case "Yellow":
		return "🟡"
	case "Blue":
		return "🔵"
	default:
		return "⚠️"
	}
}

// formatTime formats ISO8601 time to a more readable format
func formatTime(isoTime string) string {
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		return isoTime
	}
	return t.Format("2006-01-02 15:04")
}
