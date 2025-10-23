# 런타임 환경변수 설정 가이드

이 문서는 프론트엔드의 런타임 환경변수 설정 방법을 설명합니다.

## 📋 변경 사항 개요

기존의 하드코딩된 URL들을 **런타임 환경변수**로 변경하여, 하나의 컨테이너 이미지로 여러 환경에 배포할 수 있도록 개선했습니다.

### 주요 변경점

1. **환경변수 기반 설정**: 모든 URL이 환경변수로 관리됨
2. **런타임 적용**: 컨테이너 시작 시 `env.js` 파일 생성
3. **ConfigMap 지원**: Kubernetes ConfigMap으로 환경별 설정 관리
4. **단일 이미지**: 프로토타입/고객 환경에 동일한 이미지 사용 가능

## 🔧 환경변수 목록

| 환경변수 | 설명 | 기본값 (프로토타입) |
|---------|------|-------------------|
| `KEYCLOAK_URL` | Keycloak 서버 URL | `https://keycloak.miribit.cloud` |
| `KEYCLOAK_REALM` | Keycloak Realm 이름 | `sso-demo` |
| `KEYCLOAK_CLIENT_ID` | OIDC Client ID | `frontend` |
| `KEYCLOAK_CLIENT_SECRET` | OIDC Client Secret | `aSnWDRHlSNITRlME6uYgIkdTRmIxZk7j` |
| `GRAFANA_URL` | Grafana 대시보드 URL | `https://grafana.miribit.cloud` |
| `JENKINS_URL` | Jenkins CI/CD URL | `https://jenkins.miribit.cloud` |
| `ARGOCD_URL` | ArgoCD GitOps URL | `https://argocd.miribit.cloud` |
| `BACKEND_URL` | 백엔드 API URL | `https://portal.miribit.cloud` |
| `PORTAL_URL` | 포털 자체 URL (리디렉션용) | `https://portal.miribit.cloud` |

## 🚀 배포 방법

### 1. ConfigMap 생성

**프로토타입 환경:**
```bash
kubectl apply -f deployment/frontend-configmap-prototype.yaml
```

**고객 환경:**
```bash
# 고객 환경에 맞게 수정 후 적용
kubectl apply -f deployment/frontend-configmap-customer.yaml
```

### 2. Deployment 배포

```bash
kubectl apply -f deployment/user-portal-frontend.yaml
```

### 3. 확인

```bash
# Pod 로그 확인 (env.js 생성 확인)
kubectl logs -n user-portal deployment/user-portal-frontend

# 생성된 환경변수 확인
kubectl exec -n user-portal deployment/user-portal-frontend -- cat /usr/share/nginx/html/env.js
```

## 🔍 작동 원리

### 1. 컨테이너 시작 시

`entrypoint.sh` 스크립트가 실행되어 환경변수를 `/usr/share/nginx/html/env.js` 파일로 생성:

```javascript
window.ENV = {
  KEYCLOAK_URL: "https://customer-keycloak.internal",
  KEYCLOAK_REALM: "production",
  // ... 기타 환경변수
};
```

### 2. 브라우저에서 로드

`index.html`에서 `env.js`를 먼저 로드:

```html
<script src="/env.js"></script>
```

### 3. React 앱에서 사용

`src/config/env.ts`에서 환경변수 접근:

```typescript
export const env = {
  KEYCLOAK_URL: window.ENV?.KEYCLOAK_URL || 'https://keycloak.miribit.cloud',
  // ...
};
```

## ⚙️ 로컬 개발

로컬 개발 시에는 `env.js` 파일이 없으므로 기본값이 사용됩니다.

### 로컬에서 테스트하려면:

1. `public/env.js` 파일 생성:
```javascript
window.ENV = {
  KEYCLOAK_URL: "http://localhost:8080",
  KEYCLOAK_REALM: "sso-demo",
  // ...
};
```

2. 개발 서버 실행:
```bash
npm run dev
```

## 🔒 보안 고려사항

### 1. Client Secret 관리

현재는 ConfigMap에 저장되어 있지만, **Kubernetes Secret**으로 관리하는 것이 더 안전합니다:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: frontend-secret
  namespace: user-portal
type: Opaque
stringData:
  KEYCLOAK_CLIENT_SECRET: "your-secret-here"
```

Deployment에 추가:
```yaml
envFrom:
- configMapRef:
    name: frontend-env
- secretRef:
    name: frontend-secret  # Secret 추가
```

### 2. PKCE 사용 권장

현재는 `pkce: false`로 설정되어 있지만, 보안을 위해 **PKCE (Proof Key for Code Exchange)**를 활성화하고 Client Secret을 제거하는 것이 권장됩니다.

## ⚠️ 고객 환경 특수 사항

### 1. Realm이 포함되지 않은 Keycloak URL

일부 환경에서 `KEYCLOAK_URL`에 realm이 포함되지 않을 수 있습니다:

```yaml
# ❌ Realm 포함 (잘못된 형식)
KEYCLOAK_URL: "https://keycloak.example.com/realms/myrealm"

