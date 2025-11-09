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
 * URL이 실제로 응답 가능한 상태인지 확인합니다.
 * @param url 확인할 URL
 * @param maxRetries 최대 재시도 횟수 (기본값: 3)
 * @param retryInterval 재시도 간격(ms) (기본값: 2000)
 * @returns Promise<boolean>
 */
export async function verifyTerminalReady(
  url: string,
  maxRetries: number = 3,
  retryInterval: number = 2000
): Promise<boolean> {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      console.log(`[TerminalService] Verifying terminal readiness (attempt ${attempt}/${maxRetries}): ${url}`);

      // HEAD 요청으로 가볍게 확인
      const response = await fetch(url, {
        method: 'HEAD',
        // mode: 'no-cors', // local dev only, dev 및 prod 사용 금지
        cache: 'no-cache',
      });
      // console.log(response);

      // local dev only, dev 및 prod 사용 금지
      // if (response.type === 'opaque') {
      //   console.log(`[TerminalService] Terminal is ready at ${url}`);
      //   return true;
      // }

      // 일반 CORS 응답인 경우
      if (response.ok) {
        console.log(`[TerminalService] Terminal is ready at ${url} (status: ${response.status})`);
        return true;
      }

      console.log(`[TerminalService] Terminal not ready yet (status: ${response.status})`);
    } catch (error) {
      console.log(`[TerminalService] Terminal verification failed (attempt ${attempt}/${maxRetries}):`, error);
    }

    // 마지막 시도가 아니면 대기 후 재시도
    if (attempt < maxRetries) {
      console.log(`[TerminalService] Waiting ${retryInterval}ms before retry...`);
      await new Promise(resolve => setTimeout(resolve, retryInterval + (attempt * 1000)));
    }
  }

  console.warn(`[TerminalService] Terminal not ready after ${maxRetries} attempts`);
  return false;
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

  const terminalUrl = data.data.url;
  const resourceId = data.data.resourceId || '';

  console.log(`[TerminalService] Terminal URL received: ${terminalUrl}`);
  console.log(`[TerminalService] Resource ID: ${resourceId}`);

  // URL 유효성 확인 (최대 3번 재시도, 2초 간격)
  const isReady = await verifyTerminalReady(terminalUrl, 3, 2000);

  if (!isReady) {
    console.warn(`[TerminalService] Terminal URL may not be fully ready, but returning anyway`);
    // 경고는 하지만 URL은 반환 (사용자가 직접 재시도할 수 있음)
  }

  return {
    url: terminalUrl,
    resourceId: resourceId,
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
