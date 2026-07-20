import './ArchDiagram.css';

export function ArchDiagram() {
  return (
    <div className="arch-diagram-visual">
      <svg viewBox="0 0 800 520" xmlns="http://www.w3.org/2000/svg" className="arch-svg">
        {/* Background */}
        <rect x="0" y="0" width="800" height="520" fill="var(--bg-subtle)" rx="12" />

        {/* Title */}
        <text x="400" y="30" textAnchor="middle" className="arch-title">NetPulse System Architecture</text>

        {/* Desktop App Container */}
        <rect x="20" y="45" width="760" height="460" fill="none" stroke="var(--border)" strokeWidth="2" rx="10" strokeDasharray="6 3" />
        <text x="40" y="68" className="arch-label">Desktop App (Wails v2 — WebKit2GTK)</text>

        {/* Frontend Box */}
        <rect x="40" y="80" width="340" height="200" fill="var(--card-bg)" stroke="#3b82f6" strokeWidth="2" rx="8" />
        <text x="210" y="105" textAnchor="middle" className="arch-box-title" fill="#3b82f6">React Frontend</text>

        {/* Frontend components */}
        <rect x="60" y="115" width="140" height="35" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="5" />
        <text x="130" y="137" textAnchor="middle" className="arch-item">Dashboard UI</text>

        <rect x="220" y="115" width="140" height="35" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="5" />
        <text x="290" y="137" textAnchor="middle" className="arch-item">Analytics (10 charts)</text>

        <rect x="60" y="160" width="140" height="35" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="5" />
        <text x="130" y="182" textAnchor="middle" className="arch-item">AI Chat</text>

        <rect x="220" y="160" width="140" height="35" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="5" />
        <text x="290" y="182" textAnchor="middle" className="arch-item">Theory / Report</text>

        <rect x="60" y="205" width="300" height="30" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="5" />
        <text x="210" y="224" textAnchor="middle" className="arch-item">Hooks (useNetworkData, useAnalytics, useWailsEvents)</text>

        <rect x="60" y="245" width="300" height="25" fill="#3b82f620" stroke="#3b82f6" strokeWidth="1" rx="5" />
        <text x="210" y="262" textAnchor="middle" className="arch-item" fill="#3b82f6">Wails JS Bindings (auto-generated)</text>

        {/* Backend Box */}
        <rect x="420" y="80" width="340" height="410" fill="var(--card-bg)" stroke="#22c55e" strokeWidth="2" rx="8" />
        <text x="590" y="105" textAnchor="middle" className="arch-box-title" fill="#22c55e">Go Backend</text>

        {/* API Layer */}
        <rect x="440" y="115" width="300" height="30" fill="#22c55e20" stroke="#22c55e" strokeWidth="1" rx="5" />
        <text x="590" y="135" textAnchor="middle" className="arch-item" fill="#22c55e">API Layer (app.go, api.go, analytics_api.go)</text>

        {/* Probe Runner */}
        <rect x="440" y="155" width="145" height="55" fill="var(--bg-subtle)" stroke="#f59e0b" strokeWidth="1.5" rx="5" />
        <text x="512" y="175" textAnchor="middle" className="arch-item-small">Probe Runner</text>
        <text x="512" y="195" textAnchor="middle" className="arch-detail">9 concurrent probes</text>
        <text x="512" y="205" textAnchor="middle" className="arch-detail">ping/dns/tcp/wifi/gw</text>

        {/* Diagnosis Engine */}
        <rect x="595" y="155" width="145" height="55" fill="var(--bg-subtle)" stroke="#ef4444" strokeWidth="1.5" rx="5" />
        <text x="667" y="175" textAnchor="middle" className="arch-item-small">Diagnosis Engine</text>
        <text x="667" y="195" textAnchor="middle" className="arch-detail">5-layer decision tree</text>
        <text x="667" y="205" textAnchor="middle" className="arch-detail">evidence + confidence</text>

        {/* Arrow from probes to diagnosis */}
        <path d="M 585 182 L 595 182" stroke="var(--text-muted)" strokeWidth="1.5" fill="none" markerEnd="url(#arrowhead)" />

        {/* Speed Test */}
        <rect x="440" y="220" width="145" height="40" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="5" />
        <text x="512" y="237" textAnchor="middle" className="arch-item-small">Speed Test</text>
        <text x="512" y="252" textAnchor="middle" className="arch-detail">Cloudflare CDN</text>

        {/* Network Watcher */}
        <rect x="595" y="220" width="145" height="40" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="5" />
        <text x="667" y="237" textAnchor="middle" className="arch-item-small">Network Watcher</text>
        <text x="667" y="252" textAnchor="middle" className="arch-detail">SSID/band change detect</text>

        {/* LAN Scanner */}
        <rect x="440" y="270" width="145" height="40" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="5" />
        <text x="512" y="287" textAnchor="middle" className="arch-item-small">LAN Scanner</text>
        <text x="512" y="302" textAnchor="middle" className="arch-detail">ARP + ping sweep</text>

        {/* Notifier */}
        <rect x="595" y="270" width="145" height="40" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="5" />
        <text x="667" y="287" textAnchor="middle" className="arch-item-small">Notifier</text>
        <text x="667" y="302" textAnchor="middle" className="arch-detail">notify-send (Linux)</text>

        {/* SQLite */}
        <rect x="440" y="325" width="300" height="50" fill="#8b5cf620" stroke="#8b5cf6" strokeWidth="2" rx="8" />
        <text x="590" y="347" textAnchor="middle" className="arch-item-small" fill="#8b5cf6">SQLite (WAL mode)</text>
        <text x="590" y="365" textAnchor="middle" className="arch-detail">probe_results | diagnoses | speed_tests | wifi_snapshots | baselines | network_events</text>

        {/* Maintenance */}
        <rect x="440" y="385" width="145" height="35" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="5" />
        <text x="512" y="405" textAnchor="middle" className="arch-item-small">Maintenance</text>
        <text x="512" y="418" textAnchor="middle" className="arch-detail">purge + baselines</text>

        {/* Batched Writer */}
        <rect x="595" y="385" width="145" height="35" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="5" />
        <text x="667" y="405" textAnchor="middle" className="arch-item-small">Batched Writer</text>
        <text x="667" y="418" textAnchor="middle" className="arch-detail">single goroutine</text>

        {/* Wails Events arrow */}
        <path d="M 420 230 L 380 230" stroke="#22c55e" strokeWidth="2" fill="none" markerEnd="url(#arrowGreen)" />
        <text x="400" y="220" textAnchor="middle" className="arch-detail" fill="#22c55e">events</text>

        {/* Wails Bindings arrow */}
        <path d="M 380 262 L 420 262" stroke="#3b82f6" strokeWidth="2" fill="none" markerEnd="url(#arrowBlue)" />
        <text x="400" y="275" textAnchor="middle" className="arch-detail" fill="#3b82f6">calls</text>

        {/* External: Ollama */}
        <rect x="40" y="310" width="160" height="50" fill="var(--card-bg)" stroke="#f59e0b" strokeWidth="2" rx="8" />
        <text x="120" y="335" textAnchor="middle" className="arch-item-small" fill="#f59e0b">Ollama (Local LLM)</text>
        <text x="120" y="350" textAnchor="middle" className="arch-detail">qwen2.5-coder:3b</text>

        {/* External: Network */}
        <rect x="40" y="380" width="340" height="110" fill="var(--card-bg)" stroke="var(--border)" strokeWidth="1" rx="8" />
        <text x="210" y="400" textAnchor="middle" className="arch-item-small">Network Targets</text>

        <rect x="60" y="410" width="90" height="30" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="4" />
        <text x="105" y="429" textAnchor="middle" className="arch-detail">Gateway</text>

        <rect x="160" y="410" width="90" height="30" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="4" />
        <text x="205" y="429" textAnchor="middle" className="arch-detail">8.8.8.8 / 1.1.1.1</text>

        <rect x="260" y="410" width="100" height="30" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="4" />
        <text x="310" y="429" textAnchor="middle" className="arch-detail">DNS Resolvers</text>

        <rect x="60" y="450" width="90" height="30" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="4" />
        <text x="105" y="469" textAnchor="middle" className="arch-detail">Cloudflare CDN</text>

        <rect x="160" y="450" width="90" height="30" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="4" />
        <text x="205" y="469" textAnchor="middle" className="arch-detail">LAN Devices</text>

        <rect x="260" y="450" width="100" height="30" fill="var(--bg-subtle)" stroke="var(--border)" strokeWidth="1" rx="4" />
        <text x="310" y="469" textAnchor="middle" className="arch-detail">nmcli / iwconfig</text>

        {/* Arrow definitions */}
        <defs>
          <marker id="arrowhead" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto">
            <polygon points="0 0, 8 3, 0 6" fill="var(--text-muted)" />
          </marker>
          <marker id="arrowGreen" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto">
            <polygon points="0 0, 8 3, 0 6" fill="#22c55e" />
          </marker>
          <marker id="arrowBlue" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto">
            <polygon points="0 0, 8 3, 0 6" fill="#3b82f6" />
          </marker>
        </defs>
      </svg>
    </div>
  );
}
