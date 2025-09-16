// 백엔드 인증 서비스 - OIDC 토큰을 백엔드 JWT 세션으로 변환
export class BackendAuthService {
  private static instance: BackendAuthService;
  private isSessionCreated: boolean = false;

  static getInstance(): BackendAuthService {
    if (!BackendAuthService.instance) {
      BackendAuthService.instance = new BackendAuthService();
    }
    return BackendAuthService.instance;
  }

  /**
   * OIDC 토큰으로 백엔드 세션 생성
   * @param oidcAccessToken - react-oidc-context에서 받은 access_token
   */
  async createBackendSession(oidcAccessToken: string): Promise<void> {
    if (this.isSessionCreated) {
      console.log('Backend session already created, skipping...');
      return;
    }

    try {
      const response = await fetch('/api/auth/session', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${oidcAccessToken}`,
          'Content-Type': 'application/json',
        },
        credentials: 'include' // 쿠키 설정을 위해 필요
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(`Failed to create backend session: ${errorData.error?.message || response.statusText}`);
      }

      const data = await response.json();
      this.isSessionCreated = true;
      
      console.log('Backend session created successfully:', data.data);
    } catch (error) {
      console.error('Error creating backend session:', error);
      throw error;
    }
  }

  /**
   * 백엔드 세션 유효성 검증
   */
  async validateBackendSession(): Promise<boolean> {
    try {
      const response = await fetch('/api/user', {
        credentials: 'include'
      });
      
      return response.ok;
    } catch (error) {
      console.error('Error validating backend session:', error);
      return false;
    }
  }

  /**
   * Web Console 실행
   */
  async launchWebConsole(): Promise<{ url: string; resourceId: string }> {
    try {
      const response = await fetch('/api/launch-console', {
        credentials: 'include'
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(`Failed to launch web console: ${errorData.error?.message || response.statusText}`);
      }

      const data = await response.json();
      
      if (!data.data?.url) {
        throw new Error('Console URL not received from server');
      }

      return {
        url: data.data.url,
        resourceId: data.data.resource_id
      };
    } catch (error) {
      console.error('Error launching web console:', error);
      throw error;
    }
  }

  /**
   * 백엔드 로그아웃
   */
  async logout(): Promise<string | null> {
    try {
      const response = await fetch('/api/logout', {
        method: 'POST',
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        this.isSessionCreated = false;
        return data.logout_url || null;
      } else {
        console.error('Backend logout failed');
        this.isSessionCreated = false;
        return null;
      }
    } catch (error) {
      console.error('Logout request error:', error);
      this.isSessionCreated = false;
      return null;
    }
  }

  /**
   * 세션 상태 리셋 (개발/테스트용)
   */
  resetSessionState(): void {
    this.isSessionCreated = false;
  }
}

export const backendAuthService = BackendAuthService.getInstance();
