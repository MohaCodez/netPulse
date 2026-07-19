import type { WifiInfo } from '../hooks/useNetworkData';
import './WifiCard.css';

interface Props {
  wifi: WifiInfo | null;
}

export function WifiCard({ wifi }: Props) {
  if (!wifi) {
    return (
      <div className="wifi-card">
        <h3>Wi-Fi</h3>
        <p className="wifi-empty">No Wi-Fi data available</p>
      </div>
    );
  }

  const signalBars = getSignalBars(wifi.signal_dbm);
  const qualityClass = wifi.signal_quality.toLowerCase().replace(' ', '-');

  return (
    <div className="wifi-card">
      <div className="wifi-header">
        <h3>Wi-Fi</h3>
        <span className={`wifi-quality quality-${qualityClass}`}>{wifi.signal_quality}</span>
      </div>

      <div className="wifi-ssid-row">
        <div className="wifi-signal-bars" title={`${wifi.signal_dbm} dBm`}>
          {[1, 2, 3, 4].map((bar) => (
            <div key={bar} className={`signal-bar ${bar <= signalBars ? 'active' : ''}`} />
          ))}
        </div>
        <span className="wifi-ssid">{wifi.ssid || 'Unknown'}</span>
      </div>

      <div className="wifi-details">
        <div className="wifi-detail">
          <span className="detail-label">Signal</span>
          <span className="detail-value">{wifi.signal_dbm} dBm</span>
        </div>
        <div className="wifi-detail">
          <span className="detail-label">Channel</span>
          <span className="detail-value">{wifi.channel || '—'}</span>
        </div>
        <div className="wifi-detail">
          <span className="detail-label">Band</span>
          <span className="detail-value">{wifi.band}</span>
        </div>
        <div className="wifi-detail">
          <span className="detail-label">Link Speed</span>
          <span className="detail-value">{wifi.link_speed_mbps ? `${wifi.link_speed_mbps} Mbps` : '—'}</span>
        </div>
      </div>
    </div>
  );
}

function getSignalBars(dbm: number): number {
  if (dbm >= -50) return 4;
  if (dbm >= -60) return 3;
  if (dbm >= -70) return 2;
  if (dbm >= -80) return 1;
  return 0;
}
