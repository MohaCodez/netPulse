package main

import (
	"fmt"
	"time"

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

// AskAI sends a question to the local Ollama instance with network context.
func (a *App) AskAI(question string) (string, error) {
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

	// Current status
	v := a.diagMonitor.LastVerdict()
	if v != nil {
		ctx += fmt.Sprintf("- Diagnosis: %s (%s) — %s\n", v.Title, v.Category, v.Severity)
	}

	// Wi-Fi
	if wSnap, err := a.db.GetLatestWifiSnapshot(); err == nil && wSnap != nil {
		band := "2.4 GHz"
		if wSnap.FrequencyMHz >= 5000 {
			band = "5 GHz"
		}
		ctx += fmt.Sprintf("- Wi-Fi: %s, %s Ch%d, Signal %d dBm, Link %0.f Mbps\n",
			wSnap.SSID, band, wSnap.Channel, wSnap.SignalDBm, wSnap.LinkSpeedMbps)
	}

	// Recent latency
	since := time.Now().Add(-1 * time.Minute)
	if results, err := a.db.GetProbeResultsSince(since); err == nil && len(results) > 0 {
		var gatewayLat, extLat float64
		var gCount, eCount int
		for _, r := range results {
			if r.Success {
				if r.ProbeType == "gateway" {
					gatewayLat += r.LatencyMs
					gCount++
				} else if r.ProbeType == "ping" {
					extLat += r.LatencyMs
					eCount++
				}
			}
		}
		if gCount > 0 {
			ctx += fmt.Sprintf("- Gateway latency (last 1min): %.0fms avg\n", gatewayLat/float64(gCount))
		}
		if eCount > 0 {
			ctx += fmt.Sprintf("- External latency (last 1min): %.0fms avg\n", extLat/float64(eCount))
		}
	}

	// Speed test
	if tests, err := a.db.GetRecentSpeedTests(1); err == nil && len(tests) > 0 {
		ctx += fmt.Sprintf("- Last speed test: ↓%.1f Mbps ↑%.1f Mbps\n", tests[0].DownloadMbps, tests[0].UploadMbps)
	}

	// Network events
	evSince := time.Now().Add(-1 * time.Hour)
	if events, err := a.db.GetNetworkEvents(evSince); err == nil && len(events) > 0 {
		ctx += fmt.Sprintf("- Network changes in last hour: %d\n", len(events))
		ctx += fmt.Sprintf("  Last change: %s → %s (%s)\n", events[0].PrevSSID, events[0].CurrSSID, events[0].Reason)
	}

	return ctx
}
