package handler

import (
	"github.com/imightbuyaboat/TaskFlow/pkg/task"
)

type Queue interface {
	Publish(t *task.Task) error
}
