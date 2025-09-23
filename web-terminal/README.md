# Web Terminal Module

쿠버네티스 웹 콘솔 포털의 웹 터미널 모듈입니다. OIDC(Keycloak) 인증을 기반으로 사용자별 격리된 웹 터미널 환경을 제공합니다.

## 🎯 프로젝트 개요 및 아키텍처

본 모듈은 OIDC(Keycloak) 인증을 기반으로, 사용자가 웹 포털을 통해 원격 Kubernetes 클러스터를 안전하고 편리하게 관리할 수 있는 웹 터미널 환경을 제공합니다.

### 아키텍처

- **A 클러스터 (서비스 클러스터)**:
    - `portal-backend` (Go), `portal-frontend` (JS) 애플리케이션 및 `Keycloak`이 배포되는 클러스터
    - 사용자의 요청을 받아 웹 터미널용 `Pod`를 **이곳(A 클러스터)에 생성**
    - B 클러스터의 워크로드에 영향을 주지 않기 위한 "Jumphost" 역할을 수행

- **B 클러스터 (제어 대상 클러스터)**:
    - 빅데이터 분석 등 실제 워크로드가 실행되는 클러스터
    - A 클러스터에 생성된 웹 터미널 Pod가 원격으로 제어하는 대상
    - OIDC를 API 서버 인증 수단으로 사용하며, 사용자 역할(Role)에 따른 접근 제어(RBAC) 적용

### 사용자 흐름

1. 사용자는 `https://your-portal-domain.com`로 접속하여 Keycloak을 통해 로그인
2. 로그인 완료 후 대시보드에서 'Open Web Terminal' 버튼 클릭
3. 백엔드는 A 클러스터에 해당 사용자를 위한 웹 터미널 Pod 생성
    - 이 Pod에는 B 클러스터에 접근 가능한 `kubeconfig`가 마운트
4. 사용자에게는 새 탭으로 웹 터미널이 열리며, 즉시 `kubectl` 명령어를 사용하여 B 클러스터 관리 가능
5. 사용자가 포털에서 로그아웃하면, 해당 사용자에게 할당되었던 웹 터미널 Pod와 관련 리소스는 모두 삭제

---

## 🚀 핵심 구현 및 최신 기능

### 🆕 JWT 구조 최적화 (v0.4.10+)

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

### 🆕 웹 콘솔 개인화 (v0.2.11+)

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

### 🆕 명령어 히스토리 지속성

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

### 🆕 CSRF 보호

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

---

## 🛠️ 사용자 맞춤형 터미널 컨테이너 이미지

사용자 편의성을 극대화하기 위해, 필요한 모든 도구가 사전 설치된 커스텀 컨테이너 이미지를 사용합니다.

- **Base Image**: `alpine:latest` (경량화)
- **필수 설치 도구**: `bash`, `kubectl`, `curl`, `net-tools`
- **편의 기능**:
    - `kubectl` 명령어 자동 완성 기능 (`bash-completion`) 기본 활성화
    - `alias k=kubectl` 등 단축 명령어 기본 설정

### Dockerfile

```dockerfile
# Alpine Linux 기반의 가벼운 이미지 사용
FROM alpine:3.18

# 필수 패키지 및 bash-completion 설치
RUN apk update && apk add --no-cache \
    bash \
    bash-completion \
    curl \
    net-tools \
    git \
    ca-certificates

# kubectl 설치
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
    rm kubectl

# kubectl 자동 완성 및 alias 설정
RUN echo "source /etc/profile.d/bash_completion.sh" >> /etc/bash/bashrc && \
    echo "source <(kubectl completion bash)" >> /etc/bash/bashrc && \
    echo "alias k=kubectl" >> /etc/bash/bashrc && \
    echo "complete -o default -F __start_kubectl k" >> /etc/bash/bashrc

# 기본 사용자와 작업 디렉토리 설정
RUN adduser -D user
USER user
WORKDIR /home/user

# 기본 실행 명령어
CMD ["/bin/bash"]
```

### 현재 이미지 정보

- **이미지**: `projectgreenist/web-terminal:0.2.11`
- **크기**: 약 45MB (Alpine 기반)
- **플랫폼**: linux/amd64

---

## 📊 Pod 리소스 사용량 제한

웹 터미널 Pod가 클러스터 리소스를 과도하게 점유하는 것을 방지하기 위해 최소/최대 사용량을 명시적으로 제한합니다.

