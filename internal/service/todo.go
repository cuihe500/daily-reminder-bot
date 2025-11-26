package service

import (
	"fmt"
	"strings"

	"github.com/cuichanghe/daily-reminder-bot/internal/model"
	"github.com/cuichanghe/daily-reminder-bot/internal/repository"
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
	todo := &model.Todo{
		UserID:  userID,
		Content: content,
	}
	return s.todoRepo.Create(todo)
}

// GetUserTodos retrieves all todos for a user
func (s *TodoService) GetUserTodos(userID uint) ([]model.Todo, error) {
	return s.todoRepo.FindByUserID(userID)
}

// GetIncompleteTodos retrieves incomplete todos for a user
func (s *TodoService) GetIncompleteTodos(userID uint) ([]model.Todo, error) {
	return s.todoRepo.FindIncompleteByUserID(userID)
}

// CompleteTodo marks a todo as completed
func (s *TodoService) CompleteTodo(todoID uint, userID uint) error {
	todo, err := s.todoRepo.FindByID(todoID)
	if err != nil {
		return err
	}
	if todo == nil {
		return fmt.Errorf("todo not found")
	}
	if todo.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	todo.Completed = true
	return s.todoRepo.Update(todo)
}

// DeleteTodo deletes a todo item
func (s *TodoService) DeleteTodo(todoID uint, userID uint) error {
	todo, err := s.todoRepo.FindByID(todoID)
	if err != nil {
		return err
	}
	if todo == nil {
		return fmt.Errorf("todo not found")
	}
	if todo.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	return s.todoRepo.Delete(todoID)
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
