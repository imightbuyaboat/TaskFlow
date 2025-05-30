package schedular

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/imightbuyaboat/TaskFlow/pkg/queue"
	"github.com/imightbuyaboat/TaskFlow/task-schedular/internal/db"
	"go.uber.org/zap"
)

type Schedular struct {
	interval   time.Duration
	IntervalMs int `json:"schedularIntervalMs"`
	db         db.DB
	queue      queue.Queue
	logger     *zap.Logger
}

func NewSchedular(logger *zap.Logger) (*Schedular, error) {
	f, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}

	schedular := &Schedular{}
	if err = json.NewDecoder(f).Decode(schedular); err != nil {
		return nil, err
	}

	if schedular.IntervalMs <= 0 {
		return nil, fmt.Errorf("missing or incorrect interval")
	}

	postgresURL := fmt.Sprintf("postgres://%s:%s@db:%s/%s",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"))

	db, err := db.NewPostgresDB(postgresURL)
	if err != nil {
		return nil, err
	}

	amqpURL := fmt.Sprintf("amqp://%s:%s@rabbitmq:%s/",
		os.Getenv("AMQP_USER"), os.Getenv("AMQP_PASSWORD"),
		os.Getenv("AMQP_PORT"))

	queue, err := queue.NewRabbitMQQueue(amqpURL)
	if err != nil {
		return nil, err
	}

	schedular.interval = time.Duration(schedular.IntervalMs) * time.Millisecond
	schedular.db = db
	schedular.queue = queue
	schedular.logger = logger

	return schedular, nil
}

func (s *Schedular) EnterLoop() {
	for {
		time.Sleep(s.interval)

		s.processTasks()
	}
}

func (s *Schedular) processTasks() {
	tasks, err := s.db.GetPostponedTasks()
	if err != nil {
		s.logger.Error("failed to select postponed tasks", zap.Error(err))
		return
	}
	s.logger.Info("successfully select postponed tasks", zap.Int("count", len(tasks)))

	for _, t := range tasks {
		if err := s.db.UpdateStatusOfTask(t.ID, "queued"); err != nil {
			s.logger.Error("failed to update status of task", zap.Error(err), zap.String("task_id", t.ID.String()))
			continue
		}

		if err := s.queue.Publish(&t); err != nil {
			s.logger.Error("failed to publish task", zap.Error(err), zap.String("task_id", t.ID.String()))
		}

		s.logger.Info("successfully publish task", zap.String("task_id", t.ID.String()))
	}
}
