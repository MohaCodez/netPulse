import { useMemo } from 'react';
import type { HeatmapCell } from '../hooks/useAnalytics';
import './Heatmap.css';

interface Props {
  data: HeatmapCell[] | null;
}

export function Heatmap({ data }: Props) {
  const { grid, probeTypes } = useMemo(() => {
    if (!data) return { grid: new Map(), probeTypes: [] };

    const types = new Set<string>();
    const grid = new Map<string, HeatmapCell>();

    for (const cell of data) {
      if (cell.probe_type === 'wifi') continue; // skip wifi from heatmap
      types.add(cell.probe_type);
      grid.set(`${cell.hour}-${cell.probe_type}`, cell);
    }

    return { grid, probeTypes: Array.from(types).sort() };
  }, [data]);

  if (!data || data.length === 0) {
    return <div className="chart-container"><h3>🗓️ Reliability Heatmap (by hour)</h3><p className="chart-empty">Need more data (run for a full day)</p></div>;
  }

  const hours = Array.from({ length: 24 }, (_, i) => i);

  const getColor = (cell: HeatmapCell | undefined) => {
    if (!cell) return 'var(--bg-subtle)';
    const rate = cell.success_rate;
    if (rate >= 99.5) return '#22c55e';
    if (rate >= 95) return '#84cc16';
    if (rate >= 90) return '#f59e0b';
    if (rate >= 80) return '#f97316';
    return '#ef4444';
  };

  return (
    <div className="chart-container">
      <h3>🗓️ Reliability Heatmap (by hour of day)</h3>
      <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', margin: '0 0 12px' }}>
        Color = success rate. Shows which hours have the most issues.
      </p>
      <div className="heatmap-grid">
        <div className="heatmap-row heatmap-header">
          <div className="heatmap-label"></div>
          {hours.map((h) => (
            <div key={h} className="heatmap-hour">{h}</div>
          ))}
        </div>
        {probeTypes.map((type) => (
          <div key={type} className="heatmap-row">
            <div className="heatmap-label">{type}</div>
            {hours.map((h) => {
              const cell = grid.get(`${h}-${type}`);
              return (
                <div
                  key={h}
                  className="heatmap-cell"
                  style={{ backgroundColor: getColor(cell) }}
                  title={cell ? `${type} @ ${h}:00 — ${cell.success_rate.toFixed(1)}% success, ${cell.avg_latency.toFixed(0)}ms avg, ${cell.sample_count} samples` : 'No data'}
                />
              );
            })}
          </div>
        ))}
      </div>
      <div className="heatmap-scale">
        <span>80%</span>
        <div className="scale-bar">
          <span style={{ background: '#ef4444' }} />
          <span style={{ background: '#f97316' }} />
          <span style={{ background: '#f59e0b' }} />
          <span style={{ background: '#84cc16' }} />
          <span style={{ background: '#22c55e' }} />
        </div>
        <span>100%</span>
      </div>
    </div>
  );
}
