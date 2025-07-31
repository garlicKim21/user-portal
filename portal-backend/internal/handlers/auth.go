package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"portal-backend/internal/auth"
	"portal-backend/internal/logger"
	"portal-backend/internal/models"
	"portal-backend/internal/utils"
)

// AuthHandler 인증 핸들러
type AuthHandler struct {
	oidcProvider *auth.OIDCProvider
	jwtManager   *auth.JWTManager
	sessions     map[string]*models.Session // 임시 state 저장용 (JWT 도입 후 state만 사용)
}

// NewAuthHandler 새로운 인증 핸들러 생성
func NewAuthHandler(oidcProvider *auth.OIDCProvider) (*AuthHandler, error) {
	jwtManager, err := auth.NewJWTManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT manager: %v", err)
	}

	return &AuthHandler{
		oidcProvider: oidcProvider,
		jwtManager:   jwtManager,
		sessions:     make(map[string]*models.Session),
	}, nil
}

// HandleLogin 로그인 처리
func (h *AuthHandler) HandleLogin(c *gin.Context) {
	// CSRF 보호를 위한 state 생성
	state, err := auth.GenerateRandomString(32)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to generate state", err, nil)
		utils.Response.InternalError(c, fmt.Errorf("failed to generate security token: %w", err))
		return
	}

	// 세션에 state 저장 (실제로는 Redis나 DB 사용 권장)
	h.sessions[state] = &models.Session{
		State:     state,
		CreatedAt: time.Now(),
	}

	// OAuth2 인증 URL 생성
	authURL := h.oidcProvider.GetAuthURL(state)

	logger.InfoWithContext(c.Request.Context(), "Redirecting to OAuth2 provider", map[string]interface{}{
		"auth_url": authURL,
		"state":    state,
	})
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// HandleCallback OAuth2 콜백 처리
func (h *AuthHandler) HandleCallback(c *gin.Context) {
	ctx := context.Background()

	// 에러 체크
	if errParam := c.Query("error"); errParam != "" {
		utils.Response.Error(c, models.ErrInvalidCredentials.WithDetails("OAuth error: "+errParam))
		return
	}

	// Authorization code 가져오기
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		utils.Response.ValidationError(c, "code", "Authorization code is required")
		return
	}

	// State 검증 및 보안 강화
	stateSession, exists := h.sessions[state]
	if !exists {
		utils.Response.Error(c, models.ErrInvalidCredentials.WithDetails("Invalid or expired state parameter"))
		return
	}

	// State 만료 시간 확인 (5분)
	if time.Since(stateSession.CreatedAt) > 5*time.Minute {
		delete(h.sessions, state) // 만료된 state 정리
		utils.Response.Error(c, models.ErrTokenExpired.WithDetails("State parameter expired"))
		return
	}

	// 사용된 state 즉시 삭제 (재사용 방지)
	delete(h.sessions, state)

	// 토큰 교환
	token, err := h.oidcProvider.ExchangeCode(ctx, code)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Token exchange failed", err, map[string]interface{}{
			"state": state,
		})
		utils.Response.Error(c, models.ErrInvalidCredentials.WithDetails("Failed to exchange authorization code").WithCause(err))
		return
	}

	// ID Token 검증
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		utils.Response.Error(c, models.ErrTokenInvalid.WithDetails("ID token not found in OAuth response"))
		return
	}

	idToken, err := h.oidcProvider.VerifyIDToken(ctx, rawIDToken)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "ID token verification failed", err, nil)
		utils.Response.Error(c, models.ErrTokenInvalid.WithDetails("ID token verification failed").WithCause(err))
		return
	}

	// 사용자 정보 추출
	var claims auth.TokenClaims
	if err := idToken.Claims(&claims); err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to parse ID token claims", err, nil)
		utils.Response.Error(c, models.ErrTokenInvalid.WithDetails("Failed to parse token claims").WithCause(err))
		return
	}

	// 세션 정보는 JWT 토큰에 포함되므로 별도 저장 불필요

	// JWT 토큰 생성
	jwtToken, err := h.jwtManager.GenerateToken(
		claims.Subject,
		token.AccessToken,
		rawIDToken,
		token.RefreshToken,
		token.Expiry,
	)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to generate JWT token", err, map[string]interface{}{
			"user_id": claims.Subject,
		})
		utils.Response.InternalError(c, fmt.Errorf("failed to generate session token: %w", err))
		return
	}

	// 인증 성공 로그 (사용자 ID를 컨텍스트에 추가)
	ctx = context.WithValue(c.Request.Context(), logger.UserIDKey, claims.Subject)
	c.Set("user_id", claims.Subject)

	logger.InfoWithContext(ctx, "User authenticated successfully", map[string]interface{}{
		"user_id": claims.Subject,
		"email":   claims.Email,
		"name":    claims.Name,
	})

	// JWT 토큰을 쿠키로 설정 (보안 강화)
	c.SetCookie(
		"session_token", // 쿠키 이름
		jwtToken,        // 쿠키 값
		3600*24,         // 만료 시간 (24시간)
		"/",             // 경로
		"",              // 도메인 (현재 도메인)
		false,           // HTTPS only (개발환경에서는 false)
		true,            // HTTP only (XSS 방지)
	)

	// 프론트엔드로 리디렉션 (쿠키 기반이므로 파라미터 불필요)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// HandleGetUser 사용자 정보 조회
func (h *AuthHandler) HandleGetUser(c *gin.Context) {
	// 쿠키에서 JWT 토큰 가져오기
	tokenString, err := c.Cookie("session_token")
	if err != nil {
		utils.Response.Error(c, models.ErrSessionNotFound.WithDetails("Session token cookie not found"))
		return
	}

	// JWT 토큰 검증
	claims, err := h.jwtManager.ValidateToken(tokenString)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Invalid JWT token received", map[string]interface{}{
			"error": err.Error(),
		})
		utils.Response.Error(c, models.ErrTokenInvalid.WithDetails("Session token validation failed").WithCause(err))
		return
	}

	// 사용자 ID를 컨텍스트에 설정
	c.Set("user_id", claims.UserID)
	ctx := context.WithValue(c.Request.Context(), logger.UserIDKey, claims.UserID)
	c.Request = c.Request.WithContext(ctx)

	utils.Response.Success(c, models.UserInfo{
		UserID:   claims.UserID,
		LoggedIn: true,
	})
}

// GetSessionFromJWT JWT에서 세션 정보 조회
func (h *AuthHandler) GetSessionFromJWT(c *gin.Context) (*models.Session, error) {
	// 쿠키에서 JWT 토큰 가져오기
	tokenString, err := c.Cookie("session_token")
	if err != nil {
		return nil, models.ErrSessionNotFound.WithDetails("Session token cookie not found")
	}

	// JWT 토큰 검증
	claims, err := h.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, models.ErrTokenInvalid.WithDetails("Session token validation failed").WithCause(err)
	}

	// JWT 클레임에서 세션 정보 생성
	session := &models.Session{
		AccessToken:  claims.AccessToken,
		IDToken:      claims.IDToken,
		RefreshToken: claims.RefreshToken,
		UserID:       claims.UserID,
		ExpiresAt:    claims.ExpiresAt,
		CreatedAt:    claims.RegisteredClaims.IssuedAt.Time,
	}

	return session, nil
}
