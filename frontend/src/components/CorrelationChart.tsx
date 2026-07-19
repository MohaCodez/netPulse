import { useMemo } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import type { WifiTimeseriesPoint, ProbeLatencyPoint } from '../hooks/useAnalytics';

interface Props {
  wifiData: WifiTimeseriesPoint[] | null;
  latencyData: ProbeLatencyPoint[] | null;
}

export function CorrelationChart({ wifiData, latencyData }: Props) {
  const chartData = useMemo(() => {
    if (!wifiData || !latencyData) return [];

    // Merge wifi signal and latency on closest timestamps (5s buckets)
    const wifiBuckets = new Map<number, number>();
    for (const w of wifiData) {
      const bucket = Math.floor(new Date(w.timestamp).getTime() / 5000) * 5000;
      wifiBuckets.set(bucket, w.signal_dbm);
    }

    const latBuckets = new Map<number, number>();
    for (const l of latencyData) {
      const bucket = Math.floor(new Date(l.timestamp).getTime() / 5000) * 5000;
      if (l.gateway_ms > 0) {
        latBuckets.set(bucket, l.gateway_ms);
      }
    }

    // Find overlapping buckets
    const allBuckets = new Set([...wifiBuckets.keys(), ...latBuckets.keys()]);
    const sorted = Array.from(allBuckets).sort();

    let lastSignal = -60;
    let lastLatency = 0;

    return sorted.map((bucket) => {
      const signal = wifiBuckets.get(bucket) ?? lastSignal;
      const latency = latBuckets.get(bucket) ?? lastLatency;
      lastSignal = signal;
      lastLatency = latency;
      return {
        time: bucket,
        signal,
        latency,
        // Invert signal for visual correlation (more negative = worse, show as higher)
        signalInv: Math.abs(signal),
      };
    }).slice(-120); // last 120 points
  }, [wifiData, latencyData]);

  if (chartData.length === 0) {
    return <div className="chart-container"><h3>📊 Signal ↔ Latency Correlation</h3><p className="chart-empty">Collecting data...</p></div>;
  }

  return (
    <div className="chart-container">
      <h3>📊 Signal ↔ Latency Correlation</h3>
      <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', margin: '0 0 12px' }}>
        When signal drops (purple dips down), latency spikes (green goes up) — proving Wi-Fi is your bottleneck.
      </p>
      <ResponsiveContainer width="100%" height={240}>
        <LineChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
          <XAxis dataKey="time" type="number" domain={['auto', 'auto']}
            tickFormatter={(val) => new Date(val).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            stroke="var(--text-muted)" fontSize={11} />
          <YAxis yAxisId="signal" domain={[-90, -30]}
            stroke="#8b5cf6" fontSize={11}
            tickFormatter={(val) => `${val}dBm`}
            label={{ value: 'Signal', angle: -90, position: 'insideLeft', fill: '#8b5cf6', fontSize: 10 }}
          />
          <YAxis yAxisId="latency" orientation="right"
            stroke="#22c55e" fontSize={11}
            tickFormatter={(val) => `${val}ms`}
            label={{ value: 'Latency', angle: 90, position: 'insideRight', fill: '#22c55e', fontSize: 10 }}
          />
          <Tooltip
            contentStyle={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: 8 }}
            labelFormatter={(val) => new Date(val as number).toLocaleTimeString()}
            formatter={(val: any, name: any) => [
              name === 'signal' ? `${val} dBm` : `${Number(val).toFixed(1)} ms`
            ]}
          />
          <Legend />
          <Line yAxisId="signal" type="monotone" dataKey="signal" stroke="#8b5cf6" strokeWidth={2} dot={false} name="Wi-Fi Signal (dBm)" />
          <Line yAxisId="latency" type="monotone" dataKey="latency" stroke="#22c55e" strokeWidth={2} dot={false} name="Gateway Latency (ms)" />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
