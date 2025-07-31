package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// MultiClusterClient 다중 클러스터 관리를 위한 클라이언트
type MultiClusterClient struct {
	// 현재 클러스터 (A 클러스터 - 포털이 실행되는 곳)
	LocalClientset *kubernetes.Clientset
	LocalConfig    *rest.Config

	// 타겟 클러스터 (B 클러스터 - 웹 콘솔이 생성될 곳)
	TargetClientset *kubernetes.Clientset
	TargetConfig    *rest.Config
}

// NewMultiClusterClient 다중 클러스터 클라이언트 생성
func NewMultiClusterClient() (*MultiClusterClient, error) {
	client := &MultiClusterClient{}

	// 로컬 클러스터 (A 클러스터) 설정 - in-cluster config 사용
	localConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config for local cluster: %v", err)
	}

	localClientset, err := kubernetes.NewForConfig(localConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create local cluster clientset: %v", err)
	}

	client.LocalConfig = localConfig
	client.LocalClientset = localClientset

	// 타겟 클러스터 (B 클러스터) 설정
	targetConfig, err := createTargetClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create target cluster config: %v", err)
	}

	targetClientset, err := kubernetes.NewForConfig(targetConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create target cluster clientset: %v", err)
	}

	client.TargetConfig = targetConfig
	client.TargetClientset = targetClientset

	return client, nil
}

// createTargetClusterConfig 타겟 클러스터 설정 생성
func createTargetClusterConfig() (*rest.Config, error) {
	// 환경 변수에서 타겟 클러스터 정보 읽기
	targetServer := os.Getenv("TARGET_CLUSTER_SERVER")
	targetToken := os.Getenv("TARGET_CLUSTER_TOKEN")
	targetCAPath := os.Getenv("TARGET_CLUSTER_CA_CERT_PATH")
	targetKubeconfigPath := os.Getenv("TARGET_KUBECONFIG_PATH")

	// Method 1: kubeconfig 파일 사용
	if targetKubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", targetKubeconfigPath)
	}

	// Method 2: 환경 변수로 직접 설정
	if targetServer != "" && targetToken != "" {
		config := &rest.Config{
			Host:        targetServer,
			BearerToken: targetToken,
		}

		// CA 인증서 설정
		if targetCAPath != "" {
			config.TLSClientConfig = rest.TLSClientConfig{
				CAFile: targetCAPath,
			}
		} else {
			// CA 검증 비활성화 (개발/테스트 환경용)
			config.TLSClientConfig = rest.TLSClientConfig{
				Insecure: true,
			}
		}

		return config, nil
	}

	// Method 3: 기본 kubeconfig 파일 시도 (개발 환경용)
	if home := homedir.HomeDir(); home != "" {
		kubeconfigPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeconfigPath); err == nil {
			return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		}
	}

	return nil, fmt.Errorf("no valid target cluster configuration found. Please set TARGET_CLUSTER_SERVER and TARGET_CLUSTER_TOKEN environment variables, or provide TARGET_KUBECONFIG_PATH")
}

// GetLocalClient 로컬 클러스터 클라이언트 반환 (A 클러스터)
func (mc *MultiClusterClient) GetLocalClient() *kubernetes.Clientset {
	return mc.LocalClientset
}

// GetTargetClient 타겟 클러스터 클라이언트 반환 (B 클러스터)
func (mc *MultiClusterClient) GetTargetClient() *kubernetes.Clientset {
	return mc.TargetClientset
}

// TestTargetClusterConnection 타겟 클러스터 연결 테스트
func (mc *MultiClusterClient) TestTargetClusterConnection() error {
	// 타겟 클러스터에서 네임스페이스 목록 조회로 연결 테스트
	_, err := mc.TargetClientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("failed to connect to target cluster: %v", err)
	}
	return nil
}
