import { useState } from 'react';
import { RunSpeedTest } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';
import './SpeedTest.css';

interface Props {
  tests: main.SpeedTestResponse[];
  onTestComplete: () => void;
}

export function SpeedTest({ tests, onTestComplete }: Props) {
  const [running, setRunning] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleRunTest = async () => {
    setRunning(true);
    setError(null);
    try {
      await RunSpeedTest();
      onTestComplete();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Speed test failed');
    } finally {
      setRunning(false);
    }
  };

  const lastTest = tests.length > 0 ? tests[0] : null;

  return (
    <div className="speedtest-container">
      <div className="speedtest-header">
        <h3>Speed Test</h3>
        <button
          className="speedtest-btn"
          onClick={handleRunTest}
          disabled={running}
        >
          {running ? 'Testing...' : 'Run Test'}
        </button>
      </div>

      {error && <p className="speedtest-error">{error}</p>}

      {lastTest && (
        <div className="speedtest-results">
          <div className="speedtest-metric">
            <span className="metric-label">Download</span>
            <span className="metric-value">{lastTest.download_mbps.toFixed(1)}</span>
            <span className="metric-unit">Mbps</span>
          </div>
          <div className="speedtest-metric">
            <span className="metric-label">Upload</span>
            <span className="metric-value">{lastTest.upload_mbps.toFixed(1)}</span>
            <span className="metric-unit">Mbps</span>
          </div>
          <div className="speedtest-metric">
            <span className="metric-label">Latency</span>
            <span className="metric-value">{lastTest.latency_ms.toFixed(0)}</span>
            <span className="metric-unit">ms</span>
          </div>
          <div className="speedtest-metric">
            <span className="metric-label">Jitter</span>
            <span className="metric-value">{lastTest.jitter_ms.toFixed(0)}</span>
            <span className="metric-unit">ms</span>
          </div>
        </div>
      )}

      {lastTest && (
        <p className="speedtest-meta">
          Last test: {new Date(lastTest.timestamp).toLocaleString()} ({lastTest.triggered_by})
        </p>
      )}
    </div>
  );
}
