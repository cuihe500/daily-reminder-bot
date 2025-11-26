package model

import (
	"time"

	"gorm.io/gorm"
)

// Todo represents a user's todo item
type Todo struct {
	ID        uint           `gorm:"primarykey"`
	UserID    uint           `gorm:"not null;index"` // Foreign key to User
	User      User           `gorm:"foreignKey:UserID"`
	Content   string         `gorm:"not null"`               // Todo item content
	Completed bool           `gorm:"not null;default:false"` // Whether the todo is completed
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for Todo model
func (Todo) TableName() string {
	return "todos"
}
