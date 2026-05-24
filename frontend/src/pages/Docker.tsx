import { useEffect, useState, useCallback } from 'react';
import { Table, Card, Tag, Button, Typography, Space } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { docker } from '../api/docker';
import type { ContainerInfo } from '../api/docker';

const stateColors: Record<string, string> = {
  running: 'success',
  exited: 'error',
  paused: 'warning',
  created: 'default',
  restarting: 'processing',
  removing: 'error',
  dead: 'error',
};

export default function Docker() {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<ContainerInfo[]>([]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const list = await docker.listContainers();
      setData(list);
    } catch {
      // handled by api client
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
    const timer = setInterval(fetchData, 30000);
    return () => clearInterval(timer);
  }, [fetchData]);

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      render: (name: string) => (
        <span style={{ fontWeight: 500, color: 'var(--text-primary)' }}>{name || '-'}</span>
      ),
    },
    {
      title: '镜像',
      dataIndex: 'image',
      key: 'image',
      width: 220,
      render: (image: string) => (
        <span style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: 'var(--text-secondary)' }}>
          {image}
        </span>
      ),
    },
    {
      title: '状态',
      key: 'state',
      width: 100,
      render: (_: unknown, record: ContainerInfo) => (
        <Tag color={stateColors[record.state] || 'default'} style={{ borderRadius: 4, fontSize: 12 }}>
          {record.state}
        </Tag>
      ),
    },
    {
      title: '运行信息',
      dataIndex: 'status',
      key: 'status',
      width: 200,
      render: (status: string) => (
        <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>{status}</span>
      ),
    },
    {
      title: '端口',
      key: 'ports',
      width: 200,
      render: (_: unknown, record: ContainerInfo) =>
        record.ports?.length
          ? record.ports.map((p) => (
              <Tag key={p} style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 12, borderRadius: 4, marginBottom: 4 }}>
                {p}
              </Tag>
            ))
          : '-',
    },
    {
      title: '运行时长',
      key: 'uptime',
      width: 150,
      render: (_: unknown, record: ContainerInfo) => (
        <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>{record.uptime || '-'}</span>
      ),
    },
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 130,
      render: (id: string) => (
        <span style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: 'var(--text-muted)' }}>
          {id}
        </span>
      ),
    },
  ];

  return (
    <div style={{ animation: 'fade-up 0.5s ease-out' }}>
      <div style={{ marginBottom: 24 }}>
        <Typography.Title level={4} style={{ margin: 0, fontWeight: 600 }}>
          Docker 容器
        </Typography.Title>
        <Typography.Text style={{ color: 'var(--text-muted)', marginTop: 4, display: 'block', fontSize: 14 }}>
          管理系统中的 Docker 容器状态
        </Typography.Text>
      </div>

      <Card
        style={{
          borderRadius: 12,
          border: '1px solid var(--border)',
        }}
        styles={{ body: { padding: 0 } }}
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={fetchData}>
              刷新
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
    </div>
  );
}
