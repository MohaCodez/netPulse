package storage

import (
	"fmt"
	"sort"
	"time"
)

// Baseline represents aggregated stats for comparison.
type Baseline struct {
	ID           int64     `json:"id"`
	ProbeType    string    `json:"probe_type"`
	Target       string    `json:"target"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	P50Latency   float64   `json:"p50_latency_ms"`
	P95Latency   float64   `json:"p95_latency_ms"`
	AvgLatency   float64   `json:"avg_latency_ms"`
	PacketLoss   float64   `json:"packet_loss_rate"`
	SampleCount  int       `json:"sample_count"`
}

// ComputeBaseline calculates p50/p95/avg latency and loss rate from recent probe results.
func (db *DB) ComputeBaseline(probeType, target string, windowDays int) (*Baseline, error) {
	since := time.Now().AddDate(0, 0, -windowDays)

	rows, err := db.conn.Query(`
		SELECT latency_ms, packet_loss
		FROM probe_results
		WHERE probe_type = ? AND target = ? AND success = 1 AND timestamp >= ?
		ORDER BY latency_ms ASC`,
		probeType, target, since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var latencies []float64
	var totalLoss float64
	var count int

	for rows.Next() {
		var lat, loss float64
		if err := rows.Scan(&lat, &loss); err != nil {
			return nil, err
		}
		latencies = append(latencies, lat)
		totalLoss += loss
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, fmt.Errorf("no data for baseline: %s/%s", probeType, target)
	}

	sort.Float64s(latencies)

	baseline := &Baseline{
		ProbeType:   probeType,
		Target:      target,
		PeriodStart: since,
		PeriodEnd:   time.Now(),
		P50Latency:  percentile(latencies, 0.50),
		P95Latency:  percentile(latencies, 0.95),
		AvgLatency:  avg(latencies),
		PacketLoss:  totalLoss / float64(count),
		SampleCount: count,
	}

	return baseline, nil
}

// StoreBaseline persists a computed baseline.
func (db *DB) StoreBaseline(b *Baseline) error {
	_, err := db.conn.Exec(`
		INSERT OR REPLACE INTO baselines (probe_type, target, period_start, period_end, p50_latency_ms, p95_latency_ms, avg_latency_ms, packet_loss_rate, sample_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ProbeType, b.Target, b.PeriodStart, b.PeriodEnd, b.P50Latency, b.P95Latency, b.AvgLatency, b.PacketLoss, b.SampleCount,
	)
	return err
}

// GetBaseline retrieves the most recent stored baseline for a probe type/target.
func (db *DB) GetBaseline(probeType, target string) (*Baseline, error) {
	row := db.conn.QueryRow(`
		SELECT id, probe_type, target, period_start, period_end, p50_latency_ms, p95_latency_ms, avg_latency_ms, packet_loss_rate, sample_count
		FROM baselines
		WHERE probe_type = ? AND target = ?
		ORDER BY period_end DESC
		LIMIT 1`,
		probeType, target,
	)

	var b Baseline
	err := row.Scan(&b.ID, &b.ProbeType, &b.Target, &b.PeriodStart, &b.PeriodEnd, &b.P50Latency, &b.P95Latency, &b.AvgLatency, &b.PacketLoss, &b.SampleCount)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// GetAllBaselines returns all stored baselines.
func (db *DB) GetAllBaselines() ([]Baseline, error) {
	rows, err := db.conn.Query(`
		SELECT id, probe_type, target, period_start, period_end, p50_latency_ms, p95_latency_ms, avg_latency_ms, packet_loss_rate, sample_count
		FROM baselines
		ORDER BY period_end DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var baselines []Baseline
	for rows.Next() {
		var b Baseline
		if err := rows.Scan(&b.ID, &b.ProbeType, &b.Target, &b.PeriodStart, &b.PeriodEnd, &b.P50Latency, &b.P95Latency, &b.AvgLatency, &b.PacketLoss, &b.SampleCount); err != nil {
			return nil, err
		}
		baselines = append(baselines, b)
	}
	return baselines, rows.Err()
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func avg(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
