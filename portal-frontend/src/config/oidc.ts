// OIDC 설정 파일 (react-oidc-context용)
import env, { getKeycloakAuthority, getKeycloakEndpoints } from './env';

// Keycloak 엔드포인트 생성
const keycloakEndpoints = getKeycloakEndpoints();

export const oidcConfig = {
  authority: getKeycloakAuthority(),
  client_id: env.KEYCLOAK_CLIENT_ID,
  client_secret: env.KEYCLOAK_CLIENT_SECRET,
  redirect_uri: window.location.origin + '/callback',
  post_logout_redirect_uri: window.location.origin,
  response_type: 'code',
  scope: 'openid profile email',
  // 자동 갱신 설정
  automaticSilentRenew: true,
  loadUserInfo: true,
  // Silent callback 설정
  silent_redirect_uri: window.location.origin + '/silent-callback',
  // 로그 레벨 설정
  logLevel: 'debug' as const,
  // 추가 설정
  checkSessionInterval: 2000,
  silentRequestTimeout: 10000,
  // PKCE 비활성화 (client secret 사용)
  pkce: false,
  // 세션 만료 시간 (30분)
  accessTokenExpiringNotificationTime: 300,
  // 로그아웃 관련 설정
  revokeTokensOnSignout: true,
  includeIdTokenInSilentRenew: true,
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
