package main

import (
	"log"

	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/config"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/database"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/handler"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/model"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/repository"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/route"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadEnv()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database")
	}

	database.AutoMigrate(db, &model.Todo{})

	todoRepo := repository.NewTodoRepository(db)
	todoSvc := service.NewTodoService(todoRepo)
	todoHandler := handler.NewTodoHandler(todoSvc)

	r := gin.Default()
	route.SetupRoutes(r, todoHandler)

	log.Println("Server running on port 8080")
	r.Run(":" + cfg.Port)
}
