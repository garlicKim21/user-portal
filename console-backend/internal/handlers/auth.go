package handlers

import (
	"portal-backend/internal/auth"
	"portal-backend/internal/kubernetes"
)

// AuthHandler 인증 핸들러
type AuthHandler struct {
	oidcProvider *auth.OIDCProvider
	k8sClient    *kubernetes.Client
}

// NewAuthHandler 새로운 인증 핸들러 생성
func NewAuthHandler(oidcProvider *auth.OIDCProvider, k8sClient *kubernetes.Client) (*AuthHandler, error) {
	return &AuthHandler{
		oidcProvider: oidcProvider,
		k8sClient:    k8sClient,
	}, nil
}
