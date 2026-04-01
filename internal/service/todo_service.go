package service

import (
	"errors"

	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/model"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/repository"
)

type TodoService struct {
	repo *repository.TodoRepository
}

func NewTodoService(repo *repository.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

func (s *TodoService) Create(todo *model.Todo) error {
	if todo.Title == "" {
		return errors.New("Title is required")
	}
	return s.repo.Create(todo)
}
