import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

interface Props {
  data: { day: string; count: number }[];
}

export default function PermitsBarChart({ data }: Props) {
  const formatted = data.map((d) => ({
    ...d,
    label: new Date(d.day).toLocaleDateString('en-US', { weekday: 'short', day: 'numeric' }),
    count: Number(d.count),
  }));

  return (
    <div className="rounded-[10px] border border-border bg-surface p-5 shadow-card">
      <h3 className="mb-4 text-sm font-medium text-text-2">Permits Added — Last 7 Days</h3>
      <ResponsiveContainer width="100%" height={180}>
        <BarChart data={formatted} barSize={28}>
          <CartesianGrid strokeDasharray="3 3" stroke="#1F2A3D" vertical={false} />
          <XAxis
            dataKey="label"
            tick={{ fontSize: 11, fill: '#4D5E7A', fontFamily: 'JetBrains Mono, monospace' }}
            axisLine={false}
            tickLine={false}
          />
          <YAxis
            tick={{ fontSize: 11, fill: '#4D5E7A', fontFamily: 'JetBrains Mono, monospace' }}
            axisLine={false}
            tickLine={false}
            width={36}
          />
          <Tooltip
            contentStyle={{
              background: '#0D1220',
              border: '1px solid #1F2A3D',
              borderRadius: 8,
              fontSize: 12,
              color: '#F0F4FF',
            }}
          />
          <Bar dataKey="count" fill="#0E76E4" radius={[4, 4, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
