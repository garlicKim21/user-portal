# 다중 클러스터 테스트 가이드

## 🎯 테스트 환경 준비

### Prerequisites
- ✅ A 클러스터: Keycloak 실행 중
- ✅ B 클러스터: API 서버 매니페스트 수정 완료
- 🔧 네트워크 연결: A 클러스터 ↔ B 클러스터

## 📋 Phase 1: B 클러스터 설정

### 1. B 클러스터 API 서버 OIDC 활성화

```bash
# B 클러스터 마스터 노드에서 실행
sudo vim /etc/kubernetes/manifests/kube-apiserver.yaml
```

추가할 설정:
```yaml
spec:
  containers:
  - command:
    - kube-apiserver
    # 기존 설정들...
    - --oidc-issuer-url=https://keycloak.basphere.dev/realms/kubernetes-portal
    - --oidc-client-id=portal-backend
    - --oidc-username-claim=preferred_username
    - --oidc-groups-claim=groups
```

### 2. B 클러스터 RBAC 설정

```bash
# B 클러스터에 적용
kubectl apply -f deployment/b-cluster-rbac.yaml
```

### 3. 웹 콘솔용 네임스페이스 생성

```bash
kubectl create namespace web-console
kubectl label namespace web-console purpose=web-console-pods
```

## 📋 Phase 2: A 클러스터 준비

### 1. 타겟 클러스터 접근 토큰 생성

```bash
# B 클러스터에서 서비스 계정 생성
kubectl create serviceaccount portal-cross-cluster -n kube-system

# 클러스터 권한 부여
kubectl create clusterrolebinding portal-cross-cluster-binding \
  --clusterrole=cluster-admin \
  --serviceaccount=kube-system:portal-cross-cluster

# 토큰 생성 (Kubernetes 1.24+)
kubectl create token portal-cross-cluster -n kube-system --duration=8760h > target-cluster-token.txt

# 또는 수동으로 토큰 시크릿 생성 (이전 버전)
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: portal-cross-cluster-token
  namespace: kube-system
  annotations:
    kubernetes.io/service-account.name: portal-cross-cluster
type: kubernetes.io/service-account-token
EOF

# 토큰 추출
kubectl get secret portal-cross-cluster-token -n kube-system -o jsonpath='{.data.token}' | base64 -d > target-cluster-token.txt
```

### 2. CA 인증서 추출

```bash
# B 클러스터의 CA 인증서 추출
kubectl get configmap kube-root-ca.crt -o jsonpath='{.data.ca\.crt}' > target-cluster-ca.crt
```

### 3. A 클러스터에 시크릿 생성

```bash
# A 클러스터에서 실행
kubectl create secret generic portal-secrets \
  --from-literal=OIDC_CLIENT_SECRET="your-keycloak-client-secret" \
  --from-literal=JWT_SECRET_KEY="your-jwt-secret-key" \
  --from-file=TARGET_CLUSTER_TOKEN=target-cluster-token.txt

kubectl create secret generic target-cluster-ca-cert \
  --from-file=ca.crt=target-cluster-ca.crt
```

## 📋 Phase 3: 포털 애플리케이션 배포

### 1. Docker 이미지 빌드

```bash
# 프로젝트 루트에서
cd portal-backend
docker build -t your-registry/portal-backend:multi-cluster .
docker push your-registry/portal-backend:multi-cluster
```

### 2. 배포 매니페스트 수정

```bash
# deployment/test-setup.yaml 편집
vim deployment/test-setup.yaml
```

수정할 부분:
```yaml
spec:
  containers:
  - name: portal-backend
    image: your-registry/portal-backend:multi-cluster  # 실제 이미지 경로
    envFrom:
    - configMapRef:
        name: portal-config
    env:
    - name: TARGET_CLUSTER_SERVER
      value: "https://B-CLUSTER-API-SERVER-IP:6443"  # 실제 B 클러스터 API 서버 주소
    - name: TARGET_CLUSTER_TOKEN
      valueFrom:
        secretKeyRef:
          name: portal-secrets
          key: TARGET_CLUSTER_TOKEN
```

