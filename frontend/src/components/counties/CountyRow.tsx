import { useNavigate } from 'react-router-dom';
import { useState } from 'react';
import Badge from '../ui/Badge';
import Button from '../ui/Button';
import {
  useSetCountyStatus,
  useTriggerRun,
} from '../../hooks/useCounties';
import type { County } from '../../types';

export default function CountyRow({ county }: { county: County }) {
  const navigate = useNavigate();
  const setStatus = useSetCountyStatus();
  const triggerRun = useTriggerRun();
  const [runTriggered, setRunTriggered] = useState(false);

  const handleToggleStatus = () => {
    const next = county.status === 'active' ? 'paused' : 'active';
    setStatus.mutate({ id: county.county_id, status: next });
  };

  const handleRun = async () => {
    await triggerRun.mutateAsync(county.county_id);
    setRunTriggered(true);
    setTimeout(() => setRunTriggered(false), 5000);
  };

  const lastRun = county.last_run_at
    ? new Date(county.last_run_at).toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
      })
    : '—';

  const runDisabled = county.status !== 'active';

  return (
    <tr className="group border-b border-border transition hover:bg-[#151D30]">
      <td className="px-4 py-3.5">
        <div>
          <p className="text-[13px] font-medium text-text-1">{county.county_name}</p>
          <p className="font-mono text-xs text-text-3">{county.county_id}</p>
        </div>
      </td>
      <td className="px-4 py-3.5 text-sm text-text-2">{county.state}</td>
      <td className="px-4 py-3.5">
        <Badge status={county.status} />
      </td>
      <td className="px-4 py-3.5 font-mono text-sm text-text-1">
        {county.permit_count.toLocaleString()}
      </td>
      <td className="px-4 py-3.5">
        <div className="flex items-center gap-1.5">
          {county.last_run_status && (
            <Badge status={county.last_run_status as 'success' | 'failed' | 'running'} />
          )}
          <span className="text-xs text-text-3">{lastRun}</span>
        </div>
      </td>
      <td className="px-4 py-3.5 font-mono text-xs capitalize text-text-3">{county.source_type}</td>
      <td className="px-4 py-3.5">
        <div className="flex items-center gap-2 opacity-0 transition group-hover:opacity-100">
          <Button
            size="sm"
            variant="ghost"
            className="!px-2"
            onClick={() => navigate(`/counties/${county.county_id}/permits`)}
          >
            Permits
          </Button>
          <Button
            size="sm"
            variant="secondary"
            loading={triggerRun.isPending}
            disabled={runDisabled}
            onClick={() => void handleRun()}
          >
            {runTriggered ? 'Queued ✓' : 'Run'}
          </Button>
          <Button
            size="sm"
            variant={county.status === 'active' ? 'danger' : 'secondary'}
            loading={setStatus.isPending}
            onClick={handleToggleStatus}
          >
            {county.status === 'active' ? 'Pause' : 'Activate'}
          </Button>
        </div>
      </td>
    </tr>
  );
}
