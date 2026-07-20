package storage

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

// DB wraps the SQLite database connection.
type DB struct {
	conn *sql.DB
}

// ProbeResult represents a single probe measurement.
type ProbeResult struct {
	ID         int64     `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	ProbeType  string    `json:"probe_type"`
	Target     string    `json:"target"`
	Success    bool      `json:"success"`
	LatencyMs  float64   `json:"latency_ms"`
	JitterMs   float64   `json:"jitter_ms"`
	PacketLoss float64   `json:"packet_loss"`
	NetworkID  string    `json:"network_id,omitempty"` // SSID or connection name
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// Diagnosis represents a root-cause diagnosis event.
type Diagnosis struct {
	ID          int64     `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Category    string    `json:"category"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Evidence    []Evidence `json:"evidence"`
	Resolved    bool      `json:"resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// Evidence is a single piece of supporting data for a diagnosis.
type Evidence struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Value       string  `json:"value"`
	Timestamp   time.Time `json:"timestamp"`
}

// SpeedTestResult represents a speed test measurement.
type SpeedTestResult struct {
	ID           int64     `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	DownloadMbps float64   `json:"download_mbps"`
	UploadMbps   float64   `json:"upload_mbps"`
	LatencyMs    float64   `json:"latency_ms"`
	JitterMs     float64   `json:"jitter_ms"`
	Server       string    `json:"server"`
	TriggeredBy  string    `json:"triggered_by"`
}

// WifiSnapshot represents a Wi-Fi measurement at a point in time.
type WifiSnapshot struct {
	ID           int64     `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	Interface    string    `json:"interface"`
	SSID         string    `json:"ssid"`
	BSSID        string    `json:"bssid"`
	FrequencyMHz int       `json:"frequency_mhz"`
	Channel      int       `json:"channel"`
	SignalDBm    int       `json:"signal_dbm"`
	NoiseDBm     int       `json:"noise_dbm"`
	LinkSpeedMbps float64  `json:"link_speed_mbps"`
}

