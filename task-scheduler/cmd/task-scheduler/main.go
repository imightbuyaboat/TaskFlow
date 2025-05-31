package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/imightbuyaboat/TaskFlow/pkg/logger"
	"github.com/imightbuyaboat/TaskFlow/pkg/queue"
	"github.com/imightbuyaboat/TaskFlow/task-scheduler/internal/db"
	"github.com/imightbuyaboat/TaskFlow/task-scheduler/internal/scheduler"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger()
	log := logger.GetLogger()

	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env file", zap.Error(err))
	}

	f, err := os.Open("config.json")
	if err != nil {
		log.Fatal("failed to open confug file", zap.Error(err))
	}

	var configsFromFile struct {
		IntervalMs int `json:"schedulerIntervalMs"`
	}

	if err = json.NewDecoder(f).Decode(&configsFromFile); err != nil {
		log.Fatal("failed to decode configs from file", zap.Error(err))
	}

	if configsFromFile.IntervalMs <= 0 {
		log.Fatal("invalid configs in file", zap.Int("interval", configsFromFile.IntervalMs))
	}

	postgresURL := fmt.Sprintf("postgres://%s:%s@db:%s/%s",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"))

	db, err := db.NewPostgresDB(postgresURL)
	if err != nil {
		log.Fatal("failed to connect to db", zap.Error(err))
	}

	amqpURL := fmt.Sprintf("amqp://%s:%s@rabbitmq:%s/",
		os.Getenv("AMQP_USER"), os.Getenv("AMQP_PASSWORD"),
		os.Getenv("AMQP_PORT"))

	queue, err := queue.NewRabbitMQQueue(amqpURL)
	if err != nil {
		log.Fatal("failed to connect to queue", zap.Error(err))
	}

	s, err := scheduler.NewScheduler(
		time.Duration(configsFromFile.IntervalMs)*time.Millisecond,
		db,
		queue,
		log,
	)
	if err != nil {
		log.Fatal("failed to create scheduler", zap.Error(err))
	}

	s.EnterLoop()
}
