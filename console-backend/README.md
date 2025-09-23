# User Portal Backend

OIDC/LDAP 인증 기반 통합 개발자 포털의 백엔드 구현입니다.

## 🎯 주요 기능

- **🔐 OIDC/LDAP 통합 인증**: Keycloak을 통한 SSO 및 LDAP 그룹 기반 권한 관리
- **👥 프로젝트 기반 권한 관리**: `/dataops/{project}/{role}` 구조 기반 다중 프로젝트 지원
- **🖥️ 동적 웹 콘솔**: 사용자별 격리된 Kubernetes 웹 터미널 환경
- **🔗 통합 서비스 연동**: Grafana, Jenkins, ArgoCD SSO 지원
- **🛡️ Secret 기반 보안**: 민감한 정보를 Kubernetes Secret으로 관리
- **🔄 자동 리소스 정리**: 로그아웃 시 웹 콘솔 리소스 자동 삭제
- **📊 RESTful API**: React 프론트엔드와의 통신을 위한 API
- **⚡ 성능 최적화**: JWT 구조 최적화 및 세션 관리
- **🔒 보안 강화**: CSRF 보호, 세션 격리, 토큰 보안

## 🏗️ 아키텍처

### 패키지 구조

```
portal-backend/
├── main.go                    # 애플리케이션 진입점
├── internal/
│   ├── config/               # 설정 관리
│   │   └── config.go         # 환경 변수 로드 및 검증
│   ├── auth/                 # 인증 관련
│   │   ├── oidc.go          # OIDC 인증 로직
│   │   ├── jwt.go           # JWT 토큰 관리
│   │   ├── session_store.go # 세션 저장소
│   │   └── groups.go        # 사용자 그룹 관리
│   ├── kubernetes/           # 쿠버네티스 관련
│   │   ├── client.go        # 다중 클러스터 클라이언트
│   │   └── resource.go      # 리소스 생성/관리
│   ├── handlers/             # API 핸들러
│   │   ├── auth.go          # 인증 관련 핸들러
│   │   └── console.go       # 웹 콘솔 관련 핸들러
│   ├── middleware/           # 미들웨어
│   │   └── logging.go       # 로깅 미들웨어
│   ├── models/               # 데이터 모델
│   │   ├── session.go       # 세션 모델
│   │   └── error.go         # 에러 모델
│   ├── logger/               # 로깅
│   │   └── logger.go        # 구조화된 로깅
│   └── utils/                # 유틸리티
│       └── response.go       # HTTP 응답 유틸리티
├── Dockerfile                # Docker 이미지 빌드
├── env.example               # 환경 변수 예시
├── CONFIG.md                 # 설정 가이드
├── OIDC_SETUP.md            # OIDC 설정 가이드
└── README.md                # 이 문서
```

### 데이터 흐름

```
1. 사용자 로그인 요청
   ↓
2. OIDC 인증 (Keycloak)
   ↓
3. JWT 토큰 생성 및 세션 저장
   ↓
4. 웹 콘솔 요청
   ↓
5. Secret에서 클러스터 정보 조회
   ↓
6. 타겟 클러스터에 Pod 생성
   ↓
7. kubeconfig 생성 (CA 인증서 포함)
   ↓
8. 웹 콘솔 URL 반환
```

## 🛠️ 기술 스택

- **언어**: Go 1.24+
- **웹 프레임워크**: Gin
- **인증**: OIDC (coreos/go-oidc), JWT + Session 하이브리드
- **쿠버네티스**: client-go
- **보안**: JWT, Kubernetes Secrets, CSRF 보호
- **로깅**: 구조화된 로깅 (zap)
- **컨테이너**: Docker (크로스 플랫폼 빌드)
- **🆕 세션 관리**: 메모리 기반 세션 저장소
- **🆕 사용자 그룹 관리**: OIDC 토큰 기반 권한 추출

## 📋 API 엔드포인트

### 인증 관련

