import { useEffect, useState } from 'react';
import { Card, Row, Col, Typography } from 'antd';
import { CloudServerOutlined } from '@ant-design/icons';
import { api } from '../api/client';

export default function Home() {
  const [status, setStatus] = useState<string>('loading');

  useEffect(() => {
    api.get<{ status: string }>('/health')
      .then((data) => setStatus(data.status))
      .catch(() => setStatus('error'));
  }, []);

  const statusConfig = {
    loading: { label: '检测中...', dotClass: 'status-dot status-dot--loading' },
    ok: { label: '运行中', dotClass: 'status-dot status-dot--ok status-dot--pulse' },
    error: { label: '异常', dotClass: 'status-dot status-dot--error' },
  };

  const current = statusConfig[status as keyof typeof statusConfig] || statusConfig.error;

  const cards = [
    {
      key: 'backend',
      title: '后端状态',
      value: current.label,
      icon: <span className={current.dotClass} />,
      delay: '0.1s',
    },
    {
      key: 'nas',
      title: 'NAS 连接',
      value: '未连接',
      icon: <CloudServerOutlined style={{ color: 'var(--text-muted)', fontSize: 22 }} />,
      delay: '0.2s',
    },
    {
      key: 'storage',
      title: '存储空间',
      value: '-- GB',
      icon: null,
      delay: '0.3s',
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 32 }}>
        <Typography.Title level={4} style={{ margin: 0, fontWeight: 600 }}>
          控制台
        </Typography.Title>
        <Typography.Text style={{ color: 'var(--text-muted)', marginTop: 4, display: 'block', fontSize: 14 }}>
          NAS Partner 系统概览
        </Typography.Text>
      </div>

      <Row gutter={[24, 24]}>
        {cards.map((card) => (
          <Col key={card.key} xs={24} sm={12} lg={8}>
            <div style={{ animation: `fade-up 0.6s ease-out ${card.delay} both` }}>
              <Card
                style={{
                  borderRadius: 12,
                  border: `1px solid var(--border)`,
                }}
                styles={{
                  body: { padding: 28 },
                }}
              >
                <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
                  <div>
                    <div style={{ fontSize: 13, color: 'var(--text-secondary)', marginBottom: 10, fontWeight: 500 }}>
                      {card.title}
                    </div>
                    <div style={{
                      fontSize: 28,
                      fontWeight: 700,
                      color: 'var(--text-primary)',
                      letterSpacing: 0.5,
                      lineHeight: 1.2,
                    }}>
                      {card.value}
                    </div>
                  </div>
                  {card.icon && (
                    <div
                      style={{
                        width: 44,
                        height: 44,
                        borderRadius: 12,
                        background: 'var(--accent-soft)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                      }}
                    >
                      {card.icon}
                    </div>
                  )}
                </div>
              </Card>
            </div>
          </Col>
        ))}
      </Row>
    </div>
  );
}
