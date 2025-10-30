# 🔍 Keycloak 리다이렉트 문제 디버깅 가이드

## 문제 상황
- localhost 포트포워딩 접근: Keycloak 리다이렉트 **작동** ✅
- 마스터 노드 IP 포트포워딩 접근: Keycloak 리다이렉트 **작동 안함** ❌

## 추가된 디버깅 로그

원래 코드 로직은 그대로 유지하고, 문제 진단을 위한 로그만 추가했습니다.

### 수정된 파일 목록
1. `portal-frontend/src/config/env.ts` - 환경변수 로드 및 Mixed Content 체크
2. `portal-frontend/src/config/oidc.ts` - OIDC 설정 확인
3. `portal-frontend/src/components/AuthWrapper.tsx` - 인증 상태 추적
4. `portal-frontend/src/components/LoginPage.tsx` - 로그인 버튼 클릭 추적

## 브라우저에서 확인할 로그

### 페이지 로드 시 자동 출력되는 로그:

```
🟠 [env.ts] ===== 환경변수 로드 시작 =====
🟠 [env.ts] window.ENV 확인: { ... }
✅ [env.ts] window.ENV 로드 성공

🟠 [env.ts] ===== Mixed Content 체크 =====
🟠 [env.ts] 현재 페이지: http://...
🟠 [env.ts] 현재 프로토콜: http:
🟠 [env.ts] Keycloak URL: https://...
🟠 [env.ts] Keycloak 프로토콜: https:

🟢 [oidc.ts] ===== OIDC 설정 생성 시작 =====
🟢 [oidc.ts] Keycloak 기본 정보:
  - KEYCLOAK_URL: https://...
  - KEYCLOAK_REALM: ...
  - KEYCLOAK_CLIENT_ID: ...
  - authority: https://.../realms/...

🟢 [oidc.ts] Keycloak 엔드포인트:
  - issuer: ...
  - authorization: ...
  - token: ...
  - userinfo: ...
  - endSession: ...

🟢 [oidc.ts] Redirect URI 설정:
  - window.location.origin: http://...
  - redirect_uri (계산값): http://.../callback
  - post_logout_redirect_uri (계산값): http://...
  - silent_redirect_uri (계산값): http://.../silent-callback

🟡 [AuthWrapper] 렌더링: { isAuthenticated: false, isLoading: false, ... }
```

### 로그인 버튼 클릭 시 출력되는 로그:

```
🔵 [LoginPage] ===== 로그인 버튼 클릭 =====
🔵 [LoginPage] window.ENV: { ... }
🔵 [LoginPage] window.location: { href: ..., origin: ..., protocol: ..., host: ... }
🔵 [LoginPage] onLogin() 함수 호출 시도...
🟡 [AuthWrapper] signinRedirect() 호출 시작
🟡 [AuthWrapper] 현재 URL: http://...
🟡 [AuthWrapper] signinRedirect() 호출 완료
🔵 [LoginPage] onLogin() 함수 호출 완료

→ 이후 Keycloak 페이지로 리다이렉트되어야 함
```

### 에러 발생 시 출력되는 로그:

```
🔴 [LoginPage] 로그인 실패: Error: ...
🔴 [LoginPage] 에러 타입: object
🔴 [LoginPage] 에러 상세: { ... }
🔴 [AuthWrapper] signinRedirect() 에러: ...
🔴 [AuthWrapper] 에러 상세: { ... }
🔴 [AuthWrapper] OIDC 에러 발생: ...
```

## 고객사 환경에서 테스트 방법

### 1단계: 브라우저 개발자 도구 열기
- Chrome/Edge: F12 또는 Ctrl+Shift+I (Windows), Cmd+Option+I (Mac)
- Console 탭 선택

### 2단계: localhost 포트포워딩 테스트 (정상 작동하는 경우)
```bash
kubectl port-forward -n user-portal svc/user-portal-frontend-service 8080:80
```

브라우저에서 `http://localhost:8080` 접근 후:
1. Console 로그 전체 복사
2. 로그인 버튼 클릭
3. Console 로그 추가 출력 확인
4. Keycloak 페이지로 리다이렉트되는지 확인

### 3단계: 마스터 노드 IP 포트포워딩 테스트 (문제 발생하는 경우)
```bash
kubectl port-forward --address=0.0.0.0 -n user-portal svc/user-portal-frontend-service 8080:80
```

