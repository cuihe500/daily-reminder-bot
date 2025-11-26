package repository

import (
	"fmt"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
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
	if err := r.db.Create(sub).Error; err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

// FindByUserID finds a subscription by user ID
func (r *SubscriptionRepository) FindByUserID(userID uint) (*model.Subscription, error) {
	var sub model.Subscription
	err := r.db.Where("user_id = ? AND active = ?", userID, true).First(&sub).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}
	return &sub, nil
}

// Update updates a subscription
func (r *SubscriptionRepository) Update(sub *model.Subscription) error {
	if err := r.db.Save(sub).Error; err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}
	return nil
}

// GetAllActive retrieves all active subscriptions
func (r *SubscriptionRepository) GetAllActive() ([]model.Subscription, error) {
	var subs []model.Subscription
	err := r.db.Preload("User").Where("active = ?", true).Find(&subs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}
	return subs, nil
}

// GetByReminderTime retrieves active subscriptions for a specific reminder time
func (r *SubscriptionRepository) GetByReminderTime(reminderTime string) ([]model.Subscription, error) {
	var subs []model.Subscription
	err := r.db.Preload("User").Where("active = ? AND reminder_time = ?", true, reminderTime).Find(&subs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions by reminder time: %w", err)
	}
	return subs, nil
}
