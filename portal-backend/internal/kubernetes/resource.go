package kubernetes

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

// ConsoleResource 웹 콘솔 리소스 정보
type ConsoleResource struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	DeploymentName string    `json:"deployment_name"`
	ServiceName    string    `json:"service_name"`
	SecretName     string    `json:"secret_name"`
	PVCName        string    `json:"pvc_name"`
	Namespace      string    `json:"namespace"`
	ConsoleURL     string    `json:"console_url"`
	CreatedAt      time.Time `json:"created_at"`
}

// ConsoleConfig 웹 콘솔 설정
type ConsoleConfig struct {
	Namespace     string
	Image         string
	ContainerPort int32
	ServicePort   int32
	TTLSeconds    int32
}

// GetDefaultConfig 기본 설정 반환
func GetDefaultConfig() *ConsoleConfig {
	namespace := os.Getenv("CONSOLE_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	image := os.Getenv("CONSOLE_IMAGE")
	if image == "" {
		// 커스텀 웹 터미널 이미지 사용 (kubectl + ttyd + 간소화된 버전)
		image = "projectgreenist/web-terminal:0.2.3"
	}

	containerPort := int32(8080)
	if port := os.Getenv("CONSOLE_CONTAINER_PORT"); port != "" {
		if p := parseInt32(port); p > 0 {
			containerPort = p
		}
	}

	servicePort := int32(80)
	if port := os.Getenv("CONSOLE_SERVICE_PORT"); port != "" {
		if p := parseInt32(port); p > 0 {
			servicePort = p
		}
	}

	ttlSeconds := int32(3600) // 기본 1시간
	if ttl := os.Getenv("CONSOLE_TTL_SECONDS"); ttl != "" {
		if t := parseInt32(ttl); t > 0 {
			ttlSeconds = t
		}
	}

	return &ConsoleConfig{
		Namespace:     namespace,
		Image:         image,
		ContainerPort: containerPort,
		ServicePort:   servicePort,
		TTLSeconds:    ttlSeconds,
	}
}

// CreateConsoleResources 웹 콘솔 리소스 생성
func (c *Client) CreateConsoleResources(userID, idToken, refreshToken string) (*ConsoleResource, error) {
	config := GetDefaultConfig()
	// 전체 UUID + timestamp로 고유성 보장
	fullUUID := uuid.New().String()
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	resourceID := fmt.Sprintf("%s-%s", fullUUID, timestamp)

	consoleResource := &ConsoleResource{
		ID:             resourceID,
		UserID:         userID,
		DeploymentName: fmt.Sprintf("console-%s-%s", userID, fullUUID),
		ServiceName:    fmt.Sprintf("console-svc-%s-%s", userID, fullUUID),
		SecretName:     fmt.Sprintf("kubeconfig-secret-%s-%s", userID, fullUUID),
		PVCName:        fmt.Sprintf("history-%s", userID), // 사용자별 히스토리는 공유
		Namespace:      config.Namespace,
		CreatedAt:      time.Now(),
	}

	ctx := context.Background()

	// 1. PVC 생성 (명령어 히스토리 영구 보존) - 사용자별로 한 개만 생성
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consoleResource.PVCName,
			Namespace: consoleResource.Namespace,
			Labels: map[string]string{
				"app":  "web-console",
				"user": userID,
				"type": "history",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("100Mi"),
				},
			},
			StorageClassName: func() *string { s := "local-path"; return &s }(),
		},
	}

	// PVC가 존재하지 않으면 생성 (사용자별 히스토리는 공유)
	_, err := c.Clientset.CoreV1().PersistentVolumeClaims(consoleResource.Namespace).Get(ctx, consoleResource.PVCName, metav1.GetOptions{})
	if err != nil {
		_, err = c.Clientset.CoreV1().PersistentVolumeClaims(consoleResource.Namespace).Create(ctx, pvc, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create PVC: %v", err)
		}
	}

	// 2. Secret 생성 (kubeconfig 보안 저장)
	kubeconfig := GenerateKubeconfig(idToken)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consoleResource.SecretName,
			Namespace: consoleResource.Namespace,
			Labels: map[string]string{
				"app":     "web-console",
				"user":    userID,
				"session": resourceID,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"config": []byte(kubeconfig),
		},
	}

	_, err = c.Clientset.CoreV1().Secrets(consoleResource.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Secret: %v", err)
	}

	// 3. Deployment 생성
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consoleResource.DeploymentName,
			Namespace: consoleResource.Namespace,
			Labels: map[string]string{
				"app":     "web-console",
				"user":    userID,
				"session": resourceID,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func(i int32) *int32 { return &i }(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "web-console",
					"user":    userID,
					"session": resourceID,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     "web-console",
						"user":    userID,
						"session": resourceID,
					},
				},
				Spec: corev1.PodSpec{
					// ActiveDeadlineSeconds는 Deployment PodTemplate에서 지원되지 않음
					// 대신 CleanupExpiredResources 함수로 주기적 정리
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:  func(i int64) *int64 { return &i }(1000),
						RunAsGroup: func(i int64) *int64 { return &i }(1000),
						FSGroup:    func(i int64) *int64 { return &i }(1000),
					},
					Containers: []corev1.Container{
						{
							Name:    "web-console",
							Image:   config.Image,
							Command: []string{"/bin/sh", "-c"},
							Args: []string{
								`
								set -e
								echo "Initializing web terminal environment..."
								
								# Check kubeconfig is available
								if [ -f /home/user/.kube/config ]; then
									echo "Kubeconfig found, checking connectivity..."
									kubectl version --client || echo "kubectl client ready"
								else
									echo "Warning: kubeconfig not found"
								fi
								
								# Start ttyd service
								echo "Starting ttyd service..."
								exec ttyd --port 8080 --writable --max-clients 1 bash
								`,
							},
							Env: []corev1.EnvVar{
								{Name: "KUBECONFIG", Value: "/home/user/.kube/config"},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "kubeconfig",
									MountPath: "/home/user/.kube/config",
									SubPath:   "config",
									//ReadOnly:  true,
								},
								{
									Name:      "history-storage",
									MountPath: "/home/user/.bash_history",
									SubPath:   "bash_history",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),  // 최소 0.1 CPU
									corev1.ResourceMemory: resource.MustParse("128Mi"), // 최소 128MB 메모리
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),  // 최대 0.25 CPU
									corev1.ResourceMemory: resource.MustParse("256Mi"), // 최대 256MB 메모리
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "kubeconfig",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: consoleResource.SecretName,
									Items: []corev1.KeyToPath{
										{
											Key:  "config",
											Path: "config",
										},
									},
								},
							},
						},
						{
							Name: "history-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: consoleResource.PVCName,
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}

	_, err = c.Clientset.AppsV1().Deployments(consoleResource.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		// Secret 정리
		c.Clientset.CoreV1().Secrets(consoleResource.Namespace).Delete(ctx, consoleResource.SecretName, metav1.DeleteOptions{})
		return nil, fmt.Errorf("failed to create Deployment: %v", err)
	}

	// 4. Service 생성
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consoleResource.ServiceName,
			Namespace: consoleResource.Namespace,
			Labels: map[string]string{
				"app":     "web-console",
				"user":    userID,
				"session": resourceID,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":     "web-console",
				"user":    userID,
				"session": resourceID,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       config.ServicePort,
					TargetPort: intstr.FromInt32(config.ContainerPort),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	_, err = c.Clientset.CoreV1().Services(consoleResource.Namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		// Deployment와 Secret 정리
		c.Clientset.AppsV1().Deployments(consoleResource.Namespace).Delete(ctx, consoleResource.DeploymentName, metav1.DeleteOptions{})
		c.Clientset.CoreV1().Secrets(consoleResource.Namespace).Delete(ctx, consoleResource.SecretName, metav1.DeleteOptions{})
		return nil, fmt.Errorf("failed to create Service: %v", err)
	}

	// Deployment가 준비될 때까지 대기
	err = c.WaitForDeploymentReady(consoleResource.DeploymentName, consoleResource.Namespace, 60*time.Second)
	if err != nil {
		log.Printf("Deployment %s not ready after timeout, but continuing: %v", consoleResource.DeploymentName, err)
	}

	// 콘솔 URL 생성 (실제로는 Ingress나 LoadBalancer를 통해 외부 접근 가능한 URL)
	// 웹 콘솔 외부 접근 URL 생성 (B 클러스터의 Ingress를 통해)
	baseURL := os.Getenv("WEB_CONSOLE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://console.basphere.dev"
	}
	consoleResource.ConsoleURL = fmt.Sprintf("%s/%s", baseURL, resourceID)

	return consoleResource, nil
}

