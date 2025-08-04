# User Portal Deployment Guide

이 문서는 User Portal의 Kubernetes 배포 및 Secret 관리에 대한 가이드를 제공합니다.

## 📋 배포 개요

User Portal은 다중 클러스터 환경에서 동작하며, 민감한 정보는 Kubernetes Secret으로 관리됩니다.

### 클러스터 구성

- **A 클러스터**: 포털 애플리케이션 실행
- **B 클러스터**: 웹 콘솔 Pod 생성 대상

## 🔐 Secret 관리

### Secret 구조

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: user-portal-secrets
  namespace: user-portal
type: Opaque
data:
  jwt-secret-key: <base64-encoded-jwt-secret>
  oidc-client-secret: <base64-encoded-oidc-secret>
  kubectl-oidc-client-secret: <base64-encoded-kubectl-secret>
  target-cluster-server: <base64-encoded-cluster-url>
  target-cluster-ca-cert-data: <base64-encoded-ca-cert>
```

### CA 인증서 추가 방법

1. **CA 인증서를 base64로 인코딩**
```bash
# CA 인증서 파일을 base64로 인코딩
cat /path/to/ca.crt | base64 -w 0
```

2. **Secret 업데이트**
```bash
# 기존 Secret에 CA 인증서 추가
kubectl patch secret user-portal-secrets -n user-portal \
  --type='json' \
  -p='[{"op": "add", "path": "/data/target-cluster-ca-cert-data", "value": "<base64-encoded-ca-cert>"}]'
```

또는 Secret을 직접 편집:
```bash
kubectl edit secret user-portal-secrets -n user-portal
```

### Secret 생성 예시

```bash
# 모든 민감 정보를 포함한 Secret 생성
kubectl create secret generic user-portal-secrets \
  --from-literal=jwt-secret-key="your-jwt-secret" \
  --from-literal=oidc-client-secret="your-oidc-secret" \
  --from-literal=kubectl-oidc-client-secret="your-kubectl-secret" \
  --from-literal=target-cluster-server="https://<your-target-server-api-address or fqdn>:6443" \
  --from-literal=target-cluster-ca-cert-data="$(cat /path/to/ca.crt | base64 -w 0)" \
  -n user-portal
```

### Secret 검증

```bash
# Secret 존재 확인
kubectl get secret user-portal-secrets -n user-portal

# Secret 내용 확인 (base64 디코딩)
kubectl get secret user-portal-secrets -n user-portal -o jsonpath='{.data.target-cluster-server}' | base64 -d

# Secret 키 목록 확인
kubectl get secret user-portal-secrets -n user-portal -o jsonpath='{.data}' | jq 'keys'
```

## 🚀 배포 단계

### 1. 네임스페이스 생성

```bash
kubectl create namespace user-portal
```

### 2. Secret 생성

```bash
# Secret 생성
kubectl apply -f user-portal-secrets.yaml

# 또는 명령어로 생성
kubectl create secret generic user-portal-secrets \
  --from-literal=jwt-secret-key="your-jwt-secret" \
  --from-literal=oidc-client-secret="your-oidc-secret" \
  --from-literal=kubectl-oidc-client-secret="your-kubectl-secret" \
  --from-literal=target-cluster-server="https://your-cluster:6443" \
  --from-literal=target-cluster-ca-cert-data="$(cat /path/to/ca.crt | base64 -w 0)" \
  -n user-portal
```

### 3. 백엔드 배포

```bash
# 백엔드 배포
kubectl apply -f user-portal-backend.yaml

# 배포 상태 확인
kubectl get pods -n user-portal
kubectl get services -n user-portal
```

### 4. 서비스 계정 설정 (필요시)

```bash
# RBAC 설정
kubectl apply -f rbac.yaml
```

## 🔧 환경 변수 보안 수준

### 🔴 민감 정보 (Secret 사용)
- `OIDC_CLIENT_SECRET` - OIDC 클라이언트 시크릿
- `JWT_SECRET_KEY` - JWT 서명 키
- `KUBECTL_OIDC_CLIENT_SECRET` - kubectl OIDC 클라이언트 시크릿
- `TARGET_CLUSTER_SERVER` - 타겟 클러스터 API 서버 URL
- `TARGET_CLUSTER_CA_CERT_DATA` - 타겟 클러스터 CA 인증서

### 🟡 중간 민감도 (환경별 고려)
- `WEB_CONSOLE_BASE_URL` - 환경별 URL

### 🟢 비민감 정보 (환경 변수)
- `OIDC_ISSUER_URL` - OIDC 발급자 URL
- `OIDC_CLIENT_ID` - OIDC 클라이언트 ID
- `OIDC_REDIRECT_URL` - OIDC 리다이렉트 URL
- `ALLOWED_ORIGINS` - CORS 허용 오리진
- `LOG_LEVEL` - 로그 레벨
- `GIN_MODE` - Gin 모드
- `CONSOLE_NAMESPACE` - 웹 콘솔 네임스페이스
- `CONSOLE_IMAGE` - 웹 콘솔 이미지

## 🔍 모니터링

### Pod 상태 확인

```bash
# Pod 상태 확인
kubectl get pods -n user-portal

