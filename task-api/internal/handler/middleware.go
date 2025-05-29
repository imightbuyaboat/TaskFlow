package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imightbuyaboat/TaskFlow/task-api/internal/auth"
	"go.uber.org/zap"
)

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			h.logger.Info("request with missing or invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := auth.ValidateToken(token)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidToken) {
				h.logger.Info("request with invalid token", zap.String("token", token))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				return
			}
			h.logger.Error("failed to validate token", zap.Error(err), zap.String("token", token))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate token"})
			return
		}

		h.logger.Info("succesfully validate token", zap.String("token", token), zap.Uint64("user_id", userID))
		c.Set(UserIDKey, userID)
		c.Next()
	}
}
