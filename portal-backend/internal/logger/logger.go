package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ContextKey 컨텍스트 키의 타입 안전성을 위한 커스텀 타입
type ContextKey string

const (
	// UserIDKey 사용자 ID를 컨텍스트에 저장할 때 사용하는 키
	UserIDKey ContextKey = "user_id"

	// RequestIDKey Request ID를 컨텍스트에 저장할 때 사용하는 키
	RequestIDKey ContextKey = "request_id"
)

// LogLevel 로그 레벨 타입
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

// LogEntry 구조화된 로그 엔트리
type LogEntry struct {
	Timestamp  time.Time      `json:"timestamp"`
	Level      LogLevel       `json:"level"`
	Message    string         `json:"message"`
	RequestID  string         `json:"request_id,omitempty"`
	UserID     string         `json:"user_id,omitempty"`
	Method     string         `json:"method,omitempty"`
	Path       string         `json:"path,omitempty"`
	StatusCode int            `json:"status_code,omitempty"`
	Duration   string         `json:"duration,omitempty"`
	Error      string         `json:"error,omitempty"`
	File       string         `json:"file,omitempty"`
	Line       int            `json:"line,omitempty"`
	Extra      map[string]any `json:"extra,omitempty"`
}

// Logger 구조화된 로거
type Logger struct {
	minLevel LogLevel
	output   *log.Logger
}

// 전역 로거 인스턴스
var defaultLogger *Logger

// Init 로거 초기화
func Init() {
	level := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if level == "" {
		level = "INFO"
	}

	defaultLogger = &Logger{
		minLevel: LogLevel(level),
		output:   log.New(os.Stdout, "", 0),
	}
}

// GetLogger 로거 인스턴스 반환
func GetLogger() *Logger {
	if defaultLogger == nil {
		Init()
	}
	return defaultLogger
}

// shouldLog 로그 레벨이 출력되어야 하는지 확인
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
		FATAL: 4,
	}

	return levels[level] >= levels[l.minLevel]
}

// getCallerInfo 호출자 정보 가져오기
func getCallerInfo(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip + 2)
	if !ok {
		return "unknown", 0
	}

	// 파일 경로를 짧게 만들기
	parts := strings.Split(file, "/")
	if len(parts) >= 2 {
		file = strings.Join(parts[len(parts)-2:], "/")
	}

	return file, line
}

// log 실제 로그 출력
func (l *Logger) log(level LogLevel, message string, ctx context.Context, extra map[string]any) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Message:   message,
		Extra:     extra,
	}

	// 호출자 정보 추가
	file, line := getCallerInfo(1)
	entry.File = file
	entry.Line = line

	// 컨텍스트에서 정보 추출
	if ctx != nil {
		if requestID, ok := ctx.Value("request_id").(string); ok {
			entry.RequestID = requestID
		}
		if userID, ok := ctx.Value("user_id").(string); ok {
			entry.UserID = userID
		}
	}

	// JSON으로 마샬링
	jsonData, err := json.Marshal(entry)
	if err != nil {
		// JSON 마샬링 실패 시 기본 로그 출력
		l.output.Printf("[%s] %s - JSON Marshal Error: %v", level, message, err)
		return
	}

	l.output.Println(string(jsonData))
}

// Debug 디버그 로그
func (l *Logger) Debug(message string) {
	l.DebugWithContext(context.Background(), message, nil)
}

// DebugWithContext 컨텍스트가 포함된 디버그 로그
func (l *Logger) DebugWithContext(ctx context.Context, message string, extra map[string]any) {
	l.log(DEBUG, message, ctx, extra)
}

// Info 정보 로그
func (l *Logger) Info(message string) {
	l.InfoWithContext(context.Background(), message, nil)
}

// InfoWithContext 컨텍스트가 포함된 정보 로그
func (l *Logger) InfoWithContext(ctx context.Context, message string, extra map[string]any) {
	l.log(INFO, message, ctx, extra)
}

// Warn 경고 로그
func (l *Logger) Warn(message string) {
	l.WarnWithContext(context.Background(), message, nil)
}

