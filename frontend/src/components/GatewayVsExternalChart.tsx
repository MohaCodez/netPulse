import { useMemo } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import type { ProbeLatencyPoint } from '../hooks/useAnalytics';

interface Props {
  data: ProbeLatencyPoint[] | null;
}

export function GatewayVsExternalChart({ data }: Props) {
  const chartData = useMemo(() => {
    if (!data) return [];
    return data.map((p) => ({
      time: new Date(p.timestamp).getTime(),
      gateway: p.gateway_ms || null,
      external: p.external_ms || null,
      dns: p.dns_ms || null,
    }));
  }, [data]);

  if (chartData.length === 0) {
    return <div className="chart-container"><h3>🔀 Gateway vs External Latency</h3><p className="chart-empty">Collecting data...</p></div>;
  }

  return (
    <div className="chart-container">
      <h3>🔀 Gateway vs External Latency</h3>
      <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', margin: '0 0 12px' }}>
        Gap between lines = ISP/internet overhead. If gateway spikes alone = local Wi-Fi issue.
      </p>
      <ResponsiveContainer width="100%" height={220}>
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
          <Line type="monotone" dataKey="gateway" stroke="#22c55e" strokeWidth={2} dot={false} connectNulls name="Gateway (local)" />
          <Line type="monotone" dataKey="external" stroke="#3b82f6" strokeWidth={2} dot={false} connectNulls name="External (internet)" />
          <Line type="monotone" dataKey="dns" stroke="#f59e0b" strokeWidth={1.5} dot={false} connectNulls name="DNS" />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
