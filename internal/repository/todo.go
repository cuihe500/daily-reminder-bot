package repository

import (
	"fmt"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TodoRepository handles todo data access
type TodoRepository struct {
	db *gorm.DB
}

// NewTodoRepository creates a new TodoRepository
func NewTodoRepository(db *gorm.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// Create creates a new todo item
func (r *TodoRepository) Create(todo *model.Todo) error {
	logger.Debug("TodoRepository.Create called",
		zap.Uint("subscription_id", todo.SubscriptionID),
		zap.String("content", todo.Content))

	if err := r.db.Create(todo).Error; err != nil {
		logger.Error("Failed to create todo",
			zap.Uint("subscription_id", todo.SubscriptionID),
			zap.Error(err))
		return fmt.Errorf("failed to create todo: %w", err)
	}

	logger.Info("Todo created successfully",
		zap.Uint("todo_id", todo.ID),
		zap.Uint("subscription_id", todo.SubscriptionID))
	return nil
}

// FindBySubscriptionID retrieves all todos for a subscription
func (r *TodoRepository) FindBySubscriptionID(subscriptionID uint) ([]model.Todo, error) {
	logger.Debug("TodoRepository.FindBySubscriptionID called",
		zap.Uint("subscription_id", subscriptionID))

	var todos []model.Todo
	err := r.db.Where("subscription_id = ?", subscriptionID).Order("created_at DESC").Find(&todos).Error
	if err != nil {
		logger.Error("Failed to find todos",
			zap.Uint("subscription_id", subscriptionID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to find todos: %w", err)
	}

	logger.Debug("Todos found",
		zap.Uint("subscription_id", subscriptionID),
		zap.Int("count", len(todos)))
	return todos, nil
}

// FindIncompleteBySubscriptionID retrieves incomplete todos for a subscription
func (r *TodoRepository) FindIncompleteBySubscriptionID(subscriptionID uint) ([]model.Todo, error) {
	logger.Debug("TodoRepository.FindIncompleteBySubscriptionID called",
		zap.Uint("subscription_id", subscriptionID))

	var todos []model.Todo
	err := r.db.Where("subscription_id = ? AND completed = ?", subscriptionID, false).Order("created_at DESC").Find(&todos).Error
	if err != nil {
		logger.Error("Failed to find incomplete todos",
			zap.Uint("subscription_id", subscriptionID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to find incomplete todos: %w", err)
	}

	logger.Debug("Incomplete todos found",
		zap.Uint("subscription_id", subscriptionID),
		zap.Int("count", len(todos)))
	return todos, nil
}

// Update updates a todo item
func (r *TodoRepository) Update(todo *model.Todo) error {
	logger.Debug("TodoRepository.Update called",
		zap.Uint("todo_id", todo.ID),
		zap.Bool("completed", todo.Completed))

	if err := r.db.Save(todo).Error; err != nil {
		logger.Error("Failed to update todo",
			zap.Uint("todo_id", todo.ID),
			zap.Error(err))
		return fmt.Errorf("failed to update todo: %w", err)
	}

	logger.Debug("Todo updated successfully",
		zap.Uint("todo_id", todo.ID))
	return nil
}

// Delete deletes a todo item
func (r *TodoRepository) Delete(id uint) error {
	logger.Debug("TodoRepository.Delete called",
		zap.Uint("todo_id", id))

	if err := r.db.Delete(&model.Todo{}, id).Error; err != nil {
		logger.Error("Failed to delete todo",
			zap.Uint("todo_id", id),
			zap.Error(err))
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	logger.Info("Todo deleted successfully",
		zap.Uint("todo_id", id))
	return nil
}

// FindByID finds a todo by ID
func (r *TodoRepository) FindByID(id uint) (*model.Todo, error) {
	logger.Debug("TodoRepository.FindByID called",
		zap.Uint("todo_id", id))

	var todo model.Todo
	err := r.db.Preload("Subscription").First(&todo, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Debug("Todo not found",
				zap.Uint("todo_id", id))
			return nil, nil
		}
		logger.Error("Failed to find todo",
			zap.Uint("todo_id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to find todo: %w", err)
	}

	logger.Debug("Todo found",
		zap.Uint("todo_id", id),
		zap.Uint("subscription_id", todo.SubscriptionID))
	return &todo, nil
}

// FindByIDAndVerifyOwnership finds a todo by ID and verifies the user owns it
func (r *TodoRepository) FindByIDAndVerifyOwnership(todoID uint, userID uint) (*model.Todo, error) {
	logger.Debug("TodoRepository.FindByIDAndVerifyOwnership called",
		zap.Uint("todo_id", todoID),
		zap.Uint("user_id", userID))

	var todo model.Todo
	err := r.db.Preload("Subscription").First(&todo, todoID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Debug("Todo not found",
				zap.Uint("todo_id", todoID))
			return nil, nil
		}
		logger.Error("Failed to find todo",
			zap.Uint("todo_id", todoID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to find todo: %w", err)
	}

	// Verify ownership
	if todo.Subscription.UserID != userID {
		logger.Warn("Unauthorized todo access",
			zap.Uint("todo_id", todoID),
			zap.Uint("user_id", userID),
			zap.Uint("owner_id", todo.Subscription.UserID))
		return nil, fmt.Errorf("unauthorized")
	}

	logger.Debug("Todo found and ownership verified",
		zap.Uint("todo_id", todoID),
		zap.Uint("user_id", userID))
	return &todo, nil
}
