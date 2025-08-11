package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT 클레임 구조체
type JWTClaims struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
	jwt.RegisteredClaims
}

// JWTManager JWT 토큰 관리자
type JWTManager struct {
	secretKey []byte
}

// NewJWTManager 새로운 JWT 관리자 생성
func NewJWTManager() (*JWTManager, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("JWT_SECRET_KEY environment variable is required")
	}

	return &JWTManager{
		secretKey: []byte(secretKey),
	}, nil
}

// GenerateToken JWT 토큰 생성
func (j *JWTManager) GenerateToken(userID, sessionID string, expiresAt time.Time) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		SessionID: sessionID,
		ExpiresAt: expiresAt,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // JWT 자체는 24시간 유효
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "portal-backend",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken JWT 토큰 검증
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// OIDC 토큰 만료 시간 확인
	if time.Now().After(claims.ExpiresAt) {
		return nil, fmt.Errorf("OIDC token expired")
	}

	return claims, nil
}