# ✅ Realm 미포함 (올바른 형식)
KEYCLOAK_URL: "https://keycloak.example.com"
KEYCLOAK_REALM: "myrealm"
```

코드는 자동으로 조합하여 사용합니다:
```typescript
const authority = `${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}`;
```

### 2. 사설 인증서 (Self-Signed Certificate)

폐쇄망 환경에서 사설 인증서를 사용하는 경우:

#### 문제점:
- 브라우저에서 "신뢰할 수 없는 인증서" 경고 발생
- OIDC 리디렉션 실패 가능성

#### 해결 방법:

**Option 1: 브라우저에 인증서 신뢰 추가 (임시)**
1. Keycloak URL 직접 방문
2. "고급" → "계속 진행" 클릭
3. 브라우저 세션에 예외 추가

**Option 2: 인증서를 OS/브라우저에 설치 (권장)**
1. 고객 환경의 CA 인증서 받기
2. OS 인증서 저장소에 설치
3. 브라우저 재시작

**Option 3: 프록시를 통한 SSL 종료**
- Nginx Ingress에서 SSL 종료
- 내부 통신은 HTTP 사용
- 외부 사용자는 HTTPS로 접근

#### 개발자 도구 확인:

브라우저 Console에서 다음 에러 확인:
```
ERR_CERT_AUTHORITY_INVALID
ERR_CERT_COMMON_NAME_INVALID
```

Network 탭에서 실패한 요청 확인:
- Status: `(failed)` 또는 `net::ERR_CERT_*`
- Keycloak 인증 엔드포인트 호출 실패

#### nginx.conf 설정 (사설 인증서 대응):

필요한 경우 nginx에서 upstream SSL 검증 비활성화:
```nginx
location /api/ {
    proxy_pass https://backend.internal;
    proxy_ssl_verify off;  # 개발 환경에만 사용
    # ...
}
```

⚠️ **주의**: `proxy_ssl_verify off`는 보안상 위험하므로 개발/테스트 환경에서만 사용하세요.

## 📝 ConfigMap 수정 예시

### 고객 환경에 맞게 수정:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: frontend-env
  namespace: user-portal
data:
  # 고객 환경의 실제 URL로 변경
  KEYCLOAK_URL: "https://keycloak.customer.internal"  # Realm 제외
  KEYCLOAK_REALM: "production"                        # Realm 별도 지정
  KEYCLOAK_CLIENT_ID: "portal-app"
  KEYCLOAK_CLIENT_SECRET: "customer-secret"
  
  GRAFANA_URL: "https://grafana.customer.internal"
  JENKINS_URL: "https://jenkins.customer.internal"
  ARGOCD_URL: "https://argocd.customer.internal"
  BACKEND_URL: "https://portal.customer.internal"
  PORTAL_URL: "https://portal.customer.internal"
```

수정 후 재시작:
```bash
# ConfigMap 적용
kubectl apply -f deployment/frontend-configmap-customer.yaml

# Pod 재시작 (새로운 환경변수 적용)
kubectl rollout restart deployment/user-portal-frontend -n user-portal

# 확인
kubectl logs -n user-portal deployment/user-portal-frontend
```

## 🐛 트러블슈팅

### 로그인 버튼 클릭 시 아무 반응이 없는 경우:

1. **브라우저 개발자 도구 확인**
   - F12 → Network 탭
   - Console 탭에서 에러 확인
   - CORS 에러, DNS 실패, 인증서 에러 등

2. **env.js 생성 확인**
   ```bash
   kubectl exec -n user-portal deployment/user-portal-frontend -- cat /usr/share/nginx/html/env.js
   ```

3. **Keycloak 접근 가능 여부 확인**
   ```bash
   # Pod에서 테스트
   kubectl exec -n user-portal deployment/user-portal-frontend -- wget -O- https://keycloak-url/realms/realm-name/.well-known/openid-configuration
   ```

4. **Keycloak Valid Redirect URIs 확인**
   - Keycloak Admin Console → Clients → frontend
   - Valid Redirect URIs에 `https://portal-url/callback` 추가

### 환경변수가 적용되지 않는 경우:

```bash
# ConfigMap 재적용
kubectl apply -f deployment/frontend-configmap-customer.yaml

# Pod 재시작 (필수!)
kubectl delete pod -n user-portal -l app=user-portal-frontend
```

### 로컬에서 env.js 로드 에러:

개발 서버에서는 `public/env.js` 파일을 생성하거나, 기본값이 사용되도록 그대로 두세요.

## 📚 참고 자료

- [Vite 환경변수 가이드](https://vitejs.dev/guide/env-and-mode.html)
- [Kubernetes ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [OIDC Authorization Code Flow](https://openid.net/specs/openid-connect-core-1_0.html)
- [Self-Signed Certificate 처리](https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/)

