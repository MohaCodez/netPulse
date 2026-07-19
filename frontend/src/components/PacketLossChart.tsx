import { useMemo } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import type { TimeseriesPoint } from '../hooks/useAnalytics';

interface Props {
  data: TimeseriesPoint[] | null;
}

export function PacketLossChart({ data }: Props) {
  const chartData = useMemo(() => {
    if (!data) return [];
    // Bucket into 30-second intervals
    const buckets = new Map<number, { time: number; gateway: number; external: number; gCount: number; eCount: number }>();

    for (const p of data) {
      const t = new Date(p.timestamp).getTime();
      const bucket = Math.floor(t / 30000) * 30000;
      if (!buckets.has(bucket)) {
        buckets.set(bucket, { time: bucket, gateway: 0, external: 0, gCount: 0, eCount: 0 });
      }
      const b = buckets.get(bucket)!;
      if (p.label === 'gateway') {
        b.gateway += p.value;
        b.gCount++;
      } else {
        b.external += p.value;
        b.eCount++;
      }
    }

    return Array.from(buckets.values())
      .map((b) => ({
        time: b.time,
        gateway: b.gCount > 0 ? (b.gateway / b.gCount) * 100 : 0,
        external: b.eCount > 0 ? (b.external / b.eCount) * 100 : 0,
      }))
      .sort((a, b) => a.time - b.time);
  }, [data]);

  if (chartData.length === 0) {
    return <div className="chart-container"><h3>📉 Packet Loss</h3><p className="chart-empty">No loss detected</p></div>;
  }

  const hasLoss = chartData.some((d) => d.gateway > 0 || d.external > 0);

  return (
    <div className="chart-container">
      <h3>📉 Packet Loss</h3>
      {!hasLoss ? (
        <p className="chart-empty" style={{ color: '#22c55e' }}>✓ Zero packet loss — connection is solid</p>
      ) : (
        <ResponsiveContainer width="100%" height={180}>
          <BarChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
            <XAxis dataKey="time" type="number" domain={['auto', 'auto']}
              tickFormatter={(val) => new Date(val).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
              stroke="var(--text-muted)" fontSize={11} />
            <YAxis stroke="var(--text-muted)" fontSize={11} tickFormatter={(val) => `${val}%`} />
            <Tooltip
              contentStyle={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: 8 }}
              labelFormatter={(val) => new Date(val as number).toLocaleTimeString()}
              formatter={(val: any) => [`${Number(val).toFixed(1)}%`]}
            />
            <Bar dataKey="gateway" fill="#f59e0b" name="Gateway" radius={[2, 2, 0, 0]} />
            <Bar dataKey="external" fill="#ef4444" name="External" radius={[2, 2, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}
