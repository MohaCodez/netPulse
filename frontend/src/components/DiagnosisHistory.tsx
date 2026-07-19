import { main } from '../../wailsjs/go/models';
import './DiagnosisHistory.css';

interface Props {
  diagnoses: main.DiagnosisResponse[];
}

export function DiagnosisHistory({ diagnoses }: Props) {
  if (diagnoses.length === 0) {
    return (
      <div className="diagnosis-container">
        <h3>Diagnosis History</h3>
        <p className="diagnosis-empty">No issues detected yet — your network is healthy!</p>
      </div>
    );
  }

  return (
    <div className="diagnosis-container">
      <h3>Diagnosis History</h3>
      <div className="diagnosis-list">
        {diagnoses.map((d) => (
          <div key={d.id} className={`diagnosis-item severity-${d.severity}`}>
            <div className="diagnosis-header">
              <span className={`severity-dot severity-${d.severity}`} />
              <span className="diagnosis-title">{d.title}</span>
              <span className="diagnosis-category">{d.category}</span>
              <span className="diagnosis-time">
                {new Date(d.timestamp).toLocaleString()}
              </span>
            </div>
            <p className="diagnosis-desc">{d.description}</p>
            {d.evidence && d.evidence.length > 0 && (
              <div className="diagnosis-evidence">
                {d.evidence.map((e: any, i: number) => (
                  <div key={i} className="evidence-row">
                    <span>{e.description}:</span>
                    <code>{e.value}</code>
                  </div>
                ))}
              </div>
            )}
            {d.resolved && (
              <span className="diagnosis-resolved">
                ✓ Resolved {d.resolved_at ? new Date(d.resolved_at).toLocaleTimeString() : ''}
              </span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
