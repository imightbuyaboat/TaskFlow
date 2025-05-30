package handler

import "time"

type createTaskReq struct {
	Type       string                 `json:"type" binding:"required"`
	Payload    map[string]interface{} `json:"payload" binding:"required"`
	MaxRetries *uint8                 `json:"max_retries"`
	RunAt      *time.Time             `json:"run_at"`
}
