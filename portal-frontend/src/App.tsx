import { AuthProvider } from 'react-oidc-context';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { oidcConfig } from './config/oidc';
import { AuthWrapper } from './components/AuthWrapper';
import { Callback } from './components/Callback';
import { SilentCallback } from './components/SilentCallback';

export default function App() {
  return (
    <AuthProvider {...oidcConfig}>
      <Router>
        <Routes>
          <Route path="/callback" element={<Callback />} />
          <Route path="/silent-callback" element={<SilentCallback />} />
          <Route path="/*" element={<AuthWrapper />} />
        </Routes>
      </Router>
    </AuthProvider>
  );
}