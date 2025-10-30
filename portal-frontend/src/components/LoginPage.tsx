import { useState } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';

interface LoginPageProps {
  onLogin: () => void;
}

export function LoginPage({ onLogin }: LoginPageProps) {
  const [isLoading, setIsLoading] = useState(false);

  const handleKeycloakLogin = async () => {
    // === 디버깅 로그 시작 ===
    console.log('🔵 [LoginPage] ===== 로그인 버튼 클릭 =====');
    console.log('🔵 [LoginPage] window.ENV:', window.ENV);
    console.log('🔵 [LoginPage] window.location:', {
      href: window.location.href,
      origin: window.location.origin,
      protocol: window.location.protocol,
      host: window.location.host
    });
    
    setIsLoading(true);
    
    try {
      console.log('🔵 [LoginPage] onLogin() 함수 호출 시도...');
      // Keycloak OIDC 로그인으로 리다이렉트
      await onLogin(); // 이 함수는 AuthWrapper에서 전달받은 login() 함수
      console.log('🔵 [LoginPage] onLogin() 함수 호출 완료');
    } catch (error) {
      console.error('🔴 [LoginPage] 로그인 실패:', error);
      console.error('🔴 [LoginPage] 에러 타입:', typeof error);
      console.error('🔴 [LoginPage] 에러 상세:', JSON.stringify(error, null, 2));
      setIsLoading(false);
    }
    // === 디버깅 로그 끝 ===
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/30">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <div className="flex justify-center mb-4">
            <img 
              src="/skhynix.png" 
              alt="SK하이닉스 로고" 
              className="h-12 w-auto"
            />
          </div>
          <CardTitle className="text-center login-title">빅데이터 분석 플랫폼 사용자 포털</CardTitle>
          <CardDescription className="text-center login-description">
            개발자 및 빅데이터 분석가 전용 포털에 로그인하세요
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="text-center space-y-2">
              <p className="text-muted-foreground">
                Keycloak SSO를 통해 로그인하세요
              </p>
              <p className="text-sm text-muted-foreground">
                로그인 버튼을 클릭하면 Keycloak 인증 페이지로 이동합니다
              </p>
            </div>
            <Button 
              onClick={handleKeycloakLogin}
              className={`w-full login-button ${isLoading ? 'loading' : ''}`}
              disabled={isLoading}
            >
              {isLoading ? '로그인 중...' : 'Keycloak으로 로그인'}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}