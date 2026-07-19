import { useMemo } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine } from 'recharts';
import type { TimeseriesPoint } from '../hooks/useAnalytics';

interface Props {
  data: TimeseriesPoint[] | null;
}

export function JitterChart({ data }: Props) {
  const chartData = useMemo(() => {
    if (!data) return [];
    return data.map((p) => ({
      time: new Date(p.timestamp).getTime(),
      jitter: p.value,
      type: p.label,
    }));
  }, [data]);

  if (chartData.length === 0) {
    return <div className="chart-container"><h3>〰️ Jitter (Latency Variance)</h3><p className="chart-empty">Collecting data...</p></div>;
  }

  return (
    <div className="chart-container">
      <h3>〰️ Jitter (Latency Variance)</h3>
      <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', margin: '0 0 12px' }}>
        High jitter = unstable for video calls &amp; gaming. &lt;30ms is ideal.
      </p>
      <ResponsiveContainer width="100%" height={180}>
        <LineChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
          <ReferenceLine y={30} stroke="#f59e0b" strokeDasharray="3 3" />
          <ReferenceLine y={100} stroke="#ef4444" strokeDasharray="3 3" />
          <XAxis dataKey="time" type="number" domain={['auto', 'auto']}
            tickFormatter={(val) => new Date(val).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            stroke="var(--text-muted)" fontSize={11} />
          <YAxis stroke="var(--text-muted)" fontSize={11} tickFormatter={(val) => `${val}ms`} />
          <Tooltip
            contentStyle={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: 8 }}
            labelFormatter={(val) => new Date(val as number).toLocaleTimeString()}
            formatter={(val: any) => [`${Number(val).toFixed(1)}ms`]}
          />
          <Line type="monotone" dataKey="jitter" stroke="#06b6d4" strokeWidth={2} dot={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
