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
		// 요청 시작 시간
		startTime := time.Now()

		// Request ID 생성
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// 컨텍스트에 Request ID 추가
		ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// 요청 처리
		c.Next()

		// 응답 시간 계산
		duration := time.Since(startTime)

		// 사용자 ID 추출 (JWT에서)
		userID, exists := c.Get("user_id")
		if exists {
			ctx = context.WithValue(ctx, logger.UserIDKey, userID)
		}

		// HTTP 요청 로그 기록
		logger.LogHTTPRequest(
			ctx,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			c,
		)

		// 에러가 있는 경우 추가 로깅
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
		// JWT에서 사용자 ID 추출하는 로직은 실제 JWT 미들웨어에서 수행
		// 여기서는 설정된 user_id를 컨텍스트에 추가하는 역할만 수행
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

		// 처리되지 않은 패닉 복구
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
