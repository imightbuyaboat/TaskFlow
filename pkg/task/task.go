package task

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID         uuid.UUID              `json:"id"`
	UserID     uint64                 `json:"user_id"`
	Type       string                 `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	Status     string                 `json:"status"`
	Retries    uint8                  `json:"retries"`
	MaxRetries uint8                  `json:"max_retries"`
	RunAt      *time.Time             `json:"run_at"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}
