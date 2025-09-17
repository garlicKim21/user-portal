import { useEffect } from 'react';
import { useAuth } from 'react-oidc-context';

export function SilentCallback() {
  const { signinSilent } = useAuth();

  useEffect(() => {
    // Silent callback 처리
    const handleSilentCallback = async () => {
      try {
        console.log('Silent 콜백 처리 시작...');
        await signinSilent();
        console.log('Silent 콜백 처리 완료');
      } catch (error) {
        console.error('Silent 콜백 처리 중 오류:', error);
      }
    };

    handleSilentCallback();
  }, [signinSilent]);

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
        <p className="text-muted-foreground">세션 갱신 중...</p>
      </div>
    </div>
  );
}