// WaitForDeploymentReady Deployment가 준비될 때까지 대기
func (c *Client) WaitForDeploymentReady(deploymentName, namespace string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return wait.PollUntilContextTimeout(ctx, 2*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		deployment, err := c.Clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		// Deployment가 Ready인지 확인
		if deployment.Status.ReadyReplicas > 0 && deployment.Status.ReadyReplicas == deployment.Status.Replicas {
			return true, nil
		}

		return false, nil
	})
}

// WaitForPodReady Pod가 준비될 때까지 대기 (기존 호환성 유지)
func (c *Client) WaitForPodReady(podName, namespace string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return wait.PollUntilContextTimeout(ctx, 2*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		pod, err := c.TargetClientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		// Pod가 Running 상태이고 모든 컨테이너가 Ready인지 확인
		if pod.Status.Phase == corev1.PodRunning {
			for _, condition := range pod.Status.Conditions {
				if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
					return true, nil
				}
			}
		}

		return false, nil
	})
}

// DeleteConsoleResources 웹 콘솔 리소스 삭제
func (c *Client) DeleteConsoleResources(resource *ConsoleResource) error {
	ctx := context.Background()
	deletePolicy := metav1.DeletePropagationForeground

	// Service 삭제
	err := c.Clientset.CoreV1().Services(resource.Namespace).Delete(ctx, resource.ServiceName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		log.Printf("Failed to delete Service %s: %v", resource.ServiceName, err)
	}

	// Deployment 삭제
	err = c.Clientset.AppsV1().Deployments(resource.Namespace).Delete(ctx, resource.DeploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		log.Printf("Failed to delete Deployment %s: %v", resource.DeploymentName, err)
	}

	// Secret 삭제
	err = c.Clientset.CoreV1().Secrets(resource.Namespace).Delete(ctx, resource.SecretName, metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Failed to delete Secret %s: %v", resource.SecretName, err)
	}

	// 참고: PVC는 사용자 히스토리 보존을 위해 삭제하지 않음

	return nil
}

