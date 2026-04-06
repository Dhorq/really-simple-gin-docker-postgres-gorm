package handler

import (
	"net/http"

	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/service"
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

	user, err := h.service.Register(input.Email, input.Password)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, user)
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

	user, accessToken, refreshToken, err := h.service.Login(input.Email, input.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid password")
		return
	}

	h.setCookies(c, accessToken, refreshToken)

	response.Success(c, gin.H{
		"user": user,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		response.Error(c, http.StatusUnauthorized, "refresh token required")
		return
	}

	accessToken, newRefreshToken, err := h.service.RefreshToken(refreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	h.setCookies(c, accessToken, newRefreshToken)

	response.Success(c, gin.H{
		"message": "token refreshed",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("userID")
	if exists {
		h.service.Logout(userID.(uint))
	}

	h.clearCookies(c)

	response.Success(c, gin.H{
		"message": "logout success",
	})
}

func (h *AuthHandler) setCookies(c *gin.Context, accessToken, refreshToken string) {
	c.SetCookie(
		"access_token",
		accessToken,
		900, // 15 menit
		"/",
		"",
		false,
		true,
	)

	c.SetCookie(
		"refresh_token",
		refreshToken,
		604800, // 7 hari
		"/",
		"",
		false,
		true,
	)
}

func (h *AuthHandler) clearCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.GetUserByID(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	response.Success(c, user)
}
