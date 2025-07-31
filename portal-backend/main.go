package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"portal-backend/internal/auth"
	"portal-backend/internal/handlers"
	"portal-backend/internal/kubernetes"
)

func main() {
	// 환경 변수에서 포트 가져오기
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// OIDC Provider 초기화
	oidcProvider, err := auth.NewOIDCProvider()
	if err != nil {
		log.Fatal("Failed to create OIDC provider:", err)
	}

	// 쿠버네티스 클라이언트 초기화
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		log.Fatal("Failed to create Kubernetes client:", err)
	}

	// 핸들러 초기화
	authHandler := handlers.NewAuthHandler(oidcProvider)
	consoleHandler := handlers.NewConsoleHandler(k8sClient, authHandler)

	// Gin 라우터 설정
	r := gin.Default()

	// CORS 설정
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

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
		api.GET("/launch-console", consoleHandler.HandleLaunchConsole)
		api.GET("/user", authHandler.HandleGetUser)
	}

	// 정적 파일 서빙 (프론트엔드용)
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// 루트 경로
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Portal Demo - Web Console PoC",
		})
	})

	log.Printf("Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