브라우저에서 `http://<마스터노드IP>:8080` 접근 후:
1. Console 로그 전체 복사
2. 로그인 버튼 클릭
3. Console 로그 추가 출력 확인
4. **어디서 멈추는지** 확인

### 4단계: 로그 비교
- 두 시나리오의 로그를 비교
- 어느 단계에서 차이가 발생하는지 확인

## 예상되는 문제 시나리오

### 시나리오 1: Mixed Content 차단 (가장 가능성 높음)

**예상 로그:**
```
🔴 [env.ts] ===== Mixed Content 경고! =====
🔴 [env.ts] HTTP 페이지에서 HTTPS Keycloak으로 리다이렉트 시도
🔴 [env.ts] 브라우저가 차단할 수 있습니다!
```

**증상:**
- 로그인 버튼 클릭 후 아무 일도 일어나지 않음
- 브라우저 Console에 Mixed Content 관련 경고/에러
- Network 탭에 차단된 요청 표시

**원인:**
- HTTP 페이지(포털)에서 HTTPS 리소스(Keycloak)로 리다이렉트 시도
- 브라우저의 Mixed Content 정책이 차단

**해결방법:**
1. **권장:** 포털도 HTTPS로 접근
   - Ingress에 TLS 설정 추가
   - 또는 Nginx 리버스 프록시에 SSL 설정
2. **임시:** Keycloak을 HTTP로 변경 (보안 취약, 비추천)
3. **임시:** 브라우저 설정에서 Mixed Content 허용 (개발/테스트용)

**localhost에서 작동하는 이유:**
- 브라우저가 localhost에 대해서는 Mixed Content 정책을 완화함
- Chrome: localhost를 "Potentially Trustworthy Origin"으로 간주

---

### 시나리오 2: 사설 인증서 신뢰 문제

**예상 로그:**
```
🟡 [AuthWrapper] signinRedirect() 호출 시작
(이후 로그 없음)
```

**증상:**
- `signinRedirect()` 호출 후 멈춤
- 브라우저 Console에 인증서 관련 에러
  - `ERR_CERT_AUTHORITY_INVALID`
  - `NET::ERR_CERT_INVALID`

**원인:**
- Keycloak이 사설 인증서 사용
- 브라우저가 인증서를 신뢰하지 않음

**해결방법:**
1. Keycloak URL에 직접 접속하여 인증서 예외 추가
   - Chrome: "고급" → "안전하지 않음(계속)" 클릭
2. 브라우저에 CA 인증서 수동 설치
3. **권장:** 유효한 SSL 인증서 사용

**localhost에서 작동하는 이유:**
- 이미 Keycloak URL에 접속하여 인증서 예외를 추가했을 가능성

---

### 시나리오 3: 환경변수 로드 실패

**예상 로그:**
```
⚠️ [env.ts] window.ENV가 undefined입니다.
⚠️ [env.ts] /env.js 파일이 로드되지 않았을 수 있습니다.
⚠️ [env.ts] 기본값을 사용합니다.
```

**증상:**
- 기본값(miribit.cloud)으로 설정됨
- 잘못된 Keycloak URL로 리다이렉트 시도

**원인:**
- `/env.js` 파일이 생성되지 않음
- 또는 Nginx가 `/env.js` 파일을 제공하지 못함

**해결방법:**
```bash
# Pod 내부에서 env.js 파일 확인
kubectl exec -n user-portal -it <frontend-pod-name> -- cat /usr/share/nginx/html/env.js

# ConfigMap 확인
kubectl get configmap frontend-env -n user-portal -o yaml

# entrypoint.sh가 제대로 실행되었는지 확인
kubectl logs -n user-portal <frontend-pod-name> | grep "Generating runtime configuration"
```

---

### 시나리오 4: CORS 정책 위반

**예상 로그:**
```
🟡 [AuthWrapper] signinRedirect() 호출 완료
```
그 후 Console에 CORS 에러:
```
Access to XMLHttpRequest at 'https://keycloak...' from origin 'http://...' has been blocked by CORS policy
```

**증상:**
- `signinRedirect()` 자체는 성공
- Keycloak으로의 리다이렉트는 발생하지만 CORS 에러

**원인:**
- Keycloak의 CORS 설정 문제
- Keycloak 클라이언트의 Web Origins 설정 누락

