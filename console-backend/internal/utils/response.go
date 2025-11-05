package utils

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"portal-backend/internal/models"
)

// ResponseHelper HTTP 응답 헬퍼
type ResponseHelper struct {
	Logger *log.Logger
}

// NewResponseHelper 새로운 응답 헬퍼 생성
func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{
		Logger: log.Default(),
	}
}

// Success 성공 응답
func (r *ResponseHelper) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"data":      data,
		"timestamp": time.Now().UTC(),
	})
}

// SuccessWithMessage 메시지가 포함된 성공 응답
func (r *ResponseHelper) SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   message,
		"data":      data,
		"timestamp": time.Now().UTC(),
	})
}

// Error 에러 응답
func (r *ResponseHelper) Error(c *gin.Context, err error) {
	var apiErr *models.APIError
	var httpStatus int
	var errorResponse models.ErrorResponse

	// APIError 타입 확인
	if e, ok := err.(*models.APIError); ok {
		apiErr = e
		httpStatus = e.HTTPStatus
	} else {
		// 일반 에러인 경우 내부 서버 에러로 처리
		apiErr = models.ErrInternalServer.WithCause(err)
		httpStatus = http.StatusInternalServerError
	}

	requestID := uuid.New().String()

	errorResponse = models.ErrorResponse{
		Error: models.ErrorInfo{
			Type:    apiErr.Type,
			Code:    apiErr.Code,
			Message: apiErr.Message,
			Details: apiErr.Details,
		},
		Timestamp: time.Now().UTC(),
		Path:      c.Request.URL.Path,
		RequestID: requestID,
	}

	// 내부 에러의 경우에만 로깅 (중요한 에러만)
	if apiErr.Type == models.ErrorTypeInternal {
		r.Logger.Printf("[ERROR] %s - %s | RequestID: %s | Path: %s | Cause: %v",
			apiErr.Code, apiErr.Message, requestID, c.Request.URL.Path, apiErr.Cause)
	}

	c.Header("X-Request-ID", requestID)
	c.JSON(httpStatus, errorResponse)
}

// ValidationError 검증 에러 응답
func (r *ResponseHelper) ValidationError(c *gin.Context, field, message string) {
	err := models.ErrInvalidInput.WithDetails("Field: " + field + ", " + message)
	r.Error(c, err)
}

// NotFound 404 에러 응답
func (r *ResponseHelper) NotFound(c *gin.Context, resource string) {
	err := models.ErrResourceNotFound.WithDetails("Resource: " + resource)
	r.Error(c, err)
}

// Unauthorized 401 에러 응답
func (r *ResponseHelper) Unauthorized(c *gin.Context, message string) {
	var err *models.APIError

	// 메시지에 "token has expired"가 포함되어 있으면 ErrTokenExpired 사용
	if strings.Contains(strings.ToLower(message), "token has expired") {
		err = models.ErrTokenExpired.WithDetails(message)
	} else {
		err = models.ErrInvalidCredentials.WithDetails(message)
	}

	r.Error(c, err)
}

// Forbidden 403 에러 응답
func (r *ResponseHelper) Forbidden(c *gin.Context, message string) {
	err := models.ErrInsufficientPermissions.WithDetails(message)
	r.Error(c, err)
}

// InternalError 500 에러 응답
func (r *ResponseHelper) InternalError(c *gin.Context, cause error) {
	err := models.ErrInternalServer.WithCause(cause)
	r.Error(c, err)
}

// KubernetesError 쿠버네티스 관련 에러 응답
func (r *ResponseHelper) KubernetesError(c *gin.Context, operation string, cause error) {
	err := models.ErrKubernetesOperation.WithDetails("Operation: " + operation).WithCause(cause)
	r.Error(c, err)
}

// 전역 응답 헬퍼 인스턴스
var Response = NewResponseHelper()
