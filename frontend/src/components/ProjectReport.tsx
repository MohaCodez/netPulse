import { useState, useEffect } from 'react';
import { ArchDiagram } from './ArchDiagram';
import './ArchDiagram.css';
import './ProjectReport.css';

interface ReportData {
  totalProbes: number;
  totalDiagnoses: number;
  totalSpeedTests: number;
  totalWifiSnapshots: number;
  totalNetworkEvents: number;
  uptimeHour: number;
  uptime24h: number;
  uptime7d: number;
  avgLatency: number;
  avgGatewayLatency: number;
  avgDnsLatency: number;
  avgDownload: number;
  avgUpload: number;
  avgSignal: number;
  bandHops: number;
  currentSSID: string;
  currentBand: string;
  currentChannel: number;
  monitoringSince: string;
}

export function ProjectReport() {
  const [report, setReport] = useState<ReportData | null>(null);

  useEffect(() => {
    if (window.go?.main?.App?.GetProjectReport) {
      window.go.main.App.GetProjectReport().then((r: ReportData) => setReport(r));
    }
  }, []);

  return (
    <div className="report-page">
      <h2>📄 Project Report</h2>
      <p className="report-subtitle">NetPulse — Architecture, Implementation &amp; Findings</p>

      <section className="report-section">
        <h3>1. Problem Statement</h3>
        <p>
          Users know their internet "feels slow" or "keeps dropping" but lack data to determine whether
          the cause is their ISP, router, Wi-Fi signal, or DNS. Existing tools (Pi-hole, UptimeRobot, PRTG)
          report <em>that</em> something broke but not <em>why</em>.
        </p>
        <p>
          NetPulse solves this by running a layered diagnostic engine that isolates the root cause
          and explains it in plain language with supporting evidence — all running locally with no cloud dependency.
        </p>
      </section>

      <section className="report-section">
        <h3>2. System Architecture</h3>
        <ArchDiagram />
      </section>

      <section className="report-section">
        <h3>3. Probe System</h3>
        <p>9 concurrent probes run every 5 seconds:</p>
        <table className="report-table">
          <thead>
            <tr><th>Probe</th><th>Method</th><th>Purpose</th></tr>
          </thead>
          <tbody>
            <tr><td>Gateway</td><td>ICMP to default route</td><td>Local network reachability</td></tr>
            <tr><td>External Ping (×3)</td><td>ICMP to 8.8.8.8, 1.1.1.1, 208.67.222.222</td><td>Internet reachability</td></tr>
            <tr><td>DNS (×3 domains)</td><td>Multi-resolver lookup</td><td>Name resolution health</td></tr>
            <tr><td>TCP</td><td>TCP connect to 8.8.8.8:443</td><td>Fallback when ICMP blocked</td></tr>
            <tr><td>Wi-Fi</td><td>nmcli / iwconfig / /proc</td><td>Signal, channel, band, link speed</td></tr>
          </tbody>
        </table>
      </section>

      <section className="report-section">
        <h3>4. Diagnosis Engine</h3>
        <p>5-layer decision tree evaluated every 10 seconds over a 30-second data window:</p>
        <div className="decision-tree">
          <div className="tree-layer layer-1">
            <strong>Layer 1: Gateway</strong> → Router reachable? If no → local network / router down
          </div>
          <div className="tree-layer layer-2">
            <strong>Layer 2: ISP</strong> → External targets reachable? If no → ISP outage
          </div>
          <div className="tree-layer layer-3">
            <strong>Layer 3: DNS</strong> → Resolvers responding? System vs alternates? If no → DNS issue
          </div>
          <div className="tree-layer layer-4">
            <strong>Layer 4: Wi-Fi</strong> → Signal adequate? If weak → Wi-Fi bottleneck
          </div>
          <div className="tree-layer layer-5">
            <strong>Layer 5: Throughput</strong> → Latency/jitter/loss vs baseline → performance degradation
          </div>
        </div>
        <p>Each diagnosis includes a confidence score (0–100%) and evidence array with supporting metrics.</p>
      </section>

      <section className="report-section">
        <h3>5. Key ECE Concepts Applied</h3>
        <ul className="concepts-list">
          <li><strong>Shannon's Theorem:</strong> Channel capacity C = B × log₂(1 + SNR) — explains why weak signal reduces throughput even when "connected"</li>
          <li><strong>CSMA/CA:</strong> Wi-Fi contention mechanism — more devices = more backoff = higher gateway latency</li>
          <li><strong>Multipath Fading:</strong> Explains signal fluctuations (-58 to -71 dBm) even when stationary</li>
          <li><strong>PHY Rate Adaptation:</strong> MCS index drops with signal, increasing airtime per packet</li>
          <li><strong>Bufferbloat:</strong> High latency + low loss = router buffers too large</li>
          <li><strong>Band Steering:</strong> 2.4 GHz ↔ 5 GHz switching behavior and its impact on performance continuity</li>
          <li><strong>TCP Congestion Control:</strong> 1% packet loss → 50%+ throughput reduction due to window halving</li>
        </ul>
      </section>

      <section className="report-section">
        <h3>6. Collected Data Summary</h3>
        {report ? (
          <div className="data-summary">
            <div className="summary-grid">
              <div className="summary-item">
                <span className="summary-value">{report.totalProbes.toLocaleString()}</span>
                <span className="summary-label">Probe Results</span>
              </div>
              <div className="summary-item">
                <span className="summary-value">{report.totalDiagnoses}</span>
                <span className="summary-label">Diagnoses</span>
              </div>
              <div className="summary-item">
                <span className="summary-value">{report.totalSpeedTests}</span>
                <span className="summary-label">Speed Tests</span>
              </div>
              <div className="summary-item">
                <span className="summary-value">{report.totalWifiSnapshots.toLocaleString()}</span>
                <span className="summary-label">Wi-Fi Snapshots</span>
              </div>
              <div className="summary-item">
                <span className="summary-value">{report.totalNetworkEvents}</span>
                <span className="summary-label">Network Changes</span>
              </div>
              <div className="summary-item">
                <span className="summary-value">{report.bandHops}</span>
                <span className="summary-label">Band Hops</span>
              </div>
            </div>

            <h4>Network Performance</h4>
            <table className="report-table">
              <tbody>
                <tr><td>Uptime (1h / 24h / 7d)</td><td>{report.uptimeHour.toFixed(1)}% / {report.uptime24h.toFixed(1)}% / {report.uptime7d.toFixed(1)}%</td></tr>
                <tr><td>Avg External Latency</td><td>{report.avgLatency.toFixed(1)} ms</td></tr>
                <tr><td>Avg Gateway Latency</td><td>{report.avgGatewayLatency.toFixed(1)} ms</td></tr>
                <tr><td>Avg DNS Resolution</td><td>{report.avgDnsLatency.toFixed(1)} ms</td></tr>
                <tr><td>Avg Download Speed</td><td>{report.avgDownload.toFixed(1)} Mbps</td></tr>
                <tr><td>Avg Upload Speed</td><td>{report.avgUpload.toFixed(1)} Mbps</td></tr>
                <tr><td>Avg Wi-Fi Signal</td><td>{report.avgSignal} dBm</td></tr>
                <tr><td>Current Network</td><td>{report.currentSSID} ({report.currentBand} Ch{report.currentChannel})</td></tr>
              </tbody>
            </table>
          </div>
        ) : (
          <p className="loading-text">Loading data...</p>
        )}
      </section>

      <section className="report-section">
        <h3>7. Key Findings</h3>
        {report && (
          <div className="findings">
            {report.avgGatewayLatency > 20 && (
              <div className="finding finding-warning">
                <strong>High Gateway Latency ({report.avgGatewayLatency.toFixed(0)}ms):</strong> Local network hop should be &lt;5ms.
                This indicates Wi-Fi contention or signal issues between device and router.
              </div>
            )}
            {report.avgSignal < -65 && (
              <div className="finding finding-warning">
                <strong>Weak Wi-Fi Signal ({report.avgSignal} dBm):</strong> Below -65 dBm causes PHY rate drops
                and increased retransmissions. Consider moving closer to router or switching to 5 GHz.
              </div>
            )}
            {report.bandHops > 5 && (
              <div className="finding finding-info">
                <strong>Frequent Band Switching ({report.bandHops} hops):</strong> Router's band steering
                is indecisive. This causes brief connectivity drops during each transition.
                Consider locking to 5 GHz if signal permits.
              </div>
            )}
            {report.avgDownload < 20 && report.avgDownload > 0 && (
              <div className="finding finding-info">
                <strong>Throughput Below Plan ({report.avgDownload.toFixed(0)} Mbps):</strong> If your ISP plan
                is higher, the bottleneck is likely your Wi-Fi link (signal: {report.avgSignal} dBm) rather than
                the ISP connection itself.
              </div>
            )}
            {report.uptime7d >= 99.5 && (
              <div className="finding finding-good">
                <strong>Excellent Reliability ({report.uptime7d.toFixed(1)}% uptime):</strong> Your connection
                is highly stable with minimal drops.
              </div>
            )}
          </div>
        )}
      </section>

      <section className="report-section">
        <h3>8. Technology Stack</h3>
        <table className="report-table">
          <thead><tr><th>Layer</th><th>Technology</th><th>Rationale</th></tr></thead>
          <tbody>
            <tr><td>Backend</td><td>Go 1.26</td><td>Concurrency for parallel probes, low resource footprint, single binary</td></tr>
            <tr><td>Storage</td><td>SQLite (WAL)</td><td>Zero-config embedded, batched writes, auto-maintenance</td></tr>
            <tr><td>Desktop</td><td>Wails v2</td><td>Native webview without Electron overhead, Go ↔ JS bindings</td></tr>
            <tr><td>Frontend</td><td>React + TypeScript</td><td>Type safety, component architecture, Recharts for visualization</td></tr>
            <tr><td>AI</td><td>Ollama (Qwen 2.5)</td><td>Local inference, no cloud dependency, network-context-aware</td></tr>
            <tr><td>Notifications</td><td>notify-send</td><td>Native Linux desktop notifications</td></tr>
          </tbody>
        </table>
      </section>

      <section className="report-section">
        <h3>9. Future Work</h3>
        <ul>
          <li>Multi-device monitoring across local network</li>
          <li>Raspberry Pi appliance mode for continuous monitoring</li>
          <li>Neighboring networks scanner for channel congestion analysis</li>
          <li>Packet loss burst detection (random vs bursty classification)</li>
          <li>Prometheus/Grafana export for advanced analytics</li>
          <li>Regional outage map via anonymous opt-in telemetry</li>
        </ul>
      </section>
    </div>
  );
}
