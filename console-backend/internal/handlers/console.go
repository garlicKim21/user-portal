package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"portal-backend/internal/auth"
	"portal-backend/internal/config"
	"portal-backend/internal/kubernetes"
	"portal-backend/internal/logger"
	"portal-backend/internal/models"
	"portal-backend/internal/utils"
)

// ConsoleHandler 웹 콘솔 핸들러
type ConsoleHandler struct {
	k8sClient   *kubernetes.Client
	authHandler *AuthHandler
	// 생성된 리소스 추적 (실제로는 Redis나 DB 사용 권장)
	resources map[string]*kubernetes.ConsoleResource
}

// NewConsoleHandler 새로운 콘솔 핸들러 생성
func NewConsoleHandler(k8sClient *kubernetes.Client, authHandler *AuthHandler) *ConsoleHandler {
	handler := &ConsoleHandler{
		k8sClient:   k8sClient,
		authHandler: authHandler,
		resources:   make(map[string]*kubernetes.ConsoleResource),
	}

	// 백그라운드에서 주기적으로 만료된 리소스 정리
	go handler.startCleanupRoutine()

	return handler
}

// HandleLaunchConsole 웹 콘솔 Pod 생성
func (h *ConsoleHandler) HandleLaunchConsole(c *gin.Context) {
	ctx := context.Background()

	// Authorization 헤더에서 Bearer 토큰(OIDC Access Token) 추출
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		utils.Response.Unauthorized(c, "Authorization header with Bearer token is required")
		return
	}

	oidcAccessToken := strings.TrimPrefix(authHeader, "Bearer ")

	// OIDC Access Token 검증 (userinfo 엔드포인트 사용)
	userInfo, err := h.validateOIDCAccessToken(ctx, oidcAccessToken)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Failed to verify OIDC token", map[string]any{
			"error": err.Error(),
		})
		utils.Response.Unauthorized(c, "Invalid OIDC token")
		return
	}

	userID := userInfo.PreferredUsername
	if userID == "" {
		userID = userInfo.Subject
	}

	// OIDC Access Token을 Kubernetes용 토큰으로 교환
	exchangeResp, err := auth.ExchangeTokenForKubernetes(oidcAccessToken)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to exchange token for kubernetes", err, map[string]any{
			"user_id": userID,
		})
		utils.Response.InternalError(c, fmt.Errorf("failed to get kubernetes token: %w", err))
		return
	}

	newK8sAccessToken := exchangeResp.AccessToken

	// 사용자 그룹 정보 확인 및 기본 네임스페이스 결정
	// ID 토큰에서 그룹 정보를 추출하기 위해 raw ID token을 사용
	rawIDToken := oidcAccessToken // 실제로는 ID token이어야 하지만, Access token에서 userinfo로 그룹 정보를 가져올 수 있음
	userGroups, err := auth.ExtractUserGroups(rawIDToken)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Failed to extract user groups", map[string]any{
			"user_id": userID,
			"error":   err.Error(),
		})
		userGroups = &auth.UserGroups{UserID: userID, Username: userID}
	}

	defaultNamespace := userGroups.DetermineDefaultNamespace()
	logger.InfoWithContext(c.Request.Context(), "Determined default namespace for user", map[string]any{
		"user_id":           userID,
		"groups":            userGroups.Groups,
		"default_namespace": defaultNamespace,
	})

	// 웹 콘솔 리소스 생성 (기본 네임스페이스 전달)
	resource, err := h.k8sClient.CreateConsoleResources(userID, newK8sAccessToken, "", defaultNamespace)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to create console resources", err, map[string]any{
			"user_id": userID,
		})
		utils.Response.Error(c, models.ErrPodCreationFailed.WithDetails("User: "+userID).WithCause(err))
		return
	}

	// 생성된 리소스 추적 저장
	h.resources[resource.ID] = resource

	logger.InfoWithContext(c.Request.Context(), "Web console created successfully", map[string]any{
		"user_id":     userID,
		"resource_id": resource.ID,
		"console_url": resource.ConsoleURL,
	})

	utils.Response.SuccessWithMessage(c, "Web console created successfully", models.LaunchConsoleResponse{
		URL:        resource.ConsoleURL,
		ResourceID: resource.ID,
	})
}

