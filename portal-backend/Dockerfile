# 빌드 스테이지
FROM golang:1.21-alpine AS builder

# 필요한 패키지 설치
RUN apk add --no-cache git ca-certificates tzdata

# 작업 디렉토리 설정
WORKDIR /app

# go.mod 및 go.sum 복사
COPY go.mod go.sum ./

# 의존성 다운로드
RUN go mod download

# 소스 코드 복사
COPY . .

# 크로스 플랫폼 빌드 (AMD64)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main .

# 실행 스테이지
FROM alpine:latest

# 필요한 패키지 설치
RUN apk --no-cache add ca-certificates tzdata

# 사용자 생성
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 작업 디렉토리 설정
WORKDIR /app

# 빌드된 바이너리 복사
COPY --from=builder /app/main .

# 템플릿 및 정적 파일 복사
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# 소유권 변경
RUN chown -R appuser:appgroup /app

# 사용자 전환
USER appuser

# 포트 노출
EXPOSE 8080

# 애플리케이션 실행
CMD ["./main"] 