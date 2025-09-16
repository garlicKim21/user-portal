package models

// LaunchConsoleResponse 웹 콘솔 실행 응답
type LaunchConsoleResponse struct {
	URL        string `json:"url"`
	ResourceID string `json:"resource_id"`
}
