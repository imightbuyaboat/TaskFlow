package scheduler

import (
	"time"

	"github.com/imightbuyaboat/TaskFlow/pkg/queue"
	"go.uber.org/zap"
)

type Scheduler struct {
	interval time.Duration
	db       DB
	queue    queue.Queue
	logger   *zap.Logger
}

func NewScheduler(interval time.Duration, db DB, queue queue.Queue, logger *zap.Logger) (*Scheduler, error) {
	return &Scheduler{
		interval: interval,
		db:       db,
		queue:    queue,
		logger:   logger,
	}, nil
}

func (s *Scheduler) EnterLoop() {
	for {
		time.Sleep(s.interval)

		s.processTasks()
	}
}

func (s *Scheduler) processTasks() {
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