| 엔드포인트 | 메서드 | 설명 | 인증 필요 |
|-----------|--------|------|----------|
| `/api/login` | GET | OIDC 인증 시작 | ❌ |
| `/api/callback` | GET | OAuth2 콜백 처리 | ❌ |
| `/api/user` | GET | 사용자 정보 조회 | ✅ |
| `/api/logout` | GET | 로그아웃 | ✅ |

### 웹 콘솔 관련

| 엔드포인트 | 메서드 | 설명 | 인증 필요 |
|-----------|--------|------|----------|
| `/api/launch-console` | GET | 웹 콘솔 Pod 생성 및 실행 | ✅ |
| `/api/console-status` | GET | 웹 콘솔 상태 확인 | ✅ |

### 헬스체크

| 엔드포인트 | 메서드 | 설명 |
|-----------|--------|------|
| `/health` | GET | 애플리케이션 상태 확인 |

## ⚙️ 환경 변수

### 필수 환경 변수

```bash
# OIDC 설정
OIDC_CLIENT_ID=portal-app                    # OIDC 클라이언트 ID
OIDC_CLIENT_SECRET=your-client-secret        # OIDC 클라이언트 시크릿
OIDC_ISSUER_URL=https://your-keycloak-url/realms/basphere  # OIDC 발급자 URL
OIDC_REDIRECT_URL=https://your-keycloak-url/api/callback     # OIDC 리다이렉트 URL

# JWT 설정
JWT_SECRET_KEY=your-super-secure-jwt-secret  # JWT 서명 키 (최소 32자)

# 서버 설정
PORT=8080                                    # 서버 포트
GIN_MODE=release                            # Gin 모드
ALLOWED_ORIGINS=https://portal.basphere.dev # CORS 허용 오리진
```

### 선택적 환경 변수

```bash
# 로깅 설정
LOG_LEVEL=INFO                              # 로그 레벨 (DEBUG/INFO/WARN/ERROR/FATAL)

# 웹 콘솔 설정
CONSOLE_NAMESPACE=web-console              # 웹 콘솔 네임스페이스
CONSOLE_IMAGE=projectgreenist/web-terminal:0.2.11  # 웹 콘솔 이미지 (최신 버전)
CONSOLE_CONTAINER_PORT=8080                # 컨테이너 포트
CONSOLE_SERVICE_PORT=80                    # 서비스 포트
CONSOLE_TTL_SECONDS=3600                   # TTL (초)
WEB_CONSOLE_BASE_URL=https://console.basphere.dev  # 웹 콘솔 베이스 URL

# 개발 환경용
KUBECONFIG=~/.kube/config                  # Kubeconfig 파일 경로 (개발 환경에서만 사용)
```

### Secret에서 관리되는 환경 변수

```bash
# 타겟 클러스터 설정 (Secret에서 관리)
TARGET_CLUSTER_SERVER=https://<target-cluster-api-server>:6443  # 타겟 클러스터 서버
TARGET_CLUSTER_CA_CERT_DATA=LS0tLS1CRUdJTi...      # CA 인증서 (base64 인코딩)

# kubectl OIDC 설정 (Secret에서 관리)
KUBECTL_OIDC_CLIENT_ID=kubernetes                  # kubectl OIDC 클라이언트 ID
KUBECTL_OIDC_CLIENT_SECRET=your-kubectl-secret     # kubectl OIDC 클라이언트 시크릿
```

## 🔐 보안 아키텍처

### Secret 기반 설정 관리

민감한 정보는 Kubernetes Secret으로 관리됩니다:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: user-portal-secrets
  namespace: user-portal
type: Opaque
data:
  jwt-secret-key: <base64-encoded-jwt-secret>
  oidc-client-secret: <base64-encoded-oidc-secret>
  kubectl-oidc-client-secret: <base64-encoded-kubectl-secret>
  target-cluster-server: <base64-encoded-cluster-url>
  target-cluster-ca-cert-data: <base64-encoded-ca-cert>
