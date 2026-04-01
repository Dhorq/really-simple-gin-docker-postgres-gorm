package route

import (
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/handler"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, todoHandler *handler.TodoHandler) {
	api := r.Group("/api")
	{
		todos := api.Group("/todos")
		{
			todos.POST("", todoHandler.Create)
		}
	}
}
