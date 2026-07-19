package main

import (
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
