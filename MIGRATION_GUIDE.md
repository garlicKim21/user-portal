# 환경변수 기반 설정으로 마이그레이션 가이드

## 📋 변경 개요

**브랜치**: `feature/runtime-env-config`

프론트엔드의 하드코딩된 URL들을 런타임 환경변수로 변경하여, 하나의 컨테이너 이미지로 프로토타입 환경과 고객 환경 모두에 배포할 수 있도록 개선했습니다.

## 🎯 주요 변경 사항

### 1. 새로운 파일 추가

| 파일 | 설명 |
|------|------|
| `portal-frontend/entrypoint.sh` | 컨테이너 시작 시 환경변수를 env.js로 생성 |
| `portal-frontend/src/config/env.ts` | 런타임 환경변수 접근 헬퍼 |
| `deployment/frontend-configmap-prototype.yaml` | 프로토타입 환경 설정 |
| `deployment/frontend-configmap-customer.yaml` | 고객 환경 설정 (수정 필요) |
| `portal-frontend/RUNTIME_ENV_CONFIG.md` | 상세 설정 가이드 |
| `portal-frontend/public/env.js.example` | 로컬 개발용 예시 |
| `portal-frontend/.gitignore` | Git 제외 파일 목록 |

### 2. 수정된 파일

| 파일 | 변경 내용 |
|------|----------|
| `portal-frontend/Dockerfile` | entrypoint.sh 추가, 런타임 환경변수 적용 |
| `portal-frontend/index.html` | env.js 로드 추가 |
| `portal-frontend/src/config/oidc.ts` | 하드코딩 제거, env 모듈 사용 |
| `portal-frontend/src/components/Dashboard.tsx` | menuItems URL을 apiEndpoints로 변경 |
| `portal-frontend/src/components/AuthWrapper.tsx` | 로그아웃 URL 동적 생성 |
| `deployment/user-portal-frontend.yaml` | ConfigMap envFrom 추가 |

## 🔧 환경변수 목록

| 변수명 | 용도 | 프로토타입 기본값 |
|--------|------|------------------|
| `KEYCLOAK_URL` | Keycloak 서버 URL | `https://keycloak.miribit.cloud` |
| `KEYCLOAK_REALM` | Keycloak Realm | `sso-demo` |
| `KEYCLOAK_CLIENT_ID` | OIDC Client ID | `frontend` |
| `KEYCLOAK_CLIENT_SECRET` | OIDC Client Secret | `aSnWDRHlSNITRlME6uYgIkdTRmIxZk7j` |
| `GRAFANA_URL` | Grafana 대시보드 | `https://grafana.miribit.cloud` |
| `JENKINS_URL` | Jenkins CI/CD | `https://jenkins.miribit.cloud` |
| `ARGOCD_URL` | ArgoCD GitOps | `https://argocd.miribit.cloud` |
| `BACKEND_URL` | 백엔드 API | `https://portal.miribit.cloud` |
| `PORTAL_URL` | 포털 자체 URL | `https://portal.miribit.cloud` |

## 🚀 배포 방법

### Step 1: 브랜치 확인

```bash
git checkout feature/runtime-env-config
```

### Step 2: 컨테이너 이미지 빌드

```bash
cd portal-frontend

# 이미지 빌드 (환경변수 없이)
docker build -t your-registry/user-portal-frontend:latest .

# 이미지 푸시
docker push your-registry/user-portal-frontend:latest
```

### Step 3: ConfigMap 생성

**프로토타입 환경:**
```bash
kubectl apply -f deployment/frontend-configmap-prototype.yaml
```

**고객 환경 (수정 필요):**
1. `deployment/frontend-configmap-customer.yaml` 파일을 열어 고객 환경에 맞게 수정:
   ```yaml
   data:
     KEYCLOAK_URL: "https://고객-keycloak-url"
     KEYCLOAK_REALM: "고객-realm"
     # ... 기타 URL 수정
   ```

2. 적용:
   ```bash
   kubectl apply -f deployment/frontend-configmap-customer.yaml
   ```

### Step 4: Deployment 업데이트

```bash
# Deployment YAML에서 이미지 태그 확인/수정
vi deployment/user-portal-frontend.yaml

# 적용
kubectl apply -f deployment/user-portal-frontend.yaml
```

### Step 5: 확인

```bash
# Pod 로그 확인 (env.js 생성 확인)
kubectl logs -n user-portal deployment/user-portal-frontend

# 출력 예시:
# === Generating runtime configuration ===
# Generated env.js with following configuration:
# window.ENV = { KEYCLOAK_URL: "...", ... }
# === Configuration complete, starting Nginx ===

# 생성된 env.js 파일 확인
kubectl exec -n user-portal deployment/user-portal-frontend -- cat /usr/share/nginx/html/env.js
```

