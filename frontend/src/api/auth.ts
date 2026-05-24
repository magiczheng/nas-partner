import { api } from './client';

export interface AuthStatus {
  initialized: boolean;
}

export interface AuthResponse {
  token: string;
  username: string;
}

export const auth = {
  status: () => api.get<AuthStatus>('/auth/status'),
  init: (username: string, password: string) =>
    api.post<AuthResponse>('/auth/init', { username, password }),
  login: (username: string, password: string) =>
    api.post<AuthResponse>('/auth/login', { username, password }),
  me: () => api.get<{ username: string }>('/me'),
  changePassword: (oldPassword: string, newPassword: string) =>
    api.put<void>('/me/password', { old_password: oldPassword, new_password: newPassword }),
  refresh: () => api.put<AuthResponse>('/auth/refresh', undefined),
};
