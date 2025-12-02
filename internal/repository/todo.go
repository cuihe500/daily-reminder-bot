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
		zap.Uint("user_id", todo.UserID),
		zap.String("content", todo.Content))

	if err := r.db.Create(todo).Error; err != nil {
		logger.Error("Failed to create todo",
			zap.Uint("user_id", todo.UserID),
			zap.Error(err))
		return fmt.Errorf("failed to create todo: %w", err)
	}

	logger.Info("Todo created successfully",
		zap.Uint("todo_id", todo.ID),
		zap.Uint("user_id", todo.UserID))
	return nil
}

// FindByUserID retrieves all todos for a user
func (r *TodoRepository) FindByUserID(userID uint) ([]model.Todo, error) {
	logger.Debug("TodoRepository.FindByUserID called",
		zap.Uint("user_id", userID))

	var todos []model.Todo
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&todos).Error
	if err != nil {
		logger.Error("Failed to find todos",
			zap.Uint("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to find todos: %w", err)
	}

	logger.Debug("Todos found",
		zap.Uint("user_id", userID),
		zap.Int("count", len(todos)))
	return todos, nil
}

// FindIncompleteByUserID retrieves incomplete todos for a user
func (r *TodoRepository) FindIncompleteByUserID(userID uint) ([]model.Todo, error) {
	logger.Debug("TodoRepository.FindIncompleteByUserID called",
		zap.Uint("user_id", userID))

	var todos []model.Todo
	err := r.db.Where("user_id = ? AND completed = ?", userID, false).Order("created_at DESC").Find(&todos).Error
	if err != nil {
		logger.Error("Failed to find incomplete todos",
			zap.Uint("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to find incomplete todos: %w", err)
	}

	logger.Debug("Incomplete todos found",
		zap.Uint("user_id", userID),
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
	err := r.db.First(&todo, id).Error
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
		zap.Uint("user_id", todo.UserID))
	return &todo, nil
}
