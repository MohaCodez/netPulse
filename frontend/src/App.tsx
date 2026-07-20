import { useState, lazy, Suspense } from 'react';
import { StatusCard } from './components/StatusCard';
import { LatencyChart } from './components/LatencyChart';
import { SpeedTest } from './components/SpeedTest';
import { DiagnosisHistory } from './components/DiagnosisHistory';
import { UptimeCard } from './components/UptimeCard';
import { WifiCard } from './components/WifiCard';
import { ExportButton } from './components/ExportButton';
import { GatewayVsExternalChart } from './components/GatewayVsExternalChart';
import { DiagnosisTimelineChart } from './components/DiagnosisTimelineChart';
import { AlertConfig } from './components/AlertConfig';
import { AIChat } from './components/AIChat';
import { NetworkDevices } from './components/NetworkDevices';
import './components/NetworkDevices.css';

// Lazy-loaded heavy components (only loaded when tab is active)
const WifiSignalChart = lazy(() => import('./components/WifiSignalChart').then(m => ({ default: m.WifiSignalChart })));
const PacketLossChart = lazy(() => import('./components/PacketLossChart').then(m => ({ default: m.PacketLossChart })));
const DNSComparisonChart = lazy(() => import('./components/DNSComparisonChart').then(m => ({ default: m.DNSComparisonChart })));
const JitterChart = lazy(() => import('./components/JitterChart').then(m => ({ default: m.JitterChart })));
const BandHopChart = lazy(() => import('./components/BandHopChart').then(m => ({ default: m.BandHopChart })));
const Heatmap = lazy(() => import('./components/Heatmap').then(m => ({ default: m.Heatmap })));
const SpeedHistoryChart = lazy(() => import('./components/SpeedHistoryChart').then(m => ({ default: m.SpeedHistoryChart })));
const CorrelationChart = lazy(() => import('./components/CorrelationChart').then(m => ({ default: m.CorrelationChart })));
const Theory = lazy(() => import('./components/Theory').then(m => ({ default: m.Theory })));
const ProjectReport = lazy(() => import('./components/ProjectReport').then(m => ({ default: m.ProjectReport })));
import './components/NetworkDevices.css';
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

type Tab = 'overview' | 'analytics' | 'theory' | 'report' | 'ai' | 'settings';

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
            <button className={`tab-btn ${tab === 'report' ? 'active' : ''}`} onClick={() => setTab('report')}>
              Report
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
                <NetworkDevices />
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
          <Suspense fallback={<div className="loading-tab">Loading analytics...</div>}>
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
          </Suspense>
        )}

        {tab === 'settings' && (
          <div className="settings-page">
            <AlertConfig />
            <ExportButton />
          </div>
        )}

        {tab === 'theory' && (
          <Suspense fallback={<div className="loading-tab">Loading...</div>}>
            <Theory />
          </Suspense>
        )}

        {tab === 'report' && (
          <Suspense fallback={<div className="loading-tab">Loading report...</div>}>
            <ProjectReport />
          </Suspense>
        )}

        {/* AI chat is always mounted but hidden when not active — preserves conversation */}
        <div className="ai-page" style={{ display: tab === 'ai' ? 'block' : 'none' }}>
          <AIChat />
        </div>
      </main>
    </div>
  );
}

export default App;
