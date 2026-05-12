import { Link } from 'react-router-dom';
import { useDashboard } from '../hooks/useDashboard';
import PermitsBarChart from '../components/dashboard/PermitsBarChart';
import RecentErrors from '../components/dashboard/RecentErrors';
import StatCard from '../components/dashboard/StatCard';
import StatusDonut from '../components/dashboard/StatusDonut';

export default function Dashboard() {
  const { data, isLoading, error } = useDashboard();

  if (isLoading) return <DashboardSkeleton />;
  if (error || !data) return <p className="text-status-broken">Failed to load dashboard.</p>;

  const errorRateDisplay = `${data.error_rate.toFixed(1)}%`;
  const lastRun = data.last_run_at
    ? new Date(data.last_run_at).toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    : 'Never';

  return (
    <div className="space-y-6">
      <div>
        <p className="text-sm text-text-2">Last batch run: {lastRun}</p>
        <p className="mt-1 text-xs text-text-3">
          <Link to="/repairs" className="font-medium text-brand hover:underline">
            View AI repairs
          </Link>{' '}
          (Claude Code runs & history)
        </p>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        <StatCard label="Total Counties" value={data.total_counties} />
        <StatCard label="Active Counties" value={data.active_counties} accent />
        <StatCard
          label="Broken Counties"
          value={data.broken_counties}
          alert={data.broken_counties > 0}
        />
        <StatCard label="Permits This Week" value={data.permits_this_week.toLocaleString()} accent />
        <StatCard label="Total Permits" value={data.total_permits.toLocaleString()} />
        <StatCard
          label="Error Rate (7d)"
          value={errorRateDisplay}
          alert={data.error_rate > 10}
          sub="% of runs failed"
        />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <PermitsBarChart data={data.permits_by_day} />
        </div>
        <StatusDonut data={data.by_status} />
      </div>

      <RecentErrors errors={data.recent_errors} />
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <div className="h-6 w-48 rounded bg-border" />
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {[...Array(6)].map((_, i) => (
          <div key={i} className="h-28 rounded-[10px] border border-border bg-surface" />
        ))}
      </div>
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2 h-52 rounded-[10px] border border-border bg-surface" />
        <div className="h-52 rounded-[10px] border border-border bg-surface" />
      </div>
    </div>
  );
}
