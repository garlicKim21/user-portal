// 터미널 관련 API 서비스 함수들

export interface TerminalLaunchResponse {
  url: string;
  resourceId: string;
}

export interface TerminalListResponse {
  consoles: Array<{
    id: string;
    userId: string;
    consoleUrl: string;
    createdAt: string;
  }>;
  count: number;
  userId: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

/**
 * 웹 터미널을 시작합니다.
 * @returns Promise<TerminalLaunchResponse>
 */
export async function launchTerminal(): Promise<TerminalLaunchResponse> {
  const response = await fetch('/api/launch-console', {
    method: 'GET',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
  }

  const data = await response.json();
  
  if (!data.data || !data.data.url) {
    throw new Error('Terminal URL not received from server');
  }

  return {
    url: data.data.url,
    resourceId: data.data.resourceId || '',
  };
}

/**
 * 사용자의 웹 터미널 목록을 조회합니다.
 * @returns Promise<TerminalListResponse>
 */
export async function listTerminals(): Promise<TerminalListResponse> {
  const response = await fetch('/api/console/list', {
    method: 'GET',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
  }

  const data = await response.json();
  return data.data;
}

/**
 * 특정 웹 터미널을 삭제합니다.
 * @param resourceId 삭제할 터미널의 리소스 ID
 * @returns Promise<void>
 */
export async function deleteTerminal(resourceId: string): Promise<void> {
  const response = await fetch(`/api/console/${resourceId}`, {
    method: 'DELETE',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
  }
}

/**
 * 사용자 인증 상태를 확인합니다.
 * @returns Promise<boolean>
 */
export async function checkAuthStatus(): Promise<boolean> {
  try {
    const response = await fetch('/api/user', {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    return response.ok;
  } catch (error) {
    console.error('Auth check failed:', error);
    return false;
  }
}
