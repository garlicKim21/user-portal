#!/usr/bin/env node

/**
 * OpenAPI 스펙에서 TypeScript 타입과 API 클라이언트 생성
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const OPENAPI_SPEC_PATH = path.join(__dirname, '../../api/openapi/user-portal-api.yaml');
const OUTPUT_DIR = path.join(__dirname, '../../../poc_front/src/generated');

async function generateTypesAndClient() {
  try {
    console.log('🚀 OpenAPI 코드 생성 시작...');
    
    // 출력 디렉토리 생성
    if (!fs.existsSync(OUTPUT_DIR)) {
      fs.mkdirSync(OUTPUT_DIR, { recursive: true });
    }
    
    // OpenAPI Generator 실행
    const command = `npx @openapitools/openapi-generator-cli generate \
      -i ${OPENAPI_SPEC_PATH} \
      -g typescript-axios \
      -o ${OUTPUT_DIR} \
      --additional-properties=typescriptThreePlus=true,withInterfaces=true,withNodeImports=true`;
    
    console.log('📝 TypeScript 타입 및 API 클라이언트 생성 중...');
    execSync(command, { stdio: 'inherit' });
    
    // 생성된 파일 정리
    cleanupGeneratedFiles();
    
    console.log('✅ 코드 생성 완료!');
    console.log(`📁 출력 위치: ${OUTPUT_DIR}`);
    
  } catch (error) {
    console.error('❌ 코드 생성 실패:', error.message);
    process.exit(1);
  }
}

function cleanupGeneratedFiles() {
  const filesToRemove = [
    'README.md',
    'package.json',
    'tsconfig.json',
    '.openapi-generator',
    '.openapi-generator-ignore'
  ];
  
  filesToRemove.forEach(file => {
    const filePath = path.join(OUTPUT_DIR, file);
    if (fs.existsSync(filePath)) {
      if (fs.statSync(filePath).isDirectory()) {
        fs.rmSync(filePath, { recursive: true });
      } else {
        fs.unlinkSync(filePath);
      }
    }
  });
  
  // index.ts 파일 생성
  const indexContent = `// 자동 생성된 API 클라이언트
export * from './api';
export * from './models';
export * from './base';
export * from './configuration';
`;
  
  fs.writeFileSync(path.join(OUTPUT_DIR, 'index.ts'), indexContent);
}

// 스크립트 실행
if (require.main === module) {
  generateTypesAndClient();
}

module.exports = { generateTypesAndClient };