// HandleDeleteConsole 웹 콘솔 리소스 삭제
func (h *ConsoleHandler) HandleDeleteConsole(c *gin.Context) {
	ctx := context.Background()

	// Authorization 헤더에서 Bearer 토큰(OIDC Access Token) 추출
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		utils.Response.Unauthorized(c, "Authorization header with Bearer token is required")
		return
	}

	oidcAccessToken := strings.TrimPrefix(authHeader, "Bearer ")

	// OIDC Access Token 검증 (userinfo 엔드포인트 사용)
	userInfo, err := h.validateOIDCAccessToken(ctx, oidcAccessToken)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Failed to verify OIDC token", map[string]any{
			"error": err.Error(),
		})
		utils.Response.Unauthorized(c, "Invalid OIDC token")
		return
	}

	userID := userInfo.PreferredUsername
	if userID == "" {
		userID = userInfo.Subject
	}

	resourceID := c.Param("resourceId")
	if resourceID == "" {
		utils.Response.ValidationError(c, "resourceId", "Resource ID is required")
		return
	}

	// 리소스 조회
	resource, exists := h.resources[resourceID]
	if !exists {
		utils.Response.Error(c, models.ErrConsoleNotFound.WithDetails("Resource ID: "+resourceID))
		return
	}

	// 사용자 권한 확인
	if resource.UserID != userID {
		utils.Response.Forbidden(c, "You can only delete your own console resources")
		return
	}

	// 리소스 삭제
	logger.InfoWithContext(c.Request.Context(), "Deleting console resources", map[string]any{
		"user_id":     userID,
		"resource_id": resourceID,
		"deployment":  resource.DeploymentName,
		"service":     resource.ServiceName,
	})

	err = h.k8sClient.DeleteConsoleResources(resource)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to delete console resources", err, map[string]any{
			"user_id":     userID,
			"resource_id": resourceID,
		})
		utils.Response.KubernetesError(c, "delete console resources", err)
		return
	}

	// 메모리에서 제거
	delete(h.resources, resourceID)

	logger.InfoWithContext(c.Request.Context(), "Web console deleted successfully", map[string]any{
		"user_id":     userID,
		"resource_id": resourceID,
		"deployment":  resource.DeploymentName,
		"service":     resource.ServiceName,
	})
	utils.Response.SuccessWithMessage(c, "Console deleted successfully", gin.H{
		"resource_id": resourceID,
		"deleted_at":  time.Now().UTC(),
	})
}

// HandleListConsoles 사용자의 웹 콘솔 목록 조회
func (h *ConsoleHandler) HandleListConsoles(c *gin.Context) {
	ctx := context.Background()

	// Authorization 헤더에서 Bearer 토큰(OIDC Access Token) 추출
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		utils.Response.Unauthorized(c, "Authorization header with Bearer token is required")
		return
	}

	oidcAccessToken := strings.TrimPrefix(authHeader, "Bearer ")

	// OIDC Access Token 검증 (userinfo 엔드포인트 사용)
	userInfo, err := h.validateOIDCAccessToken(ctx, oidcAccessToken)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Failed to verify OIDC token", map[string]any{
			"error": err.Error(),
		})
		utils.Response.Unauthorized(c, "Invalid OIDC token")
		return
	}

	userID := userInfo.PreferredUsername
	if userID == "" {
		userID = userInfo.Subject
	}

	// 사용자의 리소스 필터링
	userResources := make([]*kubernetes.ConsoleResource, 0)
	for _, resource := range h.resources {
		if resource.UserID == userID {
			userResources = append(userResources, resource)
		}
	}

	utils.Response.Success(c, gin.H{
		"consoles": userResources,
		"count":    len(userResources),
		"user_id":  userID,
	})
}

