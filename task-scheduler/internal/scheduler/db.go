package scheduler

import (
	"github.com/google/uuid"
	"github.com/imightbuyaboat/TaskFlow/pkg/task"
)

type DB interface {
	GetPostponedTasks() ([]task.Task, error)
	UpdateStatusOfTask(taskID uuid.UUID, status string) error
}
