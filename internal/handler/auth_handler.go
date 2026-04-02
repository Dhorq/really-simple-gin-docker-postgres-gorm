package handler

import (
	"net/http"

	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/model"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/service"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/pkg/auth"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{service: svc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := &model.User{
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	if err := h.service.CreateUser(user); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "user created",
		"user":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid input")
		return
	}

	user, err := h.service.GetUserByEmail(input.Email)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	c.SetCookie(
		"token", // name
		token,   // value
		86400,   // maxAge (24 jam dalam detik)
		"/",     // path
		"",      // domain
		false,   // secure (false buat dev, true buat prod)
		true,    // httpOnly
	)

	response.Success(c, gin.H{
		"user": user,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie(
		"token",
		"",
		-1, // maxAge -1 = hapus cookie
		"/",
		"",
		false,
		true,
	)

	response.Success(c, gin.H{
		"message": "logout success",
	})
}