// HandleLogout 사용자 로그아웃 처리 (K8s 리소스 정리 + 세션 삭제)
func (h *ConsoleHandler) HandleLogout(c *gin.Context) {
	ctx := context.Background()

	// JWT 쿠키에서 사용자 정보 추출
	userID, err := h.extractUserFromCookie(c)
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to extract user from cookie", err, map[string]any{
			"error": err.Error(),
		})
		utils.Response.Unauthorized(c, "Invalid or missing authentication cookie")
		return
	}

	logger.InfoWithContext(ctx, "Processing logout for user", map[string]any{
		"user_id": userID,
	})

	// 1. 사용자별 모든 Web Console 리소스 정리
	config := kubernetes.GetDefaultConfig()
	err = h.k8sClient.DeleteUserResources(userID, config.Namespace)
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to cleanup user resources", err, map[string]any{
			"user_id": userID,
		})
		// 리소스 정리 실패해도 로그아웃은 진행
	}

	// 2. 메모리에서 사용자 리소스 정리
	h.cleanupUserResourcesFromMemory(userID)

	// 3. JWT 쿠키 삭제
	c.SetCookie("portal-jwt", "", -1, "/", "", true, true)

	// 4. Keycloak 로그아웃 URL 생성
	logoutURL := h.generateKeycloakLogoutURL()

	logger.InfoWithContext(ctx, "Logout completed successfully", map[string]any{
		"user_id":    userID,
		"logout_url": logoutURL,
	})

	// 5. 응답 반환
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Logout completed successfully",
		"logout_url": logoutURL,
	})
}

// extractUserFromCookie JWT 쿠키에서 사용자 정보 추출
func (h *ConsoleHandler) extractUserFromCookie(c *gin.Context) (string, error) {
	// JWT 쿠키에서 토큰 추출
	jwtToken, err := c.Cookie("portal-jwt")
	if err != nil {
		return "", fmt.Errorf("portal-jwt cookie not found: %v", err)
	}

	// 간단한 JWT 토큰 파싱 (base64 디코딩)
	// JWT 형식: header.payload.signature
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT token format")
	}

	// payload 부분 디코딩
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT payload: %v", err)
	}

	// JSON 파싱
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse JWT claims: %v", err)
	}

	// 사용자 ID 추출
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in JWT claims")
	}

	return userID, nil
}

// cleanupUserResourcesFromMemory 메모리에서 사용자 리소스 정리
func (h *ConsoleHandler) cleanupUserResourcesFromMemory(userID string) {
	var resourcesToDelete []string

	// 사용자별 리소스 찾기
	for resourceID, resource := range h.resources {
		if resource.UserID == userID {
			resourcesToDelete = append(resourcesToDelete, resourceID)
		}
	}

	// 리소스 삭제
	for _, resourceID := range resourcesToDelete {
		delete(h.resources, resourceID)
		logger.InfoWithContext(context.TODO(), "Removed resource from memory", map[string]any{
			"resource_id": resourceID,
			"user_id":     userID,
		})
	}

	if len(resourcesToDelete) > 0 {
		logger.InfoWithContext(context.TODO(), "Memory cleanup completed", map[string]any{
			"user_id":           userID,
			"cleaned_resources": len(resourcesToDelete),
		})
	}
}

// generateKeycloakLogoutURL Keycloak 로그아웃 URL 생성
func (h *ConsoleHandler) generateKeycloakLogoutURL() string {
	config := config.Get()

	// Keycloak 로그아웃 URL 형식:
	// {issuer}/protocol/openid-connect/logout?client_id={client_id}&post_logout_redirect_uri={redirect_uri}
	logoutURL := fmt.Sprintf("%s/protocol/openid-connect/logout?client_id=%s&post_logout_redirect_uri=%s",
		config.OIDC.IssuerURL,
		config.OIDC.ClientID,
		"https://front.miribit.cloud", // 프론트엔드 URL로 리다이렉트
	)

	return logoutURL
}

