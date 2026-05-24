import { api } from './client';

export interface SystemInfo {
  hostname: string;
  uptime: number;
  os: string;
  cpu: {
    percent: number;
    cores: number;
  };
  memory: {
    total: number;
    used: number;
    available: number;
    used_percent: number;
  };
  disk: {
    path: string;
    total: number;
    used: number;
    free: number;
    used_percent: number;
  };
  load1: number;
  load5: number;
  load15: number;
}

export const system = {
  getInfo: () => api.get<SystemInfo>('/system/info'),
};
