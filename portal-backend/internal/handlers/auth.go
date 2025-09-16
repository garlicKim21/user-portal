package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"portal-backend/internal/auth"
	"portal-backend/internal/kubernetes"
	"portal-backend/internal/logger"
	"portal-backend/internal/models"
	"portal-backend/internal/utils"
)

// AuthHandler 인증 핸들러
type AuthHandler struct {
	oidcProvider *auth.OIDCProvider
	sessionStore *auth.SessionStore
	jwtManager   *auth.JWTManager
	tempSessions map[string]*models.Session
	k8sClient    *kubernetes.Client
}

// NewAuthHandler 새로운 인증 핸들러 생성
func NewAuthHandler(oidcProvider *auth.OIDCProvider, k8sClient *kubernetes.Client) (*AuthHandler, error) {
	jwtManager, err := auth.NewJWTManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT manager: %w", err)
	}

	return &AuthHandler{
		oidcProvider: oidcProvider,
		sessionStore: auth.NewSessionStore(),
		jwtManager:   jwtManager,
		tempSessions: make(map[string]*models.Session),
		k8sClient:    k8sClient,
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

	authURL := h.oidcProvider.GetAuthURL(state)
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
		logger.ErrorWithContext(c.Request.Context(), "Failed to create session", err, map[string]any{
			"user_id": claims.Subject,
		})
		utils.Response.InternalError(c, fmt.Errorf("failed to create session: %w", err))
		return
	}

	// JWT 토큰 생성 (세션 ID 포함)
	jwtToken, err := h.jwtManager.GenerateToken(userID, sessionID, token.Expiry)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to generate JWT token", err, map[string]any{
			"user_id":    claims.Subject,
			"session_id": sessionID,
		})
		utils.Response.InternalError(c, fmt.Errorf("failed to generate JWT token: %w", err))
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

	// JWT 토큰을 쿠키에 설정
	cookie := &http.Cookie{
		Name:     "portal-jwt",
		Value:    jwtToken,
		Path:     "/",
		Domain:   "",
		MaxAge:   3600,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(c.Writer, cookie)
	c.Redirect(http.StatusFound, "/dashboard")
}

// HandleGetUser 사용자 정보 조회
func (h *AuthHandler) HandleGetUser(c *gin.Context) {
	jwtToken, err := c.Cookie("portal-jwt")
	if err != nil {
		utils.Response.Error(c, models.ErrSessionNotFound.WithDetails("JWT token cookie not found"))
		return
	}

	// JWT 토큰 검증
	claims, err := h.jwtManager.ValidateToken(jwtToken)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Invalid JWT token received", map[string]any{
			"error": err.Error(),
		})
		utils.Response.Error(c, models.ErrTokenInvalid.WithDetails("JWT token validation failed").WithCause(err))
		return
	}

	// 세션에서 사용자 정보 조회
	session, err := h.sessionStore.GetSession(claims.SessionID)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Session not found", map[string]any{
			"session_id": claims.SessionID,
			"error":      err.Error(),
		})
		utils.Response.Error(c, models.ErrSessionNotFound.WithDetails("Session not found").WithCause(err))
		return
	}

	c.Set("user_id", session.UserID)
	ctx := context.WithValue(c.Request.Context(), logger.UserIDKey, session.UserID)
	c.Request = c.Request.WithContext(ctx)

	utils.Response.Success(c, models.UserInfo{
		UserID:   session.UserID,
		LoggedIn: true,
	})
}

// GetSession JWT에서 세션 정보 조회
func (h *AuthHandler) GetSession(c *gin.Context) (*models.Session, error) {
	jwtToken, err := c.Cookie("portal-jwt")
	if err != nil {
		return nil, models.ErrSessionNotFound.WithDetails("JWT token cookie not found")
	}

	// JWT 토큰 검증
	claims, err := h.jwtManager.ValidateToken(jwtToken)
	if err != nil {
		return nil, models.ErrTokenInvalid.WithDetails("JWT token validation failed").WithCause(err)
	}

	// 세션에서 사용자 정보 조회
	session, err := h.sessionStore.GetSession(claims.SessionID)
	if err != nil {
		return nil, models.ErrSessionNotFound.WithDetails("Session not found").WithCause(err)
	}

	return session, nil
}

