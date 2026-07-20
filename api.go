package main

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/amit/netpulse/internal/diagnosis"
	"github.com/amit/netpulse/internal/export"
	"github.com/amit/netpulse/internal/storage"
)

// --- Frontend API Methods ---
// These are automatically available to the React frontend via Wails bindings.

// StatusResponse is the current network status for the dashboard.
type StatusResponse struct {
	Status      string              `json:"status"`       // "healthy", "warning", "critical"
	Category    string              `json:"category"`     // diagnosis category
	Title       string              `json:"title"`        // short description
	Description string              `json:"description"`  // full explanation
	Evidence    []storage.Evidence  `json:"evidence"`     // supporting data
	Confidence  float64             `json:"confidence"`   // 0-1
	Timestamp   string              `json:"timestamp"`
}

// ProbeResultResponse is a probe result for the frontend.
type ProbeResultResponse struct {
	Timestamp  string  `json:"timestamp"`
	ProbeType  string  `json:"probe_type"`
	Target     string  `json:"target"`
	Success    bool    `json:"success"`
	LatencyMs  float64 `json:"latency_ms"`
	JitterMs   float64 `json:"jitter_ms"`
	PacketLoss float64 `json:"packet_loss"`
}

// SpeedTestResponse is a speed test result for the frontend.
type SpeedTestResponse struct {
	Timestamp    string  `json:"timestamp"`
	DownloadMbps float64 `json:"download_mbps"`
	UploadMbps   float64 `json:"upload_mbps"`
	LatencyMs    float64 `json:"latency_ms"`
	JitterMs     float64 `json:"jitter_ms"`
	Server       string  `json:"server"`
	TriggeredBy  string  `json:"triggered_by"`
}

