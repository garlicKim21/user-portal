package models

import "time"

// Session 사용자 세션 정보
type Session struct {
	AccessToken  string    `json:"access_token"`
	IDToken      string    `json:"id_token"`
	RefreshToken string    `json:"refresh_token"`
	UserID       string    `json:"user_id"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// LaunchConsoleResponse 웹 콘솔 실행 응답
type LaunchConsoleResponse struct {
	URL string `json:"url"`
}

// UserInfo 사용자 정보
type UserInfo struct {
	UserID   string `json:"user_id"`
	LoggedIn bool   `json:"logged_in"`
}
