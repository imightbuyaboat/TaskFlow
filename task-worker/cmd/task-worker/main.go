package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/imightbuyaboat/TaskFlow/pkg/logger"
	"github.com/imightbuyaboat/TaskFlow/pkg/queue"
	"github.com/imightbuyaboat/TaskFlow/task-worker/internal/db"
	"github.com/imightbuyaboat/TaskFlow/task-worker/internal/email"
	"github.com/imightbuyaboat/TaskFlow/task-worker/internal/worker"
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

	dialer, err := email.NewMailDialer()
	if err != nil {
		log.Fatal("failed to create mail dialer", zap.Error(err))
	}

	executers := map[string]worker.Executer{
		"send_email": dialer,
	}

	numOfWorkersStr := os.Getenv("NUMOFWORKERS")
	numOfWorkers, err := strconv.Atoi(numOfWorkersStr)
	if err != nil {
		log.Fatal("incorrect NUMOFWORKERS format", zap.Error(err))
	}

	for i := 0; i < numOfWorkers; i++ {
		w, err := worker.NewWorker(i+1, queue, executers, db, log)
		if err != nil {
			log.Fatal("failed to create worker", zap.Error(err))
		}
		go w.Work()
	}

	select {}
}
