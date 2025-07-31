# 아키텍처 마이그레이션 가이드

## 현재 PoC → 실제 아키텍처 수정사항

### 1. 쿠버네티스 클라이언트 설정 수정

**현재**: 단일 클러스터 (in-cluster config)
**수정**: A 클러스터에서 B 클러스터로 연결

```go
// portal-backend/internal/kubernetes/client.go 수정 필요
type Client struct {
    Clientset        *kubernetes.Clientset
    TargetClientset  *kubernetes.Clientset  // B 클러스터용 추가
    InClusterConfig  *rest.Config           // A 클러스터 (현재)
    TargetConfig     *rest.Config           // B 클러스터 (추가)
}
```

### 2. 환경 변수 추가

```bash
# A 클러스터에서 실행되는 포털 앱용 환경 변수
TARGET_CLUSTER_SERVER=https://b-cluster-api-server:6443
TARGET_CLUSTER_CA_CERT=<B클러스터-CA-인증서>
TARGET_CLUSTER_TOKEN=<서비스계정-토큰>

# 또는 kubeconfig 파일 경로
TARGET_KUBECONFIG_PATH=/etc/kubeconfig/target-cluster
```

### 3. B 클러스터 API 서버 설정

B 클러스터의 kube-apiserver 매니페스트에 OIDC 설정 추가:

```yaml
# /etc/kubernetes/manifests/kube-apiserver.yaml
spec:
  containers:
  - command:
    - kube-apiserver
    - --oidc-issuer-url=https://keycloak.basphere.dev/realms/kubernetes-portal
    - --oidc-client-id=portal-backend
    - --oidc-username-claim=preferred_username
    - --oidc-groups-claim=groups
    - --oidc-ca-file=/etc/ssl/certs/ca-certificates.crt
```

### 4. 네트워킹 고려사항

#### A → B 클러스터 통신
- A 클러스터의 포털 Pod가 B 클러스터 API 서버에 접근 가능해야 함
- 네트워크 정책, 방화벽 규칙 확인 필요

#### 사용자 → 웹 콘솔 접근
- B 클러스터의 웹 콘솔 Pod에 외부 접근 가능한 Ingress/LoadBalancer 필요
- 포트포워딩 또는 NodePort 서비스 고려

### 5. RBAC 설정

#### A 클러스터 (포털 앱용)
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: portal-service-account
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: target-cluster-manager
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps"]
  verbs: ["create", "delete", "get", "list", "watch"]
```

#### B 클러스터 (웹 콘솔용)
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: oidc-cluster-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: User
  name: testuser  # LDAP 사용자명
  apiGroup: rbac.authorization.k8s.io
```

## 테스트 단계별 진행 계획

### Phase 1: 기본 연결 테스트
1. A 클러스터에서 B 클러스터 API 서버 접근 확인
2. 포털 앱 배포 및 B 클러스터 연결 테스트

### Phase 2: 인증 통합 테스트  
1. B 클러스터 API 서버 OIDC 설정 적용
2. Keycloak 토큰으로 B 클러스터 인증 테스트

### Phase 3: 웹 콘솔 배포 테스트
1. A 클러스터 포털에서 B 클러스터에 웹 콘솔 Pod 생성
2. 네트워킹 및 접근성 확인

### Phase 4: 통합 테스트
1. 전체 사용자 플로우 테스트
2. 다중 사용자 동시 접근 테스트