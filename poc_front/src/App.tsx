// 임시로 OIDC 없이 테스트
import { SimpleAuthWrapper } from './components/SimpleAuthWrapper';

export default function App() {
  console.log('App 컴포넌트 렌더링');
  return <SimpleAuthWrapper />;
}

// OIDC 버전 (문제 해결 후 사용)
/*
import { OidcProvider } from '@axa-fr/react-oidc';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { LoginPage } from './components/LoginPage';
import { Dashboard } from './components/Dashboard';
import { oidcConfig } from './config/oidc';
import { AuthWrapper } from './components/AuthWrapper';
import { Callback } from './components/Callback';
import { SilentCallback } from './components/SilentCallback';

export default function App() {
  return (
    <OidcProvider configuration={oidcConfig}>
      <Router>
        <Routes>
          <Route path="/callback" element={<Callback />} />
          <Route path="/silent-callback" element={<SilentCallback />} />
          <Route path="/*" element={<AuthWrapper />} />
        </Routes>
      </Router>
    </OidcProvider>
  );
}
*/