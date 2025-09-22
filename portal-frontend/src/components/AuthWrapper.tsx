import { useAuth } from 'react-oidc-context';
import { LoginPage } from './LoginPage';
import { Dashboard } from './Dashboard';
import { Callback } from './Callback';
import { useLocation } from 'react-router-dom';
import { authService } from '../services/authService';

export function AuthWrapper() {
  const { signinRedirect, signoutRedirect, isAuthenticated, user, isLoading } = useAuth();
  const location = useLocation();

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

  // 완전한 로그아웃 처리 함수
  const handleLogout = async () => {
    try {
      console.log('로그아웃 시작...');
      
      // 로컬 스토리지에서 OIDC 사용자 정보 정리
      const oidcUserKey = 'oidc.user:https://keycloak.miribit.cloud/realms/sso-demo:frontend';
      localStorage.removeItem(oidcUserKey);
      sessionStorage.clear();
      
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

  // 인증된 경우 대시보드 표시
  return <Dashboard onLogout={handleLogout} user={user} />;
}