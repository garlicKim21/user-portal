import { useOidc, useOidcUser } from '@axa-fr/react-oidc';
import { LoginPage } from './LoginPage';
import { Dashboard } from './Dashboard';
import { LoadingSpinner } from './ui/loading-spinner';
import { useEffect, useState } from 'react';

export function AuthWrapper() {
  const { login, logout, isAuthenticated, loginError } = useOidc();
  const { oidcUser, isOidcLoading } = useOidcUser();
  const [isInitialized, setIsInitialized] = useState(false);

  // 초기화 완료 대기
  useEffect(() => {
    const timer = setTimeout(() => {
      setIsInitialized(true);
    }, 1000); // 1초 후 초기화 완료로 간주

    return () => clearTimeout(timer);
  }, []);

  // 초기화 중이거나 OIDC 로딩 중
  if (!isInitialized || isOidcLoading) {
    return <LoadingSpinner />;
  }

  // 인증 에러 처리
  if (loginError) {
    console.error('OIDC 로그인 에러:', loginError);
    return (
      <div className="min-h-screen flex items-center justify-center bg-red-50">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-red-600 mb-4">인증 오류</h2>
          <p className="text-red-500 mb-4">{loginError.message}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
          >
            다시 시도
          </button>
        </div>
      </div>
    );
  }

  // 인증되지 않은 경우 - 로그인 페이지 표시
  if (!isAuthenticated) {
    console.log('사용자가 인증되지 않음, 로그인 페이지 표시');
    return <LoginPage onLogin={login} />;
  }

  // 인증된 경우 - 대시보드 표시
  console.log('사용자가 인증됨, 대시보드 표시', oidcUser);
  return <Dashboard onLogout={logout} user={oidcUser} />;
}
