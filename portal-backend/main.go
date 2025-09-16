package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"portal-backend/internal/auth"
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

	oidcProvider, err := auth.NewOIDCProvider()
	if err != nil {
		logger.Fatal("Failed to create OIDC provider", err)
	}

	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		logger.Fatal("Failed to create Kubernetes client", err)
	}

	authHandler, err := handlers.NewAuthHandler(oidcProvider, k8sClient)
	if err != nil {
		logger.Fatal("Failed to create auth handler", err)
	}
	consoleHandler := handlers.NewConsoleHandler(k8sClient, authHandler)

	gin.SetMode(cfg.Server.GinMode)

	r := gin.New()

	err = r.SetTrustedProxies([]string{"192.168.1.1", "127.0.0.1"})
	if err != nil {
		logger.Fatal("Failed to set trusted proxies", err)
	}

	r.Use(middleware.RecoveryLoggingMiddleware())
	r.Use(middleware.RequestLoggingMiddleware())
	r.Use(middleware.SetUserIDMiddleware())
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
		// 기존 인증 라우트
		api.GET("/login", authHandler.HandleLogin)
		api.GET("/callback", authHandler.HandleCallback)
		api.GET("/user", authHandler.HandleGetUser)
		api.POST("/logout", authHandler.HandleLogout)

		// 새로운 인증 라우트 (poc_front용)
		auth := api.Group("/auth")
		{
			auth.POST("/session", authHandler.HandleCreateSessionFromOIDC)
		}

		console := api.Group("/console")
		{
			console.GET("/launch", consoleHandler.HandleLaunchConsole)
			console.GET("/list", consoleHandler.HandleListConsoles)
			console.DELETE("/:resourceId", consoleHandler.HandleDeleteConsole)
		}

		api.GET("/launch-console", consoleHandler.HandleLaunchConsole)
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
