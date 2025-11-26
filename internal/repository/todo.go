package repository

import (
	"fmt"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
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
	if err := r.db.Create(todo).Error; err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}
	return nil
}

// FindByUserID retrieves all todos for a user
func (r *TodoRepository) FindByUserID(userID uint) ([]model.Todo, error) {
	var todos []model.Todo
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&todos).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find todos: %w", err)
	}
	return todos, nil
}

// FindIncompleteByUserID retrieves incomplete todos for a user
func (r *TodoRepository) FindIncompleteByUserID(userID uint) ([]model.Todo, error) {
	var todos []model.Todo
	err := r.db.Where("user_id = ? AND completed = ?", userID, false).Order("created_at DESC").Find(&todos).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find incomplete todos: %w", err)
	}
	return todos, nil
}

// Update updates a todo item
func (r *TodoRepository) Update(todo *model.Todo) error {
	if err := r.db.Save(todo).Error; err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}
	return nil
}

// Delete deletes a todo item
func (r *TodoRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.Todo{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}
	return nil
}

// FindByID finds a todo by ID
func (r *TodoRepository) FindByID(id uint) (*model.Todo, error) {
	var todo model.Todo
	err := r.db.First(&todo, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find todo: %w", err)
	}
	return &todo, nil
}
