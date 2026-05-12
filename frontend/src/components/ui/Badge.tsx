type Status =
  | 'active'
  | 'broken'
  | 'paused'
  | 'running'
  | 'success'
  | 'failed'
  | 'skipped';

const styles: Record<Status, string> = {
  active: 'text-status-active bg-status-active-bg border-status-active/30',
  broken: 'text-status-broken bg-status-broken-bg border-status-broken/30',
  paused: 'text-status-paused bg-status-paused-bg border-status-paused/30',
  running: 'text-status-running bg-status-running-bg border-status-running/30',
  success: 'text-status-active bg-status-active-bg border-status-active/30',
  failed: 'text-status-broken bg-status-broken-bg border-status-broken/30',
  skipped:
    'text-text-3 bg-elevated border-border',
};

export default function Badge({ status }: { status: string }) {
  const clazz = styles[status as Status] ?? 'border-border bg-elevated text-text-3';
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-medium capitalize ${clazz}`}
    >
      <span className="h-1.5 w-1.5 rounded-full bg-current opacity-80" />
      {status}
    </span>
  );
}
