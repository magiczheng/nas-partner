import { useEffect, useState, useCallback } from 'react';
import { Table, Card, Tag, Button, Typography, Space, Progress } from 'antd';
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

const stateLabels: Record<string, string> = {
  running: '运行中',
  exited: '已停止',
  paused: '已暂停',
  created: '已创建',
  restarting: '重启中',
  removing: '删除中',
  dead: '异常',
};

function formatBytes(bytes: number): string {
  if (bytes < 0) return '0 B/s';
  const abs = bytes;
  if (abs < 1024) return abs + ' B/s';
  if (abs < 1024 * 1024) return (abs / 1024).toFixed(1) + ' KB/s';
  return (abs / 1024 / 1024).toFixed(1) + ' MB/s';
}

function formatMemory(bytes: number): string {
  const mb = bytes / 1024 / 1024;
  if (mb < 1024) return mb.toFixed(0) + ' MB';
  return (mb / 1024).toFixed(1) + ' GB';
}

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
      title: '状态',
      key: 'state',
      width: 100,
      render: (_: unknown, record: ContainerInfo) => (
        <Tag color={stateColors[record.state] || 'default'} style={{ borderRadius: 4, fontSize: 12 }}>
          {stateLabels[record.state] || record.state}
        </Tag>
      ),
    },
    {
      title: '端口',
      key: 'ports',
      width: 180,
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
      title: 'CPU',
      key: 'cpu',
      width: 100,
      sorter: (a: ContainerInfo, b: ContainerInfo) => a.cpu_percent - b.cpu_percent,
      render: (_: unknown, record: ContainerInfo) => {
        const pct = record.cpu_percent || 0;
        return (
          <span style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: 'var(--text-secondary)' }}>
            {pct.toFixed(1)}%
          </span>
        );
      },
    },
    {
      title: '内存',
      key: 'memory',
      width: 180,
      sorter: (a: ContainerInfo, b: ContainerInfo) => a.memory_usage - b.memory_usage,
      render: (_: unknown, record: ContainerInfo) => {
        const used = record.memory_usage || 0;
        const limit = record.memory_limit || 0;
        const usedMB = used / 1024 / 1024;
        const limitMB = limit / 1024 / 1024;
        const pct = limit > 0 ? (used / limit) * 100 : 0;
        return (
          <div style={{ minWidth: 120 }}>
            <div style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: 'var(--text-secondary)', marginBottom: 2 }}>
              {formatMemory(used)}
              {limit > 0 && <span style={{ color: 'var(--text-muted)' }}> / {formatMemory(limit)}</span>}
            </div>
            {limit > 0 && (
              <Progress
                percent={pct}
                size="small"
                showInfo={false}
                strokeColor={pct > 90 ? '#ef4444' : pct > 70 ? '#f59e0b' : '#14b8a6'}
                trailColor="var(--border)"
              />
            )}
          </div>
        );
      },
    },
    {
      title: '网络 ↓',
      key: 'network_rx',
      width: 110,
      sorter: (a: ContainerInfo, b: ContainerInfo) => a.network_rx - b.network_rx,
      render: (_: unknown, record: ContainerInfo) => (
        <span style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: 'var(--text-secondary)' }}>
          {formatBytes(record.network_rx || 0)}
        </span>
      ),
    },
    {
      title: '网络 ↑',
      key: 'network_tx',
      width: 110,
      sorter: (a: ContainerInfo, b: ContainerInfo) => a.network_tx - b.network_tx,
      render: (_: unknown, record: ContainerInfo) => (
        <span style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13, color: 'var(--text-secondary)' }}>
          {formatBytes(record.network_tx || 0)}
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
