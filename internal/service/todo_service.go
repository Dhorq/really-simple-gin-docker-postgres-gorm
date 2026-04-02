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

func (s *TodoService) GetAll() ([]*model.Todo, error) {
	return s.repo.GetAll()
}

func (s *TodoService) Get(id uint) (*model.Todo, error) {
	return s.repo.Get(id)
}

func (s *TodoService) Create(todo *model.Todo) error {
	if todo.Title == "" {
		return errors.New("Title is required")
	}
	return s.repo.Create(todo)
}

func (s *TodoService) Update(id uint, title string, completed bool) (*model.Todo, error) {
	if title == "" {
		return nil, errors.New("Title is required")
	}
	return s.repo.Update(id, title, completed)
}

func (s *TodoService) Delete(id uint) error {
	return s.repo.Delete(id)
}
