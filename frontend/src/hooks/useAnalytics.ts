import { useState, useEffect, useCallback } from 'react';

// Generic hook for analytics endpoints
function useAnalyticsData<T>(method: string, args: any[], pollInterval = 10000) {
  const [data, setData] = useState<T | null>(null);

  const refresh = useCallback(async () => {
    try {
      if (window.go?.main?.App?.[method]) {
        const result = await window.go.main.App[method](...args);
        setData(result);
      }
    } catch (err) {
      console.error(`Failed to fetch ${method}:`, err);
    }
  }, [method, ...args]);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, pollInterval);
    return () => clearInterval(interval);
  }, [refresh, pollInterval]);

  return { data, refresh };
}

export interface WifiTimeseriesPoint {
  timestamp: string;
  signal_dbm: number;
  channel: number;
  frequency_mhz: number;
  link_speed_mbps: number;
  band: string;
}

export interface TimeseriesPoint {
  timestamp: string;
  value: number;
  label: string;
}

export interface ProbeLatencyPoint {
  timestamp: string;
  gateway_ms: number;
  external_ms: number;
  dns_ms: number;
  tcp_ms: number;
  gateway_loss: number;
  external_loss: number;
}

export interface HeatmapCell {
  hour: number;
  probe_type: string;
  success_rate: number;
  avg_latency: number;
  sample_count: number;
}

export interface DiagnosisPeriod {
  start: string;
  end: string;
  category: string;
  severity: string;
  title: string;
}

export function useWifiSignalHistory(minutes = 30) {
  return useAnalyticsData<WifiTimeseriesPoint[]>('GetWifiSignalHistory', [minutes], 5000);
}

export function usePacketLossHistory(minutes = 30) {
  return useAnalyticsData<TimeseriesPoint[]>('GetPacketLossHistory', [minutes], 5000);
}

export function useGatewayVsExternal(minutes = 30) {
  return useAnalyticsData<ProbeLatencyPoint[]>('GetGatewayVsExternal', [minutes], 5000);
}

export function useDNSResolverComparison(minutes = 30) {
  return useAnalyticsData<TimeseriesPoint[]>('GetDNSResolverComparison', [minutes], 10000);
}

export function useJitterHistory(minutes = 30) {
  return useAnalyticsData<TimeseriesPoint[]>('GetJitterHistory', [minutes], 5000);
}

export function useHeatmap(days = 7) {
  return useAnalyticsData<HeatmapCell[]>('GetHeatmap', [days], 60000);
}

export function useDiagnosisTimeline(hours = 24) {
  return useAnalyticsData<DiagnosisPeriod[]>('GetDiagnosisTimeline', [hours], 10000);
}
