import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Table, Card, Tag, Button, Space, Typography, message, Popconfirm } from 'antd';
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
      .catch(() => message.error('获取日志失败'))
      .finally(() => setLoading(false));
  };

  useEffect(() => { fetchLogs(); }, [id]);

  const handleClear = async () => {
    if (!id) return;
    try {
      await ddns.clearLogs(Number(id));
      message.success('已清理所有日志');
      fetchLogs();
    } catch {
      message.error('清理失败');
    }
  };

  const columns = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (v: string) => new Date(v).toLocaleString('zh-CN'),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (v: string) => <Tag color={statusColors[v] || 'default'}>{v || '-'}</Tag>,
    },
    {
      title: '信息',
      dataIndex: 'message',
      key: 'message',
      render: (v: string) => v ? (
        <pre style={{ margin: 0, whiteSpace: 'pre-wrap', fontFamily: 'inherit', fontSize: 'inherit' }}>{v}</pre>
      ) : '-',
    },
  ];

  return (
    <Card
      title={
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/ddns')} type="text" />
          <Typography.Title level={5} style={{ margin: 0 }}>执行日志</Typography.Title>
        </Space>
      }
      extra={
        <Popconfirm title="确认清理所有日志？" onConfirm={handleClear}>
          <Button icon={<DeleteOutlined />} danger>清理日志</Button>
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
  );
}
