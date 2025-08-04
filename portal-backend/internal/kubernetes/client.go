package kubernetes

import (
	"encoding/base64"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"portal-backend/internal/config"
)

// Client 쿠버네티스 클라이언트 래퍼 (다중 클러스터 지원)
type Client struct {
	Clientset       *kubernetes.Clientset // 로컬 클러스터 (A)
	TargetClientset *kubernetes.Clientset // 타겟 클러스터 (B) - 웹 콘솔 생성용
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

	// 타겟 클러스터 (B 클러스터) 설정 - 웹 콘솔에서 제어할 클러스터
	targetClientset := clientset // 기본값은 동일한 클러스터 사용

	// 별도의 타겟 클러스터 설정이 있는 경우
	targetConfig, err := createTargetClusterConfig()
	if err == nil {
		targetClientset, err = kubernetes.NewForConfig(targetConfig)
		if err != nil {
			// 타겟 클러스터 설정 실패 시 로컬 클러스터 사용
			targetClientset = clientset
		}
	}

	return &Client{
		Clientset:       clientset,       // A 클러스터 (로컬)
		TargetClientset: targetClientset, // B 클러스터 (타겟)
	}, nil
}

// createTargetClusterConfig 타겟 클러스터 설정 생성
func createTargetClusterConfig() (*rest.Config, error) {
	cfg := config.Get()

	// 타겟 클러스터 설정이 있는 경우에만 연결 시도
	if cfg.Kubernetes.TargetServer != "" {
		config := &rest.Config{
			Host: cfg.Kubernetes.TargetServer,
		}

		// CA 인증서 설정 (base64 인코딩된 데이터 사용)
		if cfg.Kubernetes.TargetCAData != "" {
			config.TLSClientConfig = rest.TLSClientConfig{
				CAData: []byte(cfg.Kubernetes.TargetCAData),
			}
		} else {
			// CA 검증 비활성화 (개발/테스트 환경용)
			config.TLSClientConfig = rest.TLSClientConfig{
				Insecure: true,
			}
		}

		return config, nil
	}

	return nil, fmt.Errorf("TARGET_CLUSTER_SERVER environment variable is required for multi-cluster setup")
}

// GenerateKubeconfig B 클러스터 접근을 위한 Bearer Token 설정이 포함된 kubeconfig 생성
func GenerateKubeconfig(accessToken string) string {
	cfg := config.Get()

	// B 클러스터 정보 가져오기 (웹 터미널에서 제어할 타겟 클러스터)
	clusterServer := cfg.Kubernetes.TargetServer
	if clusterServer == "" {
		// 개발 환경에서는 동일한 클러스터 사용
		clusterServer = "https://kubernetes.default.svc"
	}

	clusterCAData := cfg.Kubernetes.TargetCAData
	if clusterCAData == "" {
		clusterCAData = "" // 기본값은 빈 문자열 (insecure-skip-tls-verify 사용)
	}

	// Bearer Token 기반 kubeconfig 생성
	var clusterConfig string
	if clusterCAData != "" {
		clusterConfig = fmt.Sprintf(`    server: %s
    certificate-authority-data: %s`, clusterServer, clusterCAData)
	} else {
		clusterConfig = fmt.Sprintf(`    server: %s
    insecure-skip-tls-verify: true`, clusterServer)
	}

	return fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: kubernetes
  cluster:
%s
contexts:
- name: token-context
  context:
    cluster: kubernetes
    user: token-user
current-context: token-context
users:
- name: token-user
  user:
    token: %s
`, clusterConfig, accessToken)
}

// EncodeCACertToBase64 CA 인증서 파일을 base64로 인코딩
func EncodeCACertToBase64(caCertPath string) (string, error) {
	data, err := os.ReadFile(caCertPath)
	if err != nil {
		return "", fmt.Errorf("failed to read CA certificate file: %v", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// DecodeBase64ToCAData base64로 인코딩된 CA 인증서를 디코딩
func DecodeBase64ToCAData(base64Data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64Data)
}
