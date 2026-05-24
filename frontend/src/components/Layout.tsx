import { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Button, Dropdown, Avatar, Modal, Form, Input, message } from 'antd';
import {
  DashboardOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  UserOutlined,
  LogoutOutlined,
  LockOutlined,
  CloudServerOutlined,
  ContainerOutlined,
  SafetyCertificateOutlined,
  SunOutlined,
  MoonOutlined,
} from '@ant-design/icons';
import { useTheme } from '../theme/ThemeContext';
import { auth } from '../api/auth';

const { Header, Sider, Content } = Layout;

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '控制台' },
  { key: '/ddns', icon: <CloudServerOutlined />, label: 'DDNS' },
  { key: '/docker', icon: <ContainerOutlined />, label: 'Docker' },
  { key: '/audit', icon: <SafetyCertificateOutlined />, label: '审计日志' },
];

export default function AdminLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const [passwordModalOpen, setPasswordModalOpen] = useState(false);
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordForm] = Form.useForm();
  const navigate = useNavigate();
  const location = useLocation();
  const { isDark, toggleTheme } = useTheme();

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login', { replace: true });
    window.location.reload();
  };

  const handleChangePassword = async (values: { oldPassword: string; newPassword: string }) => {
    setPasswordLoading(true);
    try {
      await auth.changePassword(values.oldPassword, values.newPassword);
      message.success('密码修改成功');
      setPasswordModalOpen(false);
      passwordForm.resetFields();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '修改失败');
    } finally {
      setPasswordLoading(false);
    }
  };

  const userDropdownItems = {
    items: [
      { key: 'profile', icon: <UserOutlined />, label: '修改密码' },
      { type: 'divider' as const },
      { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true },
    ],
    onClick: ({ key }: { key: string }) => {
      if (key === 'profile') {
        setPasswordModalOpen(true);
      } else if (key === 'logout') {
        handleLogout();
      }
    },
  };

  return (
    <Layout style={{ height: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        width={240}
        style={{
          borderRight: `1px solid var(--border)`,
          transition: 'background-color 0.3s ease, border-color 0.3s ease',
        }}
      >
        {/* Logo */}
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            gap: collapsed ? 0 : 12,
            padding: collapsed ? '0 16px' : '0 20px',
            borderBottom: `1px solid var(--border)`,
            justifyContent: collapsed ? 'center' : 'flex-start',
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              width: 34,
              height: 34,
              borderRadius: 10,
              background: isDark
                ? 'linear-gradient(135deg, #2dd4bf, #0d9488)'
                : 'linear-gradient(135deg, #14b8a6, #0d9488)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 14,
              fontWeight: 700,
              color: '#fff',
              fontFamily: "'DM Sans', sans-serif",
              flexShrink: 0,
            }}
          >
            NP
          </div>
          {!collapsed && (
            <span
              style={{
                fontSize: 18,
                fontWeight: 700,
                color: 'var(--text-primary)',
                letterSpacing: 0.5,
                whiteSpace: 'nowrap',
                transition: 'color 0.3s ease',
              }}
            >
              NAS Partner
            </span>
          )}
        </div>

        {/* Menu */}
        <Menu
          mode="inline"
          selectedKeys={[location.pathname === '/ddns' ? '/ddns' : location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
          style={{
            borderRight: 0,
            marginTop: 4,
            background: 'transparent',
          }}
        />

        {/* Bottom status */}
        {!collapsed && (
          <div
            style={{
              position: 'absolute',
              bottom: 0,
              left: 0,
              right: 0,
              padding: '16px 20px',
              borderTop: `1px solid var(--border)`,
              display: 'flex',
              alignItems: 'center',
              gap: 8,
            }}
          >
            <span className="status-dot status-dot--ok status-dot--pulse" />
            <span style={{ fontSize: 12, color: 'var(--text-muted)', transition: 'color 0.3s ease' }}>
              系统运行正常
            </span>
          </div>
        )}
      </Sider>

      <Layout>
        {/* Glass Header */}
        <Header
          style={{
            padding: '0 24px',
            background: isDark ? 'rgba(13, 15, 22, 0.7)' : 'rgba(255, 255, 255, 0.8)',
            backdropFilter: 'blur(16px)',
            WebkitBackdropFilter: 'blur(16px)',
            borderBottom: `1px solid var(--border)`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            height: 64,
            position: 'sticky',
            top: 0,
            zIndex: 10,
            transition: 'background-color 0.3s ease',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              style={{
                width: 40,
                height: 40,
                color: 'var(--text-secondary)',
                fontSize: 18,
              }}
            />

            {/* Theme toggle */}
            <Button
              type="text"
              icon={isDark ? <SunOutlined /> : <MoonOutlined />}
              onClick={toggleTheme}
              style={{
                width: 40,
                height: 40,
                color: 'var(--text-secondary)',
                fontSize: 18,
              }}
            />
          </div>

          <Dropdown menu={userDropdownItems} placement="bottomRight">
            <div
              style={{
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                padding: '4px 12px',
                borderRadius: 8,
                transition: 'background 0.2s',
              }}
              onMouseEnter={(e) => { e.currentTarget.style.background = isDark ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.04)'; }}
              onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
            >
              <Avatar
                size={32}
                icon={<UserOutlined />}
                style={{
                  background: isDark ? 'rgba(45, 212, 191, 0.15)' : 'rgba(20, 184, 166, 0.12)',
                  color: isDark ? '#2dd4bf' : '#14b8a6',
                }}
              />
              <span style={{ color: 'var(--text-primary)', fontSize: 14, transition: 'color 0.3s ease' }}>
                管理员
              </span>
            </div>
          </Dropdown>
        </Header>

        {/* Content with grid bg */}
        <Content
          className="admin-content"
          style={{
            padding: 28,
            overflow: 'auto',
            position: 'relative',
            transition: 'background-color 0.3s ease',
          }}
        >
          <div style={{ position: 'relative', zIndex: 1, minHeight: '100%' }}>
            <Outlet />
          </div>
        </Content>
      </Layout>

      {/* Password Change Modal */}
      <Modal
        title="修改密码"
        open={passwordModalOpen}
        onCancel={() => {
          setPasswordModalOpen(false);
          passwordForm.resetFields();
        }}
        footer={null}
        destroyOnClose
      >
        <Form
          form={passwordForm}
          layout="vertical"
          onFinish={handleChangePassword}
          style={{ marginTop: 8 }}
        >
          <Form.Item
            name="oldPassword"
            label="当前密码"
            rules={[{ required: true, message: '请输入当前密码' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="当前密码" />
          </Form.Item>
          <Form.Item
            name="newPassword"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少 6 位' },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="新密码（至少 6 位）" />
          </Form.Item>
          <Form.Item
            name="confirm"
            label="确认新密码"
            dependencies={['newPassword']}
            rules={[
              { required: true, message: '请确认新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) return Promise.resolve();
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="确认新密码" />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Button onClick={() => {
              setPasswordModalOpen(false);
              passwordForm.resetFields();
            }}>
              取消
            </Button>
            <Button type="primary" htmlType="submit" loading={passwordLoading} style={{ marginLeft: 8 }}>
              确认修改
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </Layout>
  );
}
