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

func (r *TodoRepository) GetAll(userID uint, page, limit int, completed *bool) ([]*model.Todo, int64, error) {
	var todos []*model.Todo
	var total int64

	query := r.db.Model(&model.Todo{}).Where("user_id = ?", userID)

	if completed != nil {
		query = query.Where("completed = ?", *completed)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&todos).Error

	return todos, total, err
}

func (r *TodoRepository) Get(id uint, userID uint) (*model.Todo, error) {
	var todo model.Todo
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&todo).Error
	return &todo, err
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

	if err := r.db.Model(&todo).Updates(&todo).Error; err != nil {
		return nil, err
	}
	return todo, nil
}

func (r *TodoRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Todo{}).Error
}
