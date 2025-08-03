FROM alpine:3.18
RUN apk update && apk add --no-cache \
    bash \
    bash-completion \
    curl \
    net-tools \
    git \
    ca-certificates

RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
    rm kubectl

RUN echo "source <(kubectl completion bash)" >> /etc/bash/bashrc && \
    echo "alias k=kubectl" >> /etc/bash/bashrc && \
    echo "complete -o default -F __start_kubectl k" >> /etc/bash/bashrc

RUN adduser -D user
USER user
WORKDIR /home/user

CMD ["/bin/bash"]