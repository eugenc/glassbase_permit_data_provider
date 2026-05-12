import {
  CartesianGrid,
  Line,
  LineChart,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import type { ErrorRateRow } from '../../types';

interface Props {
  data: ErrorRateRow[];
}

export default function ErrorRateChart({ data }: Props) {
  const formatted = data.map((d) => ({
    ...d,
    label: new Date(d.day).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
  }));

  return (
    <div className="rounded-[10px] border border-border bg-surface p-5 shadow-card">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-sm font-medium text-text-2">Error Rate — Last 30 Days</h3>
        <span className="text-xs text-text-3">% of runs failed</span>
      </div>
      <ResponsiveContainer width="100%" height={160}>
        <LineChart data={formatted}>
          <CartesianGrid strokeDasharray="3 3" stroke="#1F2A3D" vertical={false} />
          <XAxis
            dataKey="label"
            tick={{ fontSize: 10, fill: '#4D5E7A', fontFamily: 'JetBrains Mono, monospace' }}
            axisLine={false}
            tickLine={false}
            interval={4}
          />
          <YAxis
            tick={{ fontSize: 10, fill: '#4D5E7A', fontFamily: 'JetBrains Mono, monospace' }}
            axisLine={false}
            tickLine={false}
            width={36}
            domain={[0, 100]}
            tickFormatter={(v) => `${v}%`}
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
          <ReferenceLine y={10} stroke="#F59E0B" strokeDasharray="4 4" />
          <Line
            type="monotone"
            dataKey="error_pct"
            stroke="#EF4444"
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4, fill: '#EF4444' }}
          />
        </LineChart>
      </ResponsiveContainer>
      <p className="mt-2 text-xs text-text-3">Amber line = 10% threshold</p>
    </div>
  );
}
