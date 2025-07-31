# Kubernetes Web Console Portal Backend

OIDC와 동적 Pod를 이용한 쿠버네티스 웹 콘솔 포털의 백엔드 구현입니다.

## 기능

- OIDC 기반 사용자 인증 (Keycloak 지원)
- 동적 웹 콘솔 Pod 생성
- 쿠버네티스 리소스 관리 (ConfigMap, Pod, Service)
- RESTful API 제공

## 기술 스택

- **언어**: Go 1.21+
- **웹 프레임워크**: Gin
- **인증**: OIDC (coreos/go-oidc)
- **쿠버네티스**: client-go
- **컨테이너**: Docker (크로스 플랫폼 빌드)

## API 엔드포인트

### 인증 관련

- `GET /api/login` - OIDC 인증 시작
- `GET /api/callback` - OAuth2 콜백 처리
- `GET /api/user` - 사용자 정보 조회

### 웹 콘솔 관련

- `GET /api/launch-console` - 웹 콘솔 Pod 생성 및 실행

## 환경 변수

다음 환경 변수들을 설정해야 합니다:

```bash
# OIDC 설정
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
OIDC_ISSUER_URL=https://keycloak.example.com/auth/realms/your-realm
OIDC_REDIRECT_URL=http://localhost:8080/api/callback

# 서버 설정
PORT=8080

# 쿠버네티스 설정 (개발 환경용)
KUBECONFIG=~/.kube/config
```

## 개발 환경 설정

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
go run main.go
```

## Docker 빌드

### 크로스 플랫폼 빌드 (AMD64)

```bash
docker buildx build --platform linux/amd64 -t portal-backend:latest .
```

### 로컬 빌드

```bash
docker build -t portal-backend:latest .
```

## 배포

### 쿠버네티스 배포

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: portal-backend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: portal-backend
  template:
    metadata:
      labels:
        app: portal-backend
    spec:
      containers:
      - name: portal-backend
        image: portal-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: OIDC_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: oidc-secret
              key: client-id
        - name: OIDC_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: oidc-secret
              key: client-secret
        - name: OIDC_ISSUER_URL
          value: "https://keycloak.example.com/auth/realms/your-realm"
        - name: OIDC_REDIRECT_URL
          value: "https://portal.example.com/api/callback"
```

## 보안 고려사항

1. **세션 관리**: 현재는 메모리 기반 세션을 사용하지만, 프로덕션에서는 Redis나 데이터베이스를 사용해야 합니다.
2. **토큰 보안**: ID 토큰은 안전하게 저장하고 전송해야 합니다.
3. **RBAC**: 쿠버네티스 클러스터에서 적절한 RBAC 설정이 필요합니다.
4. **네트워크 정책**: Pod 간 통신을 제한하는 네트워크 정책을 설정해야 합니다.

## 아키텍처

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Browser  │    │  Portal Backend │    │  Kubernetes    │
│                 │    │                 │    │   Cluster      │
│ 1. Login       │───▶│ 2. OIDC Auth   │───▶│ 3. Create Pod  │
│ 4. Launch      │    │ 5. Session Mgmt │    │ 4. ConfigMap   │
│ 6. Web Console │◀───│ 6. Return URL  │◀───│ 5. Service     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 라이선스

MIT License 