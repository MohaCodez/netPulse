package storage

import (
	"fmt"
	"time"
)

// TimeseriesPoint is a generic time+value point for charts.
type TimeseriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label,omitempty"`
}

// WifiTimeseriesPoint includes band/channel info.
type WifiTimeseriesPoint struct {
	Timestamp     time.Time `json:"timestamp"`
	SignalDBm     int       `json:"signal_dbm"`
	Channel       int       `json:"channel"`
	FrequencyMHz  int       `json:"frequency_mhz"`
	LinkSpeedMbps float64   `json:"link_speed_mbps"`
	Band          string    `json:"band"`
}

// ProbeLatencyPoint holds latency data grouped by type.
type ProbeLatencyPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	GatewayMs      float64   `json:"gateway_ms"`
	ExternalMs     float64   `json:"external_ms"`
	DNSMs          float64   `json:"dns_ms"`
	TCPMs          float64   `json:"tcp_ms"`
	GatewayLoss    float64   `json:"gateway_loss"`
	ExternalLoss   float64   `json:"external_loss"`
}

// HeatmapCell represents one cell in a time-of-day heatmap.
type HeatmapCell struct {
	Hour      int     `json:"hour"`
	ProbeType string  `json:"probe_type"`
	SuccessRate float64 `json:"success_rate"`
	AvgLatency float64 `json:"avg_latency"`
	SampleCount int    `json:"sample_count"`
}

// DiagnosisPeriod represents a time block with a specific status.
type DiagnosisPeriod struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Category string    `json:"category"`
	Severity string    `json:"severity"`
	Title    string    `json:"title"`
}

