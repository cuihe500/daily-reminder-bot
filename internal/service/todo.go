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

// AddTodo adds a new todo item for a subscription
func (s *TodoService) AddTodo(subscriptionID uint, content string) error {
	logger.Debug("AddTodo called",
		zap.Uint("subscription_id", subscriptionID),
		zap.String("content", content))

	todo := &model.Todo{
		SubscriptionID: subscriptionID,
		Content:        content,
	}
	if err := s.todoRepo.Create(todo); err != nil {
		logger.Error("Failed to add todo",
			zap.Uint("subscription_id", subscriptionID),
			zap.String("content", content),
			zap.Error(err))
		return err
	}

	logger.Info("Todo added successfully",
		zap.Uint("subscription_id", subscriptionID),
		zap.Uint("todo_id", todo.ID))
	return nil
}

// GetSubscriptionTodos retrieves all todos for a subscription
func (s *TodoService) GetSubscriptionTodos(subscriptionID uint) ([]model.Todo, error) {
	logger.Debug("GetSubscriptionTodos called", zap.Uint("subscription_id", subscriptionID))

	todos, err := s.todoRepo.FindBySubscriptionID(subscriptionID)
	if err != nil {
		logger.Error("Failed to get subscription todos",
			zap.Uint("subscription_id", subscriptionID),
			zap.Error(err))
		return nil, err
	}

	logger.Debug("Subscription todos retrieved",
		zap.Uint("subscription_id", subscriptionID),
		zap.Int("count", len(todos)))
	return todos, nil
}

// GetIncompleteTodos retrieves incomplete todos for a subscription
func (s *TodoService) GetIncompleteTodos(subscriptionID uint) ([]model.Todo, error) {
	logger.Debug("GetIncompleteTodos called", zap.Uint("subscription_id", subscriptionID))

	todos, err := s.todoRepo.FindIncompleteBySubscriptionID(subscriptionID)
	if err != nil {
		logger.Error("Failed to get incomplete todos",
			zap.Uint("subscription_id", subscriptionID),
			zap.Error(err))
		return nil, err
	}

	logger.Debug("Incomplete todos retrieved",
		zap.Uint("subscription_id", subscriptionID),
		zap.Int("count", len(todos)))
	return todos, nil
}

// CompleteTodo marks a todo as completed
func (s *TodoService) CompleteTodo(todoID uint, userID uint) error {
	logger.Debug("CompleteTodo called",
		zap.Uint("todo_id", todoID),
		zap.Uint("user_id", userID))

	todo, err := s.todoRepo.FindByIDAndVerifyOwnership(todoID, userID)
	if err != nil {
		if err.Error() == "unauthorized" {
			logger.Warn("Unauthorized todo access",
				zap.Uint("todo_id", todoID),
				zap.Uint("user_id", userID))
			return fmt.Errorf("unauthorized")
		}
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

	todo, err := s.todoRepo.FindByIDAndVerifyOwnership(todoID, userID)
	if err != nil {
		if err.Error() == "unauthorized" {
			logger.Warn("Unauthorized todo access",
				zap.Uint("todo_id", todoID),
				zap.Uint("user_id", userID))
			return fmt.Errorf("unauthorized")
		}
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

// FormatTodoListWithCity formats a list of todos for display with city information
func (s *TodoService) FormatTodoListWithCity(todos []model.Todo, city string) string {
	if len(todos) == 0 {
		return fmt.Sprintf("📝 %s - 暂无待办事项", city)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📝 %s - 待办事项列表：\n\n", city))

	for i, todo := range todos {
		status := "⬜"
		if todo.Completed {
			status = "✅"
		}
		builder.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, status, todo.Content))
	}

	return builder.String()
}
