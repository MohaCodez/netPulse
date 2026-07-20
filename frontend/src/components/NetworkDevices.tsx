import { useState } from 'react';
import './NetworkDevices.css';

interface Device {
  ip: string;
  mac: string;
  vendor: string;
  hostname: string;
  interface: string;
  first_seen: string;
  last_seen: string;
  is_gateway: boolean;
  is_local: boolean;
}

export function NetworkDevices() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [scanning, setScanning] = useState(false);
  const [lastScan, setLastScan] = useState<string | null>(null);

  const handleScan = async () => {
    setScanning(true);
    try {
      if (window.go?.main?.App?.ScanNetwork) {
        const result = await window.go.main.App.ScanNetwork();
        setDevices(result || []);
        setLastScan(new Date().toLocaleTimeString());
      }
    } catch (err) {
      console.error('Scan failed:', err);
    } finally {
      setScanning(false);
    }
  };

  const sortedDevices = [...devices].sort((a, b) => {
    // Gateway first, then local, then by IP
    if (a.is_gateway) return -1;
    if (b.is_gateway) return 1;
    if (a.is_local) return -1;
    if (b.is_local) return 1;
    return a.ip.localeCompare(b.ip, undefined, { numeric: true });
  });

  return (
    <div className="devices-container">
      <div className="devices-header">
        <div>
          <h3>🖥️ Network Devices</h3>
          {lastScan && <span className="last-scan">Last scan: {lastScan}</span>}
        </div>
        <button className="scan-btn" onClick={handleScan} disabled={scanning}>
          {scanning ? 'Scanning...' : '🔍 Scan Network'}
        </button>
      </div>

      {devices.length === 0 && !scanning && (
        <div className="devices-empty">
          <p>Click "Scan Network" to discover devices on your LAN</p>
          <p className="devices-note">Uses ARP + ping sweep on your local subnet</p>
        </div>
      )}

      {scanning && (
        <div className="devices-scanning">
          <div className="scan-animation" />
          <p>Scanning 192.168.x.x subnet...</p>
        </div>
      )}

      {devices.length > 0 && (
        <>
          <div className="devices-summary">
            <span className="device-count">{devices.length} devices found</span>
          </div>
          <div className="devices-list">
            {sortedDevices.map((d) => (
              <div key={d.mac} className={`device-row ${d.is_gateway ? 'is-gateway' : ''} ${d.is_local ? 'is-local' : ''}`}>
                <div className="device-icon">
                  {d.is_gateway ? '🌐' : d.is_local ? '💻' : getDeviceIcon(d.vendor)}
                </div>
                <div className="device-info">
                  <div className="device-name">
                    {d.hostname || d.vendor || 'Unknown Device'}
                    {d.is_gateway && <span className="device-tag tag-gateway">Gateway</span>}
                    {d.is_local && <span className="device-tag tag-local">This Device</span>}
                  </div>
                  <div className="device-details">
                    <span className="device-ip">{d.ip}</span>
                    <span className="device-mac">{d.mac}</span>
                    {d.vendor && d.vendor !== 'Unknown' && <span className="device-vendor">{d.vendor}</span>}
                  </div>
                </div>
                <div className="device-seen">
                  <span className="seen-label">Last seen</span>
                  <span className="seen-time">{formatTime(d.last_seen)}</span>
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
}

function getDeviceIcon(vendor: string): string {
  const v = vendor.toLowerCase();
  if (v.includes('apple') || v.includes('iphone') || v.includes('ipad')) return '🍎';
  if (v.includes('samsung')) return '📱';
  if (v.includes('xiaomi') || v.includes('oneplus')) return '📱';
  if (v.includes('google')) return '🔵';
  if (v.includes('amazon') || v.includes('alexa')) return '🔊';
  if (v.includes('intel') || v.includes('dell') || v.includes('hp') || v.includes('asus')) return '💻';
  if (v.includes('tp-link') || v.includes('netgear') || v.includes('cisco') || v.includes('d-link')) return '📡';
  if (v.includes('vmware') || v.includes('virtual') || v.includes('qemu') || v.includes('docker')) return '🐳';
  return '📟';
}

function formatTime(ts: string): string {
  if (!ts) return '—';
  const d = new Date(ts);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}
