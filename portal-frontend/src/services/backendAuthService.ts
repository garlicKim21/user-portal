// 백엔드 인증 서비스 - OIDC 토큰을 직접 사용하여 백엔드 API 호출
export class BackendAuthService {
  private static instance: BackendAuthService;

  static getInstance(): BackendAuthService {
    if (!BackendAuthService.instance) {
      BackendAuthService.instance = new BackendAuthService();
    }
    return BackendAuthService.instance;
  }

  /**
   * Web Console 실행 (OIDC Access Token 사용)
   * @param oidcAccessToken - react-oidc-context에서 받은 access_token
   */
  async launchWebConsole(oidcAccessToken: string): Promise<{ url: string; resourceId: string }> {
    try {
      console.log('Launching web console with OIDC token...');
      
      const response = await fetch('/api/launch-console', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${oidcAccessToken}`,
          'Content-Type': 'application/json',
        }
      });

      console.log('Launch console response status:', response.status);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ 
          error: { 
            message: response.statusText,
            code: 'NETWORK_ERROR',
            type: 'AUTHENTICATION_ERROR'
          } 
        }));
        console.error('Web console launch failed:', errorData);
        throw new Error(`Failed to launch web console: ${errorData.error?.message || response.statusText}`);
      }

      const data = await response.json();
      
      if (!data.data?.url) {
        throw new Error('Console URL not received from server');
      }

      console.log('Web console launched successfully:', data.data);
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
   * 콘솔 목록 조회 (OIDC Access Token 사용)
   * @param oidcAccessToken - react-oidc-context에서 받은 access_token
   */
  async listConsoles(oidcAccessToken: string): Promise<any[]> {
    try {
      const response = await fetch('/api/console/list', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${oidcAccessToken}`,
          'Content-Type': 'application/json',
        }
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ 
          error: { message: response.statusText } 
        }));
        throw new Error(`Failed to list consoles: ${errorData.error?.message || response.statusText}`);
      }

      const data = await response.json();
      return data.consoles || [];
    } catch (error) {
      console.error('Error listing consoles:', error);
      throw error;
    }
  }

  /**
   * 콘솔 삭제 (OIDC Access Token 사용)
   * @param oidcAccessToken - react-oidc-context에서 받은 access_token
   * @param resourceId - 삭제할 리소스 ID
   */
  async deleteConsole(oidcAccessToken: string, resourceId: string): Promise<void> {
    try {
      const response = await fetch(`/api/console/${resourceId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${oidcAccessToken}`,
          'Content-Type': 'application/json',
        }
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ 
          error: { message: response.statusText } 
        }));
        throw new Error(`Failed to delete console: ${errorData.error?.message || response.statusText}`);
      }
    } catch (error) {
      console.error('Error deleting console:', error);
      throw error;
    }
  }

  /**
   * 세션 상태 리셋 (호환성을 위해 유지하지만 아무 작업 안 함)
   */
  resetSessionState(): void {
    // 더 이상 세션을 관리하지 않으므로 아무 작업 안 함
  }
}

export const backendAuthService = BackendAuthService.getInstance();
