import { useState } from 'react';
import { LoginPage } from './LoginPage';
import { Dashboard } from './Dashboard';

export function SimpleAuthWrapper() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  const handleLogin = () => {
    console.log('로그인 시도');
    setIsLoggedIn(true);
  };

  const handleLogout = () => {
    console.log('로그아웃 시도');
    setIsLoggedIn(false);
  };

  console.log('SimpleAuthWrapper 렌더링, isLoggedIn:', isLoggedIn);

  if (isLoggedIn) {
    return <Dashboard onLogout={handleLogout} />;
  }

  return <LoginPage onLogin={handleLogin} />;
}
