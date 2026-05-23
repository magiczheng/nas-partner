import { useEffect, useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, Spin } from 'antd';
import { auth } from './api/auth';
import AdminLayout from './components/Layout';
import Home from './pages/Home';
import LoginPage from './pages/LoginPage';
import InitPage from './pages/InitPage';
import DDNSList from './pages/DDNSList';
import DDNSLogs from './pages/DDNSLogs';

type AppState = 'loading' | 'init' | 'login' | 'ready';

export default function App() {
  const [state, setState] = useState<AppState>('loading');

  useEffect(() => {
    (async () => {
      try {
        const { initialized } = await auth.status();
        if (!initialized) { setState('init'); return; }
        const token = localStorage.getItem('token');
        if (!token) { setState('login'); return; }
        try {
          await auth.me();
          setState('ready');
        } catch {
          localStorage.removeItem('token');
          setState('login');
        }
      } catch {
        setState('login');
      }
    })();
  }, []);

  if (state === 'loading') {
    return (
      <ConfigProvider
        theme={{ token: { fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif" } }}
      >
        <div style={{ height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Spin size="large" />
        </div>
      </ConfigProvider>
    );
  }

  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#1677ff',
          borderRadius: 8,
          fontSize: 16,
          fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
        },
        components: {
          Layout: { headerBg: '#fff', siderBg: '#fff', bodyBg: '#f5f5f5' },
          Menu: { itemBg: 'transparent', itemFontSize: 16 },
          Table: { headerFontSize: 15, cellFontSize: 15 },
          Button: { fontSize: 15 },
          Form: { itemMarginBottom: 20, labelFontSize: 15 },
          Card: { headerFontSize: 18 },
          Select: { fontSize: 15 },
          Input: { fontSize: 15 },
          InputNumber: { fontSize: 15 },
          Tag: { fontSize: 13 },
        },
      }}
    >
      <BrowserRouter>
        <Routes>
          <Route
            path="/init"
            element={
              state === 'init'
                ? <InitPage onComplete={() => setState('login')} />
                : <Navigate to={state === 'ready' ? '/' : '/login'} replace />
            }
          />
          <Route
            path="/login"
            element={
              state === 'login'
                ? <LoginPage onComplete={() => setState('ready')} />
                : <Navigate to={state === 'init' ? '/init' : '/'} replace />
            }
          />
          <Route
            path="/"
            element={
              state === 'ready'
                ? <AdminLayout />
                : <Navigate to={state === 'init' ? '/init' : '/login'} replace />
            }
          >
            <Route index element={<Home />} />
            <Route path="ddns" element={<DDNSList />} />
            <Route path="ddns/:id/logs" element={<DDNSLogs />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  );
}
