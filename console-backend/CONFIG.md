# 환경 변수 설정 가이드

이 문서는 Portal Backend 애플리케이션의 환경 변수 설정 방법을 설명합니다.

## 설정 구조

애플리케이션은 `internal/config` 패키지를 통해 중앙 집중식 설정 관리를 사용합니다.

### 주요 설정 그룹

#### 1. 서버 설정 (Server Config)
```bash
PORT=8080                    # 서버 포트 (기본값: 8080)
GIN_MODE=release            # Gin 모드 (debug/release, 기본값: release)
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000  # CORS 허용 오리진
```

#### 2. OIDC 설정 (OIDC Config)
```bash
OIDC_CLIENT_ID=portal-backend                               # OIDC 클라이언트 ID (필수)
OIDC_CLIENT_SECRET=your-client-secret                      # OIDC 클라이언트 시크릿 (필수)
OIDC_ISSUER_URL=http://localhost:8080/realms/portal        # OIDC 발급자 URL (필수)
OIDC_REDIRECT_URL=http://localhost:8080/api/callback       # OIDC 리다이렉트 URL (필수)
KUBERNETES_CLIENT_ID=kubernetes-client                     # Kubernetes 토큰 교환용 클라이언트 ID
```

#### 3. JWT 설정 (JWT Config)
```bash
JWT_SECRET_KEY=your-super-secure-secret-key                # JWT 서명 키 (필수, 최소 32자)
```

**🆕 JWT 구조 최적화 (v0.4.10+)**
- **토큰 중첩 제거**: OIDC 토큰을 JWT에 포함하지 않음
- **하이브리드 인증**: JWT + Session 기반으로 보안 강화
- **성능 향상**: JWT 크기 95% 감소, 파싱 속도 3-5배 향상
- **보안 강화**: 민감한 토큰 정보를 클라이언트에 노출하지 않음

#### 4. 쿠버네티스 설정 (Kubernetes Config)
```bash
# 로컬 클러스터 (개발 환경용)
KUBECONFIG=~/.kube/config                                  # Kubeconfig 파일 경로 (개발 환경에서만 사용)

# 타겟 클러스터 (다중 클러스터 환경)
TARGET_CLUSTER_SERVER=https://target-cluster:6443         # 타겟 클러스터 서버
TARGET_CLUSTER_TOKEN=eyJhbGciOiJSUzI1NiIs...             # 타겟 클러스터 토큰
TARGET_CLUSTER_CA_CERT_DATA=LS0tLS1CRUdJTi...            # 타겟 클러스터 CA 인증서 (base64 인코딩)
```

#### 5. 웹 콘솔 설정 (Console Config)
```bash
CONSOLE_NAMESPACE=web-console                              # 콘솔 네임스페이스 (기본값: web-console)
CONSOLE_IMAGE=projectgreenist/web-terminal:0.2.11         # 콘솔 이미지 (최신 버전)
CONSOLE_CONTAINER_PORT=8080                               # 컨테이너 포트 (기본값: 8080)
CONSOLE_SERVICE_PORT=80                                   # 서비스 포트 (기본값: 80)
CONSOLE_TTL_SECONDS=3600                                  # TTL 초 (기본값: 3600)
WEB_CONSOLE_BASE_URL=https://console.example.com         # 웹 콘솔 베이스 URL
INGRESS_CLASS=cilium                                     # Ingress Controller 클래스명 (기본값: cilium)
```

**🆕 웹 콘솔 개인화 기능 (v0.2.11+)**
- **동적 사용자 정보**: 실제 로그인 ID, 네임스페이스, 권한 표시
- **맞춤형 프롬프트**: `user@secure-terminal-{username}:~$` 형태
- **권한 정보 표시**: `blue-admin/red-developer` 형태로 역할 표시
- **명령어 히스토리**: PVC 기반 사용자별 히스토리 지속성 (100Mi)
- **CSRF 보호**: State 기반 보안 강화

#### 6. 로깅 설정 (Logging Config)
```bash
LOG_LEVEL=INFO                                            # 로그 레벨 (DEBUG/INFO/WARN/ERROR/FATAL, 기본값: INFO)
```

## 설정 파일 사용법

### 1. 환경 변수 파일 생성
```bash
cp env.example .env
```

### 2. 필수 환경 변수 설정
다음 환경 변수들은 반드시 설정해야 합니다:
- `OIDC_CLIENT_ID`
- `OIDC_CLIENT_SECRET`
- `OIDC_ISSUER_URL`
- `OIDC_REDIRECT_URL`
- `JWT_SECRET_KEY`

### 3. 애플리케이션에서 설정 사용
```go
import "portal-backend/internal/config"

// 애플리케이션 시작 시 설정 로드
cfg, err := config.Load()
if err != nil {
    log.Fatal("Failed to load configuration:", err)
}

// 설정 사용
port := cfg.Server.Port
oidcClientID := cfg.OIDC.ClientID
```

## 환경별 설정 예시

### 개발 환경
```bash
PORT=8080
GIN_MODE=debug
LOG_LEVEL=DEBUG
OIDC_ISSUER_URL=http://localhost:8080/realms/portal
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
CONSOLE_NAMESPACE=web-console
CONSOLE_IMAGE=projectgreenist/web-terminal:0.2.11
```

### 프로덕션 환경
```bash
PORT=8080
GIN_MODE=release
LOG_LEVEL=INFO
OIDC_ISSUER_URL=https://auth.yourdomain.com/realms/portal
ALLOWED_ORIGINS=https://portal.yourdomain.com
JWT_SECRET_KEY=your-production-secret-key
CONSOLE_NAMESPACE=web-console
CONSOLE_IMAGE=projectgreenist/web-terminal:0.2.11
CONSOLE_TTL_SECONDS=7200
```