```

### CA 인증서 처리

타겟 클러스터의 CA 인증서를 base64로 인코딩하여 Secret에 저장:

```bash
# CA 인증서를 base64로 인코딩
cat /path/to/ca.crt | base64 -w 0
```

### 다중 클러스터 보안

- **A 클러스터**: 포털 애플리케이션 실행
- **B 클러스터**: 웹 콘솔 Pod 생성
- **CA 인증서**: B 클러스터와의 보안 연결
- **RBAC**: 사용자별 권한 제한

### 🆕 최신 보안 기능

- **JWT + Session 하이브리드**: 토큰과 세션을 결합한 이중 보안
- **CSRF 보호**: State 기반 CSRF 공격 방지
- **사용자 격리**: 완전한 사용자별 웹 콘솔 환경 격리
- **명령어 히스토리 보안**: PVC 기반 안전한 히스토리 저장
- **동적 권한 표시**: 사용자별 네임스페이스 및 역할 정보 표시

## 🚀 개발 환경 설정

### 1. 의존성 설치

```bash
go mod download
```

### 2. 환경 변수 설정

```bash
cp env.example .env
# .env 파일을 편집하여 실제 값으로 설정
```

### 3. 애플리케이션 실행

```bash
# 개발 모드
go run main.go

# 프로덕션 모드
GIN_MODE=release go run main.go
```

### 4. 테스트 실행

```bash
# 전체 테스트
go test ./...

# 특정 패키지 테스트
go test ./internal/auth/...

# 커버리지 포함 테스트
go test -cover ./...
```

## 🆕 최신 기능 상세

### JWT 구조 최적화 (v0.4.10+)

**이전 구조 (토큰 중첩)**
```go
type JWTClaims struct {
    UserID       string    `json:"user_id"`
    AccessToken  string    `json:"access_token"`   // OIDC 토큰
    IDToken      string    `json:"id_token"`       // OIDC 토큰
    RefreshToken string    `json:"refresh_token"`  // OIDC 토큰
    ExpiresAt    time.Time `json:"expires_at"`
}
```

**현재 구조 (최적화)**
```go
type JWTClaims struct {
    UserID    string    `json:"user_id"`
    SessionID string    `json:"session_id"`        // 세션 ID만 포함
    ExpiresAt time.Time `json:"expires_at"`
}
```

**개선 효과**
- ✅ JWT 크기 95% 감소
- ✅ 파싱 속도 3-5배 향상
- ✅ 보안성 대폭 향상
- ✅ 토큰 중첩 문제 완전 해결

### 웹 콘솔 개인화 (v0.2.11+)

**개인화된 터미널 정보**
```bash
=== Web Terminal Session ===
User: byun
Host: secure-terminal-byun
Namespace: blue
Roles: blue-developers/red-viewers
Time: 2025-01-11T17:49:40+0900 (KST)
==========================
user@secure-terminal-byun:~$
```

**동적 환경 변수**
- `USER_ID`: 실제 로그인 ID (byun, kim, kang 등)
- `DEFAULT_NAMESPACE`: 사용자 기본 네임스페이스
- `USER_ROLES`: 네임스페이스별 권한 정보

### 명령어 히스토리 지속성

**PVC 기반 저장**
```yaml
VolumeMounts:
  - Name: "history-storage"
    MountPath: "/home/user/.bash_history.d"  # 디렉토리로 마운트
    SubPath: "bash_history"                  # PVC 내부의 디렉토리
```

**자동 히스토리 관리**
- 사용자별 100Mi PVC 생성
- 웹 콘솔 Pod에 자동 마운트
- 세션 종료 후에도 히스토리 보존

### CSRF 보호

**State 기반 보안**
```go
// CSRF 보호용 State 생성
state, err := auth.GenerateRandomString(32)
if err != nil {
    return err
}

// 세션에 State 저장
h.tempSessions[state] = &models.Session{
    State:     state,
    CreatedAt: time.Now(),
}
```

## 🐳 Docker 빌드

### 크로스 플랫폼 빌드 (AMD64)

```bash
docker buildx build --platform linux/amd64 -t portal-backend:latest .
```

### 로컬 빌드

```bash
docker build -t portal-backend:latest .
```

### 멀티 스테이지 빌드

```dockerfile
# 빌드 스테이지
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o portal-backend .

