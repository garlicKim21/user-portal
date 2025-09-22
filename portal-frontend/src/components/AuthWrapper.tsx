import { useState, useEffect } from 'react';
import { useAuth } from 'react-oidc-context';
import { LoginPage } from './LoginPage';
import { Dashboard } from './Dashboard';
import { Callback } from './Callback';
import { useLocation } from 'react-router-dom';
import { AppUser, UserProject, AuthState, mockProjects } from '../types/user';

export function AuthWrapper() {
  const { signinRedirect, isAuthenticated, user, isLoading } = useAuth();
  const location = useLocation();
  
  // 상태 관리
  const [authState, setAuthState] = useState<AuthState>({
    user: null,
    currentProject: null
  });

  // 사용자 로그인 시 실제 OIDC 토큰 정보와 프로젝트 데이터 결합
  useEffect(() => {
    if (isAuthenticated && user && !authState.user) {
      console.log('OIDC User Info:', user); // 디버깅용
      
      // 실제 OIDC 사용자 정보를 AppUser 타입으로 변환
      const appUserData: AppUser = {
        // OIDC 기본 정보
        sub: user.sub || '',
        preferred_username: user.preferred_username || user.sub || '',
        name: user.name || (user.given_name && user.family_name ? `${user.family_name}${user.given_name}` : user.preferred_username) || 'Unknown User',
        email: user.email,
        given_name: user.given_name,
        family_name: user.family_name,
        email_verified: user.email_verified,
        auth_time: user.auth_time,
        access_token: user.access_token,
        id_token: user.id_token,
        
        // 앱에서 관리하는 프로젝트 정보 (실제 환경에서는 LDAP API에서 가져옴)
        // 현재는 사용자명을 기반으로 프로젝트를 동적으로 생성
        projects: user.preferred_username?.includes('eds') 
          ? [{ id: 'eds', name: 'EDS 프로젝트', role: user.preferred_username.includes('adm') ? 'adm' : 'dev', roleLabel: user.preferred_username.includes('adm') ? '관리자' : '개발자' }]
          : user.preferred_username?.includes('fdc')
          ? [{ id: 'fdc', name: 'FDC 프로젝트', role: user.preferred_username.includes('adm') ? 'adm' : 'dev', roleLabel: user.preferred_username.includes('adm') ? '관리자' : '개발자' }]
          : mockProjects
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

  // 로그아웃 핸들러
  const handleLogout = async () => {
    try {
      console.log('로그아웃 시작...');
      
      // 로컬 스토리지에서 OIDC 사용자 정보 정리
      const oidcUserKey = 'oidc.user:https://keycloak.miribit.cloud/realms/sso-demo:frontend';
      localStorage.removeItem(oidcUserKey);
      sessionStorage.clear();
      
      // 상태 초기화
      setAuthState({
        user: null,
        currentProject: null
      });
      
      // 직접 Keycloak 로그아웃 URL로 리다이렉트 (확인 페이지 없이 바로 로그아웃)
      const logoutUrl = `https://keycloak.miribit.cloud/realms/sso-demo/protocol/openid-connect/logout?client_id=frontend&post_logout_redirect_uri=${encodeURIComponent('https://portal.miribit.cloud')}&prompt=none`;
      console.log('Keycloak 로그아웃 URL로 리다이렉트 (확인 페이지 생략):', logoutUrl);
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