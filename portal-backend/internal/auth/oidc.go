package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"portal-backend/internal/config"
)

// OIDCConfig OIDC 설정
type OIDCConfig struct {
	ClientID     string
	ClientSecret string
	IssuerURL    string
	RedirectURL  string
}

// OIDCProvider OIDC 제공자
type OIDCProvider struct {
	config       *OIDCConfig
	oauth2Config *oauth2.Config
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
}

// NewOIDCProvider 새로운 OIDC 제공자 생성
func NewOIDCProvider() (*OIDCProvider, error) {
	cfg := config.Get()
	oidcConfig := &OIDCConfig{
		ClientID:     cfg.OIDC.ClientID,
		ClientSecret: cfg.OIDC.ClientSecret,
		IssuerURL:    cfg.OIDC.IssuerURL,
		RedirectURL:  cfg.OIDC.RedirectURL,
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, oidcConfig.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("OIDC provider 생성 실패: %v", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     oidcConfig.ClientID,
		ClientSecret: oidcConfig.ClientSecret,
		RedirectURL:  oidcConfig.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "offline_access"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: oidcConfig.ClientID})

	return &OIDCProvider{
		config:       oidcConfig,
		oauth2Config: oauth2Config,
		provider:     provider,
		verifier:     verifier,
	}, nil
}

// GetAuthURL 인증 URL 생성
func (p *OIDCProvider) GetAuthURL(state string) string {
	return p.oauth2Config.AuthCodeURL(state)
}

// ExchangeCode 토큰 교환
func (p *OIDCProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.oauth2Config.Exchange(ctx, code)
}

// VerifyIDToken ID 토큰 검증
func (p *OIDCProvider) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return p.verifier.Verify(ctx, rawIDToken)
}

// GenerateRandomString 랜덤 문자열 생성
func GenerateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random string: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// TokenClaims 토큰 클레임 구조체
type TokenClaims struct {
	Subject           string `json:"sub"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
}

// TokenExchangeResponse Token Exchange 응답 구조체
type TokenExchangeResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	RefreshExpiresIn int    `json:"refresh_expires_in,omitempty"`
	TokenType        string `json:"token_type"`
	IssuedTokenType  string `json:"issued_token_type,omitempty"`
	Scope            string `json:"scope,omitempty"`
}

// ExchangeTokenForKubernetes portal-app 토큰을 kubernetes 클라이언트용 토큰으로 교환
func ExchangeTokenForKubernetes(subjectToken string) (*TokenExchangeResponse, error) {
	cfg := config.Get()
	clientID := cfg.OIDC.ClientID
	clientSecret := cfg.OIDC.ClientSecret
	targetAudience := cfg.OIDC.KubernetesClientID
	tokenEndpoint := cfg.OIDC.IssuerURL + "/protocol/openid-connect/token"

	if targetAudience == "" {
		return nil, fmt.Errorf("KUBERNETES_CLIENT_ID environment variable is required for token exchange")
	}

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	data.Set("subject_token", subjectToken)
	data.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
	data.Set("requested_token_type", "urn:ietf:params:oauth:token-type:access_token")
	data.Set("audience", targetAudience)

	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token exchange request: %v", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform token exchange: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token exchange response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenExchangeResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token exchange response: %v", err)
	}

	return &tokenResp, nil
}
