package kubernetes

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

// GenerateKubeconfig B 클러스터 접근을 위한 Bearer Token 설정이 포함된 kubeconfig 생성
func GenerateKubeconfig(accessToken string) string {
	// B 클러스터 정보 가져오기 (웹 터미널에서 제어할 타겟 클러스터)
	clusterServer := os.Getenv("TARGET_CLUSTER_SERVER")
	if clusterServer == "" {
		// 개발 환경에서는 동일한 클러스터 사용
		clusterServer = os.Getenv("K8S_CLUSTER_SERVER")
		if clusterServer == "" {
			clusterServer = "https://kubernetes.default.svc"
		}
	}

	clusterCA := os.Getenv("TARGET_CLUSTER_CA")
	if clusterCA == "" {
		// Fallback to general cluster CA
		clusterCA = os.Getenv("K8S_CLUSTER_CA")
		if clusterCA == "" {
			clusterCA = "" // 기본값은 빈 문자열 (insecure-skip-tls-verify 사용)
		}
	}

	// Bearer Token 기반 kubeconfig 생성
	var clusterConfig string
	if clusterCA != "" {
		clusterConfig = fmt.Sprintf(`    server: %s
    certificate-authority-data: %s`, clusterServer, clusterCA)
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
