package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a Telegram user in the system
type User struct {
	ID        uint           `gorm:"primarykey"`
	ChatID    int64          `gorm:"uniqueIndex;not null"` // Telegram chat ID
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}
