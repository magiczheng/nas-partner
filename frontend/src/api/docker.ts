import { api } from './client';

export interface ContainerInfo {
  id: string;
  name: string;
  status: string;
  state: string;
  ports: string[];
  cpu_percent: number;
  memory_usage: number;
  memory_limit: number;
  network_rx: number;
  network_tx: number;
}

export const docker = {
  listContainers: () => api.get<ContainerInfo[]>('/docker/containers'),
};
