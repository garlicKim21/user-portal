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
	logger.Init()
	logger.Info("Starting Portal Backend application")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.InfoWithContext(context.TODO(), "Server configuration loaded", map[string]any{
		"port": port,
	})

	logger.Info("Initializing OIDC provider")
	oidcProvider, err := auth.NewOIDCProvider()
	if err != nil {
		logger.Fatal("Failed to create OIDC provider", err)
	}
	logger.Info("OIDC provider initialized successfully")

	logger.Info("Initializing Kubernetes client")
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		logger.Fatal("Failed to create Kubernetes client", err)
	}
	logger.Info("Kubernetes client initialized successfully")

	logger.Info("Initializing handlers")
	authHandler, err := handlers.NewAuthHandler(oidcProvider)
	if err != nil {
		logger.Fatal("Failed to create auth handler", err)
	}
	consoleHandler := handlers.NewConsoleHandler(k8sClient, authHandler)
	logger.Info("Handlers initialized successfully")

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

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
		api.GET("/logout", authHandler.HandleLogout)

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
		"port":    port,
		"version": "1.0.0",
	})

	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start HTTP server", err)
	}
}
