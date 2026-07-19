import { main } from '../../wailsjs/go/models';
import './StatusCard.css';

interface Props {
  status: main.StatusResponse | null;
  loading: boolean;
}

export function StatusCard({ status, loading }: Props) {
  if (loading || !status) {
    return (
      <div className="status-card status-unknown">
        <div className="status-indicator" />
        <div className="status-content">
          <h2>Initializing...</h2>
          <p>Collecting network data</p>
        </div>
      </div>
    );
  }

  return (
    <div className={`status-card status-${status.status}`}>
      <div className="status-indicator" />
      <div className="status-content">
        <div className="status-header">
          <h2>{status.title}</h2>
          <span className="status-badge">{status.status}</span>
        </div>
        <p className="status-description">{status.description}</p>
        {status.evidence && status.evidence.length > 0 && (
          <div className="evidence-list">
            {status.evidence.map((e: any, i: number) => (
              <div key={i} className="evidence-item">
                <span className="evidence-label">{e.description}</span>
                <span className="evidence-value">{e.value}</span>
              </div>
            ))}
          </div>
        )}
        <div className="status-meta">
          <span>Confidence: {Math.round(status.confidence * 100)}%</span>
          <span>{new Date(status.timestamp).toLocaleTimeString()}</span>
        </div>
      </div>
    </div>
  );
}
