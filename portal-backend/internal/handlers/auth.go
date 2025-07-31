package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"portal-backend/internal/auth"
	"portal-backend/internal/models"
)

// AuthHandler 인증 핸들러
type AuthHandler struct {
	oidcProvider *auth.OIDCProvider
	sessions     map[string]*models.Session
}

// NewAuthHandler 새로운 인증 핸들러 생성
func NewAuthHandler(oidcProvider *auth.OIDCProvider) *AuthHandler {
	return &AuthHandler{
		oidcProvider: oidcProvider,
		sessions:     make(map[string]*models.Session),
	}
}

// HandleLogin 로그인 처리
func (h *AuthHandler) HandleLogin(c *gin.Context) {
	// CSRF 보호를 위한 state 생성
	state := auth.GenerateRandomString(32)

	// 세션에 state 저장 (실제로는 Redis나 DB 사용 권장)
	h.sessions[state] = &models.Session{}

	// OAuth2 인증 URL 생성
	authURL := h.oidcProvider.GetAuthURL(state)

	log.Printf("Redirecting to OAuth2 provider: %s", authURL)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// HandleCallback OAuth2 콜백 처리
func (h *AuthHandler) HandleCallback(c *gin.Context) {
	ctx := context.Background()

	// 에러 체크
	if err := c.Query("error"); err != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	// Authorization code 가져오기
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not found"})
		return
	}

	// State 검증 (실제로는 세션에서 확인)
	if _, exists := h.sessions[state]; !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	// 토큰 교환
	token, err := h.oidcProvider.ExchangeCode(ctx, code)
	if err != nil {
		log.Printf("Token exchange failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token exchange failed"})
		return
	}

	// ID Token 검증
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID token not found"})
		return
	}

	idToken, err := h.oidcProvider.VerifyIDToken(ctx, rawIDToken)
	if err != nil {
		log.Printf("ID token verification failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID token verification failed"})
		return
	}

	// 사용자 정보 추출
	var claims auth.TokenClaims
	if err := idToken.Claims(&claims); err != nil {
		log.Printf("Failed to parse ID token claims: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token claims"})
		return
	}

	// 세션 저장
	session := &models.Session{
		AccessToken:  token.AccessToken,
		IDToken:      rawIDToken,
		RefreshToken: token.RefreshToken,
		UserID:       claims.Subject,
		ExpiresAt:    token.Expiry,
	}

	// 간단한 세션 ID 생성 (실제로는 JWT나 세션 쿠키 사용)
	sessionID := auth.GenerateRandomString(32)
	h.sessions[sessionID] = session

	log.Printf("User %s authenticated successfully", claims.Subject)

	// 프론트엔드로 리디렉션 (세션 ID 포함)
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/?session=%s", sessionID))
}

// HandleGetUser 사용자 정보 조회
func (h *AuthHandler) HandleGetUser(c *gin.Context) {
	sessionID := c.Query("session")
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No session provided"})
		return
	}

	session, exists := h.sessions[sessionID]
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	c.JSON(http.StatusOK, models.UserInfo{
		UserID:   session.UserID,
		LoggedIn: true,
	})
}

// GetSession 세션 조회
func (h *AuthHandler) GetSession(sessionID string) (*models.Session, bool) {
	session, exists := h.sessions[sessionID]
	return session, exists
} 