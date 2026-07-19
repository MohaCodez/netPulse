import { useState } from 'react';
import './ExportButton.css';

export function ExportButton() {
  const [exporting, setExporting] = useState(false);
  const [showConfirm, setShowConfirm] = useState<'clear' | 'export-clear' | null>(null);

  const downloadJson = (data: string, filename: string) => {
    const blob = new Blob([data], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const handleExportProbes = async () => {
    setExporting(true);
    try {
      if (window.go?.main?.App?.ExportData) {
        const data = await window.go.main.App.ExportData(1440);
        downloadJson(data, `netpulse-probes-${dateStr()}.json`);
      }
    } catch (err) {
      console.error('Export failed:', err);
    } finally {
      setExporting(false);
    }
  };

  const handleExportFull = async () => {
    setExporting(true);
    try {
      if (window.go?.main?.App?.ExportFullReport) {
        const data = await window.go.main.App.ExportFullReport();
        downloadJson(data, `netpulse-full-report-${dateStr()}.json`);
      }
    } catch (err) {
      console.error('Export failed:', err);
    } finally {
      setExporting(false);
    }
  };

  const handleClear = async () => {
    try {
      if (window.go?.main?.App?.ClearAllData) {
        await window.go.main.App.ClearAllData();
        setShowConfirm(null);
      }
    } catch (err) {
      console.error('Clear failed:', err);
    }
  };

  const handleExportAndClear = async () => {
    setExporting(true);
    try {
      if (window.go?.main?.App?.ExportAndClear) {
        const data = await window.go.main.App.ExportAndClear();
        downloadJson(data, `netpulse-full-report-${dateStr()}.json`);
        setShowConfirm(null);
      }
    } catch (err) {
      console.error('Export and clear failed:', err);
    } finally {
      setExporting(false);
    }
  };

  return (
    <div className="export-section">
      <h3>Data Management</h3>
      <div className="export-buttons">
        <button className="export-btn" onClick={handleExportProbes} disabled={exporting}>
          📊 Export Probes (24h)
        </button>
        <button className="export-btn" onClick={handleExportFull} disabled={exporting}>
          📋 Export Full Report
        </button>
        <button className="export-btn btn-warning" onClick={() => setShowConfirm('export-clear')} disabled={exporting}>
          💾 Download Report &amp; Clear
        </button>
        <button className="export-btn btn-danger" onClick={() => setShowConfirm('clear')} disabled={exporting}>
          🗑️ Clear All Data
        </button>
      </div>

      {showConfirm && (
        <div className="confirm-overlay">
          <div className="confirm-box">
            <p>
              {showConfirm === 'clear'
                ? 'Are you sure? This will permanently delete all stored data.'
                : 'This will download a full report then clear all data.'}
            </p>
            <div className="confirm-actions">
              <button className="confirm-btn btn-cancel" onClick={() => setShowConfirm(null)}>
                Cancel
              </button>
              <button
                className="confirm-btn btn-confirm"
                onClick={showConfirm === 'clear' ? handleClear : handleExportAndClear}
              >
                {showConfirm === 'clear' ? 'Clear Data' : 'Export & Clear'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function dateStr() {
  return new Date().toISOString().split('T')[0];
}