// HandleLogout 로그아웃 처리
func (h *AuthHandler) HandleLogout(c *gin.Context) {
	var idTokenHint string
	var userID string

	// JWT 토큰에서 세션 정보 조회
	jwtToken, err := c.Cookie("portal-jwt")
	if err == nil {
		claims, jwtErr := h.jwtManager.ValidateToken(jwtToken)
		if jwtErr == nil {
			session, sessionErr := h.sessionStore.GetSession(claims.SessionID)
			if sessionErr == nil {
				idTokenHint = session.IDToken
				userID = session.UserID
				// 세션 삭제
				h.sessionStore.DeleteSession(claims.SessionID)
			}
		}
	}

	// 사용자 관련 쿠버네티스 리소스 정리
	if userID != "" && h.k8sClient != nil {
		logger.InfoWithContext(c.Request.Context(), "Cleaning up user resources on logout", map[string]any{
			"user_id": userID,
		})
		err := h.cleanupUserResources(userID)
		if err != nil {
			logger.ErrorWithContext(c.Request.Context(), "Failed to cleanup user resources", err, map[string]any{
				"user_id": userID,
			})
		}
	}

	// JWT 쿠키 삭제
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "portal-jwt",
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

// cleanupUserResources 사용자 관련 쿠버네티스 리소스 정리
func (h *AuthHandler) cleanupUserResources(userID string) error {
	namespace := os.Getenv("CONSOLE_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	ctx := context.Background()
	labelSelector := fmt.Sprintf("app=web-console,user=%s", userID)

	logger.InfoWithContext(ctx, "Starting resource cleanup", map[string]any{
		"user_id":        userID,
		"namespace":      namespace,
		"label_selector": labelSelector,
	})

	// Deployment 삭제
	deployments, err := h.k8sClient.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list deployments for cleanup", err, map[string]any{
			"user_id": userID,
		})
	} else {
		for _, deployment := range deployments.Items {
			err := h.k8sClient.Clientset.AppsV1().Deployments(namespace).Delete(ctx, deployment.Name, metav1.DeleteOptions{})
			if err != nil {
				logger.ErrorWithContext(ctx, "Failed to delete deployment", err, map[string]any{
					"user_id":         userID,
					"deployment_name": deployment.Name,
				})
			} else {
				logger.InfoWithContext(ctx, "Successfully deleted deployment", map[string]any{
					"user_id":         userID,
					"deployment_name": deployment.Name,
				})
			}
		}
	}

	// Service 삭제
	services, err := h.k8sClient.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list services for cleanup", err, map[string]any{
			"user_id": userID,
		})
	} else {
		for _, service := range services.Items {
			err := h.k8sClient.Clientset.CoreV1().Services(namespace).Delete(ctx, service.Name, metav1.DeleteOptions{})
			if err != nil {
				logger.ErrorWithContext(ctx, "Failed to delete service", err, map[string]any{
					"user_id":      userID,
					"service_name": service.Name,
				})
			} else {
				logger.InfoWithContext(ctx, "Successfully deleted service", map[string]any{
					"user_id":      userID,
					"service_name": service.Name,
				})
			}
		}
	}

	// Secret 삭제
	secrets, err := h.k8sClient.Clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list secrets for cleanup", err, map[string]any{
			"user_id": userID,
		})
	} else {
		for _, secret := range secrets.Items {
			err := h.k8sClient.Clientset.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
			if err != nil {
				logger.ErrorWithContext(ctx, "Failed to delete secret", err, map[string]any{
					"user_id":     userID,
					"secret_name": secret.Name,
				})
			} else {
				logger.InfoWithContext(ctx, "Successfully deleted secret", map[string]any{
					"user_id":     userID,
					"secret_name": secret.Name,
				})
			}
		}
	}

	// Ingress 삭제
	ingresses, err := h.k8sClient.Clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list ingresses for cleanup", err, map[string]any{
			"user_id": userID,
		})
	} else {
		for _, ingress := range ingresses.Items {
			err := h.k8sClient.Clientset.NetworkingV1().Ingresses(namespace).Delete(ctx, ingress.Name, metav1.DeleteOptions{})
			if err != nil {
				logger.ErrorWithContext(ctx, "Failed to delete ingress", err, map[string]any{
					"user_id":      userID,
					"ingress_name": ingress.Name,
				})
			} else {
				logger.InfoWithContext(ctx, "Successfully deleted ingress", map[string]any{
					"user_id":      userID,
					"ingress_name": ingress.Name,
				})
			}
		}
	}

	return nil
}

