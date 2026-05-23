import { api } from './client';

export interface DDNSConfig {
  id: number;
  name: string;
  enabled: boolean;
  dns_provider: string;
  access_key_id: string;
  access_key_secret: string;
  extra_params: string;
  ipv4_enabled: boolean;
  ipv4_get_type: string;
  ipv4_url: string;
  ipv4_net_interface: string;
  ipv4_cmd: string;
  ipv6_enabled: boolean;
  ipv6_get_type: string;
  ipv6_url: string;
  ipv6_net_interface: string;
  ipv6_cmd: string;
  ipv4_addr: string;
  ipv6_addr: string;
  current_ipv4: string;
  current_ipv6: string;
  domains: string[];
  ttl: string;
  interval: number;
  created_at: string;
  updated_at: string;
}

export interface DDNSConfigWithLog extends DDNSConfig {
  latest_log: DDNSRunLog | null;
}

export interface DDNSRunLog {
  id: number;
  config_id: number;
  status: string;
  message: string;
  ipv4_addr: string;
  ipv6_addr: string;
  created_at: string;
}

export interface CreateDDNSConfig {
  name: string;
  dns_provider: string;
  access_key_id?: string;
  access_key_secret?: string;
  extra_params?: string;
  ipv4_enabled?: boolean;
  ipv4_get_type?: string;
  ipv4_url?: string;
  ipv4_net_interface?: string;
  ipv4_cmd?: string;
  ipv6_enabled?: boolean;
  ipv6_get_type?: string;
  ipv6_url?: string;
  ipv6_net_interface?: string;
  ipv6_cmd?: string;
  ipv4_addr?: string;
  ipv6_addr?: string;
  domains?: string[];
  ttl?: string;
  interval?: number;
}

export interface AddressInfo {
  address: string;
  type: 'permanent' | 'temporary';
}

export interface NetInterface {
  name: string;
  address: string[];
  address_detail?: AddressInfo[];
}

export interface TestIPRequest {
  ipv4_enabled: boolean;
  ipv4_get_type: string;
  ipv4_url: string;
  ipv4_net_interface: string;
  ipv4_cmd: string;
  ipv6_enabled: boolean;
  ipv6_get_type: string;
  ipv6_url: string;
  ipv6_net_interface: string;
  ipv6_cmd: string;
  ipv4_addr: string;
  ipv6_addr: string;
}

export const ddns = {
  list: () => api.get<DDNSConfig[]>('/ddns'),
  listWithLogs: () => api.get<DDNSConfigWithLog[]>('/ddns/logs/latest'),
  get: (id: number) => api.get<DDNSConfig>(`/ddns/${id}`),
  create: (data: CreateDDNSConfig) => api.post<DDNSConfig>('/ddns', data),
  update: (id: number, data: Partial<DDNSConfig>) => api.put<DDNSConfig>(`/ddns/${id}`, data),
  delete: (id: number) => api.delete<void>(`/ddns/${id}`),
  toggle: (id: number) => api.post<DDNSConfig>(`/ddns/${id}/toggle`),
  run: (id: number) => api.post<DDNSRunLog>(`/ddns/${id}/run`),
  testIP: (data: TestIPRequest) => api.post<{ ipv4: string; ipv6: string }>('/ddns/test-ip', data),
  netInterfaces: () => api.get<{ ipv4: NetInterface[]; ipv6: NetInterface[] }>('/ddns/net-interfaces'),
  listLogs: (configId: number, limit?: number) => api.get<DDNSRunLog[]>(`/ddns/${configId}/logs`, { params: { limit: limit || 20 } }),
  clearLogs: (configId: number) => api.delete<void>(`/ddns/${configId}/logs`),
  cleanupLogs: () => api.post<{ deleted: number }>('/ddns/logs/cleanup'),
};
