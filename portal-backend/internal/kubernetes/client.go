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

// GenerateKubeconfig OIDC 설정이 포함된 kubeconfig 생성
func GenerateKubeconfig(idToken string) string {
	// 환경변수에서 쿠버네티스 클러스터 정보 가져오기
	clusterServer := os.Getenv("K8S_CLUSTER_SERVER")
	if clusterServer == "" {
		clusterServer = "https://kubernetes.default.svc"
	}

	clusterCA := os.Getenv("K8S_CLUSTER_CA")
	if clusterCA == "" {
		clusterCA = "" // 기본값은 빈 문자열 (insecure-skip-tls-verify 사용)
	}

	oidcIssuerURL := os.Getenv("OIDC_ISSUER_URL")
	oidcClientID := os.Getenv("OIDC_CLIENT_ID")

	// OIDC 설정이 있는 경우 OIDC 기반 kubeconfig 생성
	if oidcIssuerURL != "" && oidcClientID != "" {
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
- name: oidc-context
  context:
    cluster: kubernetes
    user: oidc-user
current-context: oidc-context
users:
- name: oidc-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: kubectl
      args:
      - oidc-login
      - get-token
      - --oidc-issuer-url=%s
      - --oidc-client-id=%s
      - --oidc-extra-scope=profile
      - --oidc-extra-scope=email
      - --token=%s
      env: null
      provideClusterInfo: false
`, clusterConfig, oidcIssuerURL, oidcClientID, idToken)
	}

	// OIDC 설정이 없는 경우 토큰 기반 kubeconfig 생성 (Fallback)
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
`, clusterConfig, idToken)
}
