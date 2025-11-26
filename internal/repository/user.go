package repository

import (
	"fmt"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"gorm.io/gorm"
)

// UserRepository handles user data access
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *model.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// FindByChatID finds a user by Telegram chat ID
func (r *UserRepository) FindByChatID(chatID int64) (*model.User, error) {
	var user model.User
	err := r.db.Where("chat_id = ?", chatID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

// GetOrCreate finds a user by chat ID or creates a new one
func (r *UserRepository) GetOrCreate(chatID int64) (*model.User, error) {
	user, err := r.FindByChatID(chatID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	// Create new user
	user = &model.User{ChatID: chatID}
	if err := r.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}
