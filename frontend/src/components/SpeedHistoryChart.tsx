import { useMemo } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { main } from '../../wailsjs/go/models';

interface Props {
  tests: main.SpeedTestResponse[];
}

export function SpeedHistoryChart({ tests }: Props) {
  const chartData = useMemo(() => {
    if (!tests || tests.length === 0) return [];
    return [...tests]
      .reverse()
      .map((t) => ({
        time: new Date(t.timestamp).getTime(),
        download: t.download_mbps,
        upload: t.upload_mbps,
        latency: t.latency_ms,
      }));
  }, [tests]);

  if (chartData.length === 0) {
    return <div className="chart-container"><h3>🚀 Speed Test History</h3><p className="chart-empty">No speed tests yet</p></div>;
  }

  return (
    <div className="chart-container">
      <h3>🚀 Speed Test History</h3>
      <ResponsiveContainer width="100%" height={220}>
        <LineChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
          <XAxis dataKey="time" type="number" domain={['auto', 'auto']}
            tickFormatter={(val) => new Date(val).toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
            stroke="var(--text-muted)" fontSize={10} />
          <YAxis yAxisId="speed" stroke="var(--text-muted)" fontSize={11}
            tickFormatter={(val) => `${val} Mbps`} />
          <YAxis yAxisId="latency" orientation="right" stroke="var(--text-muted)" fontSize={11}
            tickFormatter={(val) => `${val}ms`} />
          <Tooltip
            contentStyle={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: 8 }}
            labelFormatter={(val) => new Date(val as number).toLocaleString()}
            formatter={(val: any, name: any) => [
              name === 'Latency' ? `${Number(val).toFixed(0)} ms` : `${Number(val).toFixed(1)} Mbps`
            ]}
          />
          <Legend />
          <Line yAxisId="speed" type="monotone" dataKey="download" stroke="#22c55e" strokeWidth={2.5} dot={{ r: 4 }} name="Download" />
          <Line yAxisId="speed" type="monotone" dataKey="upload" stroke="#3b82f6" strokeWidth={2} dot={{ r: 3 }} name="Upload" />
          <Line yAxisId="latency" type="monotone" dataKey="latency" stroke="#f59e0b" strokeWidth={1.5} dot={{ r: 3 }} strokeDasharray="5 5" name="Latency" />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
