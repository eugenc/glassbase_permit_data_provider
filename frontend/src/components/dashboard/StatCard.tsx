interface StatCardProps {
  label: string;
  value: string | number;
  sub?: string;
  accent?: boolean;
  alert?: boolean;
}

export default function StatCard({ label, value, sub, accent, alert }: StatCardProps) {
  return (
    <div
      className={`flex flex-col gap-1 rounded-[10px] border px-6 py-5 bg-surface ${
        accent ? 'border-brand' : alert ? 'border-status-broken' : 'border-border'
      }`}
    >
      <span className="text-[11px] font-semibold uppercase tracking-[0.06em] text-text-2">{label}</span>
      <span
        className={`font-mono text-[28px] font-medium ${
          accent ? 'text-brand' : alert ? 'text-status-broken' : 'text-text-1'
        }`}
      >
        {value}
      </span>
      {sub && <span className="text-xs text-text-2">{sub}</span>}
    </div>
  );
}