// DiagnosisResponse is a diagnosis record for the frontend.
type DiagnosisResponse struct {
	ID          int64              `json:"id"`
	Timestamp   string             `json:"timestamp"`
	Category    string             `json:"category"`
	Severity    string             `json:"severity"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Evidence    []storage.Evidence `json:"evidence"`
	Resolved    bool               `json:"resolved"`
	ResolvedAt  string             `json:"resolved_at,omitempty"`
}

// GetCurrentStatus returns the current network health verdict.
func (a *App) GetCurrentStatus() StatusResponse {
	v := a.diagMonitor.LastVerdict()
	if v == nil {
		return StatusResponse{
			Status:    "unknown",
			Category:  "unknown",
			Title:     "Initializing...",
			Description: "Waiting for probe data to accumulate.",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	status := "healthy"
	switch v.Severity {
	case diagnosis.SeverityWarning:
		status = "warning"
	case diagnosis.SeverityCritical:
		status = "critical"
	}

	return StatusResponse{
		Status:      status,
		Category:    string(v.Category),
		Title:       v.Title,
		Description: v.Description,
		Evidence:    v.Evidence,
		Confidence:  v.Confidence,
		Timestamp:   v.Timestamp.Format(time.RFC3339),
	}
}

// GetRecentProbes returns probe results from the last N minutes.
func (a *App) GetRecentProbes(minutes int) []ProbeResultResponse {
	if minutes <= 0 {
		minutes = 5
	}
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	results, err := a.db.GetProbeResultsSince(since)
	if err != nil {
		return nil
	}

	var resp []ProbeResultResponse
	for _, r := range results {
		resp = append(resp, ProbeResultResponse{
			Timestamp:  r.Timestamp.Format(time.RFC3339),
			ProbeType:  r.ProbeType,
			Target:     r.Target,
			Success:    r.Success,
			LatencyMs:  r.LatencyMs,
			JitterMs:   r.JitterMs,
			PacketLoss: r.PacketLoss,
		})
	}
	return resp
}

// GetDiagnosisHistory returns recent diagnoses.
func (a *App) GetDiagnosisHistory(limit int) []DiagnosisResponse {
	if limit <= 0 {
		limit = 50
	}

	diagnoses, err := a.db.GetDiagnosisHistory(limit)
	if err != nil {
		return nil
	}

	var resp []DiagnosisResponse
	for i, d := range diagnoses {
		if i >= limit {
			break
		}
		dr := DiagnosisResponse{
			ID:          d.ID,
			Timestamp:   d.Timestamp.Format(time.RFC3339),
			Category:    d.Category,
			Severity:    d.Severity,
			Title:       d.Title,
			Description: d.Description,
			Evidence:    d.Evidence,
			Resolved:    d.Resolved,
		}
		if d.ResolvedAt != nil {
			dr.ResolvedAt = d.ResolvedAt.Format(time.RFC3339)
		}
		resp = append(resp, dr)
	}
	return resp
}

// GetSpeedTestResults returns recent speed test results.
func (a *App) GetSpeedTestResults(limit int) []SpeedTestResponse {
	if limit <= 0 {
		limit = 20
	}

	// Query from DB
	tests, err := a.db.GetRecentSpeedTests(limit)
	if err != nil {
		return nil
	}

	var resp []SpeedTestResponse
	for _, t := range tests {
		resp = append(resp, SpeedTestResponse{
			Timestamp:    t.Timestamp.Format(time.RFC3339),
			DownloadMbps: t.DownloadMbps,
			UploadMbps:   t.UploadMbps,
			LatencyMs:    t.LatencyMs,
			JitterMs:     t.JitterMs,
			Server:       t.Server,
			TriggeredBy:  t.TriggeredBy,
		})
	}
	return resp
}

// RunSpeedTest triggers an on-demand speed test and returns the result.
func (a *App) RunSpeedTest() (*SpeedTestResponse, error) {
	if a.speedRunner == nil {
		return nil, nil
	}

	// Pause diagnosis during speed test
	a.diagMonitor.SetSpeedTesting(true)
	result, err := a.speedRunner.RunNow(context.Background())
	a.diagMonitor.SetSpeedTesting(false)

	if err != nil {
		return nil, err
	}

	// Store it
	st := &storage.SpeedTestResult{
		Timestamp:    result.Timestamp,
		DownloadMbps: result.DownloadMbps,
		UploadMbps:   result.UploadMbps,
		LatencyMs:    result.LatencyMs,
		JitterMs:     result.JitterMs,
		Server:       result.Server,
		TriggeredBy:  "manual",
	}
	a.db.InsertSpeedTest(st)

	return &SpeedTestResponse{
		Timestamp:    result.Timestamp.Format(time.RFC3339),
		DownloadMbps: result.DownloadMbps,
		UploadMbps:   result.UploadMbps,
		LatencyMs:    result.LatencyMs,
		JitterMs:     result.JitterMs,
		Server:       result.Server,
		TriggeredBy:  "manual",
	}, nil
}

// GetBaselines returns stored baseline data.
func (a *App) GetBaselines() []storage.Baseline {
	baselines, err := a.db.GetAllBaselines()
	if err != nil {
		return nil
	}
	return baselines
}

// UptimeStats holds uptime percentages for different windows.
type UptimeStats struct {
	OneHour      float64 `json:"one_hour"`
	TwentyFourH  float64 `json:"twenty_four_h"`
	SevenDays    float64 `json:"seven_days"`
}

// WifiInfo holds current Wi-Fi details for the dashboard.
type WifiInfo struct {
	Interface     string  `json:"interface"`
	SSID          string  `json:"ssid"`
	BSSID         string  `json:"bssid"`
	FrequencyMHz  int     `json:"frequency_mhz"`
	Channel       int     `json:"channel"`
	SignalDBm     int     `json:"signal_dbm"`
	NoiseDBm      int     `json:"noise_dbm"`
	LinkSpeedMbps float64 `json:"link_speed_mbps"`
	Band          string  `json:"band"`
	SignalQuality string  `json:"signal_quality"`
}

// GetUptimeStats returns uptime percentages.
func (a *App) GetUptimeStats() UptimeStats {
	stats, err := a.db.GetUptimeStats()
	if err != nil {
		return UptimeStats{OneHour: 100, TwentyFourH: 100, SevenDays: 100}
	}
	return UptimeStats{
		OneHour:     stats["1h"],
		TwentyFourH: stats["24h"],
		SevenDays:   stats["7d"],
	}
}

// GetWifiInfo returns current Wi-Fi connection details.
func (a *App) GetWifiInfo() *WifiInfo {
	snap, err := a.db.GetLatestWifiSnapshot()
	if err != nil || snap == nil {
		return nil
	}

	band := "2.4 GHz"
	if snap.FrequencyMHz >= 5000 {
		band = "5 GHz"
	} else if snap.FrequencyMHz >= 5925 {
		band = "6 GHz"
	}

	quality := "Excellent"
	if snap.SignalDBm <= -80 {
		quality = "Very Poor"
	} else if snap.SignalDBm <= -70 {
		quality = "Poor"
	} else if snap.SignalDBm <= -60 {
		quality = "Fair"
	} else if snap.SignalDBm <= -50 {
		quality = "Good"
	}

	return &WifiInfo{
		Interface:     snap.Interface,
		SSID:          snap.SSID,
		BSSID:         snap.BSSID,
		FrequencyMHz:  snap.FrequencyMHz,
		Channel:       snap.Channel,
		SignalDBm:     snap.SignalDBm,
		NoiseDBm:      snap.NoiseDBm,
		LinkSpeedMbps: snap.LinkSpeedMbps,
		Band:          band,
		SignalQuality: quality,
	}
}

// ExportData returns all probe data as JSON for the given time range (minutes back).
func (a *App) ExportData(minutes int) string {
	if minutes <= 0 {
		minutes = 60
	}
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	results, err := a.db.GetProbeResultsSince(since)
	if err != nil {
		return "[]"
	}

	var buf bytes.Buffer
	export.ExportProbeResults(&buf, results, export.FormatJSON)
	return buf.String()
}

// ExportDiagnosesData returns diagnosis history as JSON.
func (a *App) ExportDiagnosesData(limit int) string {
	if limit <= 0 {
		limit = 100
	}
	diagnoses, err := a.db.GetDiagnosisHistory(limit)
	if err != nil {
		return "[]"
	}

	var buf bytes.Buffer
	export.ExportDiagnoses(&buf, diagnoses, export.FormatJSON)
	return buf.String()
}

// ExportFullReport returns a comprehensive JSON report of all data.
func (a *App) ExportFullReport() string {
	type FullReport struct {
		ExportedAt string                   `json:"exported_at"`
		Probes     []storage.ProbeResult    `json:"probes"`
		Diagnoses  []storage.Diagnosis      `json:"diagnoses"`
		SpeedTests []storage.SpeedTestResult `json:"speed_tests"`
		Baselines  []storage.Baseline       `json:"baselines"`
	}

	since := time.Now().AddDate(0, 0, -7) // last 7 days
	probes, _ := a.db.GetProbeResultsSince(since)
	diagnoses, _ := a.db.GetDiagnosisHistory(500)
	speedTests, _ := a.db.GetRecentSpeedTests(100)
	baselines, _ := a.db.GetAllBaselines()

	report := FullReport{
		ExportedAt: time.Now().Format(time.RFC3339),
		Probes:     probes,
		Diagnoses:  diagnoses,
		SpeedTests: speedTests,
		Baselines:  baselines,
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.Encode(report)
	return buf.String()
}

// ClearAllData wipes all stored data for a fresh start.
func (a *App) ClearAllData() error {
	return a.db.ClearAllData()
}

// ExportAndClear exports a full report then clears all data.
func (a *App) ExportAndClear() (string, error) {
	report := a.ExportFullReport()
	err := a.db.ClearAllData()
	if err != nil {
		return report, err
	}
	return report, nil
}
