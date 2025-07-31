package kubernetes

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client 쿠버네티스 클라이언트 래퍼
type Client struct {
	Clientset *kubernetes.Clientset
}

// NewClient 새로운 쿠버네티스 클라이언트 생성
func NewClient() (*Client, error) {
	// 클러스터 내부에서 실행되는 경우
	config, err := rest.InClusterConfig()
	if err != nil {
		// 클러스터 외부에서 실행되는 경우 (개발 환경)
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = os.Getenv("HOME") + "/.kube/config"
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	return &Client{
		Clientset: clientset,
	}, nil
}

// GenerateKubeconfig kubeconfig 생성
func GenerateKubeconfig(idToken string) string {
	// 실제 구현에서는 OIDC 설정이 포함된 kubeconfig를 생성
	// 여기서는 간단한 예시만 제공
	return fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: cluster
  cluster:
    server: https://kubernetes.default.svc
contexts:
- name: context
  context:
    cluster: cluster
    user: user
current-context: context
users:
- name: user
  user:
    token: %s
`, idToken)
} 