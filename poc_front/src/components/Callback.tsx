import { useOidc } from '@axa-fr/react-oidc';
import { useEffect } from 'react';
import { LoadingSpinner } from './ui/loading-spinner';

export function Callback() {
  const { login } = useOidc();

  useEffect(() => {
    // OIDC 콜백 처리
    login();
  }, [login]);

  return <LoadingSpinner />;
}
