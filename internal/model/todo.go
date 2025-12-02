package model

import (
	"time"

	"gorm.io/gorm"
)

// Todo represents a user's todo item
type Todo struct {
	ID             uint           `gorm:"primarykey"`
	SubscriptionID uint           `gorm:"not null;index:idx_subscription_completed"` // Foreign key to Subscription
	Subscription   Subscription   `gorm:"foreignKey:SubscriptionID"`
	Content        string         `gorm:"not null"`                                                // Todo item content
	Completed      bool           `gorm:"not null;default:false;index:idx_subscription_completed"` // Whether the todo is completed
	CreatedAt      time.Time      `gorm:"not null"`
	UpdatedAt      time.Time      `gorm:"not null"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for Todo model
func (Todo) TableName() string {
	return "todos"
}
