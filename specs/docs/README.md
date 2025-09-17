# User Portal Spec Driven Development 가이드

## 개요
이 디렉토리는 User Portal 프로젝트의 Spec Driven Development를 위한 모든 스펙과 문서를 포함합니다.

## 디렉토리 구조

```
specs/
├── api/                    # API 스펙
│   ├── openapi/           # OpenAPI 3.0 스펙
│   ├── postman/           # Postman 컬렉션
│   └── contracts/         # API 계약서 (Pact)
├── frontend/              # 프론트엔드 스펙
│   ├── component-specs/   # 컴포넌트 스펙
│   ├── storybook/         # Storybook 스토리
│   └── user-stories/      # 사용자 스토리
├── integration/           # 통합 테스트 스펙
│   ├── e2e/              # E2E 테스트
│   ├── api-tests/        # API 테스트
│   └── performance/      # 성능 테스트
├── deployment/            # 배포 스펙
│   ├── k8s/              # Kubernetes 매니페스트
│   ├── helm/             # Helm 차트
│   └── docker/           # Docker 설정
├── tools/                # 개발 도구
│   ├── codegen/          # 코드 생성 도구
│   ├── validation/       # 스펙 검증 도구
│   └── testing/          # 테스트 도구
├── docs/                 # 문서
└── scripts/              # 스크립트
```

## 개발 워크플로우

### 1. 스펙 작성
1. **API 스펙**: `api/openapi/`에서 OpenAPI 스펙 작성
2. **컴포넌트 스펙**: `frontend/component-specs/`에서 컴포넌트 요구사항 정의
3. **사용자 스토리**: `frontend/user-stories/`에서 사용자 요구사항 정의

### 2. 코드 생성
```bash
# TypeScript 타입 및 API 클라이언트 생성
npm run generate:api

# 컴포넌트 스토리 생성
npm run generate:stories
```

### 3. 테스트 작성
```bash
# 단위 테스트
npm run test:unit

# 통합 테스트
npm run test:integration

# E2E 테스트
npm run test:e2e
```

### 4. 스펙 검증
```bash
# API 스펙 검증
npm run validate:specs

# Contract 테스트
npm run test:contracts
```

## 주요 도구

### OpenAPI Generator
- **위치**: `tools/codegen/openapi-generator.js`
- **기능**: OpenAPI 스펙에서 TypeScript 타입과 API 클라이언트 생성
- **사용법**: `node tools/codegen/openapi-generator.js`

### Spec Validator
- **위치**: `tools/validation/spec-validator.js`
- **기능**: API 스펙과 Contract 검증
- **사용법**: `node tools/validation/spec-validator.js`

## 스펙 관리 원칙

### 1. 단일 진실의 원천 (Single Source of Truth)
- 모든 API 스펙은 OpenAPI 파일에서 관리
- 컴포넌트 스펙은 Markdown 파일로 관리
- 사용자 스토리는 이슈 트래커와 연동

### 2. 버전 관리
- 모든 스펙 파일은 Git으로 버전 관리
- 변경사항은 Pull Request를 통해 검토
- 주요 변경사항은 CHANGELOG에 기록

### 3. 자동화
- 스펙 변경 시 자동으로 코드 생성
- CI/CD 파이프라인에서 스펙 검증
- 테스트 자동 실행

## 품질 보증

### 코드 품질
- ESLint, Prettier 설정
- TypeScript strict 모드
- 코드 커버리지 80% 이상

### 테스트 품질
- 단위 테스트: 90% 이상 커버리지
- 통합 테스트: 주요 플로우 100% 커버
- E2E 테스트: 사용자 시나리오 100% 커버

### 문서 품질
- 모든 API 엔드포인트 문서화
- 모든 컴포넌트 스토리북 문서화
- 사용자 가이드 및 개발자 가이드 제공

## 협업 가이드

### 개발자
1. 기능 개발 전 스펙 검토
2. 스펙에 따른 테스트 작성
3. 코드 리뷰 시 스펙 준수 확인

### QA
1. 스펙 기반 테스트 케이스 작성
2. E2E 테스트 시나리오 검증
3. 사용자 스토리 검증

### 기획자
1. 사용자 스토리 작성 및 검토
2. API 스펙 요구사항 정의
3. 사용자 시나리오 검증

## 문제 해결

### 자주 발생하는 문제
1. **스펙 불일치**: 스펙과 구현 간 차이 발생 시 스펙 우선
2. **테스트 실패**: 스펙 변경 시 테스트 업데이트 필요
3. **코드 생성 실패**: OpenAPI 스펙 문법 오류 확인

### 지원 채널
- 기술 문의: 개발팀 Slack 채널
- 버그 리포트: GitHub Issues
- 기능 요청: GitHub Discussions

## 참고 자료

- [OpenAPI Specification](https://swagger.io/specification/)
- [Storybook Documentation](https://storybook.js.org/docs/)
- [Playwright Testing](https://playwright.dev/)
- [Pact Contract Testing](https://pact.io/)
