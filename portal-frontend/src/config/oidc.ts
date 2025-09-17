// OIDC 설정 파일 (react-oidc-context용)
export const oidcConfig = {
  authority: 'https://keycloak.miribit.cloud/realms/sso-demo',
  client_id: 'frontend',
  client_secret: 'aSnWDRHlSNITRlME6uYgIkdTRmIxZk7j',
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
    issuer: 'https://keycloak.miribit.cloud/realms/sso-demo',
    authorization_endpoint: 'https://keycloak.miribit.cloud/realms/sso-demo/protocol/openid-connect/auth',
    token_endpoint: 'https://keycloak.miribit.cloud/realms/sso-demo/protocol/openid-connect/token',
    userinfo_endpoint: 'https://keycloak.miribit.cloud/realms/sso-demo/protocol/openid-connect/userinfo',
    end_session_endpoint: 'https://keycloak.miribit.cloud/realms/sso-demo/protocol/openid-connect/logout',
    jwks_uri: 'https://keycloak.miribit.cloud/realms/sso-demo/protocol/openid-connect/certs',
  },
};

// ArgoCD와 Grafana용 추가 스코프
export const additionalScopes = {
  argocd: 'argocd',
  grafana: 'grafana',
  // 필요에 따라 추가 스코프 정의
};

// API 엔드포인트 설정
export const apiEndpoints = {
  argocd: import.meta.env.VITE_ARGOCD_URL || 'https://argocd.miribit.cloud',
  grafana: import.meta.env.VITE_GRAFANA_URL || 'https://grafana.miribit.cloud',
  keycloak: 'https://keycloak.miribit.cloud',
};
