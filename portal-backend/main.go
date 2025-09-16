package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"portal-backend/internal/config"
	"portal-backend/internal/handlers"
	"portal-backend/internal/kubernetes"
	"portal-backend/internal/logger"
	"portal-backend/internal/middleware"
)

func main() {
	// 설정 로드
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", err)
	}

	logger.Init()
	logger.Info("Starting Portal Backend application")

	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		logger.Fatal("Failed to create Kubernetes client", err)
	}

	consoleHandler := handlers.NewConsoleHandler(k8sClient)

	gin.SetMode(cfg.Server.GinMode)

	r := gin.New()

	err = r.SetTrustedProxies([]string{"192.168.1.1", "127.0.0.1"})
	if err != nil {
		logger.Fatal("Failed to set trusted proxies", err)
	}

	r.Use(middleware.RecoveryLoggingMiddleware())
	r.Use(middleware.RequestLoggingMiddleware())
	r.Use(middleware.ErrorLoggingMiddleware())

	// CORS 설정
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, allowedOrigin := range cfg.Server.AllowedOrigins {
			if allowedOrigin == origin {
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
		console := api.Group("/console")
		{
			console.POST("/launch", consoleHandler.HandleLaunchConsole)
			console.POST("/list", consoleHandler.HandleListConsoles)
			console.POST("/delete/:resourceId", consoleHandler.HandleDeleteConsole)
		}

		api.POST("/launch-console", consoleHandler.HandleLaunchConsole)
	}

	// 헬스체크 엔드포인트
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		})
	})

	logger.InfoWithContext(context.TODO(), "Starting HTTP server", map[string]any{
		"port": cfg.Server.Port,
	})

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		logger.Fatal("Failed to start HTTP server", err)
	}
}
