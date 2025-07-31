package kubernetes

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

// ConsoleResource 웹 콘솔 리소스 정보
type ConsoleResource struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	PodName       string    `json:"pod_name"`
	ServiceName   string    `json:"service_name"`
	ConfigMapName string    `json:"config_map_name"`
	Namespace     string    `json:"namespace"`
	ConsoleURL    string    `json:"console_url"`
	CreatedAt     time.Time `json:"created_at"`
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
		image = "tsl0922/ttyd:latest"
	}

	containerPort := int32(7681)
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
func (c *Client) CreateConsoleResources(userID, idToken string) (*ConsoleResource, error) {
	config := GetDefaultConfig()
	resourceID := uuid.New().String()[:8] // 8자리 UUID 사용

	resource := &ConsoleResource{
		ID:            resourceID,
		UserID:        userID,
		PodName:       fmt.Sprintf("console-%s-%s", userID, resourceID),
		ServiceName:   fmt.Sprintf("console-svc-%s-%s", userID, resourceID),
		ConfigMapName: fmt.Sprintf("kubeconfig-%s-%s", userID, resourceID),
		Namespace:     config.Namespace,
		CreatedAt:     time.Now(),
	}

	ctx := context.Background()

	// 1. ConfigMap 생성
	kubeconfig := GenerateKubeconfig(idToken)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.ConfigMapName,
			Namespace: resource.Namespace,
			Labels: map[string]string{
				"app":     "web-console",
				"user":    userID,
				"session": resourceID,
			},
		},
		Data: map[string]string{
			"config": kubeconfig,
		},
	}

	_, err := c.Clientset.CoreV1().ConfigMaps(resource.Namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create ConfigMap: %v", err)
	}

	// 2. Pod 생성
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.PodName,
			Namespace: resource.Namespace,
			Labels: map[string]string{
				"app":     "web-console",
				"user":    userID,
				"session": resourceID,
			},
		},
		Spec: corev1.PodSpec{
			ActiveDeadlineSeconds: func(i int32) *int64 { v := int64(i); return &v }(config.TTLSeconds), // 자동 종료 시간
			Containers: []corev1.Container{
				{
					Name:  "web-console",
					Image: config.Image,
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: config.ContainerPort,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "kubeconfig",
							MountPath: "/home/user/.kube",
						},
					},
					Command: []string{"ttyd"},
					Args:    []string{"-p", fmt.Sprintf("%d", config.ContainerPort), "sh"},
					Env: []corev1.EnvVar{
						{
							Name:  "KUBECONFIG",
							Value: "/home/user/.kube/config",
						},
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/",
								Port: intstr.FromInt32(config.ContainerPort),
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "kubeconfig",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: resource.ConfigMapName,
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
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	_, err = c.Clientset.CoreV1().Pods(resource.Namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		// ConfigMap 정리
		c.Clientset.CoreV1().ConfigMaps(resource.Namespace).Delete(ctx, resource.ConfigMapName, metav1.DeleteOptions{})
		return nil, fmt.Errorf("failed to create Pod: %v", err)
	}

	// 3. Service 생성
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.ServiceName,
			Namespace: resource.Namespace,
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

	_, err = c.Clientset.CoreV1().Services(resource.Namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		// Pod와 ConfigMap 정리
		c.Clientset.CoreV1().Pods(resource.Namespace).Delete(ctx, resource.PodName, metav1.DeleteOptions{})
		c.Clientset.CoreV1().ConfigMaps(resource.Namespace).Delete(ctx, resource.ConfigMapName, metav1.DeleteOptions{})
		return nil, fmt.Errorf("failed to create Service: %v", err)
	}

	// Pod가 준비될 때까지 대기
	err = c.WaitForPodReady(resource.PodName, resource.Namespace, 60*time.Second)
	if err != nil {
		log.Printf("Pod %s not ready after timeout, but continuing: %v", resource.PodName, err)
	}

	// 콘솔 URL 생성 (실제로는 Ingress나 LoadBalancer를 통해 외부 접근 가능한 URL)
	// 웹 콘솔 외부 접근 URL 생성 (B 클러스터의 Ingress를 통해)
	baseURL := os.Getenv("WEB_CONSOLE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://console.basphere.dev"
	}
	resource.ConsoleURL = fmt.Sprintf("%s/%s", baseURL, resourceID)

	return resource, nil
}

// WaitForPodReady Pod가 준비될 때까지 대기
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
	err := c.TargetClientset.CoreV1().Services(resource.Namespace).Delete(ctx, resource.ServiceName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		log.Printf("Failed to delete Service %s: %v", resource.ServiceName, err)
	}

	// Pod 삭제
	err = c.Clientset.CoreV1().Pods(resource.Namespace).Delete(ctx, resource.PodName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		log.Printf("Failed to delete Pod %s: %v", resource.PodName, err)
	}

	// ConfigMap 삭제
	err = c.Clientset.CoreV1().ConfigMaps(resource.Namespace).Delete(ctx, resource.ConfigMapName, metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Failed to delete ConfigMap %s: %v", resource.ConfigMapName, err)
	}

	return nil
}

// CleanupExpiredResources 만료된 리소스 정리
func (c *Client) CleanupExpiredResources(namespace string) error {
	ctx := context.Background()

	// 웹 콘솔 관련 리소스들을 라벨로 찾아서 정리
	labelSelector := "app=web-console"

	// 만료된 Pod 찾기 및 삭제
	pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	for _, pod := range pods.Items {
		// Pod가 종료되었거나 오래된 경우 관련 리소스 정리
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			if err := c.cleanupResourcesByLabels(namespace, pod.Labels); err != nil {
				log.Printf("Failed to cleanup resources for pod %s: %v", pod.Name, err)
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

		// ConfigMap 삭제
		configMaps, err := c.Clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err == nil {
			for _, cm := range configMaps.Items {
				c.Clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, cm.Name, metav1.DeleteOptions{})
			}
		}

		// Pod 삭제
		pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err == nil {
			for _, pod := range pods.Items {
				c.Clientset.CoreV1().Pods(namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
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