# Pod 로그 확인
kubectl logs -f deployment/user-portal-backend -n user-portal

# Pod 상세 정보
kubectl describe pod <pod-name> -n user-portal
```

### 서비스 상태 확인

```bash
# 서비스 상태 확인
kubectl get services -n user-portal

# 엔드포인트 확인
kubectl get endpoints -n user-portal
```

### Secret 상태 확인

```bash
# Secret 상태 확인
kubectl get secrets -n user-portal

# Secret 상세 정보
kubectl describe secret user-portal-secrets -n user-portal
```

## 🚨 문제 해결

### 일반적인 문제

1. **Pod가 시작되지 않음**
   ```bash
   # Pod 이벤트 확인
   kubectl describe pod <pod-name> -n user-portal
   
   # Pod 로그 확인
   kubectl logs <pod-name> -n user-portal
   ```

2. **Secret 관련 오류**
   ```bash
   # Secret 존재 확인
   kubectl get secret user-portal-secrets -n user-portal
   
   # Secret 키 확인
   kubectl get secret user-portal-secrets -n user-portal -o jsonpath='{.data}' | jq 'keys'
   ```

3. **OIDC 연결 실패**
   ```bash
   # 환경 변수 확인
   kubectl exec <pod-name> -n user-portal -- env | grep OIDC
   
   # 로그에서 OIDC 관련 오류 확인
   kubectl logs <pod-name> -n user-portal | grep -i oidc
   ```

4. **쿠버네티스 연결 실패**
   ```bash
   # 타겟 클러스터 연결 확인
   kubectl exec <pod-name> -n user-portal -- env | grep TARGET
   
   # CA 인증서 확인
   kubectl get secret user-portal-secrets -n user-portal -o jsonpath='{.data.target-cluster-ca-cert-data}' | base64 -d
   ```

### 로그 분석

```bash
# 실시간 로그 모니터링
kubectl logs -f deployment/user-portal-backend -n user-portal

# 특정 시간대 로그
kubectl logs --since=1h deployment/user-portal-backend -n user-portal

# 에러 로그만 확인
kubectl logs deployment/user-portal-backend -n user-portal | grep ERROR

# 특정 키워드 검색
kubectl logs deployment/user-portal-backend -n user-portal | grep -i "console\|auth\|cluster"
```

## 🔄 업데이트 및 롤백

### 배포 업데이트

```bash
# 새 이미지로 업데이트
kubectl set image deployment/user-portal-backend user-portal-backend=projectgreenist/user-portal-backend:0.3.20 -n user-portal

# 업데이트 상태 확인
kubectl rollout status deployment/user-portal-backend -n user-portal
```

### 롤백

```bash
# 이전 버전으로 롤백
kubectl rollout undo deployment/user-portal-backend -n user-portal

# 롤백 상태 확인
kubectl rollout status deployment/user-portal-backend -n user-portal
```

### Secret 업데이트

```bash
# Secret 업데이트
kubectl patch secret user-portal-secrets -n user-portal \
  --type='json' \
  -p='[{"op": "replace", "path": "/data/target-cluster-server", "value": "<new-base64-value>"}]'

# Pod 재시작 (Secret 변경사항 적용)
kubectl rollout restart deployment/user-portal-backend -n user-portal
```

## 📊 성능 모니터링

### 리소스 사용량 확인

```bash
# Pod 리소스 사용량
kubectl top pods -n user-portal

# 노드 리소스 사용량
kubectl top nodes
```

### 메트릭 수집

```bash
# Pod 메트릭
kubectl exec <pod-name> -n user-portal -- curl http://localhost:8080/debug/pprof/heap

# 고루틴 상태
kubectl exec <pod-name> -n user-portal -- curl http://localhost:8080/debug/pprof/goroutine
```

## 🔒 보안 체크리스트

- [ ] Secret이 올바르게 생성되었는지 확인
- [ ] CA 인증서가 올바르게 인코딩되었는지 확인
- [ ] RBAC 권한이 적절히 설정되었는지 확인
- [ ] 네트워크 정책이 설정되었는지 확인
- [ ] 리소스 제한이 설정되었는지 확인
- [ ] 헬스체크가 작동하는지 확인
- [ ] 로그 레벨이 적절히 설정되었는지 확인

## 📚 관련 문서

- **[Backend README](../portal-backend/README.md)** - 백엔드 상세 가이드
- **[Configuration Guide](../portal-backend/CONFIG.md)** - 설정 가이드
- **[OIDC Setup](../portal-backend/OIDC_SETUP.md)** - OIDC 설정 가이드