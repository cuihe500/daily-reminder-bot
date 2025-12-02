package model

import "time"

// WarningLog stores information about sent warning notifications to avoid duplicates
type WarningLog struct {
	ID         uint      `gorm:"primarykey"`
	WarningID  string    `gorm:"uniqueIndex;not null"` // QWeather warning ID
	LocationID string    `gorm:"index;not null"`
	City       string    `gorm:"not null"`
	Type       string    `gorm:"not null"`
	Level      string    `gorm:"not null"`
	Title      string    `gorm:"not null"`
	StartTime  time.Time `gorm:"not null"`
	EndTime    time.Time
	Status     string    `gorm:"not null"` // active/update/cancel
	NotifiedAt time.Time // When the notification was sent
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
