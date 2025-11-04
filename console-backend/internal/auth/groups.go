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
	"adm":  1, // 최고 권한
	"dev":  2, // 중간 권한
	"view": 3, // 최저 권한
}

// ParseGroupInfo 그룹 문자열에서 네임스페이스와 권한 키를 파싱합니다.
// 파싱 규칙:
// - 그룹은 '/'로 계층화되고 깊이는 가변
// - 마지막 바로 앞 토큰이 네임스페이스
// - 마지막 토큰은 언더바('_')로 구분되며, 언더바 뒤가 실제 권한 키(adm|dev|view)
// 반환값: (namespace, roleKey, ok)
func ParseGroupInfo(group string) (string, string, bool) {
	if !strings.HasPrefix(group, "/") {
		return "", "", false
	}

	parts := strings.Split(strings.TrimPrefix(group, "/"), "/")
	if len(parts) < 2 {
		return "", "", false
	}

	namespace := parts[len(parts)-2]
	last := parts[len(parts)-1]

	underscoreIdx := strings.LastIndex(last, "_")
	if underscoreIdx == -1 || underscoreIdx == len(last)-1 {
		return "", "", false
	}

	roleKey := last[underscoreIdx+1:]

	// 역할 키가 유효한지 확인
	if _, ok := RolePriority[roleKey]; !ok {
		return "", "", false
	}

	return namespace, roleKey, true
}

// DetermineDefaultNamespace 사용자의 그룹 목록을 기반으로 기본 네임스페이스를 결정합니다.
// 여러 그룹에 속해 있을 경우 가장 높은 우선순위의 네임스페이스를 선택합니다.
// 동일한 우선순위가 여러 개이거나 유효한 그룹이 없으면 "default"를 반환합니다.
func (ug *UserGroups) DetermineDefaultNamespace() string {
	// /dataops/serviceroles/platform/platform_* 패턴의 그룹이 있으면 default 반환
	for _, group := range ug.Groups {
		if strings.HasPrefix(group, "/dataops/serviceroles/platform/") {
			return "default"
		}
	}

	highestPriority := 99
	candidateNamespaces := make([]string, 0)

	for _, group := range ug.Groups {
		namespace, roleKey, ok := ParseGroupInfo(group)
		if !ok {
			continue
		}

		priority := RolePriority[roleKey]

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
