# 멀티 스테이지 빌드를 위한 Node.js 베이스 이미지
FROM --platform=$BUILDPLATFORM node:18-alpine AS builder

# 작업 디렉토리 설정
WORKDIR /app
RUN mkdir -p dist

# package.json과 package-lock.json 복사
COPY package*.json ./

# 의존성 설치
RUN npm ci

# 소스 코드 복사
COPY . .

# 프로덕션 빌드
RUN npm run build
RUN ls -l /app

# Nginx를 사용한 정적 파일 서빙
FROM nginx:alpine

# 커스텀 nginx 설정 복사
COPY nginx.conf /etc/nginx/nginx.conf

# 빌드된 파일을 nginx 디렉토리로 복사
COPY --from=builder /app/build /usr/share/nginx/html

# 포트 80 노출
EXPOSE 80

# Nginx 실행
CMD ["nginx", "-g", "daemon off;"]
