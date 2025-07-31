package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"portal-backend/internal/auth"
	"portal-backend/internal/handlers"
	"portal-backend/internal/kubernetes"
	"portal-backend/internal/logger"
	"portal-backend/internal/middleware"
)

func main() {
	// 로거 초기화
	logger.Init()
	logger.Info("Starting Portal Backend application")

	// 환경 변수에서 포트 가져오기
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.InfoWithContext(context.TODO(), "Server configuration loaded", map[string]interface{}{
		"port": port,
	})

	// OIDC Provider 초기화
	logger.Info("Initializing OIDC provider")
	oidcProvider, err := auth.NewOIDCProvider()
	if err != nil {
		logger.Fatal("Failed to create OIDC provider", err)
	}
	logger.Info("OIDC provider initialized successfully")

	// 쿠버네티스 클라이언트 초기화
	logger.Info("Initializing Kubernetes client")
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		logger.Fatal("Failed to create Kubernetes client", err)
	}
	logger.Info("Kubernetes client initialized successfully")

	// 핸들러 초기화
	logger.Info("Initializing handlers")
	authHandler, err := handlers.NewAuthHandler(oidcProvider)
	if err != nil {
		logger.Fatal("Failed to create auth handler", err)
	}
	consoleHandler := handlers.NewConsoleHandler(k8sClient, authHandler)
	logger.Info("Handlers initialized successfully")

	// Gin 모드 설정
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode) // 프로덕션 모드 기본값
	}

	// Gin 라우터 설정
	r := gin.New()

	// 미들웨어 설정
	r.Use(middleware.RecoveryLoggingMiddleware())
	r.Use(middleware.RequestLoggingMiddleware())
	r.Use(middleware.SetUserIDMiddleware())
	r.Use(middleware.ErrorLoggingMiddleware())

	// CORS 설정 - 보안 강화
	r.Use(func(c *gin.Context) {
		// 환경변수에서 허용된 오리진 가져오기, 기본값은 localhost
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "http://localhost:5173,http://localhost:3000,http://localhost:8080"
		}

		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, allowedOrigin := range strings.Split(allowedOrigins, ",") {
			if strings.TrimSpace(allowedOrigin) == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API 라우트 설정
	api := r.Group("/api")
	{
		api.GET("/login", authHandler.HandleLogin)
		api.GET("/callback", authHandler.HandleCallback)
		api.GET("/user", authHandler.HandleGetUser)

		// 웹 콘솔 관련 라우트
		console := api.Group("/console")
		{
			console.GET("/launch", consoleHandler.HandleLaunchConsole)
			console.GET("/list", consoleHandler.HandleListConsoles)
			console.DELETE("/:resourceId", consoleHandler.HandleDeleteConsole)
		}

		// 호환성을 위한 기존 엔드포인트 유지
		api.GET("/launch-console", consoleHandler.HandleLaunchConsole)
	}

	// 헬스체크 엔드포인트 (인증 불필요)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		})
	})

	// API 전용 서버 (프론트엔드 분리)
	// 정적 파일 서빙 제거 - 프론트엔드는 별도 서비스로 분리

	logger.InfoWithContext(context.TODO(), "Starting HTTP server", map[string]interface{}{
		"port":    port,
		"version": "1.0.0",
	})

	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start HTTP server", err)
	}
}
