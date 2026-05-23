import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Table, Button, Card, Tag, Space, Switch, Popconfirm, Tooltip, Typography,
} from 'antd';
import {
  PlusOutlined, PlayCircleOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, HistoryOutlined,
} from '@ant-design/icons';
import { ddns } from '../api/ddns';
import type { DDNSConfigWithLog } from '../api/ddns';
import DDNSFormModal from './DDNSForm';

const providerLabels: Record<string, string> = {
  alidns: '阿里云 DNS',
  cloudflare: 'Cloudflare',
  dnspod: 'DNSPod',
  tencentcloud: '腾讯云 DNS',
  huaweicloud: '华为云 DNS',
};

const statusColors: Record<string, string> = {
  成功: 'success',
  失败: 'error',
  未改变: 'default',
};

export default function DDNSList() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<DDNSConfigWithLog[]>([]);
  const [formOpen, setFormOpen] = useState(false);
  const [formEditId, setFormEditId] = useState<number | null>(null);

  const fetchData = async () => {
    setLoading(true);
    try {
      const list = await ddns.listWithLogs();
      setData(list);
    } catch {
      // handled by api client
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  const handleToggle = async (id: number) => {
    try {
      await ddns.toggle(id);
      fetchData();
    } catch { /* handled */ }
  };

  const handleRun = async (id: number) => {
    try {
      await ddns.run(id);
      fetchData();
    } catch { /* handled */ }
  };

  const handleDelete = async (id: number) => {
    try {
      await ddns.delete(id);
      fetchData();
    } catch { /* handled */ }
  };

  const openNewForm = () => {
    setFormEditId(null);
    setFormOpen(true);
  };

  const openEditForm = (id: number) => {
    setFormEditId(id);
    setFormOpen(true);
  };

  const accentColor = 'var(--accent)';

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
      render: (name: string, record: DDNSConfigWithLog) => (
        <a onClick={() => openEditForm(record.id)} style={{ wordBreak: 'break-word', color: accentColor }}>
          {name}
        </a>
      ),
    },
    {
      title: '当前 IPv4',
      key: 'current_ipv4',
      width: 140,
      render: (_: unknown, record: DDNSConfigWithLog) => (
        <span style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: 'var(--text-secondary)' }}>
          {record.current_ipv4 || '-'}
        </span>
      ),
    },
    {
      title: '当前 IPv6',
      key: 'current_ipv6',
      width: 280,
      render: (_: unknown, record: DDNSConfigWithLog) => (
        <span style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: 'var(--text-secondary)' }}>
          {record.current_ipv6 || '-'}
        </span>
      ),
    },
    {
      title: 'DNS 服务商',
      dataIndex: 'dns_provider',
      key: 'dns_provider',
      width: 140,
      render: (v: string) => (
        <Tag style={{ borderRadius: 4, fontSize: 12 }}>{providerLabels[v] || v}</Tag>
      ),
    },
    {
      title: 'IPv4',
      key: 'ipv4',
      width: 60,
      render: (_: unknown, record: DDNSConfigWithLog) => (
        <Tag color={record.ipv4_enabled ? 'success' : 'default'} style={{ borderRadius: 4, fontSize: 12 }}>
          {record.ipv4_enabled ? 'ON' : 'OFF'}
        </Tag>
      ),
    },
    {
      title: 'IPv6',
      key: 'ipv6',
      width: 60,
      render: (_: unknown, record: DDNSConfigWithLog) => (
        <Tag color={record.ipv6_enabled ? 'success' : 'default'} style={{ borderRadius: 4, fontSize: 12 }}>
          {record.ipv6_enabled ? 'ON' : 'OFF'}
        </Tag>
      ),
    },
    {
      title: '域名',
      dataIndex: 'domains',
      key: 'domains',
      render: (domains: string[]) => domains?.length
        ? domains.map(d =>
            <div key={d} style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.6 }}>
              {d}
            </div>
          )
        : '-',
    },
    {
      title: '上次运行',
      key: 'last_run',
      width: 170,
      render: (_: unknown, record: DDNSConfigWithLog) => {
        const time = record.latest_log ? new Date(record.latest_log.created_at).toLocaleString('zh-CN') : '-';
        return <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>{time}</span>;
      },
    },
    {
      title: '状态',
      key: 'status',
      width: 90,
      render: (_: unknown, record: DDNSConfigWithLog) => {
        const status = record.latest_log?.status;
        const msg = record.latest_log?.message;
        return (
          <Tooltip title={msg || ''}>
            <Tag color={statusColors[status!] || 'default'} style={{ borderRadius: 4, fontSize: 12 }}>
              {status || '-'}
            </Tag>
          </Tooltip>
        );
      },
    },
    {
      title: '操作',
      key: 'actions',
      width: 260,
      render: (_: unknown, record: DDNSConfigWithLog) => (
        <Space size="small">
          <Tooltip title={record.enabled ? '禁用' : '启用'}>
            <Switch size="small" checked={record.enabled} onChange={() => handleToggle(record.id)} />
          </Tooltip>
          <Tooltip title="立即执行">
            <Button type="text" size="small" icon={<PlayCircleOutlined />} onClick={() => handleRun(record.id)} style={{ color: 'var(--text-secondary)' }} />
          </Tooltip>
          <Tooltip title="执行日志">
            <Button type="text" size="small" icon={<HistoryOutlined />} onClick={() => navigate(`/ddns/${record.id}/logs`)} style={{ color: 'var(--text-secondary)' }} />
          </Tooltip>
          <Tooltip title="编辑">
            <Button type="text" size="small" icon={<EditOutlined />} onClick={() => openEditForm(record.id)} style={{ color: 'var(--text-secondary)' }} />
          </Tooltip>
          <Popconfirm title="确认删除?">
            <Tooltip title="删除">
              <Button type="text" size="small" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ animation: 'fade-up 0.5s ease-out' }}>
      <div style={{ marginBottom: 24 }}>
        <Typography.Title level={4} style={{ margin: 0, fontWeight: 600 }}>
          DDNS 配置
        </Typography.Title>
        <Typography.Text style={{ color: 'var(--text-muted)', marginTop: 4, display: 'block', fontSize: 14 }}>
          管理动态 DNS 解析记录
        </Typography.Text>
      </div>

      <Card
        style={{
          borderRadius: 12,
          border: `1px solid var(--border)`,
        }}
        styles={{
          body: { padding: 0 },
        }}
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={fetchData}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={openNewForm}>
              新增配置
            </Button>
          </Space>
        }
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

      <DDNSFormModal
        open={formOpen}
        editId={formEditId}
        onClose={() => setFormOpen(false)}
        onSuccess={fetchData}
      />
    </div>
  );
}
