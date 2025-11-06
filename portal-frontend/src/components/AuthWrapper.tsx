import { useState, useEffect } from 'react';
import { useAuth } from 'react-oidc-context';
import { LoginPage } from './LoginPage';
import { Dashboard } from './Dashboard';
import { Callback } from './Callback';
import { useLocation } from 'react-router-dom';
import { AppUser, UserProject, AuthState, mockProjects } from '../types/user';

// Keycloak groups에서 프로젝트 정보 파싱 함수
function parseUserProjectsFromGroups(groups: string[] | undefined): UserProject[] {
  console.log('parseUserProjectsFromGroups called with groups:', groups);
  
  if (!groups || groups.length === 0) {
    console.log('No groups found, returning empty array');
    return [];
  }

  const projects: UserProject[] = [];
  
  groups.forEach(group => {
    console.log('Processing group:', group);
    // /dataops/{project}/{role} 형식 파싱
    const match = group.match(/^\/dataops\/([^\/]+)\/([^\/]+)$/);
    if (match) {
      const [, projectId, role] = match;
      console.log('Parsed project:', projectId, 'role:', role);
      
      // 프로젝트 ID를 기반으로 프로젝트명 매핑
      const projectName = getProjectName(projectId);
      const roleLabel = getRoleLabel(role);
      
      const project = {
        id: projectId,
        name: projectName,
        role: role as 'dev' | 'adm' | 'viewer',
        roleLabel: roleLabel
      };
      
      console.log('Created project:', project);
      projects.push(project);
    } else {
      console.log('Group does not match pattern:', group);
    }
  });

  console.log('Final projects array:', projects);
  return projects;
}

// 프로젝트 ID를 프로젝트명으로 매핑 (동적 생성)
function getProjectName(projectId: string): string {
  // 프로젝트 ID를 대문자로 변환하고 '프로젝트' 추가
  return `${projectId.toUpperCase()} 프로젝트`;
}

// 역할 코드를 역할명으로 매핑
function getRoleLabel(role: string): string {
  const roleLabels: Record<string, string> = {
    'dev': '개발자',
    'adm': '관리자',
    'viewer': '조회자'
  };
  return roleLabels[role] || role;
}

