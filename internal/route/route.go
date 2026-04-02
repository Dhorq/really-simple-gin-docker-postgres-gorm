package route

import (
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/handler"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, todoHandler *handler.TodoHandler, authHandler *handler.AuthHandler) {
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		todos := api.Group("/todos")
		todos.Use(middleware.AuthMiddleware())
		{
			todos.GET("", todoHandler.GetAll)
			todos.GET(":id", todoHandler.Get)
			todos.POST("", todoHandler.Create)
			todos.PUT(":id", todoHandler.Update)
			todos.DELETE(":id", todoHandler.Delete)
		}
	}
}
