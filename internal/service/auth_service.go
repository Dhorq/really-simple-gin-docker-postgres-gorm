package service

import (
	"errors"

	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/model"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/repository"
)

type AuthService struct {
	repo *repository.AuthRepository
}

func NewAuthHandler(repo *repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) CreateUser(user *model.User) error {
	if user.Email == "" {
		return errors.New("Email is required")
	}
	if user.Password == "" {
		return errors.New("Password is required")
	}

	existingUser, _ := s.repo.GetUserByEmail(user.Email)
	if existingUser != nil {
		return errors.New("Email already registered")
	}

	return s.repo.CreateUser(user)
}

func (s *AuthService) GetUserByEmail(email string) (*model.User, error) {
	if email == "" {
		return nil, errors.New("Email is required")
	}
	return s.repo.GetUserByEmail(email)
}
