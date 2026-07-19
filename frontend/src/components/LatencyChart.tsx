import { useMemo } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { main } from '../../wailsjs/go/models';
import './LatencyChart.css';

interface Props {
  probes: main.ProbeResultResponse[];
}

export function LatencyChart({ probes }: Props) {
  const chartData = useMemo(() => {
    const grouped = new Map<string, Record<string, number>>();

    for (const p of probes) {
      if (!p.success || p.latency_ms === 0) continue;

      const time = new Date(p.timestamp);
      time.setSeconds(Math.round(time.getSeconds() / 5) * 5);
      time.setMilliseconds(0);
      const key = time.toISOString();

      if (!grouped.has(key)) {
        grouped.set(key, { time: time.getTime() });
      }
      const entry = grouped.get(key)!;

      const label = p.probe_type === 'gateway' ? 'Gateway' :
                    p.probe_type === 'dns' ? 'DNS' :
                    p.probe_type === 'tcp' ? 'TCP' : `Ping (${p.target})`;

      if (entry[label]) {
        entry[label] = (entry[label] + p.latency_ms) / 2;
      } else {
        entry[label] = p.latency_ms;
      }
    }

    return Array.from(grouped.values())
      .sort((a, b) => (a.time as number) - (b.time as number))
      .slice(-60);
  }, [probes]);

  const series = useMemo(() => {
    const keys = new Set<string>();
    for (const d of chartData) {
      for (const k of Object.keys(d)) {
        if (k !== 'time') keys.add(k);
      }
    }
    return Array.from(keys);
  }, [chartData]);

  const colors = ['#3b82f6', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4'];

  if (chartData.length === 0) {
    return (
      <div className="chart-container">
        <h3>Latency Timeline</h3>
        <p className="chart-empty">Collecting data...</p>
      </div>
    );
  }

  return (
    <div className="chart-container">
      <h3>Latency Timeline</h3>
      <ResponsiveContainer width="100%" height={250}>
        <LineChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
          <XAxis
            dataKey="time"
            type="number"
            domain={['auto', 'auto']}
            tickFormatter={(val) => new Date(val).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
            stroke="var(--text-muted)"
            fontSize={11}
          />
          <YAxis
            stroke="var(--text-muted)"
            fontSize={11}
            tickFormatter={(val) => `${val}ms`}
          />
          <Tooltip
            contentStyle={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: 8 }}
            labelFormatter={(val) => new Date(val as number).toLocaleTimeString()}
            formatter={(val) => [`${Number(val).toFixed(1)}ms`]}
          />
          <Legend />
          {series.map((s, i) => (
            <Line
              key={s}
              type="monotone"
              dataKey={s}
              stroke={colors[i % colors.length]}
              strokeWidth={2}
              dot={false}
              connectNulls
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
