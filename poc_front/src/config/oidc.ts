// OIDC 설정 파일
export const oidcConfig = {
  authority: 'https://keycloak.miribit.cloud/realms/sso-demo',
  client_id: 'portal-frontend',
  redirect_uri: window.location.origin + '/callback',
  post_logout_redirect_uri: window.location.origin,
  response_type: 'code',
  scope: 'openid profile email',
  // 자동 갱신 비활성화 (문제 해결을 위해)
  automaticSilentRenew: false,
  loadUserInfo: true,
  // Silent callback 설정
  silent_redirect_uri: window.location.origin + '/silent-callback',
  // 로그 레벨을 error로 설정하여 불필요한 로그 줄이기
  logLevel: 'error' as const,
  // 추가 설정
  checkSessionInterval: 2000,
  silentRequestTimeout: 10000,
};

// ArgoCD와 Grafana용 추가 스코프
export const additionalScopes = {
  argocd: 'argocd',
  grafana: 'grafana',
  // 필요에 따라 추가 스코프 정의
};

// API 엔드포인트 설정
export const apiEndpoints = {
  argocd: process.env.REACT_APP_ARGOCD_URL || 'https://argocd.miribit.cloud',
  grafana: process.env.REACT_APP_GRAFANA_URL || 'https://grafana.miribit.cloud',
  keycloak: 'https://keycloak.miribit.cloud',
};