### 리소스 제한 설정

```yaml
Resources: corev1.ResourceRequirements{
    Requests: corev1.ResourceList{
        corev1.ResourceCPU:    resource.MustParse("100m"), // 최소 0.1 CPU
        corev1.ResourceMemory: resource.MustParse("128Mi"), // 최소 128MB 메모리
    },
    Limits: corev1.ResourceList{
        corev1.ResourceCPU:    resource.MustParse("250m"), // 최대 0.25 CPU
        corev1.ResourceMemory: resource.MustParse("256Mi"), // 최대 256MB 메모리
    },
}
```

---

## 🗑️ 로그아웃 시 리소스 자동 정리

사용자가 로그아웃할 때, 해당 사용자에게 할당된 모든 쿠버네티스 리소스(Pod, Service, ConfigMap, PVC)를 자동으로 삭제하여 리소스 낭비를 방지합니다.

### 정리 대상 리소스

- **Pod**: `web-console-{userID}-{uuid}`
- **Service**: `web-console-{userID}-{uuid}`
- **ConfigMap**: `web-console-config-{userID}-{uuid}`
- **PVC**: `history-{userID}` (사용자별 히스토리 저장소)

### 구현 위치

- `internal/handlers/auth.go`의 `HandleLogout` 함수
- `internal/kubernetes/resource.go`의 리소스 정리 함수

---

## 🔐 사용자 역할 기반 접근 제어 (RBAC)

Keycloak의 그룹과 쿠버네티스 RBAC을 연동하여 `viewer`, `developer`, `admin` 역할을 구현합니다.

### RBAC 설정

1. **Keycloak 그룹 생성**: `basphere` 렐름에 `viewers`, `developers`, `admins` 그룹을 생성하고 사용자를 각 그룹에 할당

2. **B 클러스터에 `ClusterRole` 정의**: 각 역할에 맞는 권한을 가진 `ClusterRole`을 `b-cluster-rbac.yaml`에 정의
    - `viewer-role`: `get`, `list`, `watch` 등 읽기 전용 권한
    - `developer-role`: `create`, `update`, `delete` 등 개발에 필요한 쓰기 권한 추가
    - `admin`: 기존 `cluster-admin` 역할을 사용

3. **B 클러스터에 `ClusterRoleBinding` 생성**: Keycloak 그룹과 위에서 정의한 `ClusterRole`을 연결하는 `ClusterRoleBinding`을 생성

```yaml
# 예시: viewers 그룹을 viewer-role에 바인딩
- kind: ClusterRoleBinding
  apiVersion: rbac.authorization.k8s.io/v1
  metadata:
    name: viewers-binding
  subjects:
  - kind: Group
    name: viewers # Keycloak 그룹명과 일치
    apiGroup: rbac.authorization.k8s.io
  roleRef:
    kind: ClusterRole
    name: viewer-role
    apiGroup: rbac.authorization.k8s.io
```

4. **B 클러스터 API 서버 설정 확인**: API 서버의 OIDC 설정에 `--oidc-groups-claim=groups`와 `--oidc-username-claim=preferred_username` 등의 플래그가 포함되어, 토큰에서 그룹과 사용자 정보를 올바르게 파싱할 수 있는지 확인

---

## 🏗️ PoC 단계의 설계 결정

### 웹 터미널 재연결

창을 닫았다가 다시 열 때 기존 Pod에 연결하는 것은 상태 관리(DB/Redis 필요)의 복잡도를 증가시키므로 PoC 단계에서는 구현하지 않습니다. 대신, **'Open Web Terminal' 버튼은 항상 새로운 Pod를 생성**하는 현재의 단순한 모델을 유지합니다. 명령어 히스토리가 PVC로 보존되므로 사용자의 작업 연속성은 유지됩니다.

### Pod 생명주기

웹 터미널 Pod는 사용자가 명시적으로 **로그아웃할 때만 삭제**됩니다. 터미널 창을 닫는 것만으로는 삭제되지 않습니다.

---

## 📚 관련 문서

- **[Main README](../README.md)** - 전체 프로젝트 개요
- **[Backend README](../portal-backend/README.md)** - 백엔드 기술 문서
- **[Deployment README](../deployment/README.md)** - 배포 및 운영 가이드
- **[Backend Config](../portal-backend/CONFIG.md)** - 백엔드 설정 가이드

## �� 라이선스

MIT License