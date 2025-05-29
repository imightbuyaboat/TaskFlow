package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imightbuyaboat/TaskFlow/pkg/logger"
	"github.com/imightbuyaboat/TaskFlow/task-api/internal/handler"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger()
	log := logger.GetLogger()

	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env file", zap.Error(err))
	}

	h, err := handler.NewHandler(log)
	if err != nil {
		log.Fatal("failed to create handler", zap.Error(err))
	}

	r := gin.New()
	r.Use(gin.Recovery())

	r.POST("/api/register", h.RegisterHandler)
	r.POST("/api/login", h.LoginHandler)

	auth := r.Group("/api")
	auth.Use(h.AuthMiddleware())
	auth.POST("/tasks", h.CreateTaskHandler)
	auth.GET("/tasks/:id", h.GetTaskHandler)
	auth.GET("/tasks", h.GetAllTasksHandler)

	if err = http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("failed to start server", zap.Error(err))
	}
}
