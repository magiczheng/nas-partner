import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Table, Button, Card, Tag, Space, Switch, Popconfirm, message, Tooltip, Typography,
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
      message.error('获取列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  const handleToggle = async (id: number) => {
    try {
      await ddns.toggle(id);
      message.success('操作成功');
      fetchData();
    } catch {
      message.error('操作失败');
    }
  };

  const handleRun = async (id: number) => {
    try {
      const result = await ddns.run(id);
      if (result.status === '成功') {
        message.success('执行成功');
      } else if (result.status === '失败') {
        message.error('执行失败');
      } else {
        message.info('执行未改变');
      }
      fetchData();
    } catch {
      message.error('执行失败');
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await ddns.delete(id);
      message.success('已删除');
      fetchData();
    } catch {
      message.error('删除失败');
    }
  };

  const openNewForm = () => {
    setFormEditId(null);
    setFormOpen(true);
  };

  const openEditForm = (id: number) => {
    setFormEditId(id);
    setFormOpen(true);
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
      render: (name: string, record: DDNSConfigWithLog) => (
        <a onClick={() => openEditForm(record.id)} style={{ wordBreak: 'break-word' }}>{name}</a>
      ),
    },
    {
      title: '当前 IPv4',
      key: 'current_ipv4',
      width: 140,
      render: (_: unknown, record: DDNSConfigWithLog) => record.current_ipv4 || '-',
    },
    {
      title: '当前 IPv6',
      key: 'current_ipv6',
      width: 280,
      render: (_: unknown, record: DDNSConfigWithLog) => record.current_ipv6 || '-',
    },
    {
      title: 'DNS 服务商',
      dataIndex: 'dns_provider',
      key: 'dns_provider',
      width: 140,
      render: (v: string) => providerLabels[v] || v,
    },
    {
      title: 'IPv4',
      key: 'ipv4',
      width: 60,
      render: (_: unknown, record: DDNSConfigWithLog) => (
        <Tag color={record.ipv4_enabled ? 'blue' : 'default'}>
          {record.ipv4_enabled ? 'ON' : 'OFF'}
        </Tag>
      ),
    },
    {
      title: 'IPv6',
      key: 'ipv6',
      width: 60,
      render: (_: unknown, record: DDNSConfigWithLog) => (
        <Tag color={record.ipv6_enabled ? 'blue' : 'default'}>
          {record.ipv6_enabled ? 'ON' : 'OFF'}
        </Tag>
      ),
    },
    {
      title: '域名',
      dataIndex: 'domains',
      key: 'domains',
      render: (domains: string[]) => domains?.length
        ? domains.map(d => <div key={d}>{d}</div>)
        : '-',
    },
    {
      title: '上次运行',
      key: 'last_run',
      width: 170,
      render: (_: unknown, record: DDNSConfigWithLog) =>
        record.latest_log ? new Date(record.latest_log.created_at).toLocaleString('zh-CN') : '-',
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
            <Tag color={statusColors[status!] || 'default'}>{status || '-'}</Tag>
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
            <Switch
              size="small"
              checked={record.enabled}
              onChange={() => handleToggle(record.id)}
            />
          </Tooltip>
          <Tooltip title="立即执行">
            <Button
              type="text"
              size="small"
              icon={<PlayCircleOutlined />}
              onClick={() => handleRun(record.id)}
            />
          </Tooltip>
          <Tooltip title="执行日志">
            <Button
              type="text"
              size="small"
              icon={<HistoryOutlined />}
              onClick={() => navigate(`/ddns/${record.id}/logs`)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => openEditForm(record.id)}
            />
          </Tooltip>
          <Popconfirm title="确认删除?" onConfirm={() => handleDelete(record.id)}>
            <Tooltip title="删除">
              <Button type="text" size="small" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={<Typography.Title level={5} style={{ margin: 0 }}>DDNS 配置</Typography.Title>}
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
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

      <DDNSFormModal
        open={formOpen}
        editId={formEditId}
        onClose={() => setFormOpen(false)}
        onSuccess={fetchData}
      />
    </Card>
  );
}
