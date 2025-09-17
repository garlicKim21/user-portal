#!/bin/bash

# Spec Driven Development 환경 설정 스크립트

set -e

echo "🚀 Spec Driven Development 환경 설정을 시작합니다..."

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 함수 정의
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Node.js 버전 확인
check_node_version() {
    log_info "Node.js 버전 확인 중..."
    if ! command -v node &> /dev/null; then
        log_error "Node.js가 설치되어 있지 않습니다. Node.js 18+ 버전을 설치해주세요."
        exit 1
    fi
    
    NODE_VERSION=$(node --version | cut -d'v' -f2 | cut -d'.' -f1)
    if [ "$NODE_VERSION" -lt 18 ]; then
        log_error "Node.js 18+ 버전이 필요합니다. 현재 버전: $(node --version)"
        exit 1
    fi
    
    log_info "Node.js 버전: $(node --version) ✅"
}

# 필요한 패키지 설치
install_dependencies() {
    log_info "필요한 패키지 설치 중..."
    
    # 전역 패키지 설치
    npm install -g @openapitools/openapi-generator-cli
    npm install -g swagger-codegen
    npm install -g @playwright/test
    npm install -g @storybook/cli
    
    log_info "전역 패키지 설치 완료 ✅"
}

# 프로젝트별 의존성 설치
install_project_dependencies() {
    log_info "프로젝트 의존성 설치 중..."
    
    # poc_front 의존성 설치
    if [ -d "poc_front" ]; then
        log_info "poc_front 의존성 설치 중..."
        cd poc_front
        npm install
        cd ..
    fi
    
    # portal-backend 의존성 설치
    if [ -d "portal-backend" ]; then
        log_info "portal-backend 의존성 설치 중..."
        cd portal-backend
        npm install
        cd ..
    fi
    
    log_info "프로젝트 의존성 설치 완료 ✅"
}

# 스펙 검증 설정
setup_validation() {
    log_info "스펙 검증 도구 설정 중..."
    
    # 스펙 검증 스크립트 실행 권한 부여
    chmod +x specs/tools/validation/spec-validator.js
    chmod +x specs/tools/codegen/openapi-generator.js
    
    log_info "스펙 검증 도구 설정 완료 ✅"
}

# Git hooks 설정
setup_git_hooks() {
    log_info "Git hooks 설정 중..."
    
    # pre-commit hook 생성
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
echo "🔍 Pre-commit 검증 중..."

# 스펙 검증
node specs/tools/validation/spec-validator.js

# 코드 포맷팅 검사
if [ -d "poc_front" ]; then
    cd poc_front
    npm run lint
    cd ..
fi

echo "✅ Pre-commit 검증 완료"
EOF
    
    chmod +x .git/hooks/pre-commit
    
    log_info "Git hooks 설정 완료 ✅"
}

# 테스트 환경 설정
setup_testing() {
    log_info "테스트 환경 설정 중..."
    
    # Playwright 브라우저 설치
    if command -v npx &> /dev/null; then
        npx playwright install
    fi
    
    log_info "테스트 환경 설정 완료 ✅"
}

# 메인 실행
main() {
    log_info "Spec Driven Development 환경 설정을 시작합니다..."
    
    check_node_version
    install_dependencies
    install_project_dependencies
    setup_validation
    setup_git_hooks
    setup_testing
    
    log_info "🎉 환경 설정이 완료되었습니다!"
    log_info ""
    log_info "다음 명령어를 사용할 수 있습니다:"
    log_info "  npm run generate:api     # API 코드 생성"
    log_info "  npm run validate:specs   # 스펙 검증"
    log_info "  npm run test:e2e         # E2E 테스트 실행"
    log_info "  npm run storybook        # Storybook 실행"
    log_info ""
    log_info "자세한 내용은 specs/docs/README.md를 참고하세요."
}

# 스크립트 실행
main "$@"
