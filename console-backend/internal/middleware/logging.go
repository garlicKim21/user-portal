package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"portal-backend/internal/logger"
)

// RequestLoggingMiddleware HTTP 요청 로깅 미들웨어
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		duration := time.Since(startTime)

		userID, exists := c.Get("user_id")
		if exists {
			ctx = context.WithValue(ctx, logger.UserIDKey, userID)
		}

		logger.LogHTTPRequest(
			ctx,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			c,
		)

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.ErrorWithContext(ctx, "Request processing error", err.Err, map[string]any{
					"error_type": err.Type,
					"meta":       err.Meta,
				})
			}
		}

		// 느린 요청 경고 (2초 이상)
		if duration > 2*time.Second {
			logger.WarnWithContext(ctx, "Slow request detected", map[string]any{
				"duration":    duration.String(),
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"status_code": c.Writer.Status(),
			})
		}
	}
}

// SetUserIDMiddleware 사용자 ID를 컨텍스트에 설정하는 미들웨어
func SetUserIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID, exists := c.Get("user_id"); exists {
			ctx := context.WithValue(c.Request.Context(), logger.UserIDKey, userID)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}

// ErrorLoggingMiddleware 에러 로깅 미들웨어
func ErrorLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			ctx := c.Request.Context()

			for _, err := range c.Errors {
				logger.ErrorWithContext(ctx, "Unhandled error occurred", err.Err, map[string]any{
					"error_type": err.Type,
					"meta":       err.Meta,
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
				})
			}
		}
	}
}

// RecoveryLoggingMiddleware 패닉 복구 및 로깅 미들웨어
func RecoveryLoggingMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		ctx := c.Request.Context()

		logger.ErrorWithContext(ctx, "Panic recovered", nil, map[string]any{
			"panic":   recovered,
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"headers": c.Request.Header,
		})

		c.AbortWithStatus(500)
	})
}