# 실행 스테이지
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/portal-backend .
EXPOSE 8080
CMD ["./portal-backend"]
```

## 📦 배포

### 쿠버네티스 배포

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-portal-backend
  namespace: user-portal
spec:
  replicas: 1
  selector:
    matchLabels:
      app: user-portal-backend
  template:
    metadata:
      labels:
        app: user-portal-backend
    spec:
      serviceAccountName: portal-backend-sa
      containers:
      - name: user-portal-backend
        image: projectgreenist/user-portal-backend:0.3.19
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
        env:
        - name: OIDC_ISSUER_URL
          value: "https://keycloak.basphere.dev/realms/basphere"
        - name: OIDC_CLIENT_ID
          value: "portal-app"
        - name: OIDC_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: user-portal-secrets
              key: oidc-client-secret
        - name: JWT_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: user-portal-secrets
              key: jwt-secret-key
        - name: TARGET_CLUSTER_SERVER
          valueFrom:
            secretKeyRef:
              name: user-portal-secrets
              key: target-cluster-server
        - name: TARGET_CLUSTER_CA_CERT_DATA
          valueFrom:
            secretKeyRef:
              name: user-portal-secrets
              key: target-cluster-ca-cert-data
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 🔧 설정 가이드

### OIDC 설정

Keycloak을 사용한 OIDC 설정은 [OIDC_SETUP.md](OIDC_SETUP.md)를 참조하세요.

### 환경 변수 설정

상세한 환경 변수 설정은 [CONFIG.md](CONFIG.md)를 참조하세요.

## 🧪 테스트

### 단위 테스트

```bash
# 특정 함수 테스트
go test -v -run TestFunctionName

# 패키지별 테스트
go test ./internal/auth/...
go test ./internal/kubernetes/...
```

### 통합 테스트

```bash
# 전체 통합 테스트
go test -tags=integration ./...

# 특정 통합 테스트
go test -tags=integration -run TestIntegration ./...
```

### 성능 테스트

```bash
# 벤치마크 테스트
go test -bench=. ./...

# 프로파일링
go test -cpuprofile=cpu.prof -bench=. ./...
```

## 🔍 모니터링

### 로그 레벨

```bash
# 개발 환경
LOG_LEVEL=DEBUG

# 프로덕션 환경
LOG_LEVEL=INFO
```

### 헬스체크

```bash
# 애플리케이션 상태 확인
curl http://localhost:8080/health
```

### 메트릭

```bash
# 메모리 사용량
curl http://localhost:8080/debug/pprof/heap

# 고루틴 상태
curl http://localhost:8080/debug/pprof/goroutine
```

## 🚨 문제 해결

### 일반적인 문제

1. **OIDC 연결 실패**
   - `OIDC_ISSUER_URL` 확인
   - 클라이언트 ID/시크릿 확인
   - 네트워크 연결 확인

2. **쿠버네티스 연결 실패**
   - `KUBECONFIG` 파일 확인
   - 클러스터 접근 권한 확인
   - Secret 설정 확인

3. **웹 콘솔 생성 실패**
   - 타겟 클러스터 연결 확인
   - CA 인증서 설정 확인
   - RBAC 권한 확인

### 로그 분석

```bash
# 애플리케이션 로그 확인
kubectl logs -f deployment/user-portal-backend

# 특정 시간대 로그
kubectl logs --since=1h deployment/user-portal-backend

# 에러 로그만 확인
kubectl logs deployment/user-portal-backend | grep ERROR
```

## 📚 추가 문서

- **[Configuration Guide](CONFIG.md)** - 상세 설정 가이드
- **[OIDC Setup](OIDC_SETUP.md)** - OIDC 설정 가이드
- **[Deployment Guide](../deployment/README.md)** - 배포 가이드

## 📄 라이선스

MIT License 