## 설정 검증

애플리케이션 시작 시 필수 환경 변수가 누락된 경우 오류가 발생합니다:
```
config validation failed: missing required environment variables: OIDC_CLIENT_ID, JWT_SECRET_KEY
```

## 보안 고려사항

1. **JWT_SECRET_KEY**: 강력한 랜덤 키 사용 (최소 32자)
2. **OIDC_CLIENT_SECRET**: Keycloak에서 생성된 시크릿 사용
3. **환경 변수 파일**: `.env` 파일을 Git에 커밋하지 않도록 주의
4. **프로덕션 환경**: 환경 변수를 컨테이너 오케스트레이션 도구의 시크릿으로 관리

**🆕 추가 보안 기능 (v0.4.10+)**
5. **CSRF 보호**: State 기반 CSRF 공격 방지
6. **세션 격리**: 사용자별 완전한 세션 격리
7. **토큰 보안**: 민감한 OIDC 토큰을 클라이언트에 노출하지 않음
8. **웹 콘솔 격리**: 사용자별 완전한 웹 콘솔 환경 격리
9. **명령어 히스토리 보안**: PVC 기반 안전한 히스토리 저장

## Pod 환경에서의 설정

### CA 인증서 처리
Pod 환경에서는 파일 시스템에 직접 접근할 수 없으므로, CA 인증서를 base64로 인코딩하여 환경 변수로 전달합니다:

```bash
# 방법 1: 명령어로 직접 인코딩
cat /path/to/ca.crt | base64 -w 0

# 방법 2: Go 코드로 인코딩 (개발 시)
go run -c 'package main; import ("encoding/base64"; "fmt"; "os"); func main() { data, _ := os.ReadFile("ca.crt"); fmt.Println(base64.StdEncoding.EncodeToString(data)) }'

# 환경 변수로 설정
TARGET_CLUSTER_CA_CERT_DATA=LS0tLS1CRUdJTi...
```

**주의사항:**
- CA 인증서는 민감한 정보이므로 Secret으로 관리
- base64 인코딩 시 줄바꿈 문자 제거 (`-w 0` 옵션 사용)
- 인코딩된 값은 kubeconfig의 `certificate-authority-data` 필드와 동일한 형식

### Secret을 통한 설정 관리
프로덕션 환경에서는 민감한 정보를 Secret으로 관리:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: portal-backend-secrets
type: Opaque
data:
  TARGET_CLUSTER_TOKEN: <base64-encoded-token>
  TARGET_CLUSTER_CA_CERT_DATA: <base64-encoded-ca-cert>
  JWT_SECRET_KEY: <base64-encoded-jwt-secret>
  OIDC_CLIENT_SECRET: <base64-encoded-oidc-secret>
```

**Secret 생성 예시:**
```bash
# CA 인증서를 base64로 인코딩
CA_CERT_B64=$(cat /path/to/ca.crt | base64 -w 0)

# Secret 생성
kubectl create secret generic portal-backend-secrets \
  --from-literal=TARGET_CLUSTER_TOKEN="your-token" \
  --from-literal=TARGET_CLUSTER_CA_CERT_DATA="$CA_CERT_B64" \
  --from-literal=JWT_SECRET_KEY="your-jwt-secret" \
  --from-literal=OIDC_CLIENT_SECRET="your-oidc-secret"
```

## 문제 해결

### 설정 로드 실패
- 필수 환경 변수가 설정되었는지 확인
- 환경 변수 파일 경로가 올바른지 확인

### OIDC 연결 실패
- `OIDC_ISSUER_URL`이 접근 가능한지 확인
- 클라이언트 ID와 시크릿이 올바른지 확인

### 쿠버네티스 연결 실패
- `KUBECONFIG` 파일이 존재하고 유효한지 확인
- 클러스터 접근 권한이 있는지 확인

### 웹 콘솔 생성 실패
- `CONSOLE_NAMESPACE`가 존재하는지 확인
- `CONSOLE_IMAGE`가 올바른지 확인
- PVC 생성 권한이 있는지 확인
- StorageClass `local-path`가 사용 가능한지 확인

### 명령어 히스토리 문제
- PVC가 올바르게 생성되었는지 확인
- Pod에 히스토리 볼륨이 마운트되었는지 확인
- 환경 변수 `USER_ID`, `DEFAULT_NAMESPACE`, `USER_ROLES`가 설정되었는지 확인

### 웹 콘솔 개인화 문제
- 사용자 그룹 정보가 올바르게 추출되는지 확인
- OIDC ID 토큰에 `groups` 클레임이 포함되어 있는지 확인
- 네임스페이스별 역할 정보가 올바른 형식인지 확인

## 🆕 최신 기능 요약

### v0.4.10+ (Backend)
- **JWT 구조 최적화**: 토큰 중첩 제거, 하이브리드 인증
- **보안 강화**: CSRF 보호, 세션 격리, 토큰 보안
- **성능 향상**: JWT 크기 95% 감소, 파싱 속도 3-5배 향상

### v0.2.11+ (Web Terminal)
- **개인화된 터미널**: 사용자별 맞춤형 정보 표시
- **명령어 히스토리**: PVC 기반 지속성 보장
- **권한 정보 표시**: 네임스페이스별 역할 정보
- **보안 강화**: State 기반 CSRF 방지

### 권장 설정
```bash
# 최신 웹 콘솔 이미지 사용
CONSOLE_IMAGE=projectgreenist/web-terminal:0.2.11

# 웹 콘솔 전용 네임스페이스
CONSOLE_NAMESPACE=web-console

# 적절한 TTL 설정
CONSOLE_TTL_SECONDS=7200

# 보안을 위한 로그 레벨
LOG_LEVEL=INFO
```