// WarnWithContext 컨텍스트가 포함된 경고 로그
func (l *Logger) WarnWithContext(ctx context.Context, message string, extra map[string]any) {
	l.log(WARN, message, ctx, extra)
}

// Error 에러 로그
func (l *Logger) Error(message string, err error) {
	extra := make(map[string]interface{})
	if err != nil {
		extra["error_detail"] = err.Error()
	}
	l.ErrorWithContext(context.Background(), message, err, extra)
}

// ErrorWithContext 컨텍스트가 포함된 에러 로그
func (l *Logger) ErrorWithContext(ctx context.Context, message string, err error, extra map[string]any) {
	if extra == nil {
		extra = make(map[string]any)
	}
	if err != nil {
		extra["error_detail"] = err.Error()
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     ERROR,
		Message:   message,
		Extra:     extra,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// 호출자 정보 추가
	file, line := getCallerInfo(1)
	entry.File = file
	entry.Line = line

	// 컨텍스트에서 정보 추출
	if ctx != nil {
		if requestID, ok := ctx.Value("request_id").(string); ok {
			entry.RequestID = requestID
		}
		if userID, ok := ctx.Value("user_id").(string); ok {
			entry.UserID = userID
		}
	}

	if l.shouldLog(ERROR) {
		jsonData, jsonErr := json.Marshal(entry)
		if jsonErr != nil {
			l.output.Printf("[ERROR] %s - JSON Marshal Error: %v", message, jsonErr)
			return
		}
		l.output.Println(string(jsonData))
	}
}

// Fatal 치명적 에러 로그 (프로그램 종료)
func (l *Logger) Fatal(message string, err error) {
	l.Error(message, err)
	os.Exit(1)
}

// 전역 함수들 (편의성을 위해)
func Debug(message string) {
	GetLogger().Debug(message)
}

func DebugWithContext(ctx context.Context, message string, extra map[string]any) {
	GetLogger().DebugWithContext(ctx, message, extra)
}

func Info(message string) {
	GetLogger().Info(message)
}

func InfoWithContext(ctx context.Context, message string, extra map[string]any) {
	GetLogger().InfoWithContext(ctx, message, extra)
}

func Warn(message string) {
	GetLogger().Warn(message)
}

func WarnWithContext(ctx context.Context, message string, extra map[string]any) {
	GetLogger().WarnWithContext(ctx, message, extra)
}

func Error(message string, err error) {
	GetLogger().Error(message, err)
}

func ErrorWithContext(ctx context.Context, message string, err error, extra map[string]any) {
	GetLogger().ErrorWithContext(ctx, message, err, extra)
}

func Fatal(message string, err error) {
	GetLogger().Fatal(message, err)
}

// HTTPLogEntry HTTP 요청 로그 엔트리
type HTTPLogEntry struct {
	LogEntry
	RemoteAddr string `json:"remote_addr,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
	Referer    string `json:"referer,omitempty"`
}

// LogHTTPRequest HTTP 요청 로그
func LogHTTPRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration, c *gin.Context) {
	logger := GetLogger()
	if !logger.shouldLog(INFO) {
		return
	}

	entry := HTTPLogEntry{
		LogEntry: LogEntry{
			Timestamp:  time.Now().UTC(),
			Level:      INFO,
			Message:    fmt.Sprintf("%s %s", method, path),
			Method:     method,
			Path:       path,
			StatusCode: statusCode,
			Duration:   duration.String(),
		},
	}

	// 호출자 정보는 HTTP 로그에서는 제외 (의미 없음)

	// 컨텍스트에서 정보 추출
	if ctx != nil {
		if requestID, ok := ctx.Value("request_id").(string); ok {
			entry.RequestID = requestID
		}
		if userID, ok := ctx.Value("user_id").(string); ok {
			entry.UserID = userID
		}
	}

	// Gin 컨텍스트에서 추가 정보 추출
	if c != nil {
		entry.RemoteAddr = c.ClientIP()
		entry.UserAgent = c.GetHeader("User-Agent")
		entry.Referer = c.GetHeader("Referer")
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		logger.output.Printf("[INFO] HTTP %s %s - JSON Marshal Error: %v", method, path, err)
		return
	}

	logger.output.Println(string(jsonData))
}
