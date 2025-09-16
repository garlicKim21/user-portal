package kubernetes

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"text/template"

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
	cfg := config.Get()
	
	// 클러스터 내부에서 실행되는 경우
	config, err := rest.InClusterConfig()
	if err != nil {
		// 클러스터 외부에서 실행되는 경우 (개발 환경)
		kubeconfig := cfg.Kubernetes.Kubeconfig
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

// 템플릿 기반 kubeconfig
const KubeconfigTemplate = `
apiVersion: v1
kind: Config
clusters:
- name: kubernetes
  cluster:
    server: {{.ClusterServer}}
{{- if .ClusterCAData }}
    certificate-authority-data: {{.ClusterCAData}}
{{- else }}
    insecure-skip-tls-verify: true
{{- end }}
contexts:
- name: default-context
  context:
    cluster: kubernetes
    user: user
    namespace: {{.DefaultNamespace}}
current-context: default-context
users:
- name: user
  user: {}
`

type kubeconfigParams struct {
	ClusterServer    string
	ClusterCAData    string
	DefaultNamespace string
}

// GenerateUserKubeconfig 사용자의 기본 네임스페이스가 포함된 kubeconfig 생성
func GenerateUserKubeconfig(defaultNamespace string) (string, error) {
	cfg := config.Get()

	params := kubeconfigParams{
		ClusterServer:    cfg.Kubernetes.TargetServer,
		ClusterCAData:    cfg.Kubernetes.TargetCAData,
		DefaultNamespace: defaultNamespace,
	}

	if params.ClusterServer == "" {
		params.ClusterServer = "https://kubernetes.default.svc"
	}

	tmpl, err := template.New("kubeconfig").Parse(KubeconfigTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse kubeconfig template: %w", err)
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, params); err != nil {
		return "", fmt.Errorf("failed to execute kubeconfig template: %w", err)
	}

	return out.String(), nil
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
