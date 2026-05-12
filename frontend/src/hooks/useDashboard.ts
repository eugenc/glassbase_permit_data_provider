import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../api/client';
import type { DashboardStats } from '../types';

export function useDashboard() {
  return useQuery<DashboardStats>({
    queryKey: ['dashboard'],
    queryFn: async () => {
      const { data } = await apiClient.getDashboard();
      return data;
    },
    refetchInterval: 60_000,
    staleTime: 30_000,
  });
}
