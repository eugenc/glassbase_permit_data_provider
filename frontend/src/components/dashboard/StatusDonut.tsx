import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from 'recharts';

const COLORS: Record<string, string> = {
  active: '#22C55E',
  broken: '#EF4444',
  paused: '#F59E0B',
};

interface Props {
  data: { status: string; count: number }[];
}

export default function StatusDonut({ data }: Props) {
  const total = data.reduce((sum, d) => sum + d.count, 0);

  return (
    <div className="rounded-[10px] border border-border bg-surface p-5 shadow-card">
      <h3 className="mb-4 text-sm font-medium text-text-2">Counties by Status</h3>
      <div className="flex items-center gap-6">
        <div className="relative">
          <ResponsiveContainer width={120} height={120}>
            <PieChart>
              <Pie
                data={data}
                dataKey="count"
                nameKey="status"
                innerRadius={40}
                outerRadius={56}
                paddingAngle={3}
                startAngle={90}
                endAngle={-270}
              >
                {data.map((entry) => (
                  <Cell key={entry.status} fill={COLORS[entry.status] ?? '#4D5E7A'} />
                ))}
              </Pie>
              <Tooltip
                contentStyle={{
                  background: '#0D1220',
                  border: '1px solid #1F2A3D',
                  borderRadius: 8,
                  fontSize: 12,
                }}
              />
            </PieChart>
          </ResponsiveContainer>
          <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
            <span className="font-mono text-xl font-semibold text-text-1">{total}</span>
            <span className="text-[10px] text-text-3">total</span>
          </div>
        </div>

        <div className="space-y-2">
          {data.map((d) => (
            <div key={d.status} className="flex items-center gap-2">
              <div className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: COLORS[d.status] }} />
              <span className="text-xs capitalize text-text-2">{d.status}</span>
              <span className="ml-1 font-mono text-xs text-text-1">{d.count}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
