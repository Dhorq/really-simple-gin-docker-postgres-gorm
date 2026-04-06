package handler

import (
	"net/http"
	"strconv"

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

func (h *TodoHandler) GetAll(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid := userID.(uint)

	var pagination model.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid pagination params")
		return
	}
	pagination.SetDefaults()

	var completed *bool
	if cp := c.Query("completed"); cp != "" {
		b := cp == "true"
		completed = &b
	}

	todos, total, err := h.service.GetAll(uid, pagination.Page, pagination.Limit, completed)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessPaginated(c, pagination.Page, pagination.Limit, total, todos)
}

func (h *TodoHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	userID, _ := c.Get("userID")
	uid := userID.(uint)

	todo, err := h.service.Get(uint(id), uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, todo)
}

func (h *TodoHandler) Create(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input struct {
		Title     string `json:"title" binding:"required"`
		Completed bool   `json:"completed"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Title is required")
		return
	}

	todo := &model.Todo{
		Title:     input.Title,
		Completed: input.Completed,
		UserID:    userID.(uint),
	}

	if err := h.service.Create(todo); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, todo)
}

func (h *TodoHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	var input struct {
		Title     string `json:"title" binding:"required"`
		Completed bool   `json:"completed"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Title is required")
		return
	}

	todo, err := h.service.Update(uint(id), input.Title, input.Completed)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, todo)
}

func (h *TodoHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	userID, _ := c.Get("userID")
	uid := userID.(uint)

	if err := h.service.Delete(uint(id), uid); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "Todo deleted successfully")
}
