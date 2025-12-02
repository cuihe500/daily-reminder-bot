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

// FindByUserID finds all active subscriptions by user ID
func (r *SubscriptionRepository) FindByUserID(userID uint) ([]model.Subscription, error) {
	logger.Debug("SubscriptionRepository.FindByUserID called",
		zap.Uint("user_id", userID))

	var subs []model.Subscription
	err := r.db.Where("user_id = ? AND active = ?", userID, true).
		Order("created_at ASC").
		Find(&subs).Error
	if err != nil {
		logger.Error("Failed to find subscriptions",
			zap.Uint("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to find subscriptions: %w", err)
	}

	logger.Debug("Subscriptions found",
		zap.Uint("user_id", userID),
		zap.Int("count", len(subs)))
	return subs, nil
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

// FindByID finds a subscription by ID
func (r *SubscriptionRepository) FindByID(id uint) (*model.Subscription, error) {
	logger.Debug("SubscriptionRepository.FindByID called",
		zap.Uint("id", id))

	var sub model.Subscription
	err := r.db.Where("id = ?", id).First(&sub).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Debug("Subscription not found",
				zap.Uint("id", id))
			return nil, nil
		}
		logger.Error("Failed to find subscription",
			zap.Uint("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	logger.Debug("Subscription found",
		zap.Uint("id", id),
		zap.String("city", sub.City))
	return &sub, nil
}

// FindByUserAndCity finds an active subscription by user ID and city
func (r *SubscriptionRepository) FindByUserAndCity(userID uint, city string) (*model.Subscription, error) {
	logger.Debug("SubscriptionRepository.FindByUserAndCity called",
		zap.Uint("user_id", userID),
		zap.String("city", city))

	var sub model.Subscription
	err := r.db.Where("user_id = ? AND city = ? AND active = ?", userID, city, true).First(&sub).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Debug("Subscription not found",
				zap.Uint("user_id", userID),
				zap.String("city", city))
			return nil, nil
		}
		logger.Error("Failed to find subscription",
			zap.Uint("user_id", userID),
			zap.String("city", city),
			zap.Error(err))
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	logger.Debug("Subscription found",
		zap.Uint("subscription_id", sub.ID),
		zap.Uint("user_id", userID),
		zap.String("city", city))
	return &sub, nil
}

// CountActiveByUser counts active subscriptions for a user
func (r *SubscriptionRepository) CountActiveByUser(userID uint) (int64, error) {
	logger.Debug("SubscriptionRepository.CountActiveByUser called",
		zap.Uint("user_id", userID))

	var count int64
	err := r.db.Model(&model.Subscription{}).
		Where("user_id = ? AND active = ?", userID, true).
		Count(&count).Error
	if err != nil {
		logger.Error("Failed to count subscriptions",
			zap.Uint("user_id", userID),
			zap.Error(err))
		return 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	logger.Debug("Subscription count retrieved",
		zap.Uint("user_id", userID),
		zap.Int64("count", count))
	return count, nil
}

// Delete soft deletes a subscription
func (r *SubscriptionRepository) Delete(id uint) error {
	logger.Debug("SubscriptionRepository.Delete called",
		zap.Uint("id", id))

	result := r.db.Delete(&model.Subscription{}, id)
	if result.Error != nil {
		logger.Error("Failed to delete subscription",
			zap.Uint("id", id),
			zap.Error(result.Error))
		return fmt.Errorf("failed to delete subscription: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		logger.Warn("Subscription not found for deletion",
			zap.Uint("id", id))
		return fmt.Errorf("subscription not found")
	}

	logger.Info("Subscription deleted successfully",
		zap.Uint("id", id))
	return nil
}
