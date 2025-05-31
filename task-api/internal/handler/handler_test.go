package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imightbuyaboat/TaskFlow/pkg/task"
	pdb "github.com/imightbuyaboat/TaskFlow/task-api/internal/db"
	"github.com/imightbuyaboat/TaskFlow/task-api/internal/handler/mocks"
	"github.com/imightbuyaboat/TaskFlow/task-api/internal/user"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	logger := zaptest.NewLogger(t)
	h, _ := NewHandler(mockDB, nil, nil, logger)

	tests := []struct {
		name           string
		body           interface{}
		mockBDSetup    func(db *mocks.MockDB)
		expectedStatus int
		expectedBody   gin.H
	}{
		{
			name: "Successfully user creation",
			body: user.User{
				Email:    "test@test.com",
				Password: "test",
			},
			mockBDSetup: func(db *mocks.MockDB) {
				db.EXPECT().CreateUser(gomock.Any()).Return(uint64(1), nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   gin.H{"user_id": uint64(1)},
		},
		{
			name: "invalid email or password format",
			body: user.User{
				Email:    "invalid email",
				Password: "test",
			},
			mockBDSetup:    func(db *mocks.MockDB) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "Invalid email or password format"},
		},
		{
			name: "User already exists",
			body: user.User{
				Email:    "test@test.com",
				Password: "test",
			},
			mockBDSetup: func(db *mocks.MockDB) {
				db.EXPECT().CreateUser(gomock.Any()).Return(uint64(0), pdb.ErrUserAlreadyExist)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   gin.H{"error": "User already exists"},
		},
		{
			name: "Internal server error",
			body: user.User{
				Email:    "test@test.com",
				Password: "test",
			},
			mockBDSetup: func(db *mocks.MockDB) {
				db.EXPECT().CreateUser(gomock.Any()).Return(uint64(0), errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   gin.H{"error": "Internal server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBDSetup(mockDB)

			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			h.RegisterHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var responseBody map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &responseBody); err != nil {
				t.Fatal(err)
			}

			for key, expectedVal := range tt.expectedBody {
				actualVal, ok := responseBody[key]
				if !ok {
					t.Errorf("expected key %q not found in response", key)
					continue
				}

				switch v := expectedVal.(type) {
				case uint64:
					assert.Equal(t, float64(v), actualVal)
				default:
					assert.Equal(t, expectedVal, actualVal)
				}
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockTokenManager := mocks.NewMockTokenManager(ctrl)
	logger := zaptest.NewLogger(t)
	h, _ := NewHandler(mockDB, nil, mockTokenManager, logger)

	tests := []struct {
		name                  string
		body                  interface{}
		mockBDSetup           func(db *mocks.MockDB)
		mockTokenManagerSetup func(tm *mocks.MockTokenManager)
		expectedStatus        int
		expectedBody          gin.H
	}{
		{
			name: "Successfully user authorization",
			body: user.User{
				Email:    "test@test.com",
				Password: "test",
			},
			mockBDSetup: func(db *mocks.MockDB) {
				db.EXPECT().CheckUser(gomock.Any()).Return(uint64(1), nil)
			},
			mockTokenManagerSetup: func(tm *mocks.MockTokenManager) {
				tm.EXPECT().CreateToken(gomock.Any()).Return("test-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   gin.H{"token": "Bearer test-token"},
		},
		{
			name: "invalid email or password format",
			body: user.User{
				Email:    "invalid email",
				Password: "test",
			},
			mockBDSetup:           func(db *mocks.MockDB) {},
			mockTokenManagerSetup: func(tm *mocks.MockTokenManager) {},
			expectedStatus:        http.StatusBadRequest,
			expectedBody:          gin.H{"error": "Invalid email or password format"},
		},
		{
			name: "invalid credentials",
			body: user.User{
				Email:    "test@test.com",
				Password: "invalid password",
			},
			mockBDSetup: func(db *mocks.MockDB) {
				db.EXPECT().CheckUser(gomock.Any()).Return(uint64(0), pdb.ErrIncorrectPassword)
			},
			mockTokenManagerSetup: func(tm *mocks.MockTokenManager) {},
			expectedStatus:        http.StatusUnauthorized,
			expectedBody:          gin.H{"error": "Invalid credentials"},
		},
		{
			name: "db error",
			body: user.User{
				Email:    "test@test.com",
				Password: "test",
			},
			mockBDSetup: func(db *mocks.MockDB) {
				db.EXPECT().CheckUser(gomock.Any()).Return(uint64(0), errors.New("db error"))
			},
			mockTokenManagerSetup: func(tm *mocks.MockTokenManager) {},
			expectedStatus:        http.StatusInternalServerError,
			expectedBody:          gin.H{"error": "Internal server error"},
		},
		{
			name: "tokenManager error",
			body: user.User{
				Email:    "test@test.com",
				Password: "test",
			},
			mockBDSetup: func(db *mocks.MockDB) {
				db.EXPECT().CheckUser(gomock.Any()).Return(uint64(1), nil)
			},
			mockTokenManagerSetup: func(tm *mocks.MockTokenManager) {
				tm.EXPECT().CreateToken(gomock.Any()).Return("", errors.New("tokenManager error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   gin.H{"error": "Failed to create token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBDSetup(mockDB)
			tt.mockTokenManagerSetup(mockTokenManager)

			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			h.LoginHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var responseBody gin.H
			if err := json.Unmarshal(w.Body.Bytes(), &responseBody); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.expectedBody, responseBody)
		})
	}
}

func TestCreateTaskHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockQueue := mocks.NewMockQueue(ctrl)
	logger := zaptest.NewLogger(t)
	h, _ := NewHandler(mockDB, mockQueue, nil, logger)

	runAt := time.Now().Add(time.Hour)
	createdAt := time.Now()
	updatedAt := createdAt

	createdTask := task.Task{
		ID:     uuid.New(),
		UserID: 1,
		Type:   "send_email",
		Payload: map[string]interface{}{
			"to":      "test@test.com",
			"subject": "test",
		},
		Status:     "postponed",
		Retries:    0,
		MaxRetries: 3,
		RunAt:      &runAt,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	tests := []struct {
		name           string
		body           interface{}
		mockBDSetup    func(db *mocks.MockDB)
		mockQueueSetup func(q *mocks.MockQueue)
		expectedStatus int
		expectedBody   gin.H
	}{
		{
			name: "Successfully task creation",
			body: struct {
				Type       string                 `json:"type" binding:"required"`
				Payload    map[string]interface{} `json:"payload" binding:"required"`
				MaxRetries *uint8                 `json:"max_retries"`
				RunAt      *time.Time             `json:"run_at"`
			}{
				Type: "send_email",
				Payload: map[string]interface{}{
					"to":      "test@test.com",
					"subject": "test",
				},
				RunAt: &runAt,
			},
			mockBDSetup: func(db *mocks.MockDB) {
				db.EXPECT().CreateTask(gomock.Any()).Return(&createdTask, nil)
			},
			mockQueueSetup: func(q *mocks.MockQueue) {},
			expectedStatus: http.StatusCreated,
			expectedBody: gin.H{
				"id":          createdTask.ID.String(),
				"user_id":     float64(createdTask.UserID),
				"type":        createdTask.Type,
				"payload":     createdTask.Payload,
				"status":      createdTask.Status,
				"retries":     float64(createdTask.Retries),
				"max_retries": float64(createdTask.MaxRetries),
				"run_at":      createdTask.RunAt.Format(time.RFC3339Nano),
				"created_at":  createdTask.CreatedAt.Format(time.RFC3339Nano),
				"updated_at":  createdTask.UpdatedAt.Format(time.RFC3339Nano),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBDSetup(mockDB)
			tt.mockQueueSetup(mockQueue)

			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			h.CreateTaskHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var responseBody gin.H
			if err := json.Unmarshal(w.Body.Bytes(), &responseBody); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.expectedBody, responseBody)
		})
	}
}
