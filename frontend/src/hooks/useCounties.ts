import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../api/client';
import { useToast } from '../components/ui/ToastProvider';
import type { County } from '../types';

export function useCounties() {
  return useQuery<{ counties: County[]; total: number }>({
    queryKey: ['counties'],
    queryFn: async () => {
      const { data } = await apiClient.getCounties();
      return data;
    },
    refetchInterval: 30_000,
  });
}

export function useAddCounty() {
  const qc = useQueryClient();
  const toast = useToast();
  return useMutation({
    mutationFn: (payload: {
      county_id: string;
      county_name: string;
      state: string;
      url: string;
    }) => apiClient.addCounty(payload),
    onSuccess: () => {
      toast('success', 'County onboarded');
      void qc.invalidateQueries({ queryKey: ['counties'] });
      void qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
    onError: () => toast('error', 'Failed to add county'),
  });
}

export function useSetCountyStatus() {
  const qc = useQueryClient();
  const toast = useToast();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      apiClient.setCountyStatus(id, status),
    onSuccess: () => {
      toast('success', 'Status updated');
      void qc.invalidateQueries({ queryKey: ['counties'] });
      void qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
    onError: () => toast('error', 'Failed to update status'),
  });
}

export function useTriggerRun() {
  const qc = useQueryClient();
  const toast = useToast();
  return useMutation({
    mutationFn: (id: string) => apiClient.triggerRun(id),
    onSuccess: () => {
      toast('success', 'Run queued');
      void qc.invalidateQueries({ queryKey: ['counties'] });
      void qc.invalidateQueries({ queryKey: ['runs'] });
      void qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
    onError: () => toast('error', 'Failed to trigger run'),
  });
}

export function useRepairCounty() {
  const qc = useQueryClient();
  const toast = useToast();
  return useMutation({
    mutationFn: (id: string) => apiClient.repairCounty(id),
    onSuccess: () => {
      toast('success', 'Repair completed');
      void qc.invalidateQueries({ queryKey: ['counties'] });
    },
    onError: () => toast('error', 'Repair failed'),
  });
}
