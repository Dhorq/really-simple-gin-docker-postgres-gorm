package repository

import (
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/model"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *AuthRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) SaveRefreshToken(userID uint, refreshToken string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("refresh_token", refreshToken).Error
}

func (r *AuthRepository) GetRefreshToken(userID uint) (string, error) {
	var user model.User
	if err := r.db.Select("refresh_token").First(&user, userID).Error; err != nil {
		return "", err
	}
	return user.RefreshToken, nil
}

func (r *AuthRepository) DeleteRefreshToken(userID uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("refresh_token", "").Error
}
