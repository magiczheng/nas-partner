import { useEffect, useMemo, useState, useCallback } from 'react';
import { Form, Input, Select, Button, Modal, Space, message, Alert, Spin } from 'antd';
import { ddns } from '../api/ddns';
import type { CreateDDNSConfig, DDNSConfig, NetInterface } from '../api/ddns';

type ProviderKey = 'alidns' | 'cloudflare' | 'dnspod' | 'tencentcloud' | 'huaweicloud';

interface ProviderConfig {
  label: string;
  doc?: string;
  idField: { label: string; placeholder: string };
  secretField: { label: string; placeholder: string };
  extraParams?: { label: string; placeholder: string };
}

const providerConfigs: Record<ProviderKey, ProviderConfig> = {
  alidns: {
    label: '阿里云 DNS',
    doc: '使用 AccessKey 进行 HMAC-SHA1 签名认证。需要在阿里云 RAM 访问控制中创建子用户并授权 AlidnsFullAccess。',
    idField: { label: 'AccessKey ID', placeholder: 'RAM 用户 AccessKey ID' },
    secretField: { label: 'AccessKey Secret', placeholder: 'RAM 用户 AccessKey Secret' },
  },
  cloudflare: {
    label: 'Cloudflare',
    doc: '使用 API Token 认证。在 Cloudflare 仪表盘创建 API Token，权限需包含 DNS:Edit。',
    idField: { label: '(不使用)', placeholder: '' },
    secretField: { label: 'API Token', placeholder: 'Cloudflare API Token' },
    extraParams: { label: '扩展参数', placeholder: 'proxied=true&comment=备注（可选）' },
  },
  dnspod: {
    label: 'DNSPod',
    doc: '使用 API Token 认证。在 DNSPod 控制台创建 API 令牌，ID 和 Token 用逗号组合为 login_token。',
    idField: { label: 'Token ID', placeholder: 'DNSPod API Token ID' },
    secretField: { label: 'Token', placeholder: 'DNSPod API Token' },
  },
  tencentcloud: {
    label: '腾讯云 DNS',
    doc: '使用 SecretId + SecretKey 进行 TC3-HMAC-SHA256 签名认证。需授权 DNSPod 相关权限。',
    idField: { label: 'SecretId', placeholder: '腾讯云 API SecretId' },
    secretField: { label: 'SecretKey', placeholder: '腾讯云 API SecretKey' },
    extraParams: { label: '线路参数', placeholder: 'RecordLine=默认（可选）' },
  },
  huaweicloud: {
    label: '华为云 DNS',
    doc: '使用 AccessKey 进行 SDK-HMAC-SHA256 签名认证。需在 IAM 中创建 AK/SK 并授权 DNS 权限。',
    idField: { label: 'AccessKey ID', placeholder: '华为云 AK' },
    secretField: { label: 'AccessKey Secret', placeholder: '华为云 SK' },
    extraParams: { label: 'Zone/RecordSet', placeholder: 'zone_id=xxx&recordset_id=xxx（可选）' },
  },
};

const dnsProviders = Object.entries(providerConfigs).map(([value, cfg]) => ({
  value, label: cfg.label,
}));

const intervalOptions = [
  { value: 60, label: '1 分钟' },
  { value: 300, label: '5 分钟' },
  { value: 600, label: '10 分钟' },
  { value: 900, label: '15 分钟' },
  { value: 1800, label: '30 分钟' },
  { value: 3600, label: '60 分钟' },
];

// Compound value prefix for netInterface mode
const NETIF_PREFIX = 'netif:';

interface DDNSFormModalProps {
  open: boolean;
  editId: number | null;
  onClose: () => void;
  onSuccess: () => void;
}

