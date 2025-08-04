package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"portal-backend/internal/auth"
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
	// JWT에서 세션 정보 가져오기
	session, err := h.authHandler.GetSession(c)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Failed to get session from JWT", map[string]any{
			"error": err.Error(),
		})
		if apiErr, ok := err.(*models.APIError); ok {
			utils.Response.Error(c, apiErr)
		} else {
			utils.Response.Unauthorized(c, "Invalid or expired session")
		}
		return
	}

	logger.InfoWithContext(c.Request.Context(), "Performing token exchange for kubernetes access", map[string]any{
		"user_id": session.UserID,
	})

	// Access Token을 사용하여 Kubernetes 전용 토큰으로 교환합니다.
	exchangeResp, err := auth.ExchangeTokenForKubernetes(session.AccessToken)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to exchange token for kubernetes", err, map[string]any{
			"user_id": session.UserID,
		})
		utils.Response.InternalError(c, fmt.Errorf("failed to get kubernetes token: %w", err))
		return
	}

	newK8sAccessToken := exchangeResp.AccessToken

	// 사용자 그룹 정보 확인 및 로깅
	userGroups, err := auth.ExtractUserGroups(session.IDToken)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Failed to extract user groups", map[string]any{
			"user_id": session.UserID,
			"error":   err.Error(),
		})
	} else {
		logger.InfoWithContext(c.Request.Context(), "User groups extracted", map[string]any{
			"user_id":  session.UserID,
			"username": userGroups.Username,
			"groups":   userGroups.Groups,
			"role":     userGroups.GetUserRole(),
		})
	}

	// Token Exchange 수행: portal-app 토큰을 kubernetes 클라이언트용 토큰으로 교환
	logger.InfoWithContext(c.Request.Context(), "Performing token exchange for kubernetes access", map[string]any{
		"user_id": session.UserID,
	})

	tokenResponse, err := auth.ExchangeTokenForKubernetes(session.AccessToken)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to exchange token for kubernetes access", err, map[string]any{
			"user_id": session.UserID,
		})
		utils.Response.Error(c, models.ErrTokenExchangeFailed.WithDetails("User: "+session.UserID).WithCause(err))
		return
	}

	logger.InfoWithContext(c.Request.Context(), "Token exchange successful", map[string]any{
		"user_id":    session.UserID,
		"token_type": tokenResponse.TokenType,
		"expires_in": tokenResponse.ExpiresIn,
	})

	// 웹 콘솔 리소스 생성 (UUID 기반 이름으로 충돌 방지)
	logger.InfoWithContext(c.Request.Context(), "Creating console resources", map[string]any{
		"user_id": session.UserID,
		"role": func() string {
			if userGroups != nil {
				return userGroups.GetUserRole()
			} else {
				return "unknown"
			}
		}(),
	})

	// 교환된 토큰을 전달. Refresh Token은 기존 세션의 것을 사용
	resource, err := h.k8sClient.CreateConsoleResources(session.UserID, newK8sAccessToken, session.RefreshToken)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to create console resources", err, map[string]any{
			"user_id": session.UserID,
		})
		utils.Response.Error(c, models.ErrPodCreationFailed.WithDetails("User: "+session.UserID).WithCause(err))
		return
	}

	// 생성된 리소스 추적 저장
	h.resources[resource.ID] = resource

	logger.InfoWithContext(c.Request.Context(), "Web console created successfully", map[string]any{
		"user_id":         session.UserID,
		"resource_id":     resource.ID,
		"console_url":     resource.ConsoleURL,
		"deployment_name": resource.DeploymentName,
		"service_name":    resource.ServiceName,
	})

	utils.Response.SuccessWithMessage(c, "Web console created successfully", models.LaunchConsoleResponse{
		URL:        resource.ConsoleURL,
		ResourceID: resource.ID,
	})
}

// HandleDeleteConsole 웹 콘솔 리소스 삭제
func (h *ConsoleHandler) HandleDeleteConsole(c *gin.Context) {
	// JWT에서 세션 정보 가져오기
	session, err := h.authHandler.GetSession(c)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Failed to get session from JWT for delete operation", map[string]any{
			"error": err.Error(),
		})
		if apiErr, ok := err.(*models.APIError); ok {
			utils.Response.Error(c, apiErr)
		} else {
			utils.Response.Unauthorized(c, "Invalid or expired session")
		}
		return
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
	if resource.UserID != session.UserID {
		utils.Response.Forbidden(c, "You can only delete your own console resources")
		return
	}

	// 리소스 삭제
	logger.InfoWithContext(c.Request.Context(), "Deleting console resources", map[string]any{
		"user_id":     session.UserID,
		"resource_id": resourceID,
	})

	err = h.k8sClient.DeleteConsoleResources(resource)
	if err != nil {
		logger.ErrorWithContext(c.Request.Context(), "Failed to delete console resources", err, map[string]any{
			"user_id":     session.UserID,
			"resource_id": resourceID,
		})
		utils.Response.KubernetesError(c, "delete console resources", err)
		return
	}

	// 메모리에서 제거
	delete(h.resources, resourceID)

	logger.InfoWithContext(c.Request.Context(), "Web console deleted successfully", map[string]any{
		"user_id":     session.UserID,
		"resource_id": resourceID,
	})
	utils.Response.SuccessWithMessage(c, "Console deleted successfully", gin.H{
		"resource_id": resourceID,
		"deleted_at":  time.Now().UTC(),
	})
}

// HandleListConsoles 사용자의 웹 콘솔 목록 조회
func (h *ConsoleHandler) HandleListConsoles(c *gin.Context) {
	// JWT에서 세션 정보 가져오기
	session, err := h.authHandler.GetSession(c)
	if err != nil {
		logger.WarnWithContext(c.Request.Context(), "Failed to get session from JWT for list operation", map[string]any{
			"error": err.Error(),
		})
		if apiErr, ok := err.(*models.APIError); ok {
			utils.Response.Error(c, apiErr)
		} else {
			utils.Response.Unauthorized(c, "Invalid or expired session")
		}
		return
	}

	// 사용자의 리소스 필터링
	userResources := make([]*kubernetes.ConsoleResource, 0)
	for _, resource := range h.resources {
		if resource.UserID == session.UserID {
			userResources = append(userResources, resource)
		}
	}

	utils.Response.Success(c, gin.H{
		"consoles": userResources,
		"count":    len(userResources),
		"user_id":  session.UserID,
	})
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
	} else {
		logger.Debug("Expired resources cleanup completed")
	}

	// 메모리에서 오래된 리소스 정리 (1시간 이상)
	cutoff := time.Now().Add(-1 * time.Hour)
	cleanedCount := 0
	for id, resource := range h.resources {
		if resource.CreatedAt.Before(cutoff) {
			delete(h.resources, id)
			cleanedCount++
			logger.DebugWithContext(context.TODO(), "Removed expired resource from memory", map[string]any{
				"resource_id": id,
				"user_id":     resource.UserID,
				"created_at":  resource.CreatedAt,
			})
		}
	}

	if cleanedCount > 0 {
		logger.InfoWithContext(context.TODO(), "Memory cleanup completed", map[string]any{
			"cleaned_resources": cleanedCount,
		})
	}
}
