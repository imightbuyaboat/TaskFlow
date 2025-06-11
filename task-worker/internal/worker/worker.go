package worker

import (
	"encoding/json"
	"fmt"

	"github.com/imightbuyaboat/TaskFlow/pkg/queue"
	"github.com/imightbuyaboat/TaskFlow/pkg/task"
	"github.com/imightbuyaboat/TaskFlow/task-worker/internal/db"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Worker struct {
	id        int
	ch        *amqp.Channel
	msgs      <-chan amqp.Delivery
	executers map[string]Executer
	db        DB
	logger    *zap.Logger
}

func NewWorker(id int, q queue.Queue, executers map[string]Executer, db DB, logger *zap.Logger) (*Worker, error) {
	ch, msgs, err := q.NewConsumerChannel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %v", err)
	}

	return &Worker{
		id:        id,
		ch:        ch,
		msgs:      msgs,
		executers: executers,
		db:        db,
		logger:    logger,
	}, nil
}

func (w *Worker) Work() {
	defer w.ch.Close()

	for d := range w.msgs {
		w.processMsg(&d)
	}
}

func (w *Worker) processMsg(d *amqp.Delivery) {
	w.logger.Info("successfully delivered message", zap.Int("worker", w.id))

	var t task.Task
	if err := json.Unmarshal(d.Body, &t); err != nil {
		w.logger.Error("failed to unmarshal body of message", zap.Error(err), zap.Int("worker", w.id))
		d.Nack(false, false)
		return
	}

	if err := w.db.UpdateStatusOfTask(t.ID, "processing"); err != nil {
		if err == db.ErrMaxRetriesReached {
			w.logger.Info("reached max retries", zap.Int("worker", w.id), zap.String("task_id", t.ID.String()))
			d.Nack(false, false)
			return
		}

		w.logger.Error("failed to update status of task", zap.Error(err), zap.Int("worker", w.id), zap.String("task_id", t.ID.String()))
		d.Nack(false, false)
		return
	}

	if err := w.executers[t.Type].ExecuteTask(t.Payload); err != nil {
		w.logger.Error("failed to execute task", zap.Error(err), zap.Int("worker", w.id), zap.String("task_id", t.ID.String()))

		if err := w.db.UpdateStatusOfTask(t.ID, "error"); err != nil {
			w.logger.Error("failed to update status of task", zap.Error(err), zap.Int("worker", w.id), zap.String("task_id", t.ID.String()))
		}
		d.Nack(false, true)
		return
	}

	w.logger.Info("succesfully complete task", zap.Int("worker", w.id), zap.String("task_id", t.ID.String()))

	if err := w.db.UpdateStatusOfTask(t.ID, "done"); err != nil {
		w.logger.Error("failed to update status of task", zap.Error(err), zap.Int("worker", w.id), zap.String("task_id", t.ID.String()))
	}
	d.Ack(false)
}
