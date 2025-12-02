package repository

import (
	"fmt"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SubscriptionRepository handles subscription data access
type SubscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository creates a new SubscriptionRepository
func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create creates a new subscription
func (r *SubscriptionRepository) Create(sub *model.Subscription) error {
	logger.Debug("SubscriptionRepository.Create called",
		zap.Uint("user_id", sub.UserID),
		zap.String("city", sub.City),
		zap.String("reminder_time", sub.ReminderTime))

	if err := r.db.Create(sub).Error; err != nil {
		logger.Error("Failed to create subscription",
			zap.Uint("user_id", sub.UserID),
			zap.Error(err))
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	logger.Info("Subscription created successfully",
		zap.Uint("subscription_id", sub.ID),
		zap.Uint("user_id", sub.UserID),
		zap.String("city", sub.City))
	return nil
}

// FindByUserID finds a subscription by user ID
func (r *SubscriptionRepository) FindByUserID(userID uint) (*model.Subscription, error) {
	logger.Debug("SubscriptionRepository.FindByUserID called",
		zap.Uint("user_id", userID))

	var sub model.Subscription
	err := r.db.Where("user_id = ? AND active = ?", userID, true).First(&sub).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Debug("Active subscription not found",
				zap.Uint("user_id", userID))
			return nil, nil
		}
		logger.Error("Failed to find subscription",
			zap.Uint("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	logger.Debug("Subscription found",
		zap.Uint("subscription_id", sub.ID),
		zap.Uint("user_id", userID),
		zap.String("city", sub.City))
	return &sub, nil
}

// Update updates a subscription
func (r *SubscriptionRepository) Update(sub *model.Subscription) error {
	logger.Debug("SubscriptionRepository.Update called",
		zap.Uint("subscription_id", sub.ID),
		zap.Bool("active", sub.Active))

	if err := r.db.Save(sub).Error; err != nil {
		logger.Error("Failed to update subscription",
			zap.Uint("subscription_id", sub.ID),
			zap.Error(err))
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	logger.Debug("Subscription updated successfully",
		zap.Uint("subscription_id", sub.ID))
	return nil
}

// GetAllActive retrieves all active subscriptions
func (r *SubscriptionRepository) GetAllActive() ([]model.Subscription, error) {
	logger.Debug("SubscriptionRepository.GetAllActive called")

	var subs []model.Subscription
	err := r.db.Preload("User").Where("active = ?", true).Find(&subs).Error
	if err != nil {
		logger.Error("Failed to get active subscriptions",
			zap.Error(err))
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}

	logger.Debug("Active subscriptions retrieved",
		zap.Int("count", len(subs)))
	return subs, nil
}

// GetByReminderTime retrieves active subscriptions for a specific reminder time
func (r *SubscriptionRepository) GetByReminderTime(reminderTime string) ([]model.Subscription, error) {
	logger.Debug("SubscriptionRepository.GetByReminderTime called",
		zap.String("reminder_time", reminderTime))

	var subs []model.Subscription
	err := r.db.Preload("User").Where("active = ? AND reminder_time = ?", true, reminderTime).Find(&subs).Error
	if err != nil {
		logger.Error("Failed to get subscriptions by reminder time",
			zap.String("reminder_time", reminderTime),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get subscriptions by reminder time: %w", err)
	}

	logger.Debug("Subscriptions by reminder time retrieved",
		zap.String("reminder_time", reminderTime),
		zap.Int("count", len(subs)))
	return subs, nil
}
