import { api } from './client';

export interface AuditEntry {
  id: number;
  username: string;
  action: string;
  detail: string;
  ip: string;
  created_at: string;
}

export const audit = {
  listLogs: (limit?: number) => api.get<AuditEntry[]>(`/audit/logs?limit=${limit || 100}`),
};