// HandleCreateSessionFromOIDC OIDC 토큰으로 백엔드 세션 생성 (poc_front용)
func (h *AuthHandler) HandleCreateSessionFromOIDC(c *gin.Context) {
	ctx := context.Background()

	// Authorization 헤더에서 OIDC 토큰 추출
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		utils.Response.Unauthorized(c, "Authorization header with Bearer token is required")
		return
	}

	oidcAccessToken := strings.TrimPrefix(authHeader, "Bearer ")

	// OIDC 토큰 검증을 위해 userinfo 엔드포인트 호출
	userInfoResp, err := h.validateOIDCTokenViaUserInfo(ctx, oidcAccessToken)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Failed to validate OIDC token", map[string]any{
			"error": err.Error(),
		})
		utils.Response.Unauthorized(c, "Invalid OIDC token")
		return
	}

	userID := userInfoResp.PreferredUsername
	if userID == "" {
		userID = userInfoResp.Subject
	}

	// 기존 세션이 있다면 삭제 (중복 세션 방지)
	h.cleanupExistingUserSessions(userID)

	// 새 세션 생성 (OIDC 토큰 정보 저장)
	// 만료 시간은 현재 시간 + 1시간으로 설정 (OIDC 토큰 만료 시간을 정확히 알 수 없으므로)
	expiresAt := time.Now().Add(1 * time.Hour)

	sessionID, err := h.sessionStore.CreateSession(
		userID,
		oidcAccessToken, // OIDC Access Token 저장
		"",              // ID Token은 userinfo에서 받지 않으므로 빈 값
		"",              // Refresh Token도 마찬가지
		expiresAt,
	)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to create session", err, map[string]any{
			"user_id": userID,
		})
		utils.Response.InternalError(c, fmt.Errorf("failed to create session: %w", err))
		return
	}

	// JWT 토큰 생성 (세션 ID 포함)
	jwtToken, err := h.jwtManager.GenerateToken(userID, sessionID, expiresAt)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to generate JWT token", err, map[string]any{
			"user_id":    userID,
			"session_id": sessionID,
		})
		utils.Response.InternalError(c, fmt.Errorf("failed to generate JWT token: %w", err))
		return
	}

	ctx = context.WithValue(c.Request.Context(), logger.UserIDKey, userID)
	c.Set("user_id", userID)

	logger.InfoWithContext(ctx, "Session created from OIDC token", map[string]any{
		"user_id":    userID,
		"session_id": sessionID,
	})

	// JWT 토큰을 쿠키에 설정
	cookie := &http.Cookie{
		Name:     "portal-jwt",
		Value:    jwtToken,
		Path:     "/",
		Domain:   "",
		MaxAge:   3600, // 1시간
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(c.Writer, cookie)

	utils.Response.SuccessWithMessage(c, "Session created successfully", gin.H{
		"user_id": userID,
	})
}

// validateOIDCTokenViaUserInfo OIDC 토큰을 userinfo 엔드포인트로 검증
func (h *AuthHandler) validateOIDCTokenViaUserInfo(ctx context.Context, accessToken string) (*OIDCUserInfo, error) {
	// Keycloak userinfo 엔드포인트 URL 구성
	issuerURL := os.Getenv("OIDC_ISSUER_URL")
	if issuerURL == "" {
		return nil, fmt.Errorf("OIDC_ISSUER_URL environment variable is required")
	}

	userInfoURL := issuerURL + "/protocol/openid-connect/userinfo"

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call userinfo endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned status %d", resp.StatusCode)
	}

	var userInfo OIDCUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %v", err)
	}

	return &userInfo, nil
}

// cleanupExistingUserSessions 기존 사용자 세션 정리
func (h *AuthHandler) cleanupExistingUserSessions(userID string) {
	deletedSessionIDs := h.sessionStore.DeleteUserSessions(userID)

	for _, sessionID := range deletedSessionIDs {
		logger.InfoWithContext(context.TODO(), "Cleaned up existing session", map[string]any{
			"user_id":    userID,
			"session_id": sessionID,
		})
	}
}

// OIDCUserInfo OIDC userinfo 응답 구조체
type OIDCUserInfo struct {
	Subject           string `json:"sub"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	EmailVerified     bool   `json:"email_verified"`
}
