package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"portal-backend/internal/kubernetes"
	"portal-backend/internal/models"
)

// ConsoleHandler 웹 콘솔 핸들러
type ConsoleHandler struct {
	k8sClient *kubernetes.Client
	authHandler *AuthHandler
}

// NewConsoleHandler 새로운 콘솔 핸들러 생성
func NewConsoleHandler(k8sClient *kubernetes.Client, authHandler *AuthHandler) *ConsoleHandler {
	return &ConsoleHandler{
		k8sClient:   k8sClient,
		authHandler: authHandler,
	}
}

// HandleLaunchConsole 웹 콘솔 Pod 생성
func (h *ConsoleHandler) HandleLaunchConsole(c *gin.Context) {
	// 세션 ID 가져오기 (실제로는 쿠키나 헤더에서)
	sessionID := c.Query("session")
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No session provided"})
		return
	}

	// 세션 확인
	session, exists := h.authHandler.GetSession(sessionID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	// kubeconfig 생성
	kubeconfig := kubernetes.GenerateKubeconfig(session.IDToken)

	// ConfigMap 생성
	configMapName := fmt.Sprintf("kubeconfig-%s", session.UserID)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: "default",
		},
		Data: map[string]string{
			"config": kubeconfig,
		},
	}

	_, err := h.k8sClient.Clientset.CoreV1().ConfigMaps("default").Create(context.Background(), configMap, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Failed to create ConfigMap: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ConfigMap"})
		return
	}

	// Pod 생성
	podName := fmt.Sprintf("console-%s", session.UserID)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "web-console",
					Image: "tsl0922/ttyd:latest", // 웹 터미널 이미지
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 7681,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "kubeconfig",
							MountPath: "/home/user/.kube",
						},
					},
					Command: []string{"ttyd"},
					Args:    []string{"-p", "7681", "kubectl", "exec", "-it", "kubectl-pod", "--", "kubectl"},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "kubeconfig",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: configMapName,
							},
							Items: []corev1.KeyToPath{
								{
									Key:  "config",
									Path: "config",
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = h.k8sClient.Clientset.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Failed to create Pod: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Pod"})
		return
	}

	// Service 생성
	serviceName := fmt.Sprintf("console-service-%s", session.UserID)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": podName,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(7681),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	_, err = h.k8sClient.Clientset.CoreV1().Services("default").Create(context.Background(), service, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Failed to create Service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Service"})
		return
	}

	// 응답 URL 생성 (실제로는 Ingress를 통해 외부 접근 가능한 URL)
	consoleURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", serviceName, "default")

	log.Printf("Web console created for user %s: %s", session.UserID, consoleURL)

	c.JSON(http.StatusOK, models.LaunchConsoleResponse{
		URL: consoleURL,
	})
} 