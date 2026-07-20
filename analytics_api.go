package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/amit/netpulse/internal/scanner"
	"github.com/amit/netpulse/internal/storage"
)

// --- Analytics API Methods ---

// GetWifiSignalHistory returns Wi-Fi signal strength over time.
func (a *App) GetWifiSignalHistory(minutes int) []storage.WifiTimeseriesPoint {
	if minutes <= 0 {
		minutes = 30
	}
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	points, err := a.db.GetWifiTimeseries(since)
	if err != nil {
		return nil
	}
	return points
}

// GetPacketLossHistory returns packet loss data over time.
func (a *App) GetPacketLossHistory(minutes int) []storage.TimeseriesPoint {
	if minutes <= 0 {
		minutes = 30
	}
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	points, err := a.db.GetPacketLossTimeseries(since)
	if err != nil {
		return nil
	}
	return points
}

// GetGatewayVsExternal returns gateway vs external latency comparison data.
func (a *App) GetGatewayVsExternal(minutes int) []storage.ProbeLatencyPoint {
	if minutes <= 0 {
		minutes = 30
	}
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	points, err := a.db.GetGatewayVsExternalLatency(since)
	if err != nil {
		return nil
	}
	return points
}

// GetDNSResolverComparison returns DNS latency by resolver/domain.
func (a *App) GetDNSResolverComparison(minutes int) []storage.TimeseriesPoint {
	if minutes <= 0 {
		minutes = 30
	}
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	points, err := a.db.GetDNSByResolver(since)
	if err != nil {
		return nil
	}
	return points
}

// GetJitterHistory returns jitter data over time.
func (a *App) GetJitterHistory(minutes int) []storage.TimeseriesPoint {
	if minutes <= 0 {
		minutes = 30
	}
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	points, err := a.db.GetJitterTimeseries(since)
	if err != nil {
		return nil
	}
	return points
}

// GetHeatmap returns probe success rate by hour of day.
func (a *App) GetHeatmap(days int) []storage.HeatmapCell {
	if days <= 0 {
		days = 7
	}
	cells, err := a.db.GetHeatmapData(days)
	if err != nil {
		return nil
	}
	return cells
}

