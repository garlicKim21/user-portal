#!/usr/bin/env node

/**
 * API 스펙 검증 도구
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const SPECS_DIR = path.join(__dirname, '../../api');
const OPENAPI_SPEC = path.join(SPECS_DIR, 'openapi/user-portal-api.yaml');

async function validateSpecs() {
  try {
    console.log('🔍 API 스펙 검증 시작...');
    
    // OpenAPI 스펙 검증
    await validateOpenAPISpec();
    
    // Contract 검증
    await validateContracts();
    
    console.log('✅ 모든 스펙 검증 완료!');
    
  } catch (error) {
    console.error('❌ 스펙 검증 실패:', error.message);
    process.exit(1);
  }
}

async function validateOpenAPISpec() {
  console.log('📋 OpenAPI 스펙 검증 중...');
  
  try {
    // swagger-codegen으로 스펙 검증
    execSync(`npx swagger-codegen validate -i ${OPENAPI_SPEC}`, { stdio: 'inherit' });
    console.log('✅ OpenAPI 스펙이 유효합니다.');
  } catch (error) {
    console.error('❌ OpenAPI 스펙 검증 실패:', error.message);
    throw error;
  }
}

async function validateContracts() {
  console.log('📋 Contract 검증 중...');
  
  const contractsDir = path.join(SPECS_DIR, 'contracts');
  const contractFiles = fs.readdirSync(contractsDir).filter(file => file.endsWith('.json'));
  
  for (const file of contractFiles) {
    const contractPath = path.join(contractsDir, file);
    console.log(`  📄 ${file} 검증 중...`);
    
    try {
      const contract = JSON.parse(fs.readFileSync(contractPath, 'utf8'));
      
      // Contract 구조 검증
      validateContractStructure(contract);
      
      console.log(`  ✅ ${file} 유효합니다.`);
    } catch (error) {
      console.error(`  ❌ ${file} 검증 실패:`, error.message);
      throw error;
    }
  }
}

function validateContractStructure(contract) {
  const requiredFields = ['name', 'version', 'provider', 'consumer', 'interactions'];
  
  for (const field of requiredFields) {
    if (!contract[field]) {
      throw new Error(`Contract에 필수 필드 '${field}'가 없습니다.`);
    }
  }
  
  // Interactions 검증
  if (!Array.isArray(contract.interactions)) {
    throw new Error('Interactions는 배열이어야 합니다.');
  }
  
  for (const interaction of contract.interactions) {
    if (!interaction.description || !interaction.request || !interaction.response) {
      throw new Error('각 interaction은 description, request, response가 필요합니다.');
    }
  }
}

// 스크립트 실행
if (require.main === module) {
  validateSpecs();
}

module.exports = { validateSpecs };
