# OIDC 설정 가이드

이 문서는 Kubernetes Web Console Portal에서 OIDC 인증을 설정하는 방법을 설명합니다.

## 1. Keycloak 설정

### 1.1 Keycloak 설치 및 설정

#### Docker로 Keycloak 실행
```bash
docker run -d \
  --name keycloak \
  -p 8080:8080 \
  -e KEYCLOAK_ADMIN=admin \
  -e KEYCLOAK_ADMIN_PASSWORD=admin \
  quay.io/keycloak/keycloak:latest \
  start-dev
```

#### Keycloak 관리 콘솔 접속
- URL: http://localhost:8080
- 사용자명: admin
- 비밀번호: admin

### 1.2 Realm 생성

1. Keycloak 관리 콘솔에 로그인
2. 좌측 상단의 드롭다운에서 "Create Realm" 클릭
3. Realm 이름: `kubernetes-portal`
4. "Create" 클릭

### 1.3 Client 생성

1. 좌측 메뉴에서 "Clients" 클릭
2. "Create" 버튼 클릭
3. Client ID: `portal-backend`
4. Client Protocol: `openid-connect`
5. Root URL: `http://localhost:8080`
6. "Save" 클릭

### 1.4 Client 설정

#### Settings 탭
- Access Type: `confidential`
- Valid Redirect URIs: `http://localhost:8080/api/callback`
- Web Origins: `http://localhost:8080`

#### Credentials 탭
- Client Authenticator: `Client Id and Secret`
- Secret 복사 (환경 변수에 사용)

### 1.5 사용자 생성

1. 좌측 메뉴에서 "Users" 클릭
2. "Add user" 버튼 클릭
3. Username: `testuser`
4. Email: `testuser@example.com`
5. "Save" 클릭

#### 비밀번호 설정
1. "Credentials" 탭 클릭
2. Password: 원하는 비밀번호 입력
3. Temporary: `OFF` (체크 해제)
4. "Set Password" 클릭

## 2. 환경 변수 설정

### 2.1 개발 환경 (.env 파일)

```bash
# OIDC 설정
OIDC_CLIENT_ID=portal-backend
OIDC_CLIENT_SECRET=your-client-secret-from-keycloak
OIDC_ISSUER_URL=http://localhost:8080/realms/kubernetes-portal
OIDC_REDIRECT_URL=http://localhost:8080/api/callback

# 서버 설정
PORT=8080

# 쿠버네티스 설정 (개발 환경용)
KUBECONFIG=~/.kube/config
```

### 2.2 프로덕션 환경

```bash
# OIDC 설정
OIDC_CLIENT_ID=portal-backend
OIDC_CLIENT_SECRET=your-production-secret
OIDC_ISSUER_URL=https://keycloak.yourdomain.com/realms/kubernetes-portal
OIDC_REDIRECT_URL=https://portal.yourdomain.com/api/callback

# 서버 설정
PORT=8080
```

## 3. 쿠버네티스 OIDC 설정

### 3.1 API 서버 설정

쿠버네티스 API 서버에 OIDC 인증 제공자를 설정해야 합니다.

#### kube-apiserver 설정 예시
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: kube-apiserver
  namespace: kube-system
spec:
  containers:
  - name: kube-apiserver
    image: k8s.gcr.io/kube-apiserver:v1.24.0
    command:
    - kube-apiserver
    - --oidc-issuer-url=https://keycloak.yourdomain.com/realms/kubernetes-portal
    - --oidc-client-id=portal-backend
    - --oidc-username-claim=sub
    - --oidc-groups-claim=groups
    - --oidc-ca-file=/etc/ssl/certs/ca-certificates.crt
```

### 3.2 RBAC 설정

사용자에게 적절한 권한을 부여합니다.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: oidc-user-binding
subjects:
- kind: User
  name: testuser  # OIDC 토큰의 sub 클레임 값
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: view
  apiGroup: rbac.authorization.k8s.io
```

## 4. 테스트

### 4.1 애플리케이션 실행

```bash
cd portal-backend
go run main.go
```

### 4.2 브라우저에서 테스트

1. http://localhost:8080 접속
2. "Login" 버튼 클릭
3. Keycloak 로그인 페이지에서 `testuser`로 로그인
4. 인증 완료 후 포털로 리디렉션
5. "Launch Web Console" 버튼 클릭

## 5. 문제 해결

### 5.1 일반적인 오류

#### "OIDC 환경 변수가 설정되지 않았습니다"
- 모든 OIDC 환경 변수가 올바르게 설정되었는지 확인
- `.env` 파일이 로드되고 있는지 확인

#### "OIDC provider 생성 실패"
- `OIDC_ISSUER_URL`이 올바른지 확인
- Keycloak이 실행 중인지 확인
- 네트워크 연결 확인

#### "Token exchange failed"
- `OIDC_CLIENT_SECRET`이 올바른지 확인
- Keycloak Client 설정에서 Redirect URI가 올바른지 확인

#### "ID token verification failed"
- Keycloak Client의 Client ID가 올바른지 확인
- 토큰 만료 시간 확인

### 5.2 디버깅

#### 로그 확인
```bash
# 애플리케이션 로그 확인
go run main.go

# Keycloak 로그 확인
docker logs keycloak
```

#### 토큰 검증
```bash
# JWT 토큰 디코딩 (jq 필요)
echo "your-jwt-token" | jq -R 'split(".") | .[1] | @base64d | fromjson'
```

## 6. 보안 고려사항

1. **HTTPS 사용**: 프로덕션에서는 반드시 HTTPS 사용
2. **강력한 비밀번호**: Keycloak 사용자 비밀번호 강화
3. **토큰 만료**: 적절한 토큰 만료 시간 설정
4. **네트워크 정책**: Pod 간 통신 제한
5. **RBAC**: 최소 권한 원칙 적용

## 7. 고급 설정

### 7.1 LDAP 연동

Keycloak에서 LDAP을 사용자 저장소로 설정할 수 있습니다.

### 7.2 다중 클러스터 지원

여러 쿠버네티스 클러스터에 대해 OIDC를 설정할 수 있습니다.

### 7.3 토큰 갱신

Access Token이 만료되면 Refresh Token을 사용하여 자동 갱신할 수 있습니다. 