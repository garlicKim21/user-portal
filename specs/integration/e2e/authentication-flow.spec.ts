import { test, expect } from '@playwright/test';

test.describe('Authentication Flow E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    // 테스트 환경 설정
    await page.goto('https://dev-portal.miribit.cloud');
  });

  test('사용자가 성공적으로 로그인할 수 있다', async ({ page }) => {
    // Given: 로그인 페이지에 접근
    await expect(page.locator('text=빅데이터 분석 플랫폼 사용자 포털')).toBeVisible();
    
    // When: Keycloak 로그인 버튼 클릭
    await page.click('text=Keycloak으로 로그인');
    
    // Then: Keycloak 로그인 페이지로 리다이렉트
    await expect(page).toHaveURL(/keycloak\.miribit\.cloud/);
    
    // When: 사용자 인증 정보 입력
    await page.fill('input[name="username"]', 'X0157671');
    await page.fill('input[name="password"]', 'pass123');
    await page.click('input[type="submit"]');
    
    // Then: 대시보드로 리다이렉트
    await expect(page).toHaveURL(/portal\.miribit\.cloud/);
    await expect(page.locator('text=대시보드')).toBeVisible();
  });

  test('사용자가 로그아웃할 수 있다', async ({ page }) => {
    // Given: 로그인된 상태
    await loginUser(page, 'X0157671', 'pass123');
    
    // When: 로그아웃 버튼 클릭
    await page.click('text=로그아웃');
    
    // Then: 로그인 페이지로 리다이렉트
    await expect(page.locator('text=빅데이터 분석 플랫폼 사용자 포털')).toBeVisible();
  });

  test('계정 전환이 가능하다', async ({ page }) => {
    // Given: fdc 계정으로 로그인
    await loginUser(page, 'X0157671', 'pass123');
    await expect(page.locator('text=대시보드')).toBeVisible();
    
    // When: 로그아웃
    await page.click('text=로그아웃');
    
    // And: eds 계정으로 재로그인
    await page.click('text=Keycloak으로 로그인');
    await page.fill('input[name="username"]', 'X0157672');
    await page.fill('input[name="password"]', 'pass123');
    await page.click('input[type="submit"]');
    
    // Then: eds 계정으로 대시보드 접근
    await expect(page.locator('text=대시보드')).toBeVisible();
  });

  test('그라파나 접속 시 올바른 계정으로 인증된다', async ({ page }) => {
    // Given: fdc 계정으로 로그인
    await loginUser(page, 'X0157671', 'pass123');
    
    // When: 그라파나 메뉴 클릭
    await page.click('text=Grafana');
    await page.click('text=Grafana 열기');
    
    // Then: 그라파나 페이지가 새 탭에서 열림
    const [grafanaPage] = await Promise.all([
      page.waitForEvent('popup'),
      page.click('text=Grafana 열기')
    ]);
    
    await expect(grafanaPage).toHaveURL(/grafana\.miribit\.cloud/);
    
    // And: fdc 계정으로 인증됨
    await expect(grafanaPage.locator('text=X0157671')).toBeVisible();
  });

  test('토큰 만료 시 자동 갱신된다', async ({ page }) => {
    // Given: 로그인된 상태
    await loginUser(page, 'X0157671', 'pass123');
    
    // When: 토큰 만료 시뮬레이션 (개발 환경에서만)
    await page.evaluate(() => {
      // 토큰 만료 시간을 과거로 설정
      const token = JSON.parse(localStorage.getItem('oidc.user:...') || '{}');
      token.expires_at = Date.now() - 1000;
      localStorage.setItem('oidc.user:...', JSON.stringify(token));
    });
    
    // Then: 자동 갱신 시도
    await page.waitForTimeout(5000);
    
    // And: 여전히 로그인 상태 유지
    await expect(page.locator('text=대시보드')).toBeVisible();
  });

  test('네트워크 오류 시 적절한 에러 메시지를 표시한다', async ({ page }) => {
    // Given: 네트워크 오류 시뮬레이션
    await page.route('**/auth/login', route => route.abort());
    
    // When: 로그인 시도
    await page.click('text=Keycloak으로 로그인');
    
    // Then: 에러 메시지 표시
    await expect(page.locator('text=로그인 중 오류가 발생했습니다')).toBeVisible();
  });
});

// 헬퍼 함수
async function loginUser(page: any, username: string, password: string) {
  await page.click('text=Keycloak으로 로그인');
  await page.fill('input[name="username"]', username);
  await page.fill('input[name="password"]', password);
  await page.click('input[type="submit"]');
  await page.waitForURL(/portal\.miribit\.cloud/);
}
