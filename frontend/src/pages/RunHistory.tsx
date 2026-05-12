import { useState } from 'react';
import Badge from '../components/ui/Badge';
import ErrorRateChart from '../components/runs/ErrorRateChart';
import { useErrorRate, useRuns } from '../hooks/useRuns';

export default function RunHistory() {
  const [countyFilter, setCountyFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('All');

  const { data, isLoading } = useRuns({
    county_id: countyFilter || undefined,
    status: statusFilter === 'All' ? undefined : statusFilter,
  });
  const { data: errorRate } = useErrorRate();

  const runs = data?.runs ?? [];

  const duration = (run: { started_at: string; finished_at: string | null }) => {
    if (!run.finished_at) return 'running';
    const ms = new Date(run.finished_at).getTime() - new Date(run.started_at).getTime();
    return ms < 60000 ? `${Math.round(ms / 1000)}s` : `${Math.round(ms / 60000)}m`;
  };

  return (
    <div className="space-y-5">
      <div>
        <p className="mt-0.5 text-sm text-text-3">All scrape runs across every county</p>
      </div>

      {errorRate && errorRate.length > 0 && <ErrorRateChart data={errorRate} />}

      <div className="flex flex-wrap items-center gap-3">
        <input
          value={countyFilter}
          onChange={(e) => setCountyFilter(e.target.value)}
          placeholder="Filter by county ID..."
          className="w-56 rounded-md border border-border bg-surface px-3.5 py-2 font-mono text-sm text-text-1 placeholder:text-text-3 focus:border-brand focus:outline-none"
        />
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-md border border-border bg-surface px-3 py-2 text-sm text-text-1 focus:border-brand focus:outline-none"
        >
          {['All', 'success', 'failed', 'running'].map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
        <span className="ml-auto text-xs text-text-3">{runs.length} runs shown</span>
      </div>

      <div className="overflow-hidden rounded-[10px] border border-border bg-surface shadow-card">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border bg-[#0D1220]">
              {['County', 'Status', 'Found', 'Inserted', 'Duration', 'Started', 'Error'].map((h) => (
                <th
                  key={h}
                  className="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.06em] text-text-3"
                >
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {isLoading
              ? [...Array(10)].map((_, i) => (
                  <tr key={i} className="animate-pulse border-b border-border">
                    {[...Array(7)].map((_, j) => (
                      <td key={j} className="px-4 py-3.5">
                        <div className="h-4 rounded bg-border" />
                      </td>
                    ))}
                  </tr>
                ))
              : runs.map((run) => (
                  <tr key={run.id} className="border-b border-border transition hover:bg-[#151D30]">
                    <td className="px-4 py-3 font-mono text-xs text-brand">{run.county_id}</td>
                    <td className="px-4 py-3">
                      <Badge status={run.status} />
                    </td>
                    <td className="px-4 py-3 font-mono text-sm text-text-2">
                      {run.records_found.toLocaleString()}
                    </td>
                    <td className="px-4 py-3 font-mono text-sm text-text-2">
                      {run.records_inserted.toLocaleString()}
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-text-3">{duration(run)}</td>
                    <td className="px-4 py-3 text-xs text-text-3">
                      {new Date(run.started_at).toLocaleDateString('en-US', {
                        month: 'short',
                        day: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </td>
                    <td className="max-w-[200px] truncate px-4 py-3 text-xs text-status-broken">
                      {run.error_message ?? '—'}
                    </td>
                  </tr>
                ))}
          </tbody>
        </table>

        {!isLoading && runs.length === 0 && (
          <div className="py-12 text-center text-sm text-text-3">No runs found.</div>
        )}
      </div>
    </div>
  );
}
