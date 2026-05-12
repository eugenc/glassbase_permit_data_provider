import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export interface RepairRun {
  id: number;
  county_id: string;
  trigger: string;
  status: string;
  commit_sha: string | null;
  pr_url: string | null;
  started_at: string;
  finished_at: string | null;
  error_message: string | null;
  output_preview: string;
}

export function useRepairs() {
  return useQuery<{ repairs: RepairRun[] }>({
    queryKey: ['repairs'],
    queryFn: async () => {
      const { data } = await api.get('/api/repairs/recent');
      return data;
    },
    refetchInterval: 15_000,
  });
}
