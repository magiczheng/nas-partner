import { useState } from 'react';
import { Form, Input, Button, Typography, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { auth } from '../api/auth';

interface Props {
  onComplete: () => void;
}

export default function InitPage({ onComplete }: Props) {
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await auth.init(values.username, values.password);
      localStorage.removeItem('token');
      message.success('管理员账号创建成功，请登录');
      onComplete();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '创建失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-bg" style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <div
        className="glass-card"
        style={{
          width: 440,
          padding: 48,
          boxShadow: '0 16px 48px rgba(0,0,0,0.15)',
          animation: 'fade-up 0.6s ease-out',
        }}
      >
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
            创建初始管理员账号
          </Typography.Text>
        </div>

        <Form name="init" onFinish={onFinish} layout="vertical" size="large" autoComplete="off">
          <Form.Item
            name="username"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少 3 个字符' },
            ]}
          >
            <Input
              prefix={<UserOutlined style={{ color: 'var(--text-muted)' }} />}
              placeholder="管理员用户名"
            />
          </Form.Item>
          <Form.Item
            name="password"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码至少 6 位' },
            ]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: 'var(--text-muted)' }} />}
              placeholder="密码（至少 6 位）"
            />
          </Form.Item>
          <Form.Item
            name="confirm"
            dependencies={['password']}
            rules={[
              { required: true, message: '请确认密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) return Promise.resolve();
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: 'var(--text-muted)' }} />}
              placeholder="确认密码"
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
              创建管理员账号
            </Button>
          </Form.Item>
        </Form>
      </div>
    </div>
  );
}
