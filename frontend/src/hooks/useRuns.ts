import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../api/client';
import type { ErrorRateRow, ScrapeRun } from '../types';

export function useRuns(params?: { county_id?: string; status?: string }) {
  return useQuery<{ runs: ScrapeRun[] }>({
    queryKey: ['runs', params],
    queryFn: async () => {
      const { data } = await apiClient.getRuns(params);
      return data;
    },
    refetchInterval: 30_000,
  });
}

export function useErrorRate() {
  return useQuery<ErrorRateRow[]>({
    queryKey: ['error-rate'],
    queryFn: async () => {
      const { data } = await apiClient.getErrorRate();
      return data;
    },
  });
}
