package models

import (
	"net/http"
	"time"
)

// ErrorType 에러 타입 정의
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "VALIDATION_ERROR"
	ErrorTypeAuthentication ErrorType = "AUTHENTICATION_ERROR"
	ErrorTypeAuthorization  ErrorType = "AUTHORIZATION_ERROR"
	ErrorTypeNotFound       ErrorType = "NOT_FOUND_ERROR"
	ErrorTypeInternal       ErrorType = "INTERNAL_ERROR"
	ErrorTypeConflict       ErrorType = "CONFLICT_ERROR"
	ErrorTypeTimeout        ErrorType = "TIMEOUT_ERROR"
	ErrorTypeRateLimit      ErrorType = "RATE_LIMIT_ERROR"
)

// ErrorResponse 표준화된 에러 응답 구조체
type ErrorResponse struct {
	Error     ErrorInfo `json:"error"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	RequestID string    `json:"request_id,omitempty"`
}

// ErrorInfo 에러 상세 정보
type ErrorInfo struct {
	Type    ErrorType `json:"type"`
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// APIError 내부 에러 구조체
type APIError struct {
	Type       ErrorType
	Code       string
	Message    string
	Details    string
	HTTPStatus int
	Cause      error
}

// Error APIError의 에러 인터페이스 구현
func (e *APIError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap 원본 에러 반환
func (e *APIError) Unwrap() error {
	return e.Cause
}

// 사전 정의된 에러들
var (
	// 인증 관련 에러
	ErrInvalidCredentials = &APIError{
		Type:       ErrorTypeAuthentication,
		Code:       "AUTH001",
		Message:    "Invalid credentials provided",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrTokenExpired = &APIError{
		Type:       ErrorTypeAuthentication,
		Code:       "AUTH002",
		Message:    "Authentication token has expired",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrTokenInvalid = &APIError{
		Type:       ErrorTypeAuthentication,
		Code:       "AUTH003",
		Message:    "Invalid authentication token",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrSessionNotFound = &APIError{
		Type:       ErrorTypeAuthentication,
		Code:       "AUTH004",
		Message:    "Session not found or expired",
		HTTPStatus: http.StatusUnauthorized,
	}

	// 권한 관련 에러
	ErrInsufficientPermissions = &APIError{
		Type:       ErrorTypeAuthorization,
		Code:       "AUTHZ001",
		Message:    "Insufficient permissions to access this resource",
		HTTPStatus: http.StatusForbidden,
	}

	ErrResourceAccessDenied = &APIError{
		Type:       ErrorTypeAuthorization,
		Code:       "AUTHZ002",
		Message:    "Access denied to the requested resource",
		HTTPStatus: http.StatusForbidden,
	}

	// 리소스 관련 에러
	ErrResourceNotFound = &APIError{
		Type:       ErrorTypeNotFound,
		Code:       "RES001",
		Message:    "Requested resource not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrConsoleNotFound = &APIError{
		Type:       ErrorTypeNotFound,
		Code:       "RES002",
		Message:    "Web console not found",
		HTTPStatus: http.StatusNotFound,
	}

	// 쿠버네티스 관련 에러
	ErrKubernetesOperation = &APIError{
		Type:       ErrorTypeInternal,
		Code:       "K8S001",
		Message:    "Kubernetes operation failed",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrPodCreationFailed = &APIError{
		Type:       ErrorTypeInternal,
		Code:       "K8S002",
		Message:    "Failed to create web console pod",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrPodNotReady = &APIError{
		Type:       ErrorTypeTimeout,
		Code:       "K8S003",
		Message:    "Web console pod is not ready within timeout period",
		HTTPStatus: http.StatusRequestTimeout,
	}

	// 검증 관련 에러
	ErrInvalidInput = &APIError{
		Type:       ErrorTypeValidation,
		Code:       "VAL001",
		Message:    "Invalid input parameters",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrMissingParameter = &APIError{
		Type:       ErrorTypeValidation,
		Code:       "VAL002",
		Message:    "Required parameter is missing",
		HTTPStatus: http.StatusBadRequest,
	}

	// 일반적인 내부 에러
	ErrInternalServer = &APIError{
		Type:       ErrorTypeInternal,
		Code:       "SRV001",
		Message:    "Internal server error occurred",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrServiceUnavailable = &APIError{
		Type:       ErrorTypeInternal,
		Code:       "SRV002",
		Message:    "Service temporarily unavailable",
		HTTPStatus: http.StatusServiceUnavailable,
	}
)

// NewAPIError 새로운 API 에러 생성
func NewAPIError(errorType ErrorType, code, message string, httpStatus int) *APIError {
	return &APIError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// WithDetails 에러에 상세 정보 추가
func (e *APIError) WithDetails(details string) *APIError {
	newErr := *e
	newErr.Details = details
	return &newErr
}

// WithCause 원인 에러 추가
func (e *APIError) WithCause(cause error) *APIError {
	newErr := *e
	newErr.Cause = cause
	return &newErr
}
