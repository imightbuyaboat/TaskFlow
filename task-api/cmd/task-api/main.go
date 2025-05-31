package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/imightbuyaboat/TaskFlow/pkg/logger"
	"github.com/imightbuyaboat/TaskFlow/pkg/queue"
	"github.com/imightbuyaboat/TaskFlow/task-api/internal/auth"
	"github.com/imightbuyaboat/TaskFlow/task-api/internal/db"
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

	postgresURL := fmt.Sprintf("postgres://%s:%s@db:%s/%s",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"))

	db, err := db.NewPostgresDB(postgresURL)
	if err != nil {
		log.Fatal("failed to create PostgresDB connection", zap.Error(err))
	}

	amqpURL := fmt.Sprintf("amqp://%s:%s@rabbitmq:%s/",
		os.Getenv("AMQP_USER"), os.Getenv("AMQP_PASSWORD"),
		os.Getenv("AMQP_PORT"))

	queue, err := queue.NewRabbitMQQueue(amqpURL)
	if err != nil {
		log.Fatal("failed to create RabbitMQ connection", zap.Error(err))
	}

	tm := auth.NewJWTManager()

	h, err := handler.NewHandler(db, queue, tm, log)
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
