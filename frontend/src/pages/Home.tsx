import { useEffect, useState, useCallback } from 'react';
import { Card, Row, Col, Typography, Progress } from 'antd';
import {
  CloudServerOutlined,
  HddOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  DesktopOutlined,
  DashboardOutlined,
} from '@ant-design/icons';
import { system } from '../api/system';
import type { SystemInfo } from '../api/system';

function formatUptime(seconds: number): string {
  const d = Math.floor(seconds / 86400);
  const h = Math.floor((seconds % 86400) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const parts: string[] = [];
  if (d > 0) parts.push(`${d}天`);
  if (h > 0) parts.push(`${h}时`);
  if (parts.length === 0) parts.push(`${m}分`);
  return parts.join(' ');
}

function formatBytes(bytes: number): string {
  const gb = bytes / 1024 / 1024 / 1024;
  if (gb >= 1024) return (gb / 1024).toFixed(1) + ' TB';
  return gb.toFixed(1) + ' GB';
}

function CardItem({ icon, title, children }: { icon: React.ReactNode; title: string; children: React.ReactNode }) {
  return (
    <Card
      style={{
        borderRadius: 12,
        border: '1px solid var(--border)',
        height: '100%',
      }}
      styles={{ body: { padding: 24 } }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
        <div
          style={{
            width: 40,
            height: 40,
            borderRadius: 10,
            background: 'var(--accent-soft)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: 18,
            color: 'var(--color-primary)',
          }}
        >
          {icon}
        </div>
        <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-primary)' }}>{title}</span>
      </div>
      {children}
    </Card>
  );
}

export default function Home() {
  const [info, setInfo] = useState<SystemInfo | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchInfo = useCallback(async () => {
    try {
      const data = await system.getInfo();
      setInfo(data);
    } catch {
      // handled by api client
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchInfo();
    const timer = setInterval(fetchInfo, 30000);
    return () => clearInterval(timer);
  }, [fetchInfo]);

  const cpuColor =
    info && info.cpu.percent > 90 ? '#ef4444' : info && info.cpu.percent > 70 ? '#f59e0b' : '#14b8a6';
  const memColor =
    info && info.memory.used_percent > 90 ? '#ef4444' : info && info.memory.used_percent > 70 ? '#f59e0b' : '#14b8a6';
  const diskColor =
    info && info.disk.used_percent > 90 ? '#ef4444' : info && info.disk.used_percent > 70 ? '#f59e0b' : '#14b8a6';

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
        {/* CPU */}
        <Col xs={24} sm={12} lg={8}>
          <div style={{ animation: 'fade-up 0.6s ease-out 0.1s both', height: '100%' }}>
            <CardItem icon={<ThunderboltOutlined />} title="CPU">
              {info ? (
                <>
                  <div style={{ fontSize: 32, fontWeight: 700, color: cpuColor, marginBottom: 8 }}>
                    {info.cpu.percent.toFixed(1)}%
                  </div>
                  <Progress
                    percent={info.cpu.percent}
                    showInfo={false}
                    strokeColor={cpuColor}
                    trailColor="var(--border)"
                  />
                  <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 8 }}>
                    {info.cpu.cores} 核心
                  </div>
                </>
              ) : (
                <div style={{ color: 'var(--text-muted)', fontSize: 14 }}>{loading ? '加载中...' : '不可用'}</div>
              )}
            </CardItem>
          </div>
        </Col>

        {/* Memory */}
        <Col xs={24} sm={12} lg={8}>
          <div style={{ animation: 'fade-up 0.6s ease-out 0.2s both', height: '100%' }}>
            <CardItem icon={<DashboardOutlined />} title="内存">
              {info ? (
                <>
                  <div style={{ fontSize: 32, fontWeight: 700, color: memColor, marginBottom: 8 }}>
                    {info.memory.used_percent.toFixed(1)}%
                  </div>
                  <Progress
                    percent={info.memory.used_percent}
                    showInfo={false}
                    strokeColor={memColor}
                    trailColor="var(--border)"
                  />
                  <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 8 }}>
                    {formatBytes(info.memory.used)} / {formatBytes(info.memory.total)}
                  </div>
                </>
              ) : (
                <div style={{ color: 'var(--text-muted)', fontSize: 14 }}>{loading ? '加载中...' : '不可用'}</div>
              )}
            </CardItem>
          </div>
        </Col>

        {/* Disk */}
        <Col xs={24} sm={12} lg={8}>
          <div style={{ animation: 'fade-up 0.6s ease-out 0.3s both', height: '100%' }}>
            <CardItem icon={<HddOutlined />} title="磁盘">
              {info ? (
                <>
                  <div style={{ fontSize: 32, fontWeight: 700, color: diskColor, marginBottom: 8 }}>
                    {info.disk.used_percent.toFixed(1)}%
                  </div>
                  <Progress
                    percent={info.disk.used_percent}
                    showInfo={false}
                    strokeColor={diskColor}
                    trailColor="var(--border)"
                  />
                  <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 8 }}>
                    {formatBytes(info.disk.used)} / {formatBytes(info.disk.total)}
                  </div>
                  <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>路径: {info.disk.path}</div>
                </>
              ) : (
                <div style={{ color: 'var(--text-muted)', fontSize: 14 }}>{loading ? '加载中...' : '不可用'}</div>
              )}
            </CardItem>
          </div>
        </Col>

        {/* System info */}
        <Col xs={24} sm={12} lg={8}>
          <div style={{ animation: 'fade-up 0.6s ease-out 0.4s both', height: '100%' }}>
            <CardItem icon={<DesktopOutlined />} title="系统">
              {info ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)', fontSize: 13 }}>主机名</span>
                    <span style={{ color: 'var(--text-primary)', fontSize: 13, fontWeight: 500 }}>{info.hostname}</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)', fontSize: 13 }}>系统</span>
                    <span style={{ color: 'var(--text-primary)', fontSize: 13, fontWeight: 500 }}>{info.os}</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)', fontSize: 13 }}>运行时长</span>
                    <span style={{ color: 'var(--text-primary)', fontSize: 13, fontWeight: 500 }}>
                      {formatUptime(info.uptime)}
                    </span>
                  </div>
                </div>
              ) : (
                <div style={{ color: 'var(--text-muted)', fontSize: 14 }}>{loading ? '加载中...' : '不可用'}</div>
              )}
            </CardItem>
          </div>
        </Col>

        {/* Load Average */}
        <Col xs={24} sm={12} lg={8}>
          <div style={{ animation: 'fade-up 0.6s ease-out 0.5s both', height: '100%' }}>
            <CardItem icon={<ClockCircleOutlined />} title="系统负载">
              {info ? (
                <div style={{ display: 'flex', gap: 24 }}>
                  <div>
                    <div style={{ fontSize: 24, fontWeight: 700, color: 'var(--text-primary)' }}>
                      {info.load1.toFixed(2)}
                    </div>
                    <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>1分钟</div>
                  </div>
                  <div>
                    <div style={{ fontSize: 24, fontWeight: 700, color: 'var(--text-secondary)' }}>
                      {info.load5.toFixed(2)}
                    </div>
                    <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>5分钟</div>
                  </div>
                  <div>
                    <div style={{ fontSize: 24, fontWeight: 700, color: 'var(--text-muted)' }}>
                      {info.load15.toFixed(2)}
                    </div>
                    <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>15分钟</div>
                  </div>
                </div>
              ) : (
                <div style={{ color: 'var(--text-muted)', fontSize: 14 }}>{loading ? '加载中...' : '不可用'}</div>
              )}
            </CardItem>
          </div>
        </Col>

        {/* Backend Status */}
        <Col xs={24} sm={12} lg={8}>
          <div style={{ animation: 'fade-up 0.6s ease-out 0.6s both', height: '100%' }}>
            <CardItem icon={<CloudServerOutlined />} title="后端状态">
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <span className="status-dot status-dot--ok status-dot--pulse" />
                <span style={{ fontSize: 14, fontWeight: 500, color: 'var(--text-primary)' }}>
                  {info ? '运行中' : loading ? '检测中...' : '异常'}
                </span>
              </div>
              <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 8 }}>
                Ver {info?.os?.split(' ')[0] || '--'}
              </div>
            </CardItem>
          </div>
        </Col>
      </Row>
    </div>
  );
}
