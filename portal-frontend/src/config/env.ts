/**
 * 런타임 환경변수 설정
 * 
 * 이 파일은 컨테이너 시작 시 생성되는 /env.js에서 환경변수를 읽어옵니다.
 * 로컬 개발 시에는 기본값을 사용합니다.
 */

declare global {
  interface Window {
    ENV?: {
      KEYCLOAK_URL: string;
      KEYCLOAK_REALM: string;
      KEYCLOAK_CLIENT_ID: string;
      KEYCLOAK_CLIENT_SECRET: string;
      GRAFANA_URL: string;
      JENKINS_URL: string;
      ARGOCD_URL: string;
      BACKEND_URL: string;
      PORTAL_URL: string;
    };
  }
}

// 런타임 환경변수 또는 기본값 사용
export const env = {
  // Keycloak 설정
  KEYCLOAK_URL: window.ENV?.KEYCLOAK_URL || 'https://keycloak.miribit.cloud',
  KEYCLOAK_REALM: window.ENV?.KEYCLOAK_REALM || 'sso-demo',
  KEYCLOAK_CLIENT_ID: window.ENV?.KEYCLOAK_CLIENT_ID || 'frontend',
  KEYCLOAK_CLIENT_SECRET: window.ENV?.KEYCLOAK_CLIENT_SECRET || 'aSnWDRHlSNITRlME6uYgIkdTRmIxZk7j',
  
  // 외부 서비스 URL
  GRAFANA_URL: window.ENV?.GRAFANA_URL || 'https://grafana.miribit.cloud',
  JENKINS_URL: window.ENV?.JENKINS_URL || 'https://jenkins.miribit.cloud',
  ARGOCD_URL: window.ENV?.ARGOCD_URL || 'https://argocd.miribit.cloud',
  
  // 백엔드 API URL
  BACKEND_URL: window.ENV?.BACKEND_URL || 'https://portal.miribit.cloud',
  
  // 포털 자체 URL (로그아웃 리디렉션 등에 사용)
  PORTAL_URL: window.ENV?.PORTAL_URL || 'https://portal.miribit.cloud',
};

// Keycloak Authority URL 생성 헬퍼
export const getKeycloakAuthority = () => {
  return `${env.KEYCLOAK_URL}/realms/${env.KEYCLOAK_REALM}`;
};

// Keycloak 엔드포인트 생성 헬퍼
export const getKeycloakEndpoints = () => {
  const baseUrl = `${env.KEYCLOAK_URL}/realms/${env.KEYCLOAK_REALM}/protocol/openid-connect`;
  return {
    issuer: `${env.KEYCLOAK_URL}/realms/${env.KEYCLOAK_REALM}`,
    authorization: `${baseUrl}/auth`,
    token: `${baseUrl}/token`,
    userinfo: `${baseUrl}/userinfo`,
    endSession: `${baseUrl}/logout`,
    jwks: `${env.KEYCLOAK_URL}/realms/${env.KEYCLOAK_REALM}/protocol/openid-connect/certs`,
  };
};

// 로컬 스토리지 키 생성 (OIDC 라이브러리에서 사용)
export const getOidcStorageKey = () => {
  return `oidc.user:${getKeycloakAuthority()}:${env.KEYCLOAK_CLIENT_ID}`;
};

export default env;

