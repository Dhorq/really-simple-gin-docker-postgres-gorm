package route

import (
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/handler"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/pkg/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, todoHandler *handler.TodoHandler, authHandler *handler.AuthHandler) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", middleware.AuthMiddleware(), authHandler.Me)
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
