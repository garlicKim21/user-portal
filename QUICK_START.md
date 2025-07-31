# 🚀 빠른 시작 가이드

## 현재 상황
- ✅ A 클러스터: Keycloak 실행 중
- ✅ B 클러스터: API 서버 매니페스트 수정 완료
- ✅ 코드: 다중 클러스터 지원으로 수정 완료

## 🎯 다음 단계 (우선순위 순)

### 1. B 클러스터 API 서버 재시작 (가장 중요!)

```bash
# B 클러스터 마스터 노드에서
sudo systemctl restart kubelet
# 또는 API 서버 Pod가 자동으로 재시작될 때까지 대기 (수분 소요)
```

### 2. 타겟 클러스터 인증 토큰 생성

```bash
# B 클러스터에서 실행
kubectl create serviceaccount portal-cross-cluster -n kube-system
kubectl create clusterrolebinding portal-cross-cluster-binding \
  --clusterrole=cluster-admin \
  --serviceaccount=kube-system:portal-cross-cluster

# 토큰 생성 (Kubernetes 1.24+)
kubectl create token portal-cross-cluster -n kube-system --duration=8760h
```

### 3. 연결 테스트

```bash
# B 클러스터 API 서버 주소 확인
kubectl cluster-info

# A 클러스터에서 B 클러스터로 네트워크 연결 테스트
ping B-CLUSTER-API-SERVER-IP
```

### 4. 환경 변수 설정 및 테스트

```bash
# 현재 개발 환경에서 테스트
export TARGET_CLUSTER_SERVER="https://B-CLUSTER-API-SERVER-IP:6443"
export TARGET_CLUSTER_TOKEN="위에서-생성한-토큰"
export LOG_LEVEL="DEBUG"

# 포털 애플리케이션 실행 테스트
cd portal-backend
go run main.go
```

## 📞 문제 발생 시 확인사항

1. **B 클러스터 API 서버 OIDC 설정이 적용되었는지 확인**
   ```bash
   kubectl get pod kube-apiserver-* -n kube-system -o yaml | grep oidc
   ```

2. **네트워크 연결 확인**
   ```bash
   telnet B-CLUSTER-API-SERVER-IP 6443
   ```

3. **토큰 유효성 확인**
   ```bash
   curl -k -H "Authorization: Bearer $TARGET_CLUSTER_TOKEN" \
     https://B-CLUSTER-API-SERVER-IP:6443/api/v1/namespaces
   ```

## 🎉 성공하면...

포털 애플리케이션이 A 클러스터에서 실행되면서 B 클러스터에 웹 콘솔 Pod를 생성할 수 있게 됩니다!

그 다음 단계로는 실제 배포 및 프로덕션 설정을 진행하면 됩니다.