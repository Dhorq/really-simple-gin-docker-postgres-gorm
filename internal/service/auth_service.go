package service

import (
	"errors"

	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/model"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/repository"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo *repository.AuthRepository
}

func NewAuthService(repo *repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Register(email, password string) (*model.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	existing, _ := s.repo.GetUserByEmail(email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	user := &model.User{Email: email, Password: string(hashed)}
	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (*model.User, string, string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}

	if user.Password != password {
		return nil, "", "", errors.New("invalid credentials")
	}

	accessToken, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	if err := s.repo.SaveRefreshToken(user.ID, refreshToken); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	storedToken, err := s.repo.GetRefreshToken(claims.UserID)
	if err != nil || storedToken != refreshToken {
		return "", "", errors.New("invalid refresh token")
	}

	user, err := s.repo.GetUserByID(claims.UserID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	newAccessToken, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := auth.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	if err := s.repo.SaveRefreshToken(user.ID, newRefreshToken); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *AuthService) Logout(userID uint) error {
	return s.repo.DeleteRefreshToken(userID)
}
