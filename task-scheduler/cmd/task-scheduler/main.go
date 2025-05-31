package main

import (
	"github.com/imightbuyaboat/TaskFlow/pkg/logger"
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

	s, err := scheduler.NewScheduler(log)
	if err != nil {
		log.Fatal("failed to create scheduler", zap.Error(err))
	}

	s.EnterLoop()
}