## 🔍 테스트 방법

### 1. 로컬 테스트

```bash
cd portal-frontend

# env.js.example을 env.js로 복사
cp public/env.js.example public/env.js

# 필요시 수정
vi public/env.js

# 개발 서버 실행
npm run dev

# 브라우저에서 http://localhost:3000 접속
```

### 2. 컨테이너 테스트

```bash
# 로컬에서 컨테이너 실행 (환경변수 주입)
docker run -p 8080:80 \
  -e KEYCLOAK_URL="https://test-keycloak.com" \
  -e KEYCLOAK_REALM="test-realm" \
  -e KEYCLOAK_CLIENT_ID="test-client" \
  your-registry/user-portal-frontend:latest

# 브라우저에서 http://localhost:8080 접속
# 개발자 도구 Console에서 확인:
# > window.ENV
```

### 3. 쿠버네티스 테스트

```bash
# Pod에 접속하여 env.js 확인
kubectl exec -it -n user-portal deployment/user-portal-frontend -- sh

# 컨테이너 안에서
cat /usr/share/nginx/html/env.js

# Keycloak 접근 가능 여부 확인
wget -O- $KEYCLOAK_URL/realms/$KEYCLOAK_REALM/.well-known/openid-configuration
```

## ⚠️ 고객 환경 특수 사항

### 1. Realm이 포함되지 않은 Keycloak URL

고객 환경에서 Keycloak URL이 다음과 같이 제공될 수 있습니다:

```yaml
# ✅ 올바른 형식
KEYCLOAK_URL: "https://keycloak.customer.com"  # Realm 제외
KEYCLOAK_REALM: "production"                    # Realm 별도 지정

# ❌ 잘못된 형식
KEYCLOAK_URL: "https://keycloak.customer.com/realms/production"
```

코드가 자동으로 조합하므로 URL과 Realm을 분리하세요.

### 2. 사설 인증서 (Self-Signed Certificate)

폐쇄망에서 사설 인증서를 사용하는 경우:

#### 증상:
- 로그인 버튼 클릭 시 아무 반응 없음
- 브라우저 Console 에러: `ERR_CERT_AUTHORITY_INVALID`
- Network 탭에서 Keycloak 요청 실패

#### 해결 방법:

**방법 1: 브라우저에 예외 추가 (임시)**
1. Keycloak URL을 직접 브라우저에서 방문
2. "고급" → "위험을 감수하고 계속" 클릭
3. 포털 페이지로 돌아가서 로그인 시도

**방법 2: 인증서 설치 (권장)**
1. 고객 환경의 CA 인증서를 받아 OS에 설치
2. 브라우저 재시작
3. 정상 접속

**방법 3: Ingress에서 SSL 종료**
- Nginx Ingress Controller에서 SSL 처리
- 내부는 HTTP 통신
- ConfigMap에서 URL을 HTTP로 설정

### 3. Keycloak Valid Redirect URIs

**중요!** Keycloak Admin Console에서 다음 설정 필수:

```
Realm: <KEYCLOAK_REALM>
→ Clients
→ <KEYCLOAK_CLIENT_ID>
→ Settings
→ Valid Redirect URIs: 
   https://<PORTAL_URL>/callback
   https://<PORTAL_URL>/*
→ Web Origins:
   https://<PORTAL_URL>
   또는 *
```

설정하지 않으면 로그인 리디렉션이 차단됩니다!

## 🐛 트러블슈팅

### 문제 1: 로그인 버튼 클릭 시 아무 반응 없음

**확인 사항:**
1. 브라우저 개발자 도구 (F12) → Console 탭
   - JavaScript 에러 확인
   - `window.ENV` 출력하여 환경변수 확인

2. Network 탭
   - Keycloak 요청 실패 여부 확인
   - CORS 에러, DNS 실패, 인증서 에러 등

3. Pod 로그
   ```bash
   kubectl logs -n user-portal deployment/user-portal-frontend
   ```
   - env.js 생성 여부 확인

4. Keycloak 설정
   - Valid Redirect URIs 확인
   - Client가 활성화되어 있는지 확인

### 문제 2: 환경변수가 적용되지 않음

```bash
# ConfigMap 확인
kubectl get configmap frontend-env -n user-portal -o yaml

# ConfigMap 재적용
kubectl apply -f deployment/frontend-configmap-customer.yaml

# Pod 재시작 (필수!)
kubectl rollout restart deployment/user-portal-frontend -n user-portal

# 새로운 Pod에서 env.js 확인
kubectl exec -n user-portal deployment/user-portal-frontend -- cat /usr/share/nginx/html/env.js
```

