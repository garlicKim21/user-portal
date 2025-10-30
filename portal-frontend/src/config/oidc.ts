// OIDC 설정 파일 (react-oidc-context용)
import env, { getKeycloakAuthority, getKeycloakEndpoints } from './env';
import { WebStorageStateStore } from 'oidc-client-ts';

// Keycloak 엔드포인트 생성
const keycloakEndpoints = getKeycloakEndpoints();

// === 디버깅 로그 시작 ===
console.log('🟢 [oidc.ts] ===== OIDC 설정 생성 시작 =====');
console.log('🟢 [oidc.ts] Keycloak 기본 정보:');
console.log('  - KEYCLOAK_URL:', env.KEYCLOAK_URL);
console.log('  - KEYCLOAK_REALM:', env.KEYCLOAK_REALM);
console.log('  - KEYCLOAK_CLIENT_ID:', env.KEYCLOAK_CLIENT_ID);
console.log('  - authority:', getKeycloakAuthority());

console.log('🟢 [oidc.ts] Keycloak 엔드포인트:');
console.log('  - issuer:', keycloakEndpoints.issuer);
console.log('  - authorization:', keycloakEndpoints.authorization);
console.log('  - token:', keycloakEndpoints.token);
console.log('  - userinfo:', keycloakEndpoints.userinfo);
console.log('  - endSession:', keycloakEndpoints.endSession);

console.log('🟢 [oidc.ts] Redirect URI 설정:');
console.log('  - window.location.origin:', window.location.origin);
console.log('  - redirect_uri (계산값):', window.location.origin + '/callback');
console.log('  - post_logout_redirect_uri (계산값):', window.location.origin);
console.log('  - silent_redirect_uri (계산값):', window.location.origin + '/silent-callback');
// === 디버깅 로그 끝 ===

export const oidcConfig = {
  authority: getKeycloakAuthority(),
  client_id: env.KEYCLOAK_CLIENT_ID,
  client_secret: env.KEYCLOAK_CLIENT_SECRET,
  redirect_uri: window.location.origin + '/callback',
  post_logout_redirect_uri: window.location.origin,
  response_type: 'code',
  scope: 'openid profile email',
  // HTTP 개발 환경: UserInfo 로드 비활성화 (Crypto.subtle 사용 방지)
  loadUserInfo: false,
  // Silent callback 설정 (자동 갱신 비활성화했으므로 불필요하지만 유지)
  silent_redirect_uri: window.location.origin + '/silent-callback',
  // 로그 레벨 설정
  logLevel: 'debug' as const,
  // HTTP 개발 환경: 세션 체크 비활성화 (Crypto.subtle 사용 방지)
  monitorSession: false,
  checkSessionInterval: 0,  // 비활성화
  silentRequestTimeout: 10000,
  // PKCE 비활성화 (client secret 사용)
  // HTTP 환경에서는 Crypto.subtle API를 사용할 수 없으므로 PKCE 비활성화 필수
  pkce: false,
  
  // HTTP 개발 환경을 위한 보안 검증 비활성화 (프로덕션에서는 절대 사용 금지!)
  // ID Token 서명 검증 비활성화 - Crypto.subtle 사용 안 함
  validateSubOnSilentRenew: false,
  
  // 세션 만료 시간 (30분)
  accessTokenExpiringNotificationTime: 300,
  // 로그아웃 관련 설정
  revokeTokensOnSignout: false,  // HTTP 환경에서는 비활성화
  includeIdTokenInSilentRenew: false,  // HTTP 환경에서는 비활성화
  
  // HTTP 환경 지원을 위한 추가 설정
  // state 관리에서도 Crypto API를 사용하지 않도록 명시적으로 WebStorageStateStore 사용
  stateStore: new WebStorageStateStore({ store: window.sessionStorage }),
  userStore: new WebStorageStateStore({ store: window.sessionStorage }),
  
  // 자동 토큰 갱신 비활성화 (HTTP 환경에서 문제 발생 가능)
  automaticSilentRenew: false,
  // react-oidc-context 특별 설정
  metadata: {
    issuer: keycloakEndpoints.issuer,
    authorization_endpoint: keycloakEndpoints.authorization,
    token_endpoint: keycloakEndpoints.token,
    userinfo_endpoint: keycloakEndpoints.userinfo,
    end_session_endpoint: keycloakEndpoints.endSession,
    jwks_uri: keycloakEndpoints.jwks,
  },
};

// ArgoCD와 Grafana용 추가 스코프
export const additionalScopes = {
  argocd: 'argocd',
  grafana: 'grafana',
  // 필요에 따라 추가 스코프 정의
};

// API 엔드포인트 설정 (런타임 환경변수 사용)
export const apiEndpoints = {
  argocd: env.ARGOCD_URL,
  grafana: env.GRAFANA_URL,
  jenkins: env.JENKINS_URL,
  keycloak: env.KEYCLOAK_URL,
  backend: env.BACKEND_URL,
  portal: env.PORTAL_URL,
};