export default function DDNSFormModal({ open, editId, onClose, onSuccess }: DDNSFormModalProps) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [netIfaces, setNetIfaces] = useState<{ ipv4: NetInterface[]; ipv6: NetInterface[] } | null>(null);
  const [fetchingIfaces, setFetchingIfaces] = useState(false);

  // Detected IPs for auto mode
  const [ipv4Detected, setIpv4Detected] = useState<string | null>(null);
  const [ipv6Detected, setIpv6Detected] = useState<string | null>(null);
  const [ipv4Detecting, setIpv4Detecting] = useState(false);
  const [ipv6Detecting, setIpv6Detecting] = useState(false);

  const isEdit = !!editId;
  const dnsProvider = Form.useWatch('dns_provider', form) as ProviderKey | undefined;
  const ipv4Combined = Form.useWatch('ipv4_combined', form) as string | undefined;
  const ipv6Combined = Form.useWatch('ipv6_combined', form) as string | undefined;

  const providerCfg = useMemo(() => dnsProvider ? providerConfigs[dnsProvider] : undefined, [dnsProvider]);

  // Parse combined value
  const parseCombined = (val: string | undefined) => {
    if (!val || val === 'disabled') return { mode: 'disabled' as const, iface: '' };
    if (val.startsWith(NETIF_PREFIX)) return { mode: 'netInterface' as const, iface: val.slice(NETIF_PREFIX.length) };
    return { mode: 'auto' as const, iface: '' };
  };

  const ipv4Parsed = parseCombined(ipv4Combined);
  const ipv6Parsed = parseCombined(ipv6Combined);

  // Build combined options
  const buildOptions = (family: 'ipv4' | 'ipv6') => {
    const isIpv6 = family === 'ipv6';
    const detectedIP = family === 'ipv4' ? ipv4Detected : ipv6Detected;
    const detecting = family === 'ipv4' ? ipv4Detecting : ipv6Detecting;
    const ifaceList = family === 'ipv4' ? (netIfaces?.ipv4 ?? []) : (netIfaces?.ipv6 ?? []);
    const parsed = family === 'ipv4' ? ipv4Parsed : ipv6Parsed;
    const selectedIface = parsed.mode === 'netInterface' ? parsed.iface : '';

    const options: { value: string; label: string }[] = [];

    // Auto option (IPv4 only — IPv6 不再支持自动检测)
    if (!isIpv6) {
      options.push({
        value: 'auto',
        label: detectedIP ? `自动 (${detectedIP})` : detecting ? '自动 (检测中...)' : '自动',
      });
    }

    // Interface options
    for (const face of ifaceList) {
      if (isIpv6 && face.address_detail && face.address_detail.length > 0) {
        // IPv6: 每个地址独立一行，标注类型
        for (const addr of face.address_detail) {
          const isPermanent = addr.type === 'permanent';
          options.push({
            value: `${NETIF_PREFIX}${face.name}|${addr.address}`,
            label: `${face.name} — ${addr.address} (${isPermanent ? 'secured' : 'temporary'})${isPermanent ? ' ★ 推荐' : ''}`,
          });
        }
      } else {
        // IPv4 或没有类型信息的 IPv6：沿用旧格式
        const ipLabel = face.address.join(', ');
        options.push({
          value: `${NETIF_PREFIX}${face.name}`,
          label: `${face.name} (${ipLabel})`,
        });
      }
    }

    // 如果已选网卡不在当前列表中，补一个占位选项
    if (selectedIface) {
      const matchIface = selectedIface.split('|')[0];
      if (matchIface && !ifaceList.find(f => f.name === matchIface)) {
        options.push({
          value: `${NETIF_PREFIX}${selectedIface}`,
          label: `${selectedIface}（?）`,
        });
      }
    }

    // Disabled option
    options.push({ value: 'disabled', label: '禁用' });

    return options;
  };

  // Auto-detect IP
  const detectIP = useCallback(async (family: 'ipv4' | 'ipv6') => {
    const setDetecting = family === 'ipv4' ? setIpv4Detecting : setIpv6Detecting;
    const setDetected = family === 'ipv4' ? setIpv4Detected : setIpv6Detected;
    setDetecting(true);
    setDetected(null);
    try {
      const res = await ddns.testIP({
        ipv4_enabled: family === 'ipv4',
        ipv4_get_type: 'auto',
        ipv4_url: '', ipv4_net_interface: '', ipv4_cmd: '', ipv4_addr: '',
        ipv6_enabled: family === 'ipv6',
        ipv6_get_type: 'auto',
        ipv6_url: '', ipv6_net_interface: '', ipv6_cmd: '', ipv6_addr: '',
      });
      const addr = family === 'ipv4' ? res.ipv4 : res.ipv6;
      setDetected(addr || null);
    } catch {
      setDetected(null);
    } finally {
      setDetecting(false);
    }
  }, []);

  // Fetch net interfaces (needed for dropdown options)
  const fetchNetIfaces = useCallback(async () => {
    setFetchingIfaces(true);
    try {
      const res = await ddns.netInterfaces();
      setNetIfaces(res);
    } catch {
      message.error('获取网卡列表失败');
    } finally {
      setFetchingIfaces(false);
    }
  }, []);

  // Load config for edit, or set defaults for new
  useEffect(() => {
    if (!open) return;
    form.resetFields();
    setIpv4Detected(null);
    setIpv6Detected(null);
    setNetIfaces(null);
    fetchNetIfaces();

    if (isEdit && editId) {
      setLoading(true);
      ddns.get(editId)
        .then((cfg: DDNSConfig) => {
          // Map to combined value
          const v4Combined = !cfg.ipv4_enabled ? 'disabled'
            : cfg.ipv4_get_type === 'netInterface' ? `${NETIF_PREFIX}${cfg.ipv4_net_interface || ''}`
            : 'auto';
          const v6Combined = !cfg.ipv6_enabled ? 'disabled'
            : cfg.ipv6_get_type === 'netInterface' ? `${NETIF_PREFIX}${cfg.ipv6_net_interface || ''}`
            : 'auto';

          form.setFieldsValue({
            name: cfg.name,
            dns_provider: cfg.dns_provider,
            access_key_id: cfg.access_key_id,
            access_key_secret: cfg.access_key_secret,
            extra_params: cfg.extra_params,
            domains: cfg.domains,
            interval: cfg.interval,
            ipv4_combined: v4Combined,
            ipv6_combined: v6Combined,
          });

          // Auto-detect if mode is auto
          if (cfg.ipv4_enabled && cfg.ipv4_get_type === 'auto') detectIP('ipv4');
          if (cfg.ipv6_enabled && cfg.ipv6_get_type === 'auto') detectIP('ipv6');
        })
        .catch(() => message.error('加载失败'))
        .finally(() => setLoading(false));
    } else {
      form.setFieldsValue({
        interval: 300,
        ipv4_combined: 'auto',
        ipv6_combined: 'disabled',
      });
      detectIP('ipv4');
    }
  }, [open, editId, isEdit, form, detectIP, fetchNetIfaces]);

  // Auto-detect when switching to auto
  useEffect(() => {
    if (ipv4Parsed.mode === 'auto' && open) detectIP('ipv4');
  }, [ipv4Parsed.mode, open, detectIP]);


  const handleFinish = async (values: Record<string, unknown>) => {
    setSaving(true);
    try {
      const v4Combined = values.ipv4_combined as string || 'auto';
      const v6Combined = values.ipv6_combined as string || 'disabled';
      const v4Parsed = parseCombined(v4Combined);
      const v6Parsed = parseCombined(v6Combined);

      const payload: CreateDDNSConfig = {
        name: values.name as string,
        dns_provider: values.dns_provider as string,
        access_key_id: values.access_key_id as string || '',
        access_key_secret: values.access_key_secret as string || '',
        extra_params: values.extra_params as string || '',
        domains: values.domains as string[] || [],
        interval: values.interval as number || 300,
        // IPv4
        ipv4_enabled: v4Parsed.mode !== 'disabled',
        ipv4_get_type: v4Parsed.mode !== 'disabled' ? v4Parsed.mode : 'auto',
        ipv4_net_interface: v4Parsed.iface || '',
        // IPv6
        ipv6_enabled: v6Parsed.mode !== 'disabled',
        ipv6_get_type: v6Parsed.mode !== 'disabled' ? v6Parsed.mode : 'auto',
        ipv6_net_interface: v6Parsed.iface || '',
      };

      if (isEdit && editId) {
        await ddns.update(editId, payload);
        message.success('更新成功');
      } else {
        await ddns.create(payload);
        message.success('创建成功');
      }
      onSuccess();
      onClose();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '保存失败');
    } finally {
      setSaving(false);
    }
  };

  const ipSection = (family: 'ipv4' | 'ipv6') => {
    const label = family === 'ipv4' ? 'IPv4' : 'IPv6';
    const options = buildOptions(family);
    const name = `${family}_combined`;

    return (
      <Form.Item name={name} label={label} style={{ marginBottom: 16 }}>
        <Select
          options={options}
          style={{ width: 420 }}
          loading={fetchingIfaces}
          showSearch
          filterOption={(input, option) =>
            (option?.label as string ?? '').toLowerCase().includes(input.toLowerCase())
          }
        />
      </Form.Item>
    );
  };

  return (
    <Modal
      title={isEdit ? '编辑 DDNS 配置' : '新增 DDNS 配置'}
      open={open}
      onCancel={onClose}
      footer={null}
      width={680}
      destroyOnClose
    >
      {loading ? (
        <div style={{ textAlign: 'center', padding: 40 }}><Spin /></div>
      ) : (
        <Form
          form={form}
          layout="vertical"
          onFinish={handleFinish}
          style={{ maxWidth: 560 }}
        >
          <Form.Item name="name" label="配置名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="例如：主域名 DDNS" />
          </Form.Item>

          <Form.Item name="dns_provider" label="DNS 服务商" rules={[{ required: true, message: '请选择 DNS 服务商' }]}>
            <Select options={dnsProviders} placeholder="选择 DNS 服务商" onChange={() => {
              form.setFieldsValue({ access_key_id: '', access_key_secret: '', extra_params: '' });
            }} />
          </Form.Item>

          {providerCfg?.doc && (
            <Alert type="info" message={providerCfg.doc} showIcon style={{ marginBottom: 16 }} />
          )}

          {dnsProvider && dnsProvider !== 'cloudflare' && (
            <Form.Item name="access_key_id" label={providerCfg?.idField.label} rules={[{ required: true, message: `请输入${providerCfg?.idField.label}` }]}>
              <Input placeholder={providerCfg?.idField.placeholder} />
            </Form.Item>
          )}

          {dnsProvider && (
            <Form.Item name="access_key_secret" label={providerCfg?.secretField.label} rules={[{ required: true, message: `请输入${providerCfg?.secretField.label}` }]}>
              <Input.Password placeholder={providerCfg?.secretField.placeholder} />
            </Form.Item>
          )}

          {providerCfg?.extraParams && (
            <Form.Item name="extra_params" label={providerCfg.extraParams.label}>
              <Input placeholder={providerCfg.extraParams.placeholder} />
            </Form.Item>
          )}

          <Form.Item name="domains" label="域名列表" rules={[{ required: true, message: '请至少添加一个域名' }]}>
            <Select mode="tags" placeholder="输入域名后回车添加" />
          </Form.Item>

          <Form.Item name="interval" label="更新间隔">
            <Select options={intervalOptions} style={{ width: 240 }} />
          </Form.Item>

          {ipSection('ipv4')}
          {ipSection('ipv6')}

          <Form.Item style={{ marginTop: 24, textAlign: 'right' }}>
            <Space>
              <Button onClick={onClose}>取消</Button>
              <Button type="primary" htmlType="submit" loading={saving}>
                {isEdit ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      )}
    </Modal>
  );
}
