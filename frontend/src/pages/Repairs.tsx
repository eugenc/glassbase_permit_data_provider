import Badge from '../components/ui/Badge';
import { useRepairs, type RepairRun } from '../hooks/useRepairs';

const triggerLabel: Record<string, string> = {
  zero_records: 'Zero records',
  health_probe: 'Health probe',
  manual: 'Manual',
};

export default function Repairs() {
  const { data, isLoading } = useRepairs();
  const repairs = data?.repairs ?? [];

  const duration = (r: RepairRun) => {
    if (!r.finished_at) return 'running';
    const ms = new Date(r.finished_at).getTime() - new Date(r.started_at).getTime();
    const m = Math.round(ms / 60000);
    return m <= 0 ? '<1m' : `${m}m`;
  };

  return (
    <div className="space-y-5 px-8 py-10">
      <div>
        <h1 className="text-xl font-semibold text-text-1">AI repairs</h1>
        <p className="mt-0.5 text-sm text-text-3">
          Claude Code repair history (<code className="rounded bg-surface px-1">repair-cc</code> and
          scheduler)
        </p>
      </div>

      <div className="overflow-hidden rounded-[10px] border border-border bg-surface shadow-card">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border">
              {['County', 'Trigger', 'Status', 'Duration', 'Commit / PR', 'Started'].map((h) => (
                <th
                  key={h}
                  className="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-wide text-text-3"
                >
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {isLoading
              ? [...Array(5)].map((_, i) => (
                  <tr key={i} className="border-b border-border animate-pulse">
                    {[...Array(6)].map((_, j) => (
                      <td key={j} className="px-4 py-3.5">
                        <div className="h-4 w-3/4 rounded bg-border" />
                      </td>
                    ))}
                  </tr>
                ))
              : repairs.map((r) => (
                  <tr
                    key={r.id}
                    className="border-b border-border transition hover:bg-elevated/50"
                  >
                    <td className="px-4 py-3 font-mono text-xs text-brand">{r.county_id}</td>
                    <td className="px-4 py-3 text-xs text-text-3">
                      {triggerLabel[r.trigger] ?? r.trigger}
                    </td>
                    <td className="px-4 py-3">
                      <Badge status={r.status} />
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-text-3">{duration(r)}</td>
                    <td className="px-4 py-3 text-xs">
                      {r.commit_sha && (
                        <span className="font-mono text-text-2">{r.commit_sha}</span>
                      )}
                      {r.pr_url && (
                        <a
                          href={r.pr_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="ml-2 font-medium text-brand hover:underline"
                        >
                          PR
                        </a>
                      )}
                      {!r.commit_sha && !r.pr_url && (
                        <span className="text-text-3">—</span>
                      )}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-xs text-text-3">
                      {new Date(r.started_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
          </tbody>
        </table>

        {!isLoading && repairs.length === 0 ? (
          <div className="py-14 text-center text-sm text-text-3">
            No repairs recorded yet.
          </div>
        ) : null}
      </div>
    </div>
  );
}
