package model

import (
	"time"

	"gorm.io/gorm"
)

// Subscription represents a user's daily reminder subscription
type Subscription struct {
	ID           uint           `gorm:"primarykey"`
	UserID       uint           `gorm:"not null;index"` // Foreign key to User
	User         User           `gorm:"foreignKey:UserID"`
	City         string         `gorm:"not null"`              // City for weather lookup (e.g., "北京", "上海")
	ReminderTime string         `gorm:"not null"`              // Daily reminder time in HH:MM format (e.g., "08:00")
	Active       bool           `gorm:"not null;default:true"` // Whether subscription is active
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for Subscription model
func (Subscription) TableName() string {
	return "subscriptions"
}
