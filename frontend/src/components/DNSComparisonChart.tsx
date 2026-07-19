import { useMemo } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import type { TimeseriesPoint } from '../hooks/useAnalytics';

interface Props {
  data: TimeseriesPoint[] | null;
}

export function DNSComparisonChart({ data }: Props) {
  const chartData = useMemo(() => {
    if (!data) return [];
    // Group by 10-second buckets per domain
    const buckets = new Map<number, Record<string, { sum: number; count: number }>>();

    for (const p of data) {
      const t = new Date(p.timestamp).getTime();
      const bucket = Math.floor(t / 10000) * 10000;
      if (!buckets.has(bucket)) buckets.set(bucket, {});
      const b = buckets.get(bucket)!;
      const label = p.label.replace(' (multi-resolver)', '');
      if (!b[label]) b[label] = { sum: 0, count: 0 };
      b[label].sum += p.value;
      b[label].count++;
    }

    return Array.from(buckets.entries())
      .map(([time, domains]) => {
        const point: any = { time };
        for (const [domain, { sum, count }] of Object.entries(domains)) {
          point[domain] = sum / count;
        }
        return point;
      })
      .sort((a, b) => a.time - b.time);
  }, [data]);

  const domains = useMemo(() => {
    const keys = new Set<string>();
    for (const d of chartData) {
      for (const k of Object.keys(d)) {
        if (k !== 'time') keys.add(k);
      }
    }
    return Array.from(keys);
  }, [chartData]);

  const colors = ['#3b82f6', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6'];

  if (chartData.length === 0) {
    return <div className="chart-container"><h3>🌐 DNS Resolution Time</h3><p className="chart-empty">Collecting data...</p></div>;
  }

  return (
    <div className="chart-container">
      <h3>🌐 DNS Resolution by Domain</h3>
      <ResponsiveContainer width="100%" height={200}>
        <LineChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
          <XAxis dataKey="time" type="number" domain={['auto', 'auto']}
            tickFormatter={(val) => new Date(val).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            stroke="var(--text-muted)" fontSize={11} />
          <YAxis stroke="var(--text-muted)" fontSize={11} tickFormatter={(val) => `${val}ms`} />
          <Tooltip
            contentStyle={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: 8 }}
            labelFormatter={(val) => new Date(val as number).toLocaleTimeString()}
            formatter={(val: any) => [`${Number(val).toFixed(1)}ms`]}
          />
          <Legend />
          {domains.map((d, i) => (
            <Line key={d} type="monotone" dataKey={d} stroke={colors[i % colors.length]}
              strokeWidth={1.5} dot={false} connectNulls />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
