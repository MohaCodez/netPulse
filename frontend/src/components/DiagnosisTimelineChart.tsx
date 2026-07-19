import { useMemo } from 'react';
import type { DiagnosisPeriod } from '../hooks/useAnalytics';
import './DiagnosisTimeline.css';

interface Props {
  data: DiagnosisPeriod[] | null;
}

export function DiagnosisTimelineChart({ data }: Props) {
  const { blocks, totalHours } = useMemo(() => {
    if (!data) return { blocks: [], totalHours: 24 };

    const now = new Date();
    const startOfWindow = new Date(now.getTime() - 24 * 60 * 60 * 1000);
    const totalMs = now.getTime() - startOfWindow.getTime();

    const blocks = data.map((p) => {
      const start = new Date(p.start);
      const end = new Date(p.end);
      const leftPct = Math.max(0, (start.getTime() - startOfWindow.getTime()) / totalMs * 100);
      const widthPct = Math.min(100 - leftPct, (end.getTime() - start.getTime()) / totalMs * 100);

      return {
        left: leftPct,
        width: Math.max(widthPct, 0.5), // minimum visible width
        severity: p.severity,
        category: p.category,
        title: p.title,
        start: start.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        end: end.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      };
    });

    return { blocks, totalHours: 24 };
  }, [data]);

  const hourMarkers = Array.from({ length: totalHours + 1 }, (_, i) => {
    const t = new Date(Date.now() - (totalHours - i) * 60 * 60 * 1000);
    return t.getHours();
  });

  return (
    <div className="chart-container">
      <h3>🩺 Diagnosis Timeline (24h)</h3>
      <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', margin: '0 0 12px' }}>
        Green = healthy. Colored blocks = detected issues.
      </p>
      <div className="timeline-bar">
        <div className="timeline-healthy" />
        {blocks.map((b, i) => (
          <div
            key={i}
            className={`timeline-block severity-${b.severity}`}
            style={{ left: `${b.left}%`, width: `${b.width}%` }}
            title={`${b.start}–${b.end} | ${b.category}: ${b.title}`}
          />
        ))}
      </div>
      <div className="timeline-hours">
        {hourMarkers.filter((_, i) => i % 4 === 0).map((h, i) => (
          <span key={i}>{h}:00</span>
        ))}
      </div>
      {blocks.length === 0 && (
        <p className="chart-empty" style={{ color: '#22c55e', marginTop: 8 }}>✓ No issues in the last 24 hours</p>
      )}
    </div>
  );
}
