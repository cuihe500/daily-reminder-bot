package repository

import (
	"fmt"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
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
	logger.Debug("UserRepository.Create called",
		zap.Int64("chat_id", user.ChatID))

	if err := r.db.Create(user).Error; err != nil {
		logger.Error("Failed to create user",
			zap.Int64("chat_id", user.ChatID),
			zap.Error(err))
		return fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info("User created successfully",
		zap.Int64("chat_id", user.ChatID),
		zap.Uint("user_id", user.ID))
	return nil
}

// FindByChatID finds a user by Telegram chat ID
func (r *UserRepository) FindByChatID(chatID int64) (*model.User, error) {
	logger.Debug("UserRepository.FindByChatID called",
		zap.Int64("chat_id", chatID))

	var user model.User
	err := r.db.Where("chat_id = ?", chatID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Debug("User not found",
				zap.Int64("chat_id", chatID))
			return nil, nil
		}
		logger.Error("Failed to find user",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	logger.Debug("User found",
		zap.Int64("chat_id", chatID),
		zap.Uint("user_id", user.ID))
	return &user, nil
}

// GetOrCreate finds a user by chat ID or creates a new one
func (r *UserRepository) GetOrCreate(chatID int64) (*model.User, error) {
	logger.Debug("UserRepository.GetOrCreate called",
		zap.Int64("chat_id", chatID))

	user, err := r.FindByChatID(chatID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		logger.Debug("Existing user returned",
			zap.Int64("chat_id", chatID),
			zap.Uint("user_id", user.ID))
		return user, nil
	}

	// Create new user
	logger.Debug("Creating new user",
		zap.Int64("chat_id", chatID))
	user = &model.User{ChatID: chatID}
	if err := r.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}
