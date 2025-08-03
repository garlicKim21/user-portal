package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
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
	sessionStore *auth.SessionStore
	tempSessions map[string]*models.Session
}

// NewAuthHandler 새로운 인증 핸들러 생성
func NewAuthHandler(oidcProvider *auth.OIDCProvider) (*AuthHandler, error) {
	return &AuthHandler{
		oidcProvider: oidcProvider,
		sessionStore: auth.NewSessionStore(), // 세션 저장소 초기화
		tempSessions: make(map[string]*models.Session),
	}, nil
}

// HandleLogin 로그인 처리
func (h *AuthHandler) HandleLogin(c *gin.Context) {
	state, err := auth.GenerateRandomString(32)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to generate state", err, map[string]any{})
		utils.Response.InternalError(c, fmt.Errorf("failed to generate security token: %w", err))
		return
	}

	h.tempSessions[state] = &models.Session{
		State:     state,
		CreatedAt: time.Now(),
	}

	// OAuth2 인증 URL 생성
	authURL := h.oidcProvider.GetAuthURL(state)

	logger.InfoWithContext(c.Request.Context(), "Redirecting to OAuth2 provider", map[string]any{
		"auth_url": authURL,
		"state":    state,
	})
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// HandleCallback OAuth2 콜백 처리
func (h *AuthHandler) HandleCallback(c *gin.Context) {
	ctx := context.Background()

	if errParam := c.Query("error"); errParam != "" {
		utils.Response.Error(c, models.ErrInvalidCredentials.WithDetails("OAuth error: "+errParam))
		return
	}

	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		utils.Response.ValidationError(c, "code", "Authorization code is required")
		return
	}

	stateSession, exists := h.tempSessions[state]
	if !exists {
		utils.Response.Error(c, models.ErrInvalidCredentials.WithDetails("Invalid or expired state parameter"))
		return
	}

	if time.Since(stateSession.CreatedAt) > 5*time.Minute {
		delete(h.tempSessions, state)
		utils.Response.Error(c, models.ErrTokenExpired.WithDetails("State parameter expired"))
		return
	}

	delete(h.tempSessions, state)

	token, err := h.oidcProvider.ExchangeCode(ctx, code)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Token exchange failed", err, map[string]any{
			"state": state,
		})
		utils.Response.Error(c, models.ErrInvalidCredentials.WithDetails("Failed to exchange authorization code").WithCause(err))
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		utils.Response.Error(c, models.ErrTokenInvalid.WithDetails("ID token not found in OAuth response"))
		return
	}

	idToken, err := h.oidcProvider.VerifyIDToken(ctx, rawIDToken)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "ID token verification failed", err, map[string]any{})
		utils.Response.Error(c, models.ErrTokenInvalid.WithDetails("ID token verification failed").WithCause(err))
		return
	}

	var claims auth.TokenClaims
	if err := idToken.Claims(&claims); err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to parse ID token claims", err, map[string]any{})
		utils.Response.Error(c, models.ErrTokenInvalid.WithDetails("Failed to parse token claims").WithCause(err))
		return
	}
	userID := claims.PreferredUsername
	if userID == "" {
		userID = claims.Subject
	}

	sessionID, err := h.sessionStore.CreateSession(
		userID,
		token.AccessToken,
		rawIDToken,
		token.RefreshToken,
		token.Expiry,
	)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to generate JWT token", err, map[string]any{
			"user_id": claims.Subject,
		})
		utils.Response.InternalError(c, fmt.Errorf("failed to generate session token: %w", err))
		return
	}

	ctx = context.WithValue(c.Request.Context(), logger.UserIDKey, claims.Subject)
	c.Set("user_id", claims.Subject)

	logger.InfoWithContext(ctx, "User authenticated successfully", map[string]any{
		"user_id":    claims.Subject,
		"email":      claims.Email,
		"name":       claims.Name,
		"session_id": sessionID,
	})

	cookie := &http.Cookie{
		Name:     "portal-session",
		Value:    sessionID,
		Path:     "/",
		Domain:   "",
		MaxAge:   3600,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}

	logger.DebugWithContext(ctx, "Setting session cookie", map[string]any{
		"session_id":    sessionID,
		"cookie_size":   len(sessionID),
		"cookie_name":   cookie.Name,
		"cookie_secure": cookie.Secure,
		"samesite":      "None",
	})

	http.SetCookie(c.Writer, cookie)

	c.Redirect(http.StatusFound, "/dashboard")
}

// HandleGetUser 사용자 정보 조회
func (h *AuthHandler) HandleGetUser(c *gin.Context) {
	logger.DebugWithContext(c.Request.Context(), "Request headers", map[string]any{
		"user_agent": c.Request.UserAgent(),
		"referer":    c.Request.Referer(),
		"origin":     c.Request.Header.Get("Origin"),
		"host":       c.Request.Host,
	})

	cookies := make(map[string]string)
	for _, cookie := range c.Request.Cookies() {
		value := cookie.Value
		if len(value) > 20 {
			value = value[:20] + "..."
		}
		cookies[cookie.Name] = value
	}

	logger.DebugWithContext(c.Request.Context(), "All cookies received", map[string]any{
		"cookies_count": len(c.Request.Cookies()),
		"cookies":       cookies,
	})

	// 새로운 쿠키명으로 JWT 토큰 가져오기
	tokenString, err := c.Cookie("portal-session")
	if err != nil {
		logger.DebugWithContext(c.Request.Context(), "Session token cookie not found", map[string]any{
			"error": err.Error(),
		})
		utils.Response.Error(c, models.ErrSessionNotFound.WithDetails("Session token cookie not found"))
		return
	}

	// JWT 토큰 검증
	claims, err := h.sessionStore.GetSession(tokenString)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Invalid JWT token received", map[string]any{
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

// GetSession JWT에서 세션 정보 조회
func (h *AuthHandler) GetSession(c *gin.Context) (*models.Session, error) {
	// 쿠키에서 JWT 토큰 가져오기
	tokenString, err := c.Cookie("portal-session")
	if err != nil {
		return nil, models.ErrSessionNotFound.WithDetails("Session token cookie not found")
	}

	// JWT 토큰 검증
	claims, err := h.sessionStore.GetSession(tokenString)
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
		CreatedAt:    claims.CreatedAt,
	}

	return session, nil
}

// HandleLogout 로그아웃 처리
func (h *AuthHandler) HandleLogout(c *gin.Context) {
	var idTokenHint string
	sessionID, err := c.Cookie("portal-session")
	if err == nil {
		session, sessionErr := h.sessionStore.GetSession(sessionID)
		if sessionErr == nil {
			idTokenHint = session.IDToken
		}
		h.sessionStore.DeleteSession(sessionID)
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "portal-session",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})

	oidcIssuerURL := os.Getenv("OIDC_ISSUER_URL")
	postLogoutURL := os.Getenv("ALLOWED_ORIGINS")

	logoutURL, err := url.Parse(oidcIssuerURL + "/protocol/openid-connect/logout")
	if err != nil {
		utils.Response.InternalError(c, fmt.Errorf("failed to parse logout url: %w", err))
		return
	}

	params := url.Values{}
	params.Add("post_logout_redirect_uri", postLogoutURL)

	if idTokenHint != "" {
		params.Add("id_token_hint", idTokenHint)
	}

	logoutURL.RawQuery = params.Encode()

	c.JSON(http.StatusOK, gin.H{
		"message":    "Local session cleared. Redirect to Keycloak for full logout.",
		"logout_url": logoutURL.String(),
	})
}
