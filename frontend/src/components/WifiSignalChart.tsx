import { useMemo } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine, ReferenceArea } from 'recharts';
import type { WifiTimeseriesPoint } from '../hooks/useAnalytics';

interface Props {
  data: WifiTimeseriesPoint[] | null;
}

export function WifiSignalChart({ data }: Props) {
  const chartData = useMemo(() => {
    if (!data) return [];
    return data.map((p) => ({
      time: new Date(p.timestamp).getTime(),
      signal: p.signal_dbm,
      band: p.band,
      channel: p.channel,
      speed: p.link_speed_mbps,
    }));
  }, [data]);

  if (chartData.length === 0) {
    return <div className="chart-container"><h3>📶 Wi-Fi Signal Strength</h3><p className="chart-empty">Collecting data...</p></div>;
  }

  return (
    <div className="chart-container">
      <h3>📶 Wi-Fi Signal Strength</h3>
      <ResponsiveContainer width="100%" height={200}>
        <LineChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
          <ReferenceArea y1={-50} y2={0} fill="#22c55e" fillOpacity={0.05} />
          <ReferenceArea y1={-70} y2={-50} fill="#f59e0b" fillOpacity={0.05} />
          <ReferenceArea y1={-90} y2={-70} fill="#ef4444" fillOpacity={0.05} />
          <ReferenceLine y={-70} stroke="#f59e0b" strokeDasharray="3 3" label={{ value: 'Poor', fill: '#f59e0b', fontSize: 10 }} />
          <XAxis dataKey="time" type="number" domain={['auto', 'auto']}
            tickFormatter={(val) => new Date(val).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            stroke="var(--text-muted)" fontSize={11} />
          <YAxis domain={[-90, -30]} stroke="var(--text-muted)" fontSize={11}
            tickFormatter={(val) => `${val}`} />
          <Tooltip
            contentStyle={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: 8 }}
            labelFormatter={(val) => new Date(val as number).toLocaleTimeString()}
            formatter={(val: any, _name: any, entry: any) => [
              `${val} dBm | ${entry.payload.band} Ch${entry.payload.channel} | ${entry.payload.speed} Mbps`
            ]}
          />
          <Line type="monotone" dataKey="signal" stroke="#8b5cf6" strokeWidth={2} dot={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
