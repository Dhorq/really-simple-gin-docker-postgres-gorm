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

func (r *TodoRepository) GetAll() ([]*model.Todo, error) {
	var todos []*model.Todo
	err := r.db.Find(&todos).Error
	return todos, err
}

func (r *TodoRepository) Get(id uint) (*model.Todo, error) {
	var todo *model.Todo
	err := r.db.First(&todo).Error
	return todo, err
}

func (r *TodoRepository) Create(todo *model.Todo) error {
	return r.db.Create(todo).Error
}

func (r *TodoRepository) Update(id uint, title string, completed bool) (*model.Todo, error) {
	var todo *model.Todo
	if err := r.db.First(&todo, id).Error; err != nil {
		return nil, err
	}

	todo.Title = title
	todo.Completed = completed

	if err := r.db.Save(&todo).Error; err != nil {
		return nil, err
	}
	return todo, nil
}

func (r *TodoRepository) Delete(id uint) error {
	var todo model.Todo
	if err := r.db.First(&todo, id).Error; err != nil {
		return err
	}
	return r.db.Delete(&todo).Error
}
