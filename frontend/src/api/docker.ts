import { api } from './client';

export interface ContainerInfo {
  id: string;
  name: string;
  image: string;
  status: string;
  state: string;
  ports: string[];
  uptime: string;
}

export const docker = {
  listContainers: () => api.get<ContainerInfo[]>('/docker/containers'),
};
