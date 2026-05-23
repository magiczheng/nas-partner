import { useState } from 'react';
import { Form, Input, Button, Typography, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { auth } from '../api/auth';

interface Props {
  onComplete: () => void;
}

export default function LoginPage({ onComplete }: Props) {
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      const res = await auth.login(values.username, values.password);
      localStorage.setItem('token', res.token);
      onComplete();
    } catch {
      message.error('用户名或密码错误');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-bg" style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <div
        className="glass-card"
        style={{
          width: 420,
          padding: 48,
          boxShadow: '0 16px 48px rgba(0,0,0,0.15)',
          animation: 'fade-up 0.6s ease-out',
        }}
      >
        {/* Logo */}
        <div style={{ textAlign: 'center', marginBottom: 40 }}>
          <div
            style={{
              width: 52,
              height: 52,
              borderRadius: 14,
              background: 'var(--accent-gradient)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 20,
              fontWeight: 700,
              color: '#fff',
              fontFamily: "'DM Sans', sans-serif",
              margin: '0 auto 16px',
            }}
          >
            NP
          </div>
          <Typography.Title level={3} style={{ margin: 0, letterSpacing: 0.5 }}>
            NAS Partner
          </Typography.Title>
          <Typography.Text style={{ color: 'var(--text-muted)', marginTop: 6, display: 'block', fontSize: 14 }}>
            请登录以继续
          </Typography.Text>
        </div>

        <Form name="login" onFinish={onFinish} layout="vertical" size="large" autoComplete="off">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input
              prefix={<UserOutlined style={{ color: 'var(--text-muted)' }} />}
              placeholder="用户名"
            />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password
              prefix={<LockOutlined style={{ color: 'var(--text-muted)' }} />}
              placeholder="密码"
            />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, marginTop: 8 }}>
            <Button
              type="primary"
              htmlType="submit"
              block
              loading={loading}
              size="large"
              style={{ height: 48, fontSize: 16, fontWeight: 600, borderRadius: 10 }}
            >
              登 录
            </Button>
          </Form.Item>
        </Form>
      </div>
    </div>
  );
}
