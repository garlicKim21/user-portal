import { useEffect } from 'react';
import { useAuth } from 'react-oidc-context';
import { useNavigate } from 'react-router-dom';
import { LoadingSpinner } from './ui/loading-spinner';

export function Callback() {
  const { isAuthenticated, error, isLoading } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    // react-oidc-context는 자동으로 콜백을 처리합니다
    console.log('OAuth 콜백 처리 중...');
    console.log('인증 상태:', isAuthenticated);
    console.log('로딩 상태:', isLoading);
    console.log('오류:', error);
  }, [isAuthenticated, isLoading, error]);

  // 로딩이 완료되고 인증이 성공한 경우 대시보드로 리다이렉트
  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      console.log('인증 성공, 대시보드로 리다이렉트');
      navigate('/');
    }
  }, [isLoading, isAuthenticated, navigate]);

  // 로딩이 완료되고 인증이 실패한 경우 홈으로 리다이렉트
  useEffect(() => {
    if (!isLoading && !isAuthenticated && error) {
      console.error('인증 실패:', error);
      navigate('/');
    }
  }, [isLoading, isAuthenticated, error, navigate]);

  // 인증 오류가 있는 경우
  if (error) {
    console.error('인증 오류:', error);
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-red-600 mb-4">인증 오류</h2>
          <p className="text-muted-foreground mb-4">로그인 중 오류가 발생했습니다.</p>
          <button 
            onClick={() => navigate('/')}
            className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
          >
            홈으로 돌아가기
          </button>
        </div>
      </div>
    );
  }

  return <LoadingSpinner />;
}