// GetDiagnosisTimeline returns diagnosis periods for a timeline chart.
func (a *App) GetDiagnosisTimeline(hours int) []storage.DiagnosisPeriod {
	if hours <= 0 {
		hours = 24
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	periods, err := a.db.GetDiagnosisTimeline(since)
	if err != nil {
		return nil
	}
	return periods
}

// GetNetworkEvents returns recent network change events.
func (a *App) GetNetworkEvents(hours int) []storage.NetworkEvent {
	if hours <= 0 {
		hours = 24
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	events, err := a.db.GetNetworkEvents(since)
	if err != nil {
		return nil
	}
	return events
}

// GetCurrentNetwork returns the currently active network info.
func (a *App) GetCurrentNetwork() map[string]string {
	if a.networkWatcher == nil {
		return nil
	}
	info := a.networkWatcher.Current()
	if info == nil {
		return map[string]string{"status": "disconnected"}
	}
	return map[string]string{
		"status":    "connected",
		"ssid":      info.SSID,
		"type":      info.Type,
		"interface": info.Interface,
		"gateway":   info.Gateway,
	}
}

// AlertRules represents configurable alert thresholds.
type AlertRules struct {
	LatencyWarningMs      float64 `json:"latency_warning_ms"`
	LatencyCriticalMs     float64 `json:"latency_critical_ms"`
	LossWarningPct        float64 `json:"loss_warning_pct"`
	LossCriticalPct       float64 `json:"loss_critical_pct"`
	JitterWarningMs       float64 `json:"jitter_warning_ms"`
	WifiSignalWarning     int     `json:"wifi_signal_warning"`
	WifiSignalCritical    int     `json:"wifi_signal_critical"`
	NotificationsEnabled  bool    `json:"notifications_enabled"`
	NotificationCooldownSec int   `json:"notification_cooldown_sec"`
}

// GetAlertRules returns current alert threshold configuration.
func (a *App) GetAlertRules() *AlertRules {
	t := a.diagEngine.GetThresholds()
	return &AlertRules{
		LatencyWarningMs:      t.LatencyWarningMs,
		LatencyCriticalMs:     t.LatencyCriticalMs,
		LossWarningPct:        t.LossWarning * 100,
		LossCriticalPct:       t.LossCritical * 100,
		JitterWarningMs:       t.JitterWarningMs,
		WifiSignalWarning:     t.WifiSignalWarning,
		WifiSignalCritical:    t.WifiSignalCritical,
		NotificationsEnabled:  a.cfg.NotificationsEnabled,
		NotificationCooldownSec: 60,
	}
}

// SetAlertRules updates the alert thresholds.
func (a *App) SetAlertRules(rules AlertRules) {
	t := a.diagEngine.GetThresholds()
	t.LatencyWarningMs = rules.LatencyWarningMs
	t.LatencyCriticalMs = rules.LatencyCriticalMs
	t.LossWarning = rules.LossWarningPct / 100
	t.LossCritical = rules.LossCriticalPct / 100
	t.JitterWarningMs = rules.JitterWarningMs
	t.WifiSignalWarning = rules.WifiSignalWarning
	t.WifiSignalCritical = rules.WifiSignalCritical
	a.diagEngine.SetThresholds(t)

	a.cfg.NotificationsEnabled = rules.NotificationsEnabled
}

// ProjectReportData holds aggregated data for the project report section.
type ProjectReportData struct {
	TotalProbes        int     `json:"totalProbes"`
	TotalDiagnoses     int     `json:"totalDiagnoses"`
	TotalSpeedTests    int     `json:"totalSpeedTests"`
	TotalWifiSnapshots int     `json:"totalWifiSnapshots"`
	TotalNetworkEvents int     `json:"totalNetworkEvents"`
	UptimeHour         float64 `json:"uptimeHour"`
	Uptime24h          float64 `json:"uptime24h"`
	Uptime7d           float64 `json:"uptime7d"`
	AvgLatency         float64 `json:"avgLatency"`
	AvgGatewayLatency  float64 `json:"avgGatewayLatency"`
	AvgDnsLatency      float64 `json:"avgDnsLatency"`
	AvgDownload        float64 `json:"avgDownload"`
	AvgUpload          float64 `json:"avgUpload"`
	AvgSignal          int     `json:"avgSignal"`
	BandHops           int     `json:"bandHops"`
	CurrentSSID        string  `json:"currentSSID"`
	CurrentBand        string  `json:"currentBand"`
	CurrentChannel     int     `json:"currentChannel"`
	MonitoringSince    string  `json:"monitoringSince"`
}

// GetProjectReport returns aggregated data for the project report.
func (a *App) GetProjectReport() *ProjectReportData {
	r := &ProjectReportData{}

	// Counts
	a.db.QueryRow("SELECT COUNT(*) FROM probe_results").Scan(&r.TotalProbes)
	a.db.QueryRow("SELECT COUNT(*) FROM diagnoses").Scan(&r.TotalDiagnoses)
	a.db.QueryRow("SELECT COUNT(*) FROM speed_tests").Scan(&r.TotalSpeedTests)
	a.db.QueryRow("SELECT COUNT(*) FROM wifi_snapshots").Scan(&r.TotalWifiSnapshots)
	a.db.QueryRow("SELECT COUNT(*) FROM network_events").Scan(&r.TotalNetworkEvents)

	// Uptime
	uptimeStats, _ := a.db.GetUptimeStats()
	r.UptimeHour = uptimeStats["1h"]
	r.Uptime24h = uptimeStats["24h"]
	r.Uptime7d = uptimeStats["7d"]

	// Average latencies
	a.db.QueryRow("SELECT COALESCE(AVG(latency_ms),0) FROM probe_results WHERE probe_type='ping' AND success=1").Scan(&r.AvgLatency)
	a.db.QueryRow("SELECT COALESCE(AVG(latency_ms),0) FROM probe_results WHERE probe_type='gateway' AND success=1").Scan(&r.AvgGatewayLatency)
	a.db.QueryRow("SELECT COALESCE(AVG(latency_ms),0) FROM probe_results WHERE probe_type='dns' AND success=1").Scan(&r.AvgDnsLatency)

	// Speed tests
	a.db.QueryRow("SELECT COALESCE(AVG(download_mbps),0), COALESCE(AVG(upload_mbps),0) FROM speed_tests").Scan(&r.AvgDownload, &r.AvgUpload)

	// Wi-Fi signal
	var avgSignalFloat float64
	a.db.QueryRow("SELECT COALESCE(AVG(signal_dbm),0) FROM wifi_snapshots WHERE signal_dbm != 0").Scan(&avgSignalFloat)
	r.AvgSignal = int(avgSignalFloat)

	// Band hops (count channel changes in wifi_snapshots)
	a.db.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT channel, LAG(channel) OVER (ORDER BY timestamp) as prev_channel
			FROM wifi_snapshots WHERE channel > 0
		) WHERE channel != prev_channel AND prev_channel IS NOT NULL`).Scan(&r.BandHops)

	// Current network
	if a.networkWatcher != nil {
		if info := a.networkWatcher.Current(); info != nil {
			r.CurrentSSID = info.SSID
		}
	}
	if snap, err := a.db.GetLatestWifiSnapshot(); err == nil && snap != nil {
		r.CurrentChannel = snap.Channel
		if snap.FrequencyMHz >= 5000 {
			r.CurrentBand = "5 GHz"
		} else {
			r.CurrentBand = "2.4 GHz"
		}
	}

	// Monitoring since
	var earliest string
	a.db.QueryRow("SELECT MIN(timestamp) FROM probe_results").Scan(&earliest)
	r.MonitoringSince = earliest

	return r
}

// AskAI sends a question to the local Ollama instance with network context.
func (a *App) AskAI(question string) (string, error) {
	// Quick check if Ollama is reachable
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return "", fmt.Errorf("Ollama is not running. Start it with: ollama serve")
	}
	resp.Body.Close()

	// Gather current context for the AI
	context := a.buildAIContext()

	systemPrompt := `You are a network engineering assistant embedded in NetPulse, a network health monitoring app. 
You have access to real-time network metrics from the user's machine. 
Explain network concepts clearly for an ECE student. 
Use the provided metrics to give specific, actionable answers.
Keep responses concise but technically accurate.
Use bullet points for clarity when appropriate.`

	userMessage := context + "\n\nUser question: " + question

	return queryOllama(systemPrompt, userMessage)
}

func (a *App) buildAIContext() string {
	ctx := "Current Network State:\n"

	// Current diagnosis
	v := a.diagMonitor.LastVerdict()
	if v != nil {
		ctx += fmt.Sprintf("- Diagnosis: %s (%s) — %s\n", v.Title, v.Category, v.Severity)
		if len(v.Evidence) > 0 {
			for _, e := range v.Evidence {
				ctx += fmt.Sprintf("  Evidence: %s = %s\n", e.Description, e.Value)
			}
		}
	}

	// Wi-Fi
	if wSnap, err := a.db.GetLatestWifiSnapshot(); err == nil && wSnap != nil {
		band := "2.4 GHz"
		if wSnap.FrequencyMHz >= 5000 {
			band = "5 GHz"
		}
		ctx += fmt.Sprintf("- Wi-Fi: SSID=%s, %s Ch%d, Signal %d dBm, Link %.0f Mbps\n",
			wSnap.SSID, band, wSnap.Channel, wSnap.SignalDBm, wSnap.LinkSpeedMbps)
	}

	// Uptime stats
	if uptimeStats, err := a.db.GetUptimeStats(); err == nil {
		ctx += fmt.Sprintf("- Uptime: 1h=%.1f%%, 24h=%.1f%%, 7d=%.1f%%\n",
			uptimeStats["1h"], uptimeStats["24h"], uptimeStats["7d"])
	}

	// Recent latency (last 1 min)
	since := time.Now().Add(-1 * time.Minute)
	if results, err := a.db.GetProbeResultsSince(since); err == nil && len(results) > 0 {
		var gatewayLat, extLat, dnsLat, jitter float64
		var gCount, eCount, dCount, jCount int
		for _, r := range results {
			if r.Success {
				switch r.ProbeType {
				case "gateway":
					gatewayLat += r.LatencyMs
					gCount++
				case "ping":
					extLat += r.LatencyMs
					eCount++
					if r.JitterMs > 0 {
						jitter += r.JitterMs
						jCount++
					}
				case "dns":
					dnsLat += r.LatencyMs
					dCount++
				}
			}
		}
		if gCount > 0 {
			ctx += fmt.Sprintf("- Gateway latency (last 1min): %.0fms avg\n", gatewayLat/float64(gCount))
		}
		if eCount > 0 {
			ctx += fmt.Sprintf("- External latency (last 1min): %.0fms avg\n", extLat/float64(eCount))
		}
		if dCount > 0 {
			ctx += fmt.Sprintf("- DNS latency (last 1min): %.0fms avg\n", dnsLat/float64(dCount))
		}
		if jCount > 0 {
			ctx += fmt.Sprintf("- Jitter (last 1min): %.0fms avg\n", jitter/float64(jCount))
		}
	}

	// Speed test
	if tests, err := a.db.GetRecentSpeedTests(1); err == nil && len(tests) > 0 {
		ctx += fmt.Sprintf("- Last speed test: ↓%.1f Mbps ↑%.1f Mbps (latency: %.0fms)\n",
			tests[0].DownloadMbps, tests[0].UploadMbps, tests[0].LatencyMs)
	}

	// Recent diagnoses (last 24h)
	if diagnoses, err := a.db.GetDiagnosisHistory(5); err == nil && len(diagnoses) > 0 {
		ctx += fmt.Sprintf("- Recent issues (last 5 diagnoses):\n")
		for _, d := range diagnoses {
			resolved := "active"
			if d.Resolved {
				resolved = "resolved"
			}
			ctx += fmt.Sprintf("  [%s] %s — %s (%s)\n", d.Severity, d.Title, d.Category, resolved)
		}
	}

	// Band hops
	var bandHops int
	a.db.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT channel, LAG(channel) OVER (ORDER BY timestamp) as prev_channel
			FROM wifi_snapshots WHERE channel > 0 AND timestamp > datetime('now', '-1 hours')
		) WHERE channel != prev_channel AND prev_channel IS NOT NULL`).Scan(&bandHops)
	if bandHops > 0 {
		ctx += fmt.Sprintf("- Band hops in last hour: %d\n", bandHops)
	}

	// Network events
	evSince := time.Now().Add(-1 * time.Hour)
	if events, err := a.db.GetNetworkEvents(evSince); err == nil && len(events) > 0 {
		ctx += fmt.Sprintf("- Network changes in last hour: %d\n", len(events))
		ctx += fmt.Sprintf("  Last change: %s → %s (%s)\n", events[0].PrevSSID, events[0].CurrSSID, events[0].Reason)
	}

	// Current network
	if a.networkWatcher != nil {
		if info := a.networkWatcher.Current(); info != nil {
			ctx += fmt.Sprintf("- Current connection: %s (%s) via %s, gateway %s\n",
				info.SSID, info.Type, info.Interface, info.Gateway)
		}
	}

	return ctx
}

// ScanNetwork performs a LAN scan and returns discovered devices.
func (a *App) ScanNetwork() []*scanner.Device {
	if a.netScanner == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	devices, err := a.netScanner.Scan(ctx)
	if err != nil {
		return nil
	}
	return devices
}

// GetNetworkDevices returns previously discovered devices without rescanning.
func (a *App) GetNetworkDevices() []*scanner.Device {
	if a.netScanner == nil {
		return nil
	}
	return a.netScanner.GetDevices()
}
