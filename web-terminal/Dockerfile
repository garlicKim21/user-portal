# 기본 Alpine 이미지로 시작
FROM alpine:3.18

# 필요한 도구 설치 (bash, kubectl 등)
RUN apk update && apk add --no-cache \
    bash \
    curl \
    net-tools \
    git \
    ca-certificates \
    bash-completion \
    unzip

# kubectl 최신 버전 설치
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
    rm kubectl

# kubelogin (kubectl oidc-login) 플러그인 설치
RUN curl -LO "https://github.com/int128/kubelogin/releases/download/v1.33.0/kubelogin_linux_amd64.zip" && \
    unzip kubelogin_linux_amd64.zip && \
    install -o root -g root -m 0755 kubelogin /usr/local/bin/kubectl-oidc_login && \
    rm kubelogin kubelogin_linux_amd64.zip

# ttyd 바이너리 설치 (최신 1.7.7 버전)
RUN curl -L "https://github.com/tsl0922/ttyd/releases/download/1.7.7/ttyd.x86_64" -o /usr/local/bin/ttyd && \
    chmod +x /usr/local/bin/ttyd

# 보안을 위한 전용 사용자 생성
RUN adduser -D user && \
    mkdir -p /home/user/.kube && \
    chown -R user:user /home/user

# 사용자별 bash 설정 파일 생성 (Alpine에서는 수동 생성 필요)
RUN touch /home/user/.bashrc && \
    echo "# kubectl 자동완성 설정" >> /home/user/.bashrc && \
    echo "source <(kubectl completion bash)" >> /home/user/.bashrc && \
    echo "alias k=kubectl" >> /home/user/.bashrc && \
    echo "complete -o default -F __start_kubectl k" >> /home/user/.bashrc && \
    echo "" >> /home/user/.bashrc && \
    echo "# 환경변수 설정" >> /home/user/.bashrc && \
    echo "export KUBECONFIG=~/.kube/config" >> /home/user/.bashrc && \
    echo "export PS1='\\u@\\h:\\w\$ '" >> /home/user/.bashrc && \
    echo "export PATH=\$PATH:/usr/local/bin" >> /home/user/.bashrc && \
    chown user:user /home/user/.bashrc && \
    chmod 644 /home/user/.bashrc

# 사용자 전환
USER user
WORKDIR /home/user

# 기본 포트 노출
EXPOSE 8080

# 기본 실행 명령어 (Pod Spec에서 덮어쓸 수 있음)
# --writable: 쓰기 가능, 인증 없음 (Kubernetes 자체 인증 사용)
CMD ["ttyd", "--port", "8080", "--interface", "0.0.0.0", "--writable", "bash"]