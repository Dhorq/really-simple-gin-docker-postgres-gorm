package handler

import (
	"net/http"

	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/model"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/service"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/pkg/response"
	"github.com/gin-gonic/gin"
)

type TodoHandler struct {
	service *service.TodoService
}

func NewTodoHandler(svc *service.TodoService) *TodoHandler {
	return &TodoHandler{service: svc}
}

func (h *TodoHandler) Create(c *gin.Context) {
	var input struct {
		Title     string `json:"title" binding:"required"`
		Completed bool   `json:"completed"`
	}

	todo := &model.Todo{
		Title:     input.Title,
		Completed: input.Completed,
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Title is required")
		return
	}
	response.Success(c, todo)
}
