import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { apiClient } from '../api/client';
import Badge from '../components/ui/Badge';
import Button from '../components/ui/Button';
import RepairDrawer from '../components/repairs/RepairDrawer';
import { useRepairCounty, useTriggerRun } from '../hooks/useCounties';

export default function CountyDetail() {
  const { id } = useParams<{ id: string }>();
  const countyId = id ?? '';

  const navigate = useNavigate();
  const { data: county, isLoading } = useQuery({
    queryKey: ['county', countyId],
    queryFn: async () => {
      const { data } = await apiClient.getCounty(countyId);
      return data;
    },
    enabled: !!countyId,
  });

  const triggerRun = useTriggerRun();
  const repair = useRepairCounty();
  const [showAiRepair, setShowAiRepair] = useState(false);

  if (isLoading) return <p className="text-text-3">Loading...</p>;
  if (!county) return <p className="text-status-broken">County not found.</p>;

  const lastRun = county.last_run_at
    ? new Date(county.last_run_at).toLocaleString()
    : 'Never';

  const runDisabled = county.status !== 'active';

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="mb-2 flex items-center gap-2">
            <Link to="/counties" className="text-xs text-text-3 hover:text-text-1 transition">
              Counties
            </Link>
            <span className="text-border">/</span>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <h2 className="text-xl font-semibold text-text-1">{county.county_name}</h2>
            <Badge status={county.status} />
          </div>
          <p className="mt-1 font-mono text-sm text-text-3">{county.county_id}</p>
          <p className="mt-2 text-sm text-text-2">{county.url}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => navigate(`/counties/${county.county_id}/permits`)}
          >
            View permits
          </Button>
          <Button
            size="sm"
            loading={triggerRun.isPending}
            disabled={runDisabled}
            onClick={() => triggerRun.mutate(county.county_id)}
          >
            Run scrape
          </Button>
          <Button
            size="sm"
            variant="secondary"
            loading={repair.isPending}
            onClick={() => repair.mutate(county.county_id)}
          >
            Regenerate connector
          </Button>
          {(county.status === 'broken' || county.status === 'paused') && (
            <Button size="sm" variant="secondary" onClick={() => setShowAiRepair(true)}>
              Claude Code repair
            </Button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div className="rounded-[10px] border border-border bg-surface p-4 shadow-card">
          <p className="text-[11px] font-semibold uppercase tracking-[0.06em] text-text-3">
            Permit records
          </p>
          <p className="mt-1 font-mono text-2xl text-text-1">{county.permit_count.toLocaleString()}</p>
        </div>
        <div className="rounded-[10px] border border-border bg-surface p-4 shadow-card">
          <p className="text-[11px] font-semibold uppercase tracking-[0.06em] text-text-3">
            Source type
          </p>
          <p className="mt-1 text-lg capitalize text-text-1">{county.source_type}</p>
        </div>
        <div className="rounded-[10px] border border-border bg-surface p-4 shadow-card">
          <p className="text-[11px] font-semibold uppercase tracking-[0.06em] text-text-3">
            Last run
          </p>
          <p className="mt-1 text-sm text-text-2">{lastRun}</p>
          {county.last_run_status && (
            <div className="mt-2">
              <Badge status={county.last_run_status as 'success' | 'failed' | 'running'} />
            </div>
          )}
        </div>
      </div>

      {showAiRepair ? (
        <RepairDrawer
          countyId={county.county_id}
          countyName={county.county_name}
          onClose={() => setShowAiRepair(false)}
        />
      ) : null}
    </div>
  );
}
