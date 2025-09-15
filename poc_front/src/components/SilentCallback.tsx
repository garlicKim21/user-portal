import { useOidc } from '@axa-fr/react-oidc';
import { useEffect } from 'react';

export function SilentCallback() {
  const { login } = useOidc();

  useEffect(() => {
    // Silent callback 처리
    login();
  }, [login]);

  return null; // Silent callback은 UI를 표시하지 않음
}