### 문제 3: 로컬 개발 시 env.js를 찾을 수 없음

로컬 개발에서는 `public/env.js` 파일이 선택사항입니다:

```bash
# env.js 생성 (선택)
cp public/env.js.example public/env.js

# 또는 그냥 개발 서버 실행 (기본값 사용)
npm run dev
```

파일이 없으면 `src/config/env.ts`의 기본값이 사용됩니다.

### 문제 4: CORS 에러

```
Access to XMLHttpRequest at 'https://keycloak.../auth' from origin 'https://portal...' 
has been blocked by CORS policy
```

**해결:**
- Keycloak Admin Console → Clients → Web Origins에 포털 URL 추가

## 📊 변경 전후 비교

### 변경 전 (하드코딩)

```typescript
// ❌ src/config/oidc.ts
export const oidcConfig = {
  authority: 'https://keycloak.miribit.cloud/realms/sso-demo',
  client_id: 'frontend',
  // ...
};

// ❌ src/components/Dashboard.tsx
const menuItems = [
  {
    id: 'grafana',
    url: 'https://grafana.miribit.cloud'
  },
  // ...
];
```

**문제점:**
- 환경별로 다른 이미지 빌드 필요
- URL 변경 시 소스 수정 + 재빌드 필요
- 고객 환경 이식 어려움

### 변경 후 (환경변수)

```typescript
// ✅ src/config/env.ts
export const env = {
  KEYCLOAK_URL: window.ENV?.KEYCLOAK_URL || 'https://keycloak.miribit.cloud',
  // ...
};

// ✅ src/config/oidc.ts
import env, { getKeycloakAuthority } from './env';

export const oidcConfig = {
  authority: getKeycloakAuthority(),
  client_id: env.KEYCLOAK_CLIENT_ID,
  // ...
};

// ✅ src/components/Dashboard.tsx
import { apiEndpoints } from '../config/oidc';

const menuItems = [
  {
    id: 'grafana',
    url: apiEndpoints.grafana
  },
  // ...
];
```

**장점:**
- ✅ 하나의 이미지로 모든 환경 배포
- ✅ ConfigMap만 수정하면 URL 변경 가능
- ✅ 쿠버네티스 네이티브 설정 관리
- ✅ 보안 향상 (Secret으로 민감 정보 관리 가능)

## 📝 다음 단계

### 1. 프로토타입 환경에서 테스트

```bash
# 1. 현재 프로토타입 환경에 배포
kubectl apply -f deployment/frontend-configmap-prototype.yaml
kubectl apply -f deployment/user-portal-frontend.yaml

# 2. 동작 확인
# - 로그인 가능한지
# - 각 서비스 (Grafana, Jenkins 등) 링크 동작 확인
# - 로그아웃 동작 확인

# 3. 문제 없으면 이미지 태그 업데이트
```

### 2. 고객 환경 ConfigMap 작성

```bash
# 1. 고객 환경 정보 수집
# - Keycloak URL 및 Realm
# - Grafana, Jenkins, ArgoCD URL
# - 포털 URL
# - Client ID 및 Secret

# 2. ConfigMap 수정
vi deployment/frontend-configmap-customer.yaml

# 3. 고객 환경에 적용
kubectl apply -f deployment/frontend-configmap-customer.yaml
kubectl apply -f deployment/user-portal-frontend.yaml
```

### 3. 보안 강화 (선택)

Client Secret을 Secret으로 분리:

```bash
# Secret 생성
kubectl create secret generic frontend-secret \
  --from-literal=KEYCLOAK_CLIENT_SECRET='your-secret-here' \
  -n user-portal

# Deployment 수정
# envFrom에 secretRef 추가
```

### 4. 메인 브랜치 병합

```bash
# 테스트 완료 후
git checkout main
git merge feature/runtime-env-config
git push origin main
```

## 📚 관련 문서

- [RUNTIME_ENV_CONFIG.md](portal-frontend/RUNTIME_ENV_CONFIG.md) - 상세 설정 가이드
- [Dockerfile](portal-frontend/Dockerfile) - 컨테이너 이미지 빌드
- [entrypoint.sh](portal-frontend/entrypoint.sh) - 런타임 환경변수 생성 스크립트
- [env.ts](portal-frontend/src/config/env.ts) - 환경변수 접근 헬퍼

## ❓ 질문 및 지원

문제가 발생하거나 질문이 있으면:
1. 브라우저 Console 및 Network 탭 확인
2. Pod 로그 확인
3. ConfigMap 및 Deployment 설정 재확인
4. 이 가이드의 트러블슈팅 섹션 참고

