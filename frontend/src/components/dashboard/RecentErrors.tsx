interface ErrorRow {
  county_id: string;
  error_message: string;
  started_at: string;
}

export default function RecentErrors({ errors }: { errors: ErrorRow[] }) {
  if (!errors.length) {
    return (
      <div className="rounded-[10px] border border-border bg-surface p-5 shadow-card">
        <h3 className="mb-3 text-sm font-medium text-text-2">Recent Errors</h3>
        <p className="text-sm text-text-3">No errors in the last 7 days.</p>
      </div>
    );
  }

  return (
    <div className="rounded-[10px] border border-border bg-surface p-5 shadow-card">
      <h3 className="mb-4 text-sm font-medium text-text-2">Recent Errors</h3>
      <div className="space-y-2">
        {errors.map((e, i) => (
          <div
            key={`${e.county_id}-${e.started_at}-${i}`}
            className="flex items-start gap-3 border-b border-border py-2.5 last:border-0"
          >
            <div className="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-status-broken" />
            <div className="min-w-0">
              <div className="mb-0.5 flex items-center gap-2">
                <span className="font-mono text-xs text-brand">{e.county_id}</span>
                <span className="text-xs text-text-3">
                  {new Date(e.started_at).toLocaleDateString()}
                </span>
              </div>
              <p className="truncate text-xs text-text-2">{e.error_message}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
