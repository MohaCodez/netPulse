import { useState } from 'react';
import { StatusCard } from './components/StatusCard';
import { LatencyChart } from './components/LatencyChart';
import { SpeedTest } from './components/SpeedTest';
import { DiagnosisHistory } from './components/DiagnosisHistory';
import { UptimeCard } from './components/UptimeCard';
import { WifiCard } from './components/WifiCard';
import { ExportButton } from './components/ExportButton';
import { WifiSignalChart } from './components/WifiSignalChart';
import { PacketLossChart } from './components/PacketLossChart';
import { GatewayVsExternalChart } from './components/GatewayVsExternalChart';
import { DNSComparisonChart } from './components/DNSComparisonChart';
import { JitterChart } from './components/JitterChart';
import { BandHopChart } from './components/BandHopChart';
import { Heatmap } from './components/Heatmap';
import { DiagnosisTimelineChart } from './components/DiagnosisTimelineChart';
import { SpeedHistoryChart } from './components/SpeedHistoryChart';
import { CorrelationChart } from './components/CorrelationChart';
import { AlertConfig } from './components/AlertConfig';
import { Theory } from './components/Theory';
import './components/Theory.css';
import { AIChat } from './components/AIChat';
import {
  useNetworkStatus,
  useProbeResults,
  useDiagnosisHistory,
  useSpeedTests,
  useUptimeStats,
  useWifiInfo,
} from './hooks/useNetworkData';
import {
  useWifiSignalHistory,
  usePacketLossHistory,
  useGatewayVsExternal,
  useDNSResolverComparison,
  useJitterHistory,
  useHeatmap,
  useDiagnosisTimeline,
} from './hooks/useAnalytics';
import './App.css';

type Tab = 'overview' | 'analytics' | 'theory' | 'ai' | 'settings';

function App() {
  const [tab, setTab] = useState<Tab>('overview');
  const { status, loading } = useNetworkStatus();
  const { probes } = useProbeResults(5);
  const { diagnoses } = useDiagnosisHistory(20);
  const { tests, refresh: refreshTests } = useSpeedTests(10);
  const { stats: uptimeStats } = useUptimeStats();
  const { wifi } = useWifiInfo();

  // Analytics data
  const { data: wifiSignal } = useWifiSignalHistory(30);
  const { data: packetLoss } = usePacketLossHistory(30);
  const { data: gatewayVsExt } = useGatewayVsExternal(30);
  const { data: dnsData } = useDNSResolverComparison(30);
  const { data: jitterData } = useJitterHistory(30);
  const { data: heatmapData } = useHeatmap(7);
  const { data: diagTimeline } = useDiagnosisTimeline(24);

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-left">
          <h1>⚡ NetPulse</h1>
          <span className="app-subtitle">Network Health Monitor</span>
        </div>
        <div className="header-right">
          <div className="tab-switcher">
            <button className={`tab-btn ${tab === 'overview' ? 'active' : ''}`} onClick={() => setTab('overview')}>
              Overview
            </button>
            <button className={`tab-btn ${tab === 'analytics' ? 'active' : ''}`} onClick={() => setTab('analytics')}>
              Analytics
            </button>
            <button className={`tab-btn ${tab === 'theory' ? 'active' : ''}`} onClick={() => setTab('theory')}>
              Theory
            </button>
            <button className={`tab-btn ${tab === 'ai' ? 'active' : ''}`} onClick={() => setTab('ai')}>
              AI
            </button>
            <button className={`tab-btn ${tab === 'settings' ? 'active' : ''}`} onClick={() => setTab('settings')}>
              Settings
            </button>
          </div>
          <div className="live-indicator">
            <span className="live-dot" />
            <span>Live</span>
          </div>
        </div>
      </header>

      <main className="app-main">
        {tab === 'overview' && (
          <>
            <div className="top-row">
              <StatusCard status={status} loading={loading} />
              <UptimeCard stats={uptimeStats} />
            </div>

            <DiagnosisTimelineChart data={diagTimeline} />

            <div className="app-grid">
              <div className="app-grid-main">
                <LatencyChart probes={probes} />
                <GatewayVsExternalChart data={gatewayVsExt} />
                <DiagnosisHistory diagnoses={diagnoses} />
              </div>
              <div className="app-grid-side">
                <WifiCard wifi={wifi} />
                <SpeedTest tests={tests} onTestComplete={refreshTests} />
                <ExportButton />
              </div>
            </div>
          </>
        )}

        {tab === 'analytics' && (
          <div className="analytics-page">
            <div className="analytics-grid">
              <CorrelationChart wifiData={wifiSignal} latencyData={gatewayVsExt} />
              <WifiSignalChart data={wifiSignal} />
              <BandHopChart data={wifiSignal} />
              <SpeedHistoryChart tests={tests} />
              <GatewayVsExternalChart data={gatewayVsExt} />
              <PacketLossChart data={packetLoss} />
              <DNSComparisonChart data={dnsData} />
              <JitterChart data={jitterData} />
              <Heatmap data={heatmapData} />
              <DiagnosisTimelineChart data={diagTimeline} />
            </div>
          </div>
        )}

        {tab === 'settings' && (
          <div className="settings-page">
            <div className="settings-grid">
              <AlertConfig />
              <ExportButton />
            </div>
          </div>
        )}

        {tab === 'theory' && <Theory />}

        {tab === 'ai' && (
          <div className="ai-page">
            <AIChat />
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
