import axios from 'axios';
import type { County, DashboardStats, ErrorRateRow, PermitsResponse, ScrapeRun } from '../types';

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? '',
  headers: { 'Content-Type': 'application/json' },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('gb_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('gb_token');
      localStorage.removeItem('gb_user');
      const path = window.location.pathname;
      if (!path.startsWith('/login')) window.location.href = '/login';
    }
    return Promise.reject(err);
  }
);

export const apiClient = {
  login: (email: string, password: string) =>
    api.post<{ token: string; expires_at: number; name: string; role: string }>('/auth/login', {
      email,
      password,
    }),

  getDashboard: () => api.get<DashboardStats>('/api/dashboard'),

  getCounties: () => api.get<{ counties: County[]; total: number }>('/api/counties'),
  getCounty: (id: string) => api.get<County>(`/api/counties/${id}`),
  addCounty: (data: {
    county_id: string;
    county_name: string;
    state: string;
    url: string;
  }) => api.post(`/api/counties`, data),

  setCountyStatus: (id: string, status: string) =>
    api.patch(`/api/counties/${id}/status`, { status }),

  deleteCounty: (id: string) => api.delete(`/api/counties/${id}`),

  triggerRun: (id: string) => api.post(`/api/counties/${id}/run`),
  repairCounty: (id: string) => api.post(`/api/counties/${id}/repair`),

  getPermits: (countyId: string, params: Record<string, string | number>) =>
    api.get<PermitsResponse>(`/api/counties/${countyId}/permits`, { params }),

  getRuns: (params?: Record<string, string>) =>
    api.get<{ runs: ScrapeRun[] }>('/api/runs', { params }),

  getErrorRate: () => api.get<ErrorRateRow[]>('/api/runs/error-rate'),
};

export async function downloadPermitsCsv(countyId: string): Promise<void> {
  const token = localStorage.getItem('gb_token');
  const base = import.meta.env.VITE_API_URL ?? '';
  const res = await fetch(`${base}/api/counties/${countyId}/permits/export`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (!res.ok) throw new Error('Export failed');
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${countyId}_permits.csv`;
  a.click();
  URL.revokeObjectURL(url);
}
