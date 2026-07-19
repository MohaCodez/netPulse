import type { UptimeStats } from '../hooks/useNetworkData';
import './UptimeCard.css';

interface Props {
  stats: UptimeStats | null;
}

export function UptimeCard({ stats }: Props) {
  if (!stats) return null;

  const getColor = (pct: number) => {
    if (pct >= 99.9) return '#22c55e';
    if (pct >= 99) return '#84cc16';
    if (pct >= 95) return '#f59e0b';
    return '#ef4444';
  };

  const items = [
    { label: '1 Hour', value: stats.one_hour },
    { label: '24 Hours', value: stats.twenty_four_h },
    { label: '7 Days', value: stats.seven_days },
  ];

  return (
    <div className="uptime-card">
      <h3>Uptime</h3>
      <div className="uptime-items">
        {items.map((item) => (
          <div key={item.label} className="uptime-item">
            <div className="uptime-ring" style={{ '--color': getColor(item.value), '--pct': `${item.value}%` } as any}>
              <span className="uptime-value">{item.value.toFixed(1)}%</span>
            </div>
            <span className="uptime-label">{item.label}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
