# User Portal Frontend

React + TypeScript 기반의 모던 사용자 포털 프론트엔드입니다.

## 🎯 주요 기능

- **🎨 모던 React UI**: React 18 + TypeScript + shadcn/ui
- **🔐 OIDC 인증**: react-oidc-context를 통한 Keycloak SSO
- **👥 프로젝트 관리**: LDAP 그룹 기반 다중 프로젝트 선택
- **📊 통합 대시보드**: Grafana, Jenkins, ArgoCD, Web Terminal 통합 접근
- **📱 반응형 디자인**: 데스크톱 및 모바일 최적화
- **🎨 일관된 디자인**: Tailwind CSS + shadcn/ui 컴포넌트

## 🛠️ 기술 스택

- **프레임워크**: React 18 + TypeScript
- **빌드 도구**: Vite
- **UI 라이브러리**: shadcn/ui + Radix UI
- **스타일링**: Tailwind CSS
- **인증**: react-oidc-context
- **아이콘**: Lucide React
- **라우팅**: React Router DOM
- **HTTP 클라이언트**: Fetch API

## 🚀 개발 환경 설정

### 1. 의존성 설치

```bash
npm install
```

### 2. 개발 서버 실행

```bash
npm run dev
```

### 3. 빌드

```bash
npm run build
```

### 4. 미리보기

```bash
npm run preview
```

## 📁 프로젝트 구조

```
src/
├── components/              # React 컴포넌트
│   ├── ui/                 # shadcn/ui 기본 컴포넌트
│   ├── AuthWrapper.tsx     # OIDC 인증 래퍼
│   ├── Dashboard.tsx       # 메인 대시보드
│   ├── ProjectSelector.tsx # 프로젝트 선택 드롭다운
│   ├── UserInfo.tsx        # 사용자 정보 표시
│   └── LoginPage.tsx       # 로그인 페이지
├── services/               # API 서비스
│   ├── authService.ts      # 인증 관련 API
│   └── backendAuthService.ts # 백엔드 API 호출
├── types/                  # TypeScript 타입 정의
│   └── user.ts            # 사용자 및 프로젝트 타입
├── config/                 # 설정 파일
│   └── oidc.ts            # OIDC 설정
├── styles/                 # 스타일 파일
│   └── globals.css        # 전역 CSS
└── main.tsx               # React 진입점
```

## 🔧 환경 변수

### 개발 환경 (.env.local)

```bash
# Vite 개발 서버 포트
VITE_PORT=5173

# API 백엔드 URL (개발용)
VITE_API_BASE_URL=http://localhost:8080
```

### 프로덕션 환경

프로덕션에서는 Nginx를 통해 백엔드 API와 프록시됩니다.

## 🐳 Docker 배포

```bash
# 이미지 빌드
docker buildx build --platform linux/amd64 -t your-registry/user-portal-frontend:latest --push .

# 컨테이너 실행
docker run -p 80:80 your-registry/user-portal-frontend:latest
```

## 📚 주요 컴포넌트

### AuthWrapper
- OIDC 인증 상태 관리
- 로그인/로그아웃 플로우 처리
- 사용자 정보 및 프로젝트 데이터 관리

### Dashboard  
- 메인 대시보드 UI
- 서비스 메뉴 (Grafana, Jenkins, ArgoCD, Web Terminal)
- 사용자 정보 표시

### ProjectSelector
- LDAP 그룹 기반 프로젝트 선택
- 다중 프로젝트 소속 시 드롭다운 표시
- 단일 프로젝트 시 고정 표시

### UserInfo
- 사용자 정보 표시 (ID, 이름, 이메일)
- Keycloak 토큰에서 실시간 정보 추출

## 🔐 인증 플로우

1. **로그인**: Keycloak OIDC 리다이렉트
2. **토큰 수신**: ID Token, Access Token 저장
3. **사용자 정보 파싱**: Keycloak 프로필에서 정보 추출
4. **프로젝트 매핑**: LDAP 그룹에서 프로젝트 권한 파싱
5. **자동 로그아웃**: 리소스 정리 + Keycloak 세션 종료

## 📖 개발 가이드

### 새로운 컴포넌트 추가

1. `src/components/` 디렉토리에 컴포넌트 생성
2. TypeScript 인터페이스 정의
3. shadcn/ui 컴포넌트 활용
4. Tailwind CSS로 스타일링

### API 서비스 추가

1. `src/services/` 디렉토리에 서비스 파일 생성
2. Fetch API 기반 HTTP 클라이언트 구현
3. 에러 처리 및 타입 안전성 보장

### 타입 정의 추가

1. `src/types/` 디렉토리에 타입 파일 생성
2. 백엔드 API 응답과 일치하는 인터페이스 정의
3. 컴포넌트 props 타입 정의