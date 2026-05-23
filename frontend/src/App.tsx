import { useEffect, useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, theme } from 'antd';
import { auth } from './api/auth';
import { ThemeProvider, useTheme } from './theme/ThemeContext';
import AdminLayout from './components/Layout';
import Home from './pages/Home';
import LoginPage from './pages/LoginPage';
import InitPage from './pages/InitPage';
import DDNSList from './pages/DDNSList';
import DDNSLogs from './pages/DDNSLogs';

type AppState = 'loading' | 'init' | 'login' | 'ready';

/* ── Shared theme tokens ───────────────────────────── */
const sharedTokens = {
  borderRadius: 8,
  fontFamily: "'DM Sans', 'PingFang SC', 'Microsoft YaHei', sans-serif",
  wireframe: false,
};

/* ── Light Theme ───────────────────────────────────── */
const lightTheme = {
  algorithm: theme.defaultAlgorithm,
  token: {
    ...sharedTokens,
    colorPrimary: '#14b8a6',
    colorInfo: '#14b8a6',
    colorBgContainer: '#ffffff',
    colorBgElevated: '#ffffff',
    colorBorder: 'rgba(0,0,0,0.06)',
    colorText: '#0f172a',
    colorTextSecondary: '#64748b',
    colorTextHeading: '#0f172a',
    colorBgMask: 'rgba(0,0,0,0.4)',
  },
  components: {
    Layout: {
      headerBg: '#ffffff',
      siderBg: '#ffffff',
      bodyBg: '#f1f5f9',
    },
    Card: {
      colorBgContainer: '#ffffff',
      headerBg: 'transparent',
      borderRadiusLG: 12,
    },
    Menu: {
      itemBg: 'transparent',
      itemSelectedBg: 'rgba(20, 184, 166, 0.08)',
      itemHoverBg: 'rgba(0, 0, 0, 0.02)',
      itemColor: '#64748b',
      itemSelectedColor: '#14b8a6',
      itemFontSize: 15,
      subMenuItemBg: 'transparent',
    },
    Table: {
      headerBg: 'rgba(0,0,0,0.02)',
      headerColor: '#64748b',
      rowHoverBg: 'rgba(0,0,0,0.02)',
      borderColor: 'rgba(0,0,0,0.06)',
    },
    Modal: {
      contentBg: '#ffffff',
      headerBg: '#ffffff',
      titleColor: '#0f172a',
      borderRadiusLG: 12,
    },
    Button: {
      primaryShadow: '0 2px 12px rgba(20, 184, 166, 0.2)',
      primaryColor: '#ffffff',
    },
    Input: {
      colorBgContainer: '#f8fafc',
      activeBorderColor: '#14b8a6',
      hoverBorderColor: 'rgba(20,184,166,0.5)',
    },
    Select: {
      colorBgContainer: '#f8fafc',
      activeBorderColor: '#14b8a6',
      hoverBorderColor: 'rgba(20,184,166,0.5)',
      colorBorder: 'rgba(0,0,0,0.08)',
    },
    Tag: {
      colorBgContainer: 'transparent',
    },
    Switch: {
      colorPrimary: '#14b8a6',
    },
    Statistic: {
      contentFontSize: 28,
    },
    Form: {
      labelColor: '#64748b',
    },
    Tabs: {
      colorBgContainer: 'transparent',
    },
  },
};

/* ── Dark Theme ────────────────────────────────────── */
const darkTheme = {
  algorithm: theme.darkAlgorithm,
  token: {
    ...sharedTokens,
    colorPrimary: '#2dd4bf',
    colorInfo: '#2dd4bf',
    colorBgContainer: '#0d0f16',
    colorBgElevated: '#141827',
    colorBorder: 'rgba(255,255,255,0.06)',
    colorText: '#e2e8f0',
    colorTextSecondary: '#8892a8',
    colorTextHeading: '#f1f5f9',
    colorBgMask: 'rgba(0,0,0,0.6)',
  },
  components: {
    Layout: {
      headerBg: '#0d0f16',
      siderBg: '#090b12',
      bodyBg: '#07080c',
      headerHeight: 64,
    },
    Card: {
      colorBgContainer: 'rgba(13, 15, 22, 0.6)',
      headerBg: 'transparent',
      borderRadiusLG: 12,
    },
    Menu: {
      itemBg: 'transparent',
      itemSelectedBg: 'rgba(45, 212, 191, 0.1)',
      itemHoverBg: 'rgba(255, 255, 255, 0.04)',
      itemColor: '#8892a8',
      itemSelectedColor: '#2dd4bf',
      itemFontSize: 15,
      subMenuItemBg: 'transparent',
    },
    Table: {
      headerBg: 'rgba(255,255,255,0.02)',
      headerColor: '#8892a8',
      rowHoverBg: 'rgba(255,255,255,0.04)',
      borderColor: 'rgba(255,255,255,0.06)',
    },
    Modal: {
      contentBg: '#0d0f16',
      headerBg: '#0d0f16',
      titleColor: '#f1f5f9',
      borderRadiusLG: 12,
    },
    Button: {
      primaryShadow: '0 2px 12px rgba(45, 212, 191, 0.2)',
      primaryColor: '#07080c',
    },
    Input: {
      colorBgContainer: 'rgba(0,0,0,0.2)',
      activeBorderColor: '#2dd4bf',
      hoverBorderColor: 'rgba(45,212,191,0.5)',
    },
    Select: {
      colorBgContainer: 'rgba(0,0,0,0.2)',
      activeBorderColor: '#2dd4bf',
      hoverBorderColor: 'rgba(45,212,191,0.5)',
      colorBorder: 'rgba(255,255,255,0.08)',
    },
    Tag: {
      colorBgContainer: 'transparent',
    },
    Switch: {
      colorPrimary: '#2dd4bf',
    },
    Statistic: {
      contentFontSize: 28,
    },
    Form: {
      labelColor: '#8892a8',
    },
    Tabs: {
      colorBgContainer: 'transparent',
    },
  },
};

/* ── Inner app that consumes theme ─────────────────── */
function AppInner() {
  const [state, setState] = useState<AppState>('loading');
  const { isDark } = useTheme();
  const currentTheme = isDark ? darkTheme : lightTheme;

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
      <ConfigProvider theme={currentTheme}>
        <div
          style={{
            height: '100vh',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            background: isDark ? '#07080c' : '#f1f5f9',
            transition: 'background-color 0.3s ease',
            gap: 32,
          }}
        >
          <div
            style={{
              width: 60,
              height: 60,
              borderRadius: 16,
              background: isDark
                ? 'linear-gradient(135deg, #2dd4bf 0%, #0d9488 100%)'
                : 'linear-gradient(135deg, #14b8a6 0%, #0d9488 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 24,
              fontWeight: 700,
              color: '#fff',
              fontFamily: "'DM Sans', sans-serif",
              letterSpacing: 1,
              animation: 'logo-pulse 2s ease-in-out infinite',
            }}
          >
            NP
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            {[0, 1, 2].map((i) => (
              <div
                key={i}
                style={{
                  width: 8,
                  height: 8,
                  borderRadius: '50%',
                  background: isDark ? '#2dd4bf' : '#14b8a6',
                  animation: `loading-dot 1.4s ease-in-out ${i * 0.2}s infinite`,
                }}
              />
            ))}
          </div>
        </div>
      </ConfigProvider>
    );
  }

  return (
    <ConfigProvider theme={currentTheme}>
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

/* ── Root ──────────────────────────────────────────── */
export default function App() {
  return (
    <ThemeProvider>
      <AppInner />
    </ThemeProvider>
  );
}
