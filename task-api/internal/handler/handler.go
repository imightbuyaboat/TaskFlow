package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imightbuyaboat/TaskFlow/pkg/task"
	"github.com/imightbuyaboat/TaskFlow/task-api/internal/auth"
	"github.com/imightbuyaboat/TaskFlow/task-api/internal/db"
	"github.com/imightbuyaboat/TaskFlow/task-api/internal/user"
	"go.uber.org/zap"
)

const UserIDKey = "userID"

type Handler struct {
	db           DB
	queue        Queue
	tokenManager auth.TokenManager
	logger       *zap.Logger
}

func NewHandler(db DB, queue Queue, tm auth.TokenManager, logger *zap.Logger) (*Handler, error) {
	return &Handler{
		db:           db,
		queue:        queue,
		tokenManager: tm,
		logger:       logger,
	}, nil
}

func (h *Handler) RegisterHandler(c *gin.Context) {
	var u user.User
	if err := c.ShouldBindJSON(&u); err != nil {
		h.logger.Info("invalid email or password format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password format"})
		return
	}

	userID, err := h.db.CreateUser(&u)
	if err != nil {
		if errors.Is(err, db.ErrUserAlreadyExist) {
			h.logger.Info("user already exists", zap.String("email", u.Email))
			c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
			return
		}
		h.logger.Error("failed to create user", zap.Error(err), zap.String("email", u.Email))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.logger.Info("successfully create user", zap.String("email", u.Email), zap.Uint64("user_id", userID))
	c.JSON(http.StatusCreated, gin.H{"user_id": userID})
}

func (h *Handler) LoginHandler(c *gin.Context) {
	var u user.User
	if err := c.ShouldBindJSON(&u); err != nil {
		h.logger.Info("invalid email or password format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password format"})
		return
	}

	userID, err := h.db.CheckUser(&u)
	if err != nil {
		if errors.Is(err, db.ErrNoRows) || errors.Is(err, db.ErrIncorrectPassword) {
			h.logger.Info("invalid credentials", zap.Error(err), zap.String("email", u.Email))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		h.logger.Error("failed to check user", zap.Error(err), zap.String("email", u.Email))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	token, err := h.tokenManager.CreateToken(userID)
	if err != nil {
		h.logger.Error("failed to create token", zap.Error(err), zap.String("email", u.Email), zap.Uint64("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	h.logger.Info("successfully authorizate user", zap.String("email", u.Email), zap.Uint64("user_id", userID))
	c.JSON(http.StatusOK, gin.H{"token": "Bearer " + token})
}

func (h *Handler) CreateTaskHandler(c *gin.Context) {
	userID := c.GetUint64(UserIDKey)

	var req createTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Info("invalid body of request", zap.Error(err), zap.Uint64("user_id", userID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body of request"})
		return
	}

	if !task.ValidateType(req.Type) {
		h.logger.Info("invalid body of request", zap.Uint64("user_id", userID), zap.String("type", req.Type))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type of task"})
		return
	}

	err := task.ValidatePayload(req.Type, req.Payload)
	if err != nil {
		h.logger.Info("invalid body of request", zap.Uint64("user_id", userID), zap.Any("payload", req.Payload))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload of task"})
		return
	}

	if req.MaxRetries == nil {
		defaultMaxRetries := uint8(3)
		req.MaxRetries = &defaultMaxRetries
	} else if *req.MaxRetries < 1 || *req.MaxRetries > 10 {
		h.logger.Info("invalid body of request", zap.Uint64("user_id", userID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "max_retries should be between 1 and 10"})
		return
	}

	if req.RunAt != nil && req.RunAt.Before(time.Now()) {
		h.logger.Info("invalid body of request", zap.Uint64("user_id", userID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "run_at must be in the future"})
		return
	}

	taskID, err := uuid.NewUUID()
	if err != nil {
		h.logger.Error("failed to generate task_id", zap.Error(err), zap.Uint64("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate taks_id"})
		return
	}

	t := task.Task{
		ID:         taskID,
		UserID:     userID,
		Type:       req.Type,
		Payload:    req.Payload,
		MaxRetries: *req.MaxRetries,
		RunAt:      req.RunAt,
	}
	createdTask, err := h.db.CreateTask(&t)
	if err != nil {
		h.logger.Error("failed to create task", zap.Error(err), zap.Uint64("user_id", userID), zap.String("task_id", taskID.String()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	if req.RunAt == nil {
		err = h.queue.Publish(createdTask)
		if err != nil {
			h.logger.Error("failed to publish task in queue", zap.Error(err), zap.Uint64("user_id", userID), zap.String("task_id", taskID.String()))
		} else {
			h.logger.Info("successfully publish task", zap.Uint64("user_id", userID), zap.String("task_id", taskID.String()))
		}
	}

	h.logger.Info("successfully created task", zap.Uint64("user_id", userID), zap.String("task_id", taskID.String()))
	c.JSON(http.StatusCreated, createdTask)
}

func (h *Handler) GetTaskHandler(c *gin.Context) {
	userID := c.GetUint64(UserIDKey)

	rawTaskID := c.Param("id")
	taskID, err := uuid.Parse(rawTaskID)
	if err != nil {
		h.logger.Error("failed to parse task_id", zap.Error(err), zap.Uint64("user_id", userID), zap.String("task_id", rawTaskID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse task ID"})
		return
	}

	t, err := h.db.GetTask(userID, taskID)
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			h.logger.Info("incorrect task_id", zap.Uint64("user_id", userID), zap.String("task_id", rawTaskID))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect task ID"})
			return
		}
		h.logger.Error("failed to get task", zap.Error(err), zap.Uint64("user_id", userID), zap.String("task_id", rawTaskID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task"})
		return
	}

	h.logger.Info("successfully get task", zap.Uint64("user_id", userID), zap.String("task_id", rawTaskID))
	c.JSON(http.StatusOK, t)
}

func (h *Handler) GetAllTasksHandler(c *gin.Context) {
	userID := c.GetUint64(UserIDKey)

	tasks, err := h.db.GetAllTasks(userID)
	if err != nil {
		h.logger.Error("failed to get tasks", zap.Error(err), zap.Uint64("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	h.logger.Info("successfully get tasks", zap.Uint64("user_id", userID))
	c.JSON(http.StatusOK, tasks)
}
