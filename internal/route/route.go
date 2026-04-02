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
			todos.GET("", todoHandler.GetAll)
			todos.GET(":id", todoHandler.Get)
			todos.POST("", todoHandler.Create)
			todos.PUT(":id", todoHandler.Update)
			todos.DELETE(":id", todoHandler.Delete)
		}
	}
}
