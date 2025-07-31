# Portal Demo - Web Console PoC

OIDC와 동적 Pod를 이용한 쿠버네티스 웹 콘솔 포털 프로젝트입니다.

## 📋 프로젝트 개요

이 프로젝트는 사용자가 웹 포털을 통해 인증하고, 버튼 클릭 한 번으로 자신만의 격리된 웹 기반 쿠버네티스 CLI(kubectl) 환경을 동적으로 제공받는 시스템을 구축하는 것이 목표입니다.

### 🎯 주요 기능

- **OIDC 기반 인증**: Keycloak을 통한 안전한 사용자 인증
- **동적 웹 콘솔 생성**: 사용자별 격리된 웹 터미널 환경 제공
- **쿠버네티스 통합**: 동적 Pod, Service, ConfigMap 생성
- **모던 UI**: Glassmorphism 디자인의 전문적인 웹 인터페이스

## 🏗️ 시스템 아키텍처

### 클러스터 구성

```
┌─────────────────────────────────────────────────────────────┐
│                    서비스 클러스터 (A)                        │
├─────────────────────────────────────────────────────────────┤
│  • Keycloak (OIDC IdP)                                    │
│  • Portal Backend (Go + Gin)                              │
│  • Portal Frontend (Vite + Vanilla JS)                    │
│  • OPNsense 방화벽 (Nginx 리버스 프록시)                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  제어 대상 클러스터 (B)                       │
├─────────────────────────────────────────────────────────────┤
│  • 동적 생성되는 웹 콘솔 Pod                              │
│  • 사용자별 격리된 kubectl 환경                           │
│  • ConfigMap 기반 kubeconfig 관리                         │
└─────────────────────────────────────────────────────────────┘
```

### 사용자 흐름

1. 사용자가 웹 브라우저로 포털에 접속
2. "Login" 버튼 클릭 → Keycloak 로그인 페이지로 리디렉션
3. LDAP 계정으로 로그인
4. 인증 성공 후 포털로 돌아와 사용자 정보 표시
5. "Open Web Terminal" 버튼 클릭
6. 새 브라우저 탭에서 웹 기반 터미널 열림
7. 별도 로그인 없이 kubectl 명령어 사용 가능

## 🛠️ 기술 스택

### Backend
- **언어**: Go 1.21+
- **웹 프레임워크**: Gin
- **인증**: OIDC (coreos/go-oidc)
- **쿠버네티스**: client-go
- **컨테이너**: Docker (크로스 플랫폼 빌드)

### Frontend
- **빌드 도구**: Vite
- **언어**: Vanilla JavaScript
- **스타일**: CSS3 (Glassmorphism)
- **패키지 관리**: npm

### 인프라
- **인증**: Keycloak + OpenLDAP
- **컨테이너**: Docker
- **오케스트레이션**: Kubernetes
- **프록시**: Nginx (OPNsense 플러그인)

## 📁 프로젝트 구조

```
user-potal/
├── portal-backend/           # 백엔드 애플리케이션
│   ├── main.go              # 메인 애플리케이션 진입점
│   ├── internal/            # 내부 패키지
│   │   ├── models/          # 데이터 모델
│   │   ├── auth/            # OIDC 인증 로직
│   │   ├── kubernetes/      # K8s 클라이언트
│   │   └── handlers/        # API 핸들러
│   ├── templates/           # HTML 템플릿
│   ├── static/              # 정적 파일
│   ├── Dockerfile           # Docker 이미지 빌드
│   ├── env.example          # 환경 변수 예시
│   ├── OIDC_SETUP.md        # OIDC 설정 가이드
│   └── README.md            # 백엔드 상세 문서
├── portal-frontend/         # 프론트엔드 애플리케이션
│   ├── src/                 # 소스 코드
│   │   └── main.js          # 메인 JavaScript
│   ├── index.html           # 메인 HTML
│   ├── package.json         # npm 의존성
│   └── .gitignore           # Git 무시 파일
├── pdf_content.txt          # 프로젝트 요구사항 문서
└── README.md               # 프로젝트 전체 문서 (현재 파일)
```

