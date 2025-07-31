package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
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
	config := &OIDCConfig{
		ClientID:     os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		IssuerURL:    os.Getenv("OIDC_ISSUER_URL"),
		RedirectURL:  os.Getenv("OIDC_REDIRECT_URL"),
	}

	// 필수 환경 변수 검증
	if config.ClientID == "" || config.ClientSecret == "" || config.IssuerURL == "" || config.RedirectURL == "" {
		return nil, fmt.Errorf("OIDC 환경 변수가 설정되지 않았습니다")
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("OIDC provider 생성 실패: %v", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: config.ClientID})

	return &OIDCProvider{
		config:       config,
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
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// TokenClaims 토큰 클레임 구조체
type TokenClaims struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
}
