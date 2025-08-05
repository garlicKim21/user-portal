package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
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
	tempSessions map[string]*models.Session
	k8sClient    *kubernetes.Client
}

// NewAuthHandler 새로운 인증 핸들러 생성
func NewAuthHandler(oidcProvider *auth.OIDCProvider, k8sClient *kubernetes.Client) (*AuthHandler, error) {
	return &AuthHandler{
		oidcProvider: oidcProvider,
		sessionStore: auth.NewSessionStore(),
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

	http.SetCookie(c.Writer, cookie)
	c.Redirect(http.StatusFound, "/dashboard")
}

// HandleGetUser 사용자 정보 조회
func (h *AuthHandler) HandleGetUser(c *gin.Context) {
	tokenString, err := c.Cookie("portal-session")
	if err != nil {
		utils.Response.Error(c, models.ErrSessionNotFound.WithDetails("Session token cookie not found"))
		return
	}

	claims, err := h.sessionStore.GetSession(tokenString)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Invalid JWT token received", map[string]any{
			"error": err.Error(),
		})
		utils.Response.Error(c, models.ErrTokenInvalid.WithDetails("Session token validation failed").WithCause(err))
		return
	}

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
	tokenString, err := c.Cookie("portal-session")
	if err != nil {
		return nil, models.ErrSessionNotFound.WithDetails("Session token cookie not found")
	}

	claims, err := h.sessionStore.GetSession(tokenString)
	if err != nil {
		return nil, models.ErrTokenInvalid.WithDetails("Session token validation failed").WithCause(err)
	}

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
	var userID string
	sessionID, err := c.Cookie("portal-session")
	if err == nil {
		session, sessionErr := h.sessionStore.GetSession(sessionID)
		if sessionErr == nil {
			idTokenHint = session.IDToken
			userID = session.UserID
		}
		h.sessionStore.DeleteSession(sessionID)
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
