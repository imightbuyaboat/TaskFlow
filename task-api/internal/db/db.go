package db

import (
	"github.com/google/uuid"

	"github.com/imightbuyaboat/TaskFlow/pkg/task"
	"github.com/imightbuyaboat/TaskFlow/task-api/internal/user"
)

type DB interface {
	CreateTask(t *task.Task) (*task.Task, error)
	GetTask(userID uint64, taskID uuid.UUID) (*task.Task, error)
	GetAllTasks(userID uint64) ([]task.Task, error)
	CreateUser(u *user.User) (uint64, error)
	CheckUser(u *user.User) (uint64, error)
}
