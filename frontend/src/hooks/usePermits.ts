import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../api/client';
import type { Permit } from '../types';

export function normalizePermit(row: Record<string, unknown>): Permit {
  const raw = row.raw_data;
  let rawObj: Record<string, unknown> = {};
  if (raw && typeof raw === 'object' && !Array.isArray(raw)) {
    rawObj = raw as Record<string, unknown>;
  }
  return {
    ...row,
    id: Number(row.id ?? 0),
    permit_number: String(row.permit_number ?? ''),
    scraped_at: String(row.scraped_at ?? ''),
    raw_data: rawObj,
  } as Permit;
}

export function usePermits(
  countyId: string,
  params: { page: number; per_page: number; search: string }
) {
  return useQuery({
    queryKey: ['permits', countyId, params],
    queryFn: async () => {
      const q: Record<string, string | number> = {
        page: params.page,
        per_page: params.per_page,
      };
      if (params.search) q.search = params.search;
      const { data } = await apiClient.getPermits(countyId, q);
      return {
        ...data,
        permits: data.permits.map((p) => normalizePermit(p as unknown as Record<string, unknown>)),
      };
    },
    placeholderData: (prev) => prev,
    enabled: !!countyId,
  });
}