// startCleanupRoutine 백그라운드 정리 루틴 시작
func (h *ConsoleHandler) startCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute) // 5분마다 정리
	defer ticker.Stop()

	for range ticker.C {
		h.cleanupExpiredResources()
	}
}

// cleanupExpiredResources 만료된 리소스 정리
func (h *ConsoleHandler) cleanupExpiredResources() {
	config := kubernetes.GetDefaultConfig()

	// 쿠버네티스에서 만료된 리소스 정리
	err := h.k8sClient.CleanupExpiredResources(config.Namespace)
	if err != nil {
		logger.Error("Failed to cleanup expired resources", err)
	}

	// 메모리에서 오래된 리소스 정리 (1시간 이상)
	cutoff := time.Now().Add(-1 * time.Hour)
	cleanedCount := 0
	for id, resource := range h.resources {
		if resource.CreatedAt.Before(cutoff) {
			logger.InfoWithContext(context.TODO(), "Removing expired resource from memory", map[string]any{
				"resource_id": id,
				"user_id":     resource.UserID,
				"created_at":  resource.CreatedAt,
				"deployment":  resource.DeploymentName,
				"service":     resource.ServiceName,
			})
			delete(h.resources, id)
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		logger.InfoWithContext(context.TODO(), "Memory cleanup completed", map[string]any{
			"cleaned_resources": cleanedCount,
		})
	}
}

// validateOIDCAccessToken OIDC Access Token을 userinfo 엔드포인트로 검증
func (h *ConsoleHandler) validateOIDCAccessToken(ctx context.Context, accessToken string) (*OIDCUserInfo, error) {
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

// HandleDeleteUserResources 사용자별 모든 Web Console 리소스 삭제
func (h *ConsoleHandler) HandleDeleteUserResources(c *gin.Context) {
	ctx := context.Background()

	// Authorization 헤더에서 Bearer 토큰(OIDC Access Token) 추출
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		utils.Response.Unauthorized(c, "Authorization header with Bearer token is required")
		return
	}

	oidcAccessToken := strings.TrimPrefix(authHeader, "Bearer ")

	// OIDC Access Token 검증 (userinfo 엔드포인트 사용)
	userInfo, err := h.validateOIDCAccessToken(ctx, oidcAccessToken)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Failed to verify OIDC token", map[string]any{
			"error": err.Error(),
		})
		utils.Response.Unauthorized(c, "Invalid OIDC token")
		return
	}

	userID := userInfo.PreferredUsername
	if userID == "" {
		userID = userInfo.Subject
	}

	// 사용자별 모든 리소스 삭제
	logger.InfoWithContext(c.Request.Context(), "Deleting all console resources for user", map[string]any{
		"user_id": userID,
	})

	cfg := config.Get()
	err = h.k8sClient.DeleteUserResources(userID, cfg.Console.Namespace)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to delete user console resources", err, map[string]any{
			"user_id": userID,
		})
		utils.Response.KubernetesError(c, "delete user console resources", err)
		return
	}

	// 메모리에서 해당 사용자의 모든 리소스 제거
	for resourceID, resource := range h.resources {
		if resource.UserID == userID {
			delete(h.resources, resourceID)
		}
	}

	logger.InfoWithContext(c.Request.Context(), "Successfully deleted all console resources for user", map[string]any{
		"user_id": userID,
	})

	// Keycloak 로그아웃 URL 생성
	logoutURL := fmt.Sprintf("%s/protocol/openid-connect/logout?client_id=%s&post_logout_redirect_uri=%s",
		h.authHandler.oidcProvider.GetIssuerURL(),
		"frontend",
		"https://portal.miribit.cloud")

	utils.Response.Success(c, gin.H{
		"message":    "All console resources deleted successfully",
		"user_id":    userID,
		"logout_url": logoutURL,
	})
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
