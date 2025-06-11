package worker

import (
	"github.com/google/uuid"
)

type DB interface {
	UpdateStatusOfTask(taskID uuid.UUID, status string) error
}