### 3. A 클러스터에 배포

```bash
# ConfigMap 수정 및 적용
kubectl create configmap portal-config \
  --from-literal=TARGET_CLUSTER_SERVER="https://B-CLUSTER-API-SERVER-IP:6443" \
  --from-literal=WEB_CONSOLE_BASE_URL="https://console.basphere.dev" \
  --from-literal=LOG_LEVEL="DEBUG" \
  --from-literal=OIDC_ISSUER_URL="https://keycloak.basphere.dev/realms/kubernetes-portal" \
  --from-literal=OIDC_CLIENT_ID="portal-backend" \
  --from-literal=ALLOWED_ORIGINS="https://portal.basphere.dev"

# 포털 애플리케이션 배포
kubectl apply -f deployment/test-setup.yaml
```

## 📋 Phase 4: 네트워킹 설정

### 1. B 클러스터 웹 콘솔 Ingress

```yaml
# B 클러스터에 적용
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: web-console-ingress
  namespace: web-console
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/rewrite-target: /$2
spec:
  tls:
  - hosts:
    - console.basphere.dev
    secretName: console-tls-cert
  rules:
  - host: console.basphere.dev
    http:
      paths:
      - path: /(.*)
        pathType: Prefix
        backend:
          service:
            name: console-service-$1  # 동적 서비스명 지원 필요
            port:
              number: 80
```

## 📋 Phase 5: 테스트 실행

### 1. 연결 테스트

```bash
# A 클러스터 포털 Pod에서 B 클러스터 연결 테스트
kubectl exec -it $(kubectl get pod -l app=portal-backend -o jsonpath='{.items[0].metadata.name}') -- \
  curl -k -H "Authorization: Bearer $(cat /var/run/secrets/target-cluster/token)" \
  https://B-CLUSTER-API-SERVER-IP:6443/api/v1/namespaces
```

### 2. 포털 애플리케이션 접근

```bash
# 포털 서비스 확인
kubectl get svc portal-backend-service

# 포트포워딩으로 테스트 (선택사항)
kubectl port-forward svc/portal-backend-service 8080:80
```

### 3. 사용자 플로우 테스트

1. 🌐 브라우저에서 `https://portal.basphere.dev` 접속
2. 🔐 "Login" 버튼 클릭 → Keycloak 리디렉션
3. 👤 LDAP 계정으로 로그인 (예: testuser)
4. 🚀 "Launch Web Console" 버튼 클릭
5. 🖥️ 새 탭에서 웹 터미널 확인
6. ⚡ `kubectl get pods` 명령 실행 → B 클러스터 리소스 확인

## 🔍 문제 해결

### 로그 확인

```bash
# 포털 애플리케이션 로그
kubectl logs -f -l app=portal-backend

# B 클러스터 API 서버 로그
sudo tail -f /var/log/containers/kube-apiserver-*.log
```

### 일반적인 문제들

1. **네트워크 연결 실패**
   - A 클러스터에서 B 클러스터 API 서버 접근 가능한지 확인
   - 방화벽, 보안 그룹 설정 확인

2. **OIDC 인증 실패**
   - B 클러스터 API 서버 OIDC 설정 확인
   - Keycloak 클라이언트 설정 확인

3. **RBAC 권한 오류**
   - B 클러스터 RBAC 설정 확인
   - 사용자 권한 바인딩 확인

## 🎯 성공 기준

- ✅ A 클러스터 포털에서 B 클러스터에 웹 콘솔 Pod 생성
- ✅ 웹 콘솔에서 B 클러스터 리소스 조회/조작 가능
- ✅ 다중 사용자 동시 접근 정상 동작
- ✅ 세션 종료 시 B 클러스터 리소스 정리