// CleanupExpiredResources 만료된 리소스 정리
func (c *Client) CleanupExpiredResources(namespace string) error {
	ctx := context.Background()

	// 웹 콘솔 관련 리소스들을 라벨로 찾아서 정리
	labelSelector := "app=web-console"

	// 만료된 Deployment 찾기 및 삭제
	deployments, err := c.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %v", err)
	}

	for _, deployment := range deployments.Items {
		// Deployment가 실패했거나 오래된 경우 관련 리소스 정리
		if deployment.Status.ReadyReplicas == 0 && deployment.Status.Replicas > 0 {
			if err := c.cleanupResourcesByLabels(namespace, deployment.Labels); err != nil {
				log.Printf("Failed to cleanup resources for deployment %s: %v", deployment.Name, err)
			}
		}
	}

	return nil
}

// cleanupResourcesByLabels 라벨로 관련 리소스 정리
func (c *Client) cleanupResourcesByLabels(namespace string, labels map[string]string) error {
	ctx := context.Background()

	if sessionID, exists := labels["session"]; exists {
		labelSelector := fmt.Sprintf("session=%s", sessionID)

		// Service 삭제
		services, err := c.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err == nil {
			for _, svc := range services.Items {
				c.Clientset.CoreV1().Services(namespace).Delete(ctx, svc.Name, metav1.DeleteOptions{})
			}
		}

		// Secret 삭제
		secrets, err := c.Clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err == nil {
			for _, secret := range secrets.Items {
				c.Clientset.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
			}
		}

		// Deployment 삭제
		deployments, err := c.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err == nil {
			for _, deployment := range deployments.Items {
				c.Clientset.AppsV1().Deployments(namespace).Delete(ctx, deployment.Name, metav1.DeleteOptions{})
			}
		}
	}

	return nil
}

// parseInt32 문자열을 int32로 변환
func parseInt32(s string) int32 {
	var result int32
	fmt.Sscanf(s, "%d", &result)
	return result
}
