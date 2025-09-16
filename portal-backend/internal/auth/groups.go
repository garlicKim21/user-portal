package auth

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// UserGroups 사용자 그룹 정보
type UserGroups struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
	Email    string   `json:"email,omitempty"`
}

// ExtractUserGroups ID 토큰에서 사용자 그룹 정보 추출
func ExtractUserGroups(idToken string) (*UserGroups, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userGroups := &UserGroups{}

	// 사용자 ID 추출 (sub 클레임)
	if sub, ok := claims["sub"].(string); ok {
		userGroups.UserID = sub
	}

	// 사용자명 추출 (preferred_username 클레임)
	if username, ok := claims["preferred_username"].(string); ok {
		userGroups.Username = username
	}

	// 이메일 추출 (email 클레임)
	if email, ok := claims["email"].(string); ok {
		userGroups.Email = email
	}

	// 그룹 정보 추출 (groups 클레임)
	if groupsInterface, ok := claims["groups"]; ok {
		switch groups := groupsInterface.(type) {
		case []interface{}:
			// JSON 배열 형태
			for _, group := range groups {
				if groupStr, ok := group.(string); ok {
					userGroups.Groups = append(userGroups.Groups, groupStr)
				}
			}
		case []string:
			// 문자열 배열 형태
			userGroups.Groups = groups
		case string:
			// 단일 문자열 또는 쉼표로 구분된 문자열
			if strings.Contains(groups, ",") {
				userGroups.Groups = strings.Split(groups, ",")
				// 공백 제거
				for i, group := range userGroups.Groups {
					userGroups.Groups[i] = strings.TrimSpace(group)
				}
			} else {
				userGroups.Groups = []string{groups}
			}
		}
	}

	return userGroups, nil
}

// GetUserRole 사용자의 최고 권한 역할 반환
// 우선순위: adm > dev > view > user
func (ug *UserGroups) GetUserRole() string {
	for _, group := range ug.Groups {
		switch strings.ToLower(group) {
		case "admins", "admin", "adm":
			return "adm"
		}
	}

	for _, group := range ug.Groups {
		switch strings.ToLower(group) {
		case "developers", "developer", "dev":
			return "dev"
		}
	}

	for _, group := range ug.Groups {
		switch strings.ToLower(group) {
		case "viewers", "viewer", "view":
			return "view"
		}
	}
	return "user" // 기본 역할
}

// HasRole 특정 역할 권한이 있는지 확인
func (ug *UserGroups) HasRole(role string) bool {
	userRole := ug.GetUserRole()

	switch strings.ToLower(role) {
	case "adm":
		return userRole == "adm"
	case "dev":
		return userRole == "adm" || userRole == "dev"
	case "viewer":
		return userRole == "adm" || userRole == "dev" || userRole == "view"
	default:
		return true // 기본 사용자 권한
	}
}

// ToJSON 그룹 정보를 JSON 문자열로 변환
func (ug *UserGroups) ToJSON() string {
	data, err := json.Marshal(ug)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// RolePriority 역할 우선순위를 정의 (낮을수록 높음)
var RolePriority = map[string]int{
	"adm" : 1,
	"dev" : 2,
	"view": 3,
}

// DetermineDefaultNamespace 사용자의 그룹 목록을 기반으로 기본 네임스페이스를 결정합니다.
// 그룹 형태: "/cluster/namespace/role"
func (ug *UserGroups) DetermineDefaultNamespace() string {
	// cluster-admins 그룹이 있으면 무조건 default 네임스페이스를 반환합니다.
	if slices.Contains(ug.Groups, "platform-adm") {
		return "default"
	}

	highestPriority := 99
	candidateNamespaces := make([]string, 0)

	// 그룹들을 순회하며 "/cluster/namespace/role" 형태에서 namespace와 role을 추출합니다.
	for _, group := range ug.Groups {
		// "/cluster/namespace/role" 형태 파싱
		parts := strings.Split(strings.TrimPrefix(group, "/"), "/")
		if len(parts) != 3 {
			continue
		}

		cluster := parts[0]
		namespace := parts[1]
		role := parts[2]

		// cluster가 비어있거나 namespace가 비어있으면 건너뛰기
		if cluster == "" || namespace == "" || role == "" {
			continue
		}

		priority, ok := RolePriority[role]
		if !ok {
			continue
		}

		if priority < highestPriority {
			highestPriority = priority
			candidateNamespaces = []string{namespace}
			continue
		}

		if priority == highestPriority {
			if !slices.Contains(candidateNamespaces, namespace) {
				candidateNamespaces = append(candidateNamespaces, namespace)
			}
		}
	}

	if len(candidateNamespaces) == 1 {
		return candidateNamespaces[0]
	}

	// 후보가 없거나 동률(2개 이상)이면 default 반환
	return "default"
}
