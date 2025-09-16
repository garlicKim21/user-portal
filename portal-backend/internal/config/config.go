package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config 애플리케이션 전체 설정
type Config struct {
	Server     ServerConfig     `json:"server"`
	OIDC       OIDCConfig       `json:"oidc"`
	JWT        JWTConfig        `json:"jwt"`
	Kubernetes KubernetesConfig `json:"kubernetes"`
	Console    ConsoleConfig    `json:"console"`
	Logging    LoggingConfig    `json:"logging"`
}

// ServerConfig 서버 관련 설정
type ServerConfig struct {
	Port           string   `json:"port"`
	GinMode        string   `json:"gin_mode"`
	AllowedOrigins []string `json:"allowed_origins"`
}

// OIDCConfig OIDC 관련 설정
type OIDCConfig struct {
	ClientID           string `json:"client_id"`
	ClientSecret       string `json:"client_secret"`
	IssuerURL          string `json:"issuer_url"`
	RedirectURL        string `json:"redirect_url"`
	KubernetesClientID string `json:"kubernetes_client_id"`
}

// JWTConfig JWT 관련 설정
type JWTConfig struct {
	SecretKey string `json:"secret_key"`
}

// KubernetesConfig 쿠버네티스 관련 설정
type KubernetesConfig struct {
	// 로컬 클러스터 설정 (A 클러스터 - 포털이 실행되는 곳)
	Kubeconfig string `json:"kubeconfig"` // 개발 환경에서만 사용 (프로덕션에서는 InClusterConfig)

	// 타겟 클러스터 설정 (B 클러스터 - 웹 콘솔에서 제어할 클러스터)
	TargetServer string `json:"target_server"`  // B 클러스터 API 서버 URL (필수)
	TargetCAData string `json:"target_ca_data"` // B 클러스터 CA 인증서 (base64 인코딩)
}

// ConsoleConfig 웹 콘솔 관련 설정
type ConsoleConfig struct {
	Namespace     string `json:"namespace"`
	Image         string `json:"image"`
	ContainerPort int    `json:"container_port"`
	ServicePort   int    `json:"service_port"`
	TTLSeconds    int    `json:"ttl_seconds"`
	BaseURL       string `json:"base_url"`
}

// LoggingConfig 로깅 관련 설정
type LoggingConfig struct {
	Level string `json:"level"`
}

var globalConfig *Config

// Load 환경 변수에서 설정 로드
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:           getEnvWithDefault("PORT", "8080"),
			GinMode:        getEnvWithDefault("GIN_MODE", "release"),
			AllowedOrigins: parseStringSlice(getEnvWithDefault("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000,http://localhost:8080")),
		},
		OIDC: OIDCConfig{
			ClientID:           getEnvWithDefault("OIDC_CLIENT_ID", ""),
			ClientSecret:       getEnvWithDefault("OIDC_CLIENT_SECRET", ""),
			IssuerURL:          getEnvWithDefault("OIDC_ISSUER_URL", ""),
			RedirectURL:        getEnvWithDefault("OIDC_REDIRECT_URL", ""),
			KubernetesClientID: getEnvWithDefault("KUBERNETES_CLIENT_ID", ""),
		},
		JWT: JWTConfig{
			SecretKey: getEnvWithDefault("JWT_SECRET_KEY", ""),
		},
		Kubernetes: KubernetesConfig{
			Kubeconfig:   getEnvWithDefault("KUBECONFIG", ""),
			TargetServer: getEnvWithDefault("TARGET_CLUSTER_SERVER", ""),
			TargetCAData: getEnvWithDefault("TARGET_CLUSTER_CA_CERT_DATA", ""),
		},
		Console: ConsoleConfig{
			Namespace:     getEnvWithDefault("CONSOLE_NAMESPACE", "default"),
			Image:         getEnvWithDefault("CONSOLE_IMAGE", "projectgreenist/web-terminal:0.2.3"),
			ContainerPort: getEnvAsIntWithDefault("CONSOLE_CONTAINER_PORT", 8080),
			ServicePort:   getEnvAsIntWithDefault("CONSOLE_SERVICE_PORT", 80),
			TTLSeconds:    getEnvAsIntWithDefault("CONSOLE_TTL_SECONDS", 3600),
			BaseURL:       getEnvWithDefault("WEB_CONSOLE_BASE_URL", "console.basphere.dev"),
		},
		Logging: LoggingConfig{
			Level: strings.ToUpper(getEnvWithDefault("LOG_LEVEL", "INFO")),
		},
	}

	// 필수 환경 변수 검증
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	globalConfig = config
	return config, nil
}

// Get 전역 설정 반환
func Get() *Config {
	if globalConfig == nil {
		panic("config not loaded. Call config.Load() first")
	}
	return globalConfig
}

// validateConfig 필수 설정 검증
func validateConfig(config *Config) error {
	required := map[string]string{
		"OIDC_CLIENT_ID":     config.OIDC.ClientID,
		"OIDC_CLIENT_SECRET": config.OIDC.ClientSecret,
		"OIDC_ISSUER_URL":    config.OIDC.IssuerURL,
		"OIDC_REDIRECT_URL":  config.OIDC.RedirectURL,
		"JWT_SECRET_KEY":     config.JWT.SecretKey,
	}

	var missing []string
	for key, value := range required {
		if value == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// 헬퍼 함수들
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func parseStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts
}
