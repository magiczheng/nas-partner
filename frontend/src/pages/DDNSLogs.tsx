import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Table, Card, Tag, Button, Space, Typography, Popconfirm } from 'antd';
import { ArrowLeftOutlined, DeleteOutlined } from '@ant-design/icons';
import { ddns } from '../api/ddns';
import type { DDNSRunLog } from '../api/ddns';

const statusColors: Record<string, string> = {
  成功: 'success',
  失败: 'error',
  未改变: 'default',
};

export default function DDNSLogs() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [logs, setLogs] = useState<DDNSRunLog[]>([]);

  const fetchLogs = () => {
    if (!id) return;
    setLoading(true);
    ddns.listLogs(Number(id))
      .then(setLogs)
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  useEffect(() => { fetchLogs(); }, [id]);

  const handleClear = async () => {
    if (!id) return;
    try {
      await ddns.clearLogs(Number(id));
      fetchLogs();
    } catch { /* handled */ }
  };

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
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (v: string) => (
        <Tag color={statusColors[v] || 'default'} style={{ borderRadius: 4, fontSize: 12 }}>
          {v || '-'}
        </Tag>
      ),
    },
    {
      title: '信息',
      dataIndex: 'message',
      key: 'message',
      render: (v: string) => v ? (
        <pre style={{
          margin: 0,
          whiteSpace: 'pre-wrap',
          fontFamily: "'JetBrains Mono', monospace",
          fontSize: 13,
          color: 'var(--text-secondary)',
          lineHeight: 1.6,
        }}>
          {v}
        </pre>
      ) : '-',
    },
  ];

  return (
    <div style={{ animation: 'fade-up 0.5s ease-out' }}>
      <div style={{ marginBottom: 24 }}>
        <Space align="center" style={{ marginBottom: 4 }}>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/ddns')}
            type="text"
            style={{ color: 'var(--text-secondary)', marginLeft: -8 }}
          />
          <Typography.Title level={4} style={{ margin: 0, fontWeight: 600 }}>
            执行日志
          </Typography.Title>
        </Space>
        <Typography.Text style={{ color: 'var(--text-muted)', marginTop: 4, display: 'block', fontSize: 14 }}>
          DDNS 任务历史运行记录
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
          <Popconfirm title="确认清理所有日志？" onConfirm={handleClear}>
            <Button icon={<DeleteOutlined />} danger>
              清理日志
            </Button>
          </Popconfirm>
        }
      >
        <Table
          rowKey="id"
          columns={columns}
          dataSource={logs}
          loading={loading}
          pagination={{ pageSize: 20, showSizeChanger: true }}
          size="middle"
        />
      </Card>
    </div>
  );
}
