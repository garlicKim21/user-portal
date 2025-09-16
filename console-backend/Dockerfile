# Go 애플리케이션을 위한 Dockerfile
FROM golang:1.24-alpine AS builder

# 작업 디렉토리 설정
WORKDIR /app

# Go 모듈 파일 복사
COPY go.mod go.sum ./

# 의존성 다운로드
RUN go mod download

# 소스 코드 복사
COPY . .

# 애플리케이션 빌드 (AMD64 크로스 컴파일)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main .

# 최종 이미지
FROM alpine:latest

# ca-certificates 설치 (HTTPS 요청용)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 빌드된 바이너리 복사
COPY --from=builder /app/main .

# 포트 노출
EXPOSE 8080

# 애플리케이션 실행
CMD ["./main"] 