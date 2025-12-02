package model

import (
	"time"

	"gorm.io/gorm"
)

// Subscription represents a user's daily reminder subscription
type Subscription struct {
	ID            uint           `gorm:"primarykey"`
	UserID        uint           `gorm:"not null;index:idx_user_city_time"` // Foreign key to User
	User          User           `gorm:"foreignKey:UserID"`
	City          string         `gorm:"not null;index:idx_user_city_time"` // City for weather lookup (e.g., "北京", "上海")
	ReminderTime  string         `gorm:"not null;index:idx_user_city_time"` // Daily reminder time in HH:MM format (e.g., "08:00")
	Active        bool           `gorm:"not null;default:true;index"`       // Whether subscription is active
	EnableWarning bool           `gorm:"not null;default:true"`             // Whether weather warning notifications are enabled
	Todos         []Todo         `gorm:"foreignKey:SubscriptionID"`         // Associated todos for this subscription
	CreatedAt     time.Time      `gorm:"not null"`
	UpdatedAt     time.Time      `gorm:"not null"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for Subscription model
func (Subscription) TableName() string {
	return "subscriptions"
}
