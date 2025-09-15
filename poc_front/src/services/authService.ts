// 인증 서비스 - ArgoCD와 Grafana 연동을 위한 토큰 관리
import { apiEndpoints } from '../config/oidc';

export class AuthService {
  private static instance: AuthService;
  private accessToken: string | null = null;

  static getInstance(): AuthService {
    if (!AuthService.instance) {
      AuthService.instance = new AuthService();
    }
    return AuthService.instance;
  }

  // 액세스 토큰 설정
  setAccessToken(token: string) {
    this.accessToken = token;
  }

  // 액세스 토큰 가져오기
  getAccessToken(): string | null {
    return this.accessToken;
  }

  // ArgoCD URL 생성 (토큰 포함)
  getArgoCDUrl(): string {
    if (!this.accessToken) {
      throw new Error('액세스 토큰이 없습니다.');
    }
    return `${apiEndpoints.argocd}?token=${this.accessToken}`;
  }

  // Grafana URL 생성 (토큰 포함)
  getGrafanaUrl(): string {
    if (!this.accessToken) {
      throw new Error('액세스 토큰이 없습니다.');
    }
    return `${apiEndpoints.grafana}?auth_token=${this.accessToken}`;
  }

  // ArgoCD API 호출을 위한 헤더 생성
  getArgoCDHeaders(): Record<string, string> {
    if (!this.accessToken) {
      throw new Error('액세스 토큰이 없습니다.');
    }
    return {
      'Authorization': `Bearer ${this.accessToken}`,
      'Content-Type': 'application/json',
    };
  }

  // Grafana API 호출을 위한 헤더 생성
  getGrafanaHeaders(): Record<string, string> {
    if (!this.accessToken) {
      throw new Error('액세스 토큰이 없습니다.');
    }
    return {
      'Authorization': `Bearer ${this.accessToken}`,
      'Content-Type': 'application/json',
    };
  }

  // 토큰 유효성 검사
  async validateToken(): Promise<boolean> {
    if (!this.accessToken) {
      return false;
    }

    try {
      // Keycloak 토큰 검증 엔드포인트 호출
      const response = await fetch(`${apiEndpoints.keycloak}/realms/master/protocol/openid-connect/userinfo`, {
        headers: {
          'Authorization': `Bearer ${this.accessToken}`,
        },
      });
      return response.ok;
    } catch (error) {
      console.error('토큰 검증 실패:', error);
      return false;
    }
  }
}

export const authService = AuthService.getInstance();