// GetWifiTimeseries returns wifi signal data over time.
func (db *DB) GetWifiTimeseries(since time.Time) ([]WifiTimeseriesPoint, error) {
	rows, err := db.conn.Query(`
		SELECT timestamp, signal_dbm, channel, frequency_mhz, link_speed_mbps
		FROM wifi_snapshots
		WHERE timestamp >= ?
		ORDER BY timestamp ASC`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []WifiTimeseriesPoint
	for rows.Next() {
		var p WifiTimeseriesPoint
		if err := rows.Scan(&p.Timestamp, &p.SignalDBm, &p.Channel, &p.FrequencyMHz, &p.LinkSpeedMbps); err != nil {
			return nil, err
		}
		if p.FrequencyMHz >= 5925 {
			p.Band = "6 GHz"
		} else if p.FrequencyMHz >= 5000 {
			p.Band = "5 GHz"
		} else {
			p.Band = "2.4 GHz"
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

// GetPacketLossTimeseries returns packet loss data over time.
func (db *DB) GetPacketLossTimeseries(since time.Time) ([]TimeseriesPoint, error) {
	rows, err := db.conn.Query(`
		SELECT timestamp, packet_loss, probe_type
		FROM probe_results
		WHERE timestamp >= ? AND probe_type IN ('ping', 'gateway')
		ORDER BY timestamp ASC`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TimeseriesPoint
	for rows.Next() {
		var p TimeseriesPoint
		if err := rows.Scan(&p.Timestamp, &p.Value, &p.Label); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

// GetGatewayVsExternalLatency returns paired gateway and external latency for comparison.
func (db *DB) GetGatewayVsExternalLatency(since time.Time) ([]ProbeLatencyPoint, error) {
	// Get data in 10-second buckets
	rows, err := db.conn.Query(`
		SELECT 
			strftime('%Y-%m-%dT%H:%M:', timestamp) || (CAST(strftime('%S', timestamp) AS INTEGER) / 10 * 10) as bucket,
			probe_type,
			AVG(latency_ms) as avg_lat,
			AVG(packet_loss) as avg_loss
		FROM probe_results
		WHERE timestamp >= ? AND probe_type IN ('gateway', 'ping', 'dns', 'tcp') AND success = 1
		GROUP BY bucket, probe_type
		ORDER BY bucket ASC`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buckets := make(map[string]*ProbeLatencyPoint)
	var order []string

	for rows.Next() {
		var bucket, probeType string
		var avgLat, avgLoss float64
		if err := rows.Scan(&bucket, &probeType, &avgLat, &avgLoss); err != nil {
			return nil, err
		}

		if _, exists := buckets[bucket]; !exists {
			buckets[bucket] = &ProbeLatencyPoint{}
			order = append(order, bucket)
		}
		p := buckets[bucket]

		switch probeType {
		case "gateway":
			p.GatewayMs = avgLat
			p.GatewayLoss = avgLoss
		case "ping":
			p.ExternalMs = avgLat
			p.ExternalLoss = avgLoss
		case "dns":
			p.DNSMs = avgLat
		case "tcp":
			p.TCPMs = avgLat
		}
	}

	var points []ProbeLatencyPoint
	for _, bucket := range order {
		p := buckets[bucket]
		t, _ := time.Parse("2006-01-02T15:04:5", bucket)
		p.Timestamp = t
		points = append(points, *p)
	}
	return points, rows.Err()
}

// GetDNSByResolver returns DNS latency grouped by resolver.
func (db *DB) GetDNSByResolver(since time.Time) ([]TimeseriesPoint, error) {
	rows, err := db.conn.Query(`
		SELECT timestamp, latency_ms, target
		FROM probe_results
		WHERE timestamp >= ? AND probe_type = 'dns' AND success = 1
		ORDER BY timestamp ASC`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TimeseriesPoint
	for rows.Next() {
		var p TimeseriesPoint
		if err := rows.Scan(&p.Timestamp, &p.Value, &p.Label); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

// GetJitterTimeseries returns jitter data over time.
func (db *DB) GetJitterTimeseries(since time.Time) ([]TimeseriesPoint, error) {
	rows, err := db.conn.Query(`
		SELECT timestamp, jitter_ms, probe_type
		FROM probe_results
		WHERE timestamp >= ? AND probe_type IN ('ping', 'gateway') AND success = 1 AND jitter_ms > 0
		ORDER BY timestamp ASC`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TimeseriesPoint
	for rows.Next() {
		var p TimeseriesPoint
		if err := rows.Scan(&p.Timestamp, &p.Value, &p.Label); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

// GetHeatmapData returns success rate by hour-of-day and probe type.
func (db *DB) GetHeatmapData(days int) ([]HeatmapCell, error) {
	rows, err := db.conn.Query(`
		SELECT 
			CAST(strftime('%H', timestamp) AS INTEGER) as hour,
			probe_type,
			AVG(CASE WHEN success=1 THEN 1.0 ELSE 0.0 END) * 100 as success_rate,
			AVG(CASE WHEN success=1 THEN latency_ms ELSE NULL END) as avg_latency,
			COUNT(*) as samples
		FROM probe_results
		WHERE timestamp >= datetime('now', ?)
		GROUP BY hour, probe_type
		ORDER BY hour, probe_type`,
		fmt.Sprintf("-%d days", days),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cells []HeatmapCell
	for rows.Next() {
		var c HeatmapCell
		var avgLat *float64
		if err := rows.Scan(&c.Hour, &c.ProbeType, &c.SuccessRate, &avgLat, &c.SampleCount); err != nil {
			return nil, err
		}
		if avgLat != nil {
			c.AvgLatency = *avgLat
		}
		cells = append(cells, c)
	}
	return cells, rows.Err()
}

// GetDiagnosisTimeline returns diagnosis periods for a timeline view.
func (db *DB) GetDiagnosisTimeline(since time.Time) ([]DiagnosisPeriod, error) {
	rows, err := db.conn.Query(`
		SELECT timestamp, resolved_at, category, severity, title
		FROM diagnoses
		WHERE timestamp >= ?
		ORDER BY timestamp ASC`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var periods []DiagnosisPeriod
	for rows.Next() {
		var p DiagnosisPeriod
		var resolvedAt *time.Time
		if err := rows.Scan(&p.Start, &resolvedAt, &p.Category, &p.Severity, &p.Title); err != nil {
			return nil, err
		}
		if resolvedAt != nil {
			p.End = *resolvedAt
		} else {
			p.End = time.Now()
		}
		periods = append(periods, p)
	}
	return periods, rows.Err()
}
