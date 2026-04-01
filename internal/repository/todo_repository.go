package repository

import (
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/model"
	"gorm.io/gorm"
)

type TodoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) Create(todo *model.Todo) error {
	return r.db.Create(todo).Error
}
