import { AuthProvider } from 'react-oidc-context';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { oidcConfig } from './config/oidc';
import { AuthWrapper } from './components/AuthWrapper';
import { Callback } from './components/Callback';
import { SilentCallback } from './components/SilentCallback';
import Terminal from "@/components/Terminal.tsx";
import {Toaster} from "@/components/ui/sonner.tsx";

export default function App() {
  return <>
    <AuthProvider {...oidcConfig}>
      <Router>
        <Routes>
          <Route path="/callback" element={<Callback />} />
          <Route path="/silent-callback" element={<SilentCallback />} />
          <Route path="/terminal" element={<Terminal />} />
          <Route path="/*" element={<AuthWrapper />} />
        </Routes>
      </Router>
    </AuthProvider>
    <Toaster position="top-right" />
  </>
}