## 🚀 빠른 시작

### 1. 사전 요구사항

- **Node.js** 18+ 및 npm
- **Go** 1.21+
- **Docker** (선택사항)
- **Kubernetes** 클러스터 접근 권한

### 2. 프로젝트 클론

```bash
git clone git@github.com:garlicKim21/user-portal.git
cd user-portal
```

### 3. 백엔드 설정

```bash
cd portal-backend

# 의존성 설치
go mod download

# 환경 변수 설정
cp env.example .env
# .env 파일을 편집하여 실제 값으로 설정

# 애플리케이션 실행
go run main.go
```

### 4. 프론트엔드 설정

```bash
cd portal-frontend

# 의존성 설치
npm install

# 개발 서버 실행
npm run dev
```

### 5. Docker 빌드 (선택사항)

```bash
# 백엔드 크로스 플랫폼 빌드
cd portal-backend
docker buildx build --platform linux/amd64 -t portal-backend:latest .

# 프론트엔드 빌드
cd ../portal-frontend
npm run build
```

## 🔧 환경 변수 설정

### Backend 환경 변수

```bash
# OIDC 설정
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
OIDC_ISSUER_URL=https://keycloak.example.com/auth/realms/your-realm
OIDC_REDIRECT_URL=http://localhost:8080/api/callback

# 서버 설정
PORT=8080

# 쿠버네티스 설정
KUBECONFIG=~/.kube/config
```

### Frontend 환경 변수

```bash
# 개발 서버 포트 (기본값: 5173)
VITE_PORT=5173
```

## 📚 API 문서

### 인증 관련

| 엔드포인트 | 메서드 | 설명 |
|-----------|--------|------|
| `/api/login` | GET | OIDC 인증 시작 |
| `/api/callback` | GET | OAuth2 콜백 처리 |
| `/api/user` | GET | 사용자 정보 조회 |

### 웹 콘솔 관련

| 엔드포인트 | 메서드 | 설명 |
|-----------|--------|------|
| `/api/launch-console` | GET | 웹 콘솔 Pod 생성 및 실행 |

## 🔐 OIDC 설정

Keycloak을 사용한 OIDC 설정은 [portal-backend/OIDC_SETUP.md](portal-backend/OIDC_SETUP.md)를 참조하세요.

## 🐳 Docker 배포

### 백엔드 배포

```bash
# 이미지 빌드
docker buildx build --platform linux/amd64 -t portal-backend:latest ./portal-backend

# 컨테이너 실행
docker run -p 8080:8080 --env-file .env portal-backend:latest
```

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
              name: portal-secrets
              key: oidc-client-id
        - name: OIDC_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: portal-secrets
              key: oidc-client-secret
```

## 🧪 개발 및 테스트

### 로컬 개발

```bash
# 백엔드 개발 서버
cd portal-backend
go run main.go

# 프론트엔드 개발 서버
cd portal-frontend
npm run dev
```

### 테스트

```bash
# 백엔드 테스트
cd portal-backend
go test ./...

# 프론트엔드 테스트
cd portal-frontend
npm test
```

## 🤝 기여하기

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 라이선스

이 프로젝트는 MIT 라이선스 하에 배포됩니다. 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

## 📞 지원

프로젝트에 대한 질문이나 이슈가 있으시면 [Issues](https://github.com/garlicKim21/user-portal/issues)를 통해 문의해 주세요.

## 🙏 감사의 말

- [Gin](https://github.com/gin-gonic/gin) - Go 웹 프레임워크
- [Vite](https://vitejs.dev/) - 프론트엔드 빌드 도구
- [Keycloak](https://www.keycloak.org/) - OIDC 인증 서버
- [Kubernetes](https://kubernetes.io/) - 컨테이너 오케스트레이션

---

**Portal Demo - Web Console PoC** - Secure Kubernetes Terminal Access 