import { useMemo } from 'react';
import type { WifiTimeseriesPoint } from '../hooks/useAnalytics';
import './BandHopChart.css';

interface Props {
  data: WifiTimeseriesPoint[] | null;
}

export function BandHopChart({ data }: Props) {
  const events = useMemo(() => {
    if (!data || data.length < 2) return [];
    const hops: { time: string; from: string; to: string; fromChannel: number; toChannel: number }[] = [];

    for (let i = 1; i < data.length; i++) {
      if (data[i].band !== data[i - 1].band || data[i].channel !== data[i - 1].channel) {
        hops.push({
          time: new Date(data[i].timestamp).toLocaleTimeString(),
          from: `${data[i - 1].band} Ch${data[i - 1].channel}`,
          to: `${data[i].band} Ch${data[i].channel}`,
          fromChannel: data[i - 1].channel,
          toChannel: data[i].channel,
        });
      }
    }
    return hops;
  }, [data]);

  const bandTimeline = useMemo(() => {
    if (!data) return [];
    return data.map((p) => ({
      time: new Date(p.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
      band: p.band,
      channel: p.channel,
      speed: p.link_speed_mbps,
      signal: p.signal_dbm,
    }));
  }, [data]);

  if (!data || data.length === 0) {
    return <div className="chart-container"><h3>🔄 Band Hopping</h3><p className="chart-empty">Collecting data...</p></div>;
  }

  const currentBand = data[data.length - 1]?.band || '—';
  const currentChannel = data[data.length - 1]?.channel || 0;

  return (
    <div className="chart-container">
      <h3>🔄 Band Hopping Tracker</h3>
      <div className="band-current">
        <span className={`band-badge band-${currentBand.replace(/\s|GHz/g, '').replace('.', '')}`}>
          {currentBand} · Ch {currentChannel}
        </span>
        <span className="hop-count">{events.length} hops in window</span>
      </div>

      {events.length > 0 ? (
        <div className="hop-timeline">
          {events.slice(-10).map((e, i) => (
            <div key={i} className="hop-event">
              <span className="hop-time">{e.time}</span>
              <span className="hop-from">{e.from}</span>
              <span className="hop-arrow">→</span>
              <span className="hop-to">{e.to}</span>
            </div>
          ))}
        </div>
      ) : (
        <p className="chart-empty" style={{ color: '#22c55e' }}>✓ Stable — no band switches detected</p>
      )}

      {/* Visual band strip */}
      <div className="band-strip">
        {bandTimeline.slice(-60).map((p, i) => (
          <div
            key={i}
            className={`band-block band-${p.band.replace(/\s|GHz/g, '').replace('.', '')}`}
            title={`${p.time} | ${p.band} Ch${p.channel} | ${p.signal}dBm | ${p.speed}Mbps`}
          />
        ))}
      </div>
      <div className="band-legend">
        <span className="legend-item"><span className="legend-dot band-24"></span> 2.4 GHz</span>
        <span className="legend-item"><span className="legend-dot band-5"></span> 5 GHz</span>
      </div>
    </div>
  );
}