export function AuthWrapper() {
  const auth = useAuth();
  const { signinRedirect, isAuthenticated, user, isLoading, events } = auth;
  const location = useLocation();

  // 상태 관리
  const [authState, setAuthState] = useState<AuthState>({
    user: null,
    currentProject: null
  });

  // 토큰 자동 갱신 이벤트 리스너 추가
  useEffect(() => {
    if (!events) return;

    // 토큰이 만료되기 전 이벤트
    const handleAccessTokenExpiring = () => {
      console.log('[AuthWrapper] 🔄 토큰이 곧 만료됩니다. 자동 갱신 시도 중...');
      if (user?.expires_at) {
        const expiresAt = new Date(user.expires_at * 1000);
        console.log('[AuthWrapper] 현재 토큰 만료 시간:', expiresAt.toLocaleString());
      }
    };

    // 토큰이 만료된 후 이벤트
    const handleAccessTokenExpired = () => {
      console.error('[AuthWrapper] ⚠️ 토큰이 만료되었습니다!');
      console.log('[AuthWrapper] 자동 갱신 실패. 재로그인이 필요할 수 있습니다.');
    };

    // 토큰이 자동으로 갱신된 후 이벤트
    const handleUserLoaded = (user: any) => {
      console.log('[AuthWrapper] ✅ 토큰이 성공적으로 갱신되었습니다!');
      if (user?.expires_at) {
        const expiresAt = new Date(user.expires_at * 1000);
        console.log('[AuthWrapper] 새 토큰 만료 시간:', expiresAt.toLocaleString());
        const now = new Date();
        const timeUntilExpiry = Math.floor((expiresAt.getTime() - now.getTime()) / 1000 / 60);
        console.log(`[AuthWrapper] ${timeUntilExpiry}분 후에 만료됩니다`);
      }
    };

    // Silent renew 에러
    const handleSilentRenewError = (error: Error) => {
      console.error('[AuthWrapper] ❌ Silent renew 실패:', error);
      console.error('[AuthWrapper] 에러 상세:', error.message);
      console.error('[AuthWrapper] 확인사항:');
      console.error('  1. Keycloak에서 /silent-callback이 Valid Redirect URIs에 등록되어 있는지');
      console.error('  2. 브라우저에서 third-party cookies가 차단되어 있지 않은지');
      console.error('  3. Keycloak의 Access Token Lifespan 설정 확인');
    };

    // 사용자 언로드 (로그아웃)
    const handleUserUnloaded = () => {
      console.log('[AuthWrapper] 사용자가 로그아웃되었습니다');
    };

    // 이벤트 리스너 등록
    events.addAccessTokenExpiring(handleAccessTokenExpiring);
    events.addAccessTokenExpired(handleAccessTokenExpired);
    events.addUserLoaded(handleUserLoaded);
    events.addSilentRenewError(handleSilentRenewError);
    events.addUserUnloaded(handleUserUnloaded);

    console.log('[AuthWrapper] 토큰 자동 갱신 이벤트 리스너 등록 완료');

    // 클린업
    return () => {
      events.removeAccessTokenExpiring(handleAccessTokenExpiring);
      events.removeAccessTokenExpired(handleAccessTokenExpired);
      events.removeUserLoaded(handleUserLoaded);
      events.removeSilentRenewError(handleSilentRenewError);
      events.removeUserUnloaded(handleUserUnloaded);
    };
  }, [events, user]);

  // 사용자 로그인 시 실제 OIDC 토큰 정보와 프로젝트 데이터 결합
  useEffect(() => {
    if (isAuthenticated && user && !authState.user) {
      console.log('OIDC User Info:', user); // 디버깅용
      console.log('OIDC Profile:', user.profile); // 디버깅용
      console.log('OIDC Groups from profile:', user.profile?.groups); // 디버깅용
      console.log('OIDC Groups from user:', (user as any)?.groups); // 디버깅용

      // 토큰 만료 시간 로그
      if (user.expires_at) {
        const expiresAt = new Date(user.expires_at * 1000);
        console.log('[AuthWrapper] 초기 토큰 만료 시간:', expiresAt.toLocaleString());
        const now = new Date();
        const timeUntilExpiry = Math.floor((expiresAt.getTime() - now.getTime()) / 1000 / 60);
        console.log(`[AuthWrapper] ${timeUntilExpiry}분 후에 만료됩니다`);
      }

      // 실제 OIDC 사용자 정보를 AppUser 타입으로 변환
      const profile = user.profile || {};
      const appUserData: AppUser = {
        // OIDC 기본 정보 - profile 객체에서 가져오기
        sub: (profile as any)?.sub || (user as any)?.sub || '',
        preferred_username: (profile as any)?.preferred_username || (user as any)?.preferred_username || '',
        name: (profile as any)?.name || (user as any)?.name || 'Unknown User',
        email: (profile as any)?.email || (user as any)?.email || '',
        given_name: (profile as any)?.given_name || (user as any)?.given_name || '',
        family_name: (profile as any)?.family_name || (user as any)?.family_name || '',
        email_verified: (profile as any)?.email_verified || (user as any)?.email_verified || false,
        auth_time: (profile as any)?.auth_time || (user as any)?.auth_time || 0,
        access_token: (user as any)?.access_token || '',
        id_token: (user as any)?.id_token || '',
        
        // 앱에서 관리하는 프로젝트 정보
        // 실제 환경에서는 Keycloak groups 필드에서 파싱하거나 LDAP API 호출
        projects: parseUserProjectsFromGroups((profile as any)?.groups || (user as any)?.groups) || mockProjects
      };
      
      console.log('App User Data:', appUserData); // 디버깅용
      
      setAuthState({
        user: appUserData,
        currentProject: appUserData.projects[0] || null // 첫 번째 프로젝트를 기본값으로 설정
      });
    }
  }, [isAuthenticated, user, authState.user]);

  // 프로젝트 변경 핸들러
  const handleProjectChange = (project: UserProject) => {
    setAuthState(prev => ({
      ...prev,
      currentProject: project
    }));
  };

  // 로그아웃 핸들러 (웹 콘솔 리소스 삭제 + 키클락 세션 로그아웃)
  const handleLogout = async () => {
    try {
      console.log('로그아웃 시작...');
      
      // 1. 백엔드 리소스 정리 API 호출 (웹 콘솔 리소스 삭제)
      if (authState.user?.access_token) {
        try {
          const { backendAuthService } = await import('../services/backendAuthService');
          await backendAuthService.logout(authState.user.access_token);
          console.log('웹 콘솔 리소스 정리 완료');
        } catch (error) {
          console.error('웹 콘솔 리소스 정리 실패:', error);
          // 리소스 정리 실패해도 로그아웃은 계속 진행
        }
      }
      
      // 2. 로컬 스토리지에서 OIDC 사용자 정보 정리
      const oidcUserKey = 'oidc.user:https://keycloak.miribit.cloud/realms/sso-demo:frontend';
      localStorage.removeItem(oidcUserKey);
      sessionStorage.clear();
      
      // 3. 상태 초기화
      setAuthState({
        user: null,
        currentProject: null
      });
      
      // 4. 키클락 세션 로그아웃 (확인 페이지 없이 바로 로그아웃)
      const idToken = user?.id_token;
      let logoutUrl;
      
      if (idToken) {
        // id_token_hint 사용으로 확인 페이지 생략
        logoutUrl = `https://keycloak.miribit.cloud/realms/sso-demo/protocol/openid-connect/logout?client_id=frontend&post_logout_redirect_uri=${encodeURIComponent('https://portal.miribit.cloud')}&id_token_hint=${idToken}`;
      } else {
        logoutUrl = `https://keycloak.miribit.cloud/realms/sso-demo/protocol/openid-connect/logout?client_id=frontend&post_logout_redirect_uri=${encodeURIComponent('https://portal.miribit.cloud')}`;
      }
      
      console.log('키클락 로그아웃 URL로 리다이렉트:', logoutUrl);
      window.location.href = logoutUrl;
      
    } catch (error) {
      console.error('로그아웃 중 오류:', error);
      // 오류가 발생해도 강제로 홈으로 리다이렉트
      window.location.href = '/';
    }
  };

  // 로딩 중일 때
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">인증 확인 중...</p>
        </div>
      </div>
    );
  }

  // 콜백 페이지인 경우
  if (location.pathname === '/callback') {
    return <Callback />;
  }

  // 인증되지 않은 경우 로그인 페이지 표시
  if (!isAuthenticated) {
    return <LoginPage onLogin={() => signinRedirect()} />;
  }

  // 인증된 경우 대시보드 표시 (사용자 정보가 로드되지 않은 경우 로딩 표시)
  if (!authState.user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">사용자 정보를 불러오는 중...</p>
        </div>
      </div>
    );
  }

  return (
    <Dashboard 
      user={authState.user}
      currentProject={authState.currentProject}
      onProjectChange={handleProjectChange}
      onLogout={handleLogout}
    />
  );
}