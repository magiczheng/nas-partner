import { useEffect, useState } from 'react';
import { Card, Statistic, Row, Col, Typography, Spin } from 'antd';
import { CloudServerOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import { api } from '../api/client';

export default function Home() {
  const [status, setStatus] = useState<string>('loading');

  useEffect(() => {
    api.get<{ status: string }>('/health')
      .then((data) => setStatus(data.status))
      .catch(() => setStatus('error'));
  }, []);

  return (
    <div>
      <Typography.Title level={4} style={{ marginBottom: 24 }}>控制台</Typography.Title>
      <Row gutter={[24, 24]}>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="后端状态"
              value={status === 'loading' ? '检测中...' : status === 'ok' ? '运行中' : '异常'}
              prefix={status === 'loading' ? <Spin /> : status === 'ok' ? <CheckCircleOutlined style={{ color: '#52c41a' }} /> : <CloseCircleOutlined style={{ color: '#ff4d4f' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic title="NAS 连接" value="未连接" prefix={<CloudServerOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic title="存储空间" value="--" suffix="GB" />
          </Card>
        </Col>
      </Row>
    </div>
  );
}