**해결방법:**
Keycloak Admin Console에서:
1. Clients → `frontend` 선택
2. Settings 탭
3. Web Origins에 포털 URL 추가:
   - `http://<마스터노드IP>:*`
   - 또는 `*` (모든 origin 허용, 비추천)

---

### 시나리오 5: react-oidc-context 라이브러리 에러

**예상 로그:**
```
🔴 [LoginPage] 로그인 실패: Error: ...
🔴 [AuthWrapper] signinRedirect() 에러: ...
```

**증상:**
- JavaScript 에러 발생
- 에러 메시지에서 원인 파악 가능

**원인:**
- OIDC 설정 오류
- Keycloak 엔드포인트 접근 불가
- 기타 라이브러리 내부 에러

**해결방법:**
- 에러 메시지 내용에 따라 대응
- Keycloak URL이 접근 가능한지 확인
- Network 탭에서 실패하는 요청 확인

---

## 추가 확인 사항

### 1. Network 탭 확인
- 어떤 요청이 실패하는지 확인
- Keycloak URL로의 요청이 있는지 확인
- 응답 코드 확인 (200, 302, 403, 500 등)

### 2. Keycloak 클라이언트 설정 확인
```
Valid Redirect URIs:
  - * (와일드카드로 설정되어 있다고 했으므로 이건 문제 아님)

Web Origins:
  - * 또는 구체적인 origin 설정 필요
```

### 3. 브라우저 보안 설정
- Mixed Content 차단 여부
- 사설 인증서 신뢰 여부
- JavaScript 활성화 여부

## 정보 수집 체크리스트

고객사 환경에서 다음 정보를 수집하여 공유해주세요:

### 로그 정보
- [ ] localhost 접근 시 브라우저 Console 로그 (전체)
- [ ] 마스터 노드 IP 접근 시 브라우저 Console 로그 (전체)
- [ ] 로그인 버튼 클릭 후 Console 로그 (전체)
- [ ] Network 탭 스크린샷 또는 로그

### 환경 정보
- [ ] 접근 URL
  - localhost: `http://localhost:????`
  - 마스터 노드: `http://???.???.???.???:????`
- [ ] Keycloak URL: `https://...` (HTTP인지 HTTPS인지)
- [ ] 브라우저 종류 및 버전
- [ ] 사설 인증서 사용 여부

### Kubernetes 정보
```bash
# ConfigMap 확인
kubectl get configmap frontend-env -n user-portal -o yaml

# env.js 파일 확인
kubectl exec -n user-portal -it $(kubectl get pod -n user-portal -l app=user-portal-frontend -o jsonpath='{.items[0].metadata.name}') -- cat /usr/share/nginx/html/env.js

# Frontend Pod 로그
kubectl logs -n user-portal -l app=user-portal-frontend --tail=50
```

## 빠른 테스트 방법

고객사 환경에서 브라우저 Console에 직접 입력:

```javascript
// 1. 환경변수 확인
console.log('window.ENV:', window.ENV);

// 2. Mixed Content 체크
console.log('현재 프로토콜:', window.location.protocol);
console.log('Keycloak URL:', window.ENV?.KEYCLOAK_URL);

// 3. Keycloak 접근 가능 여부 확인
fetch(window.ENV.KEYCLOAK_URL + '/realms/' + window.ENV.KEYCLOAK_REALM)
  .then(r => console.log('Keycloak 접근 성공:', r.status))
  .catch(e => console.error('Keycloak 접근 실패:', e));

// 4. redirect_uri 확인
console.log('redirect_uri:', window.location.origin + '/callback');
```

## 다음 단계

로그를 확인한 후:
1. 어느 시나리오에 해당하는지 판단
2. 해당 시나리오의 해결방법 적용
3. 문제가 지속되면 로그와 함께 상세 상황 공유

---

## 참고: 디버깅 로그 제거 방법

문제 해결 후 디버깅 로그를 제거하려면:

1. `portal-frontend/src/config/env.ts`에서 `// === 디버깅 로그 시작 ===` ~ `// === 디버깅 로그 끝 ===` 블록 삭제
2. `portal-frontend/src/config/oidc.ts`에서 동일 블록 삭제
3. `portal-frontend/src/components/AuthWrapper.tsx`에서 동일 블록 삭제
4. `portal-frontend/src/components/LoginPage.tsx`에서 동일 블록 삭제

또는 프로덕션 빌드 시 자동으로 제거되도록 설정할 수 있습니다.

