import { useState, useEffect, useCallback } from 'react';
import { GetCurrentStatus, GetRecentProbes, GetDiagnosisHistory, GetSpeedTestResults } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';

// We'll call these dynamically since Wails will generate them after rebuild
declare global {
  interface Window {
    go: any;
  }
}

export function useNetworkStatus(pollInterval = 3000) {
  const [status, setStatus] = useState<main.StatusResponse | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    try {
      const s = await GetCurrentStatus();
      setStatus(s);
    } catch (err) {
      console.error('Failed to fetch status:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, pollInterval);
    return () => clearInterval(interval);
  }, [refresh, pollInterval]);

  return { status, loading, refresh };
}

export function useProbeResults(minutes = 5, pollInterval = 5000) {
  const [probes, setProbes] = useState<main.ProbeResultResponse[]>([]);

  const refresh = useCallback(async () => {
    try {
      const results = await GetRecentProbes(minutes);
      setProbes(results || []);
    } catch (err) {
      console.error('Failed to fetch probes:', err);
    }
  }, [minutes]);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, pollInterval);
    return () => clearInterval(interval);
  }, [refresh, pollInterval]);

  return { probes, refresh };
}

export function useDiagnosisHistory(limit = 20) {
  const [diagnoses, setDiagnoses] = useState<main.DiagnosisResponse[]>([]);

  const refresh = useCallback(async () => {
    try {
      const results = await GetDiagnosisHistory(limit);
      setDiagnoses(results || []);
    } catch (err) {
      console.error('Failed to fetch diagnoses:', err);
    }
  }, [limit]);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, 10000);
    return () => clearInterval(interval);
  }, [refresh]);

  return { diagnoses, refresh };
}

export function useSpeedTests(limit = 10) {
  const [tests, setTests] = useState<main.SpeedTestResponse[]>([]);

  const refresh = useCallback(async () => {
    try {
      const results = await GetSpeedTestResults(limit);
      setTests(results || []);
    } catch (err) {
      console.error('Failed to fetch speed tests:', err);
    }
  }, [limit]);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, 30000);
    return () => clearInterval(interval);
  }, [refresh]);

  return { tests, refresh };
}

export interface UptimeStats {
  one_hour: number;
  twenty_four_h: number;
  seven_days: number;
}

export interface WifiInfo {
  interface: string;
  ssid: string;
  bssid: string;
  frequency_mhz: number;
  channel: number;
  signal_dbm: number;
  noise_dbm: number;
  link_speed_mbps: number;
  band: string;
  signal_quality: string;
}

export function useUptimeStats(pollInterval = 10000) {
  const [stats, setStats] = useState<UptimeStats | null>(null);

  const refresh = useCallback(async () => {
    try {
      if (window.go?.main?.App?.GetUptimeStats) {
        const s = await window.go.main.App.GetUptimeStats();
        setStats(s);
      }
    } catch (err) {
      console.error('Failed to fetch uptime stats:', err);
    }
  }, []);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, pollInterval);
    return () => clearInterval(interval);
  }, [refresh, pollInterval]);

  return { stats, refresh };
}

export function useWifiInfo(pollInterval = 5000) {
  const [wifi, setWifi] = useState<WifiInfo | null>(null);

  const refresh = useCallback(async () => {
    try {
      if (window.go?.main?.App?.GetWifiInfo) {
        const w = await window.go.main.App.GetWifiInfo();
        setWifi(w);
      }
    } catch (err) {
      console.error('Failed to fetch wifi info:', err);
    }
  }, []);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, pollInterval);
    return () => clearInterval(interval);
  }, [refresh, pollInterval]);

  return { wifi, refresh };
}
