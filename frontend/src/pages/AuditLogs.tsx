import { useEffect, useState } from 'react';
import { Table, Card, Tag, Typography } from 'antd';
import { audit } from '../api/audit';
import type { AuditEntry } from '../api/audit';

const actionLabels: Record<string, string> = {
  login_success: '登录成功',
  login_failed: '登录失败',
  password_changed: '修改密码',
  ddns_create: '创建配置',
  ddns_update: '更新配置',
  ddns_delete: '删除配置',
  ddns_toggle: '切换启停',
  ddns_run: '手动执行',
};

const actionColors: Record<string, string> = {
  login_success: 'success',
  login_failed: 'error',
  password_changed: 'warning',
  ddns_create: 'processing',
  ddns_update: 'processing',
  ddns_delete: 'error',
  ddns_toggle: 'warning',
  ddns_run: 'cyan',
};

export default function AuditLogs() {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<AuditEntry[]>([]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const list = await audit.listLogs(200);
      setData(list);
    } catch {
      // handled by api client
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  const columns = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (v: string) => (
        <span style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: 'var(--text-secondary)' }}>
          {new Date(v).toLocaleString('zh-CN')}
        </span>
      ),
    },
    {
      title: '用户',
      dataIndex: 'username',
      key: 'username',
      width: 120,
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: unknown, record: AuditEntry) => (
        <Tag color={actionColors[record.action] || 'default'} style={{ borderRadius: 4, fontSize: 12 }}>
          {actionLabels[record.action] || record.action}
        </Tag>
      ),
    },
    {
      title: '详情',
      dataIndex: 'detail',
      key: 'detail',
    },
    {
      title: 'IP',
      dataIndex: 'ip',
      key: 'ip',
      width: 140,
      render: (ip: string) => (
        <span style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: 'var(--text-muted)' }}>
          {ip}
        </span>
      ),
    },
  ];

  return (
    <div style={{ animation: 'fade-up 0.5s ease-out' }}>
      <div style={{ marginBottom: 24 }}>
        <Typography.Title level={4} style={{ margin: 0, fontWeight: 600 }}>
          审计日志
        </Typography.Title>
        <Typography.Text style={{ color: 'var(--text-muted)', marginTop: 4, display: 'block', fontSize: 14 }}>
          系统操作记录与安全审计
        </Typography.Text>
      </div>

      <Card
        style={{
          borderRadius: 12,
          border: '1px solid var(--border)',
        }}
        styles={{ body: { padding: 0 } }}
      >
        <Table
          rowKey="id"
          columns={columns}
          dataSource={data}
          loading={loading}
          pagination={{ pageSize: 20, showSizeChanger: true }}
          size="middle"
        />
      </Card>
    </div>
  );
}
