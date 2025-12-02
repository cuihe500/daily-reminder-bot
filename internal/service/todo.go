package service

import (
	"fmt"
	"strings"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/internal/repository"
	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
)

// TodoService handles todo-related business logic
type TodoService struct {
	todoRepo *repository.TodoRepository
}

// NewTodoService creates a new TodoService
func NewTodoService(todoRepo *repository.TodoRepository) *TodoService {
	return &TodoService{todoRepo: todoRepo}
}

// AddTodo adds a new todo item for a user
func (s *TodoService) AddTodo(userID uint, content string) error {
	logger.Debug("AddTodo called",
		zap.Uint("user_id", userID),
		zap.String("content", content))

	todo := &model.Todo{
		UserID:  userID,
		Content: content,
	}
	if err := s.todoRepo.Create(todo); err != nil {
		logger.Error("Failed to add todo",
			zap.Uint("user_id", userID),
			zap.String("content", content),
			zap.Error(err))
		return err
	}

	logger.Info("Todo added successfully",
		zap.Uint("user_id", userID),
		zap.Uint("todo_id", todo.ID))
	return nil
}

// GetUserTodos retrieves all todos for a user
func (s *TodoService) GetUserTodos(userID uint) ([]model.Todo, error) {
	logger.Debug("GetUserTodos called", zap.Uint("user_id", userID))

	todos, err := s.todoRepo.FindByUserID(userID)
	if err != nil {
		logger.Error("Failed to get user todos",
			zap.Uint("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	logger.Debug("User todos retrieved",
		zap.Uint("user_id", userID),
		zap.Int("count", len(todos)))
	return todos, nil
}

// GetIncompleteTodos retrieves incomplete todos for a user
func (s *TodoService) GetIncompleteTodos(userID uint) ([]model.Todo, error) {
	logger.Debug("GetIncompleteTodos called", zap.Uint("user_id", userID))

	todos, err := s.todoRepo.FindIncompleteByUserID(userID)
	if err != nil {
		logger.Error("Failed to get incomplete todos",
			zap.Uint("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	logger.Debug("Incomplete todos retrieved",
		zap.Uint("user_id", userID),
		zap.Int("count", len(todos)))
	return todos, nil
}

// CompleteTodo marks a todo as completed
func (s *TodoService) CompleteTodo(todoID uint, userID uint) error {
	logger.Debug("CompleteTodo called",
		zap.Uint("todo_id", todoID),
		zap.Uint("user_id", userID))

	todo, err := s.todoRepo.FindByID(todoID)
	if err != nil {
		logger.Error("Failed to find todo",
			zap.Uint("todo_id", todoID),
			zap.Error(err))
		return err
	}
	if todo == nil {
		logger.Warn("Todo not found",
			zap.Uint("todo_id", todoID),
			zap.Uint("user_id", userID))
		return fmt.Errorf("todo not found")
	}
	if todo.UserID != userID {
		logger.Warn("Unauthorized todo access",
			zap.Uint("todo_id", todoID),
			zap.Uint("user_id", userID),
			zap.Uint("owner_id", todo.UserID))
		return fmt.Errorf("unauthorized")
	}

	todo.Completed = true
	if err := s.todoRepo.Update(todo); err != nil {
		logger.Error("Failed to complete todo",
			zap.Uint("todo_id", todoID),
			zap.Error(err))
		return err
	}

	logger.Info("Todo completed successfully",
		zap.Uint("todo_id", todoID),
		zap.Uint("user_id", userID))
	return nil
}

// DeleteTodo deletes a todo item
func (s *TodoService) DeleteTodo(todoID uint, userID uint) error {
	logger.Debug("DeleteTodo called",
		zap.Uint("todo_id", todoID),
		zap.Uint("user_id", userID))

	todo, err := s.todoRepo.FindByID(todoID)
	if err != nil {
		logger.Error("Failed to find todo",
			zap.Uint("todo_id", todoID),
			zap.Error(err))
		return err
	}
	if todo == nil {
		logger.Warn("Todo not found",
			zap.Uint("todo_id", todoID),
			zap.Uint("user_id", userID))
		return fmt.Errorf("todo not found")
	}
	if todo.UserID != userID {
		logger.Warn("Unauthorized todo access",
			zap.Uint("todo_id", todoID),
			zap.Uint("user_id", userID),
			zap.Uint("owner_id", todo.UserID))
		return fmt.Errorf("unauthorized")
	}

	if err := s.todoRepo.Delete(todoID); err != nil {
		logger.Error("Failed to delete todo",
			zap.Uint("todo_id", todoID),
			zap.Error(err))
		return err
	}

	logger.Info("Todo deleted successfully",
		zap.Uint("todo_id", todoID),
		zap.Uint("user_id", userID))
	return nil
}

// FormatTodoList formats a list of todos for display
func (s *TodoService) FormatTodoList(todos []model.Todo) string {
	if len(todos) == 0 {
		return "📝 暂无待办事项"
	}

	var builder strings.Builder
	builder.WriteString("📝 待办事项列表：\n\n")

	for i, todo := range todos {
		status := "⬜"
		if todo.Completed {
			status = "✅"
		}
		builder.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, status, todo.Content))
	}

	return builder.String()
}
