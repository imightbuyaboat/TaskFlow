package main

import (
	"github.com/imightbuyaboat/TaskFlow/pkg/logger"
	"github.com/imightbuyaboat/TaskFlow/task-schedular/internal/schedular"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger()
	log := logger.GetLogger()

	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env file", zap.Error(err))
	}

	s, err := schedular.NewSchedular(log)
	if err != nil {
		log.Fatal("failed to create schedular", zap.Error(err))
	}

	s.EnterLoop()
}