// Open opens (or creates) the SQLite database at the given path.
func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Apply schema
	if _, err := conn.Exec(schemaSQL); err != nil {
		conn.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	// Integrity check
	var integrity string
	if err := conn.QueryRow("PRAGMA integrity_check").Scan(&integrity); err == nil && integrity != "ok" {
		conn.Close()
		return nil, fmt.Errorf("database integrity check failed: %s", integrity)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// InsertProbeResult stores a probe measurement.
func (db *DB) InsertProbeResult(r *ProbeResult) error {
	var extraJSON *string
	if r.Extra != nil {
		data, err := json.Marshal(r.Extra)
		if err != nil {
			return err
		}
		s := string(data)
		extraJSON = &s
	}

	success := 0
	if r.Success {
		success = 1
	}

	_, err := db.conn.Exec(`
		INSERT INTO probe_results (timestamp, probe_type, target, success, latency_ms, jitter_ms, packet_loss, extra_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		r.Timestamp, r.ProbeType, r.Target, success, r.LatencyMs, r.JitterMs, r.PacketLoss, extraJSON,
	)
	return err
}

// InsertDiagnosis stores a diagnosis event.
func (db *DB) InsertDiagnosis(d *Diagnosis) (int64, error) {
	evidenceJSON, err := json.Marshal(d.Evidence)
	if err != nil {
		return 0, err
	}

	resolved := 0
	if d.Resolved {
		resolved = 1
	}

	result, err := db.conn.Exec(`
		INSERT INTO diagnoses (timestamp, category, severity, title, description, evidence_json, resolved, resolved_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		d.Timestamp, d.Category, d.Severity, d.Title, d.Description, string(evidenceJSON), resolved, d.ResolvedAt,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// InsertSpeedTest stores a speed test result.
func (db *DB) InsertSpeedTest(s *SpeedTestResult) error {
	_, err := db.conn.Exec(`
		INSERT INTO speed_tests (timestamp, download_mbps, upload_mbps, latency_ms, jitter_ms, server, triggered_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		s.Timestamp, s.DownloadMbps, s.UploadMbps, s.LatencyMs, s.JitterMs, s.Server, s.TriggeredBy,
	)
	return err
}

// InsertWifiSnapshot stores a Wi-Fi measurement.
func (db *DB) InsertWifiSnapshot(w *WifiSnapshot) error {
	_, err := db.conn.Exec(`
		INSERT INTO wifi_snapshots (timestamp, interface, ssid, bssid, frequency_mhz, channel, signal_dbm, noise_dbm, link_speed_mbps)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		w.Timestamp, w.Interface, w.SSID, w.BSSID, w.FrequencyMHz, w.Channel, w.SignalDBm, w.NoiseDBm, w.LinkSpeedMbps,
	)
	return err
}

// GetRecentProbeResults returns probe results from the last N seconds.
func (db *DB) GetRecentProbeResults(probeType string, seconds int) ([]ProbeResult, error) {
	rows, err := db.conn.Query(`
		SELECT id, timestamp, probe_type, target, success, latency_ms, jitter_ms, packet_loss, extra_json
		FROM probe_results
		WHERE probe_type = ? AND timestamp >= datetime('now', ?)
		ORDER BY timestamp DESC`,
		probeType, fmt.Sprintf("-%d seconds", seconds),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProbeResults(rows)
}

// GetProbeResultsSince returns all probe results since a given time.
func (db *DB) GetProbeResultsSince(since time.Time) ([]ProbeResult, error) {
	rows, err := db.conn.Query(`
		SELECT id, timestamp, probe_type, target, success, latency_ms, jitter_ms, packet_loss, extra_json
		FROM probe_results
		WHERE timestamp >= ?
		ORDER BY timestamp DESC`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProbeResults(rows)
}

// GetActiveDiagnoses returns unresolved diagnoses.
func (db *DB) GetActiveDiagnoses() ([]Diagnosis, error) {
	rows, err := db.conn.Query(`
		SELECT id, timestamp, category, severity, title, description, evidence_json, resolved, resolved_at
		FROM diagnoses
		WHERE resolved = 0
		ORDER BY timestamp DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDiagnoses(rows)
}

// ResolveDiagnosis marks a diagnosis as resolved.
func (db *DB) ResolveDiagnosis(id int64) error {
	_, err := db.conn.Exec(`
		UPDATE diagnoses SET resolved = 1, resolved_at = datetime('now')
		WHERE id = ?`, id,
	)
	return err
}

func scanProbeResults(rows *sql.Rows) ([]ProbeResult, error) {
	var results []ProbeResult
	for rows.Next() {
		var r ProbeResult
		var success int
		var extraJSON *string
		if err := rows.Scan(&r.ID, &r.Timestamp, &r.ProbeType, &r.Target, &success, &r.LatencyMs, &r.JitterMs, &r.PacketLoss, &extraJSON); err != nil {
			return nil, err
		}
		r.Success = success == 1
		if extraJSON != nil {
			json.Unmarshal([]byte(*extraJSON), &r.Extra)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func scanDiagnoses(rows *sql.Rows) ([]Diagnosis, error) {
	var results []Diagnosis
	for rows.Next() {
		var d Diagnosis
		var resolved int
		var evidenceJSON string
		if err := rows.Scan(&d.ID, &d.Timestamp, &d.Category, &d.Severity, &d.Title, &d.Description, &evidenceJSON, &resolved, &d.ResolvedAt); err != nil {
			return nil, err
		}
		d.Resolved = resolved == 1
		json.Unmarshal([]byte(evidenceJSON), &d.Evidence)
		results = append(results, d)
	}
	return results, rows.Err()
}

// GetRecentSpeedTests returns the most recent speed test results.
func (db *DB) GetRecentSpeedTests(limit int) ([]SpeedTestResult, error) {
	rows, err := db.conn.Query(`
		SELECT id, timestamp, download_mbps, upload_mbps, latency_ms, jitter_ms, server, triggered_by
		FROM speed_tests
		ORDER BY timestamp DESC
		LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SpeedTestResult
	for rows.Next() {
		var r SpeedTestResult
		if err := rows.Scan(&r.ID, &r.Timestamp, &r.DownloadMbps, &r.UploadMbps, &r.LatencyMs, &r.JitterMs, &r.Server, &r.TriggeredBy); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetDiagnosisHistory returns recent diagnoses (both resolved and unresolved).
func (db *DB) GetDiagnosisHistory(limit int) ([]Diagnosis, error) {
	rows, err := db.conn.Query(`
		SELECT id, timestamp, category, severity, title, description, evidence_json, resolved, resolved_at
		FROM diagnoses
		ORDER BY timestamp DESC
		LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDiagnoses(rows)
}

// GetUptimeStats returns uptime percentage for different time windows.
func (db *DB) GetUptimeStats() (map[string]float64, error) {
	stats := make(map[string]float64)

	windows := map[string]string{
		"1h":  "-1 hours",
		"24h": "-24 hours",
		"7d":  "-7 days",
	}

	for label, offset := range windows {
		var total, success int
		err := db.conn.QueryRow(`
			SELECT COUNT(*), SUM(CASE WHEN success=1 THEN 1 ELSE 0 END)
			FROM probe_results
			WHERE probe_type IN ('ping', 'gateway') AND timestamp > datetime('now', ?)`,
			offset,
		).Scan(&total, &success)
		if err != nil || total == 0 {
			stats[label] = 100.0
		} else {
			stats[label] = float64(success) / float64(total) * 100.0
		}
	}

	return stats, nil
}

// GetLatestWifiSnapshot returns the most recent wifi snapshot.
func (db *DB) GetLatestWifiSnapshot() (*WifiSnapshot, error) {
	row := db.conn.QueryRow(`
		SELECT id, timestamp, interface, ssid, bssid, frequency_mhz, channel, signal_dbm, noise_dbm, link_speed_mbps
		FROM wifi_snapshots
		ORDER BY timestamp DESC
		LIMIT 1`)

	var w WifiSnapshot
	err := row.Scan(&w.ID, &w.Timestamp, &w.Interface, &w.SSID, &w.BSSID, &w.FrequencyMHz, &w.Channel, &w.SignalDBm, &w.NoiseDBm, &w.LinkSpeedMbps)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// ClearAllData deletes all probe results, diagnoses, speed tests, wifi snapshots, and baselines.
func (db *DB) ClearAllData() error {
	tables := []string{"probe_results", "diagnoses", "speed_tests", "wifi_snapshots", "baselines"}
	for _, table := range tables {
		if _, err := db.conn.Exec("DELETE FROM " + table); err != nil {
			return err
		}
	}
	// Reclaim space
	_, err := db.conn.Exec("VACUUM")
	return err
}
