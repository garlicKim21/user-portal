package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"portal-backend/internal/models"
)

// SessionStore 서버 사이드 세션 저장소
type SessionStore struct {
	sessions map[string]*models.Session
	mutex    sync.RWMutex
}

// NewSessionStore 새로운 세션 저장소 생성
func NewSessionStore() *SessionStore {
	store := &SessionStore{
		sessions: make(map[string]*models.Session),
	}

	// 만료된 세션 정리 고루틴 시작
	go store.cleanupExpiredSessions()

	return store
}

// CreateSession 새로운 세션 생성
func (s *SessionStore) CreateSession(userID, accessToken, idToken, refreshToken string, expiresAt time.Time) (string, error) {
	// 세션 ID 생성 (32바이트 랜덤)
	sessionID, err := s.generateSessionID()
	if err != nil {
		return "", fmt.Errorf("failed to generate session ID: %v", err)
	}

	session := &models.Session{
		SessionID:    sessionID,
		UserID:       userID,
		AccessToken:  accessToken,
		IDToken:      idToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}

	s.mutex.Lock()
	s.sessions[sessionID] = session
	s.mutex.Unlock()

	return sessionID, nil
}

// GetSession 세션 조회
func (s *SessionStore) GetSession(sessionID string) (*models.Session, error) {
	s.mutex.RLock()
	session, exists := s.sessions[sessionID]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// 만료 확인
	if time.Now().After(session.ExpiresAt) {
		s.DeleteSession(sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// DeleteSession 세션 삭제
func (s *SessionStore) DeleteSession(sessionID string) {
	s.mutex.Lock()
	delete(s.sessions, sessionID)
	s.mutex.Unlock()
}

// generateSessionID 세션 ID 생성
func (s *SessionStore) generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// cleanupExpiredSessions 만료된 세션 정리
func (s *SessionStore) cleanupExpiredSessions() {
	ticker := time.NewTicker(10 * time.Minute) // 10분마다 정리
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for sessionID, session := range s.sessions {
			if now.After(session.ExpiresAt) {
				delete(s.sessions, sessionID)
			}
		}
		s.mutex.Unlock()
	}
}
