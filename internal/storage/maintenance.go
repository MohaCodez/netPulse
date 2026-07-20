package storage

import (
	"fmt"
	"log"
	"time"
)

// Maintenance performs periodic database cleanup tasks.
type Maintenance struct {
	db       *DB
	interval time.Duration
	done     chan struct{}
}

// NewMaintenance creates a maintenance runner.
func NewMaintenance(db *DB, interval time.Duration) *Maintenance {
	return &Maintenance{
		db:       db,
		interval: interval,
		done:     make(chan struct{}),
	}
}

// Start begins periodic maintenance.
func (m *Maintenance) Start() {
	go m.loop()
	log.Printf("[maintenance] started, interval=%s", m.interval)
}

// Stop halts maintenance.
func (m *Maintenance) Stop() {
	close(m.done)
}

func (m *Maintenance) loop() {
	// Run immediately on start
	m.run()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			m.run()
		}
	}
}

func (m *Maintenance) run() {
	// Purge probe_results older than 7 days (keep downsampled in baselines)
	purged, err := m.db.PurgeOldProbeResults(7)
	if err != nil {
		log.Printf("[maintenance] error purging probe results: %v", err)
	} else if purged > 0 {
		log.Printf("[maintenance] purged %d probe results older than 7 days", purged)
	}

	// Purge wifi snapshots older than 7 days
	purgedWifi, err := m.db.PurgeOldWifiSnapshots(7)
	if err != nil {
		log.Printf("[maintenance] error purging wifi snapshots: %v", err)
	} else if purgedWifi > 0 {
		log.Printf("[maintenance] purged %d wifi snapshots older than 7 days", purgedWifi)
	}

	// Recompute baselines
	m.db.RecomputeBaselines()
}

// PurgeOldProbeResults removes probe results older than N days. Returns count deleted.
func (db *DB) PurgeOldProbeResults(days int) (int64, error) {
	result, err := db.conn.Exec(`
		DELETE FROM probe_results WHERE timestamp < datetime('now', ?)`,
		fmt.Sprintf("-%d days", days),
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// PurgeOldWifiSnapshots removes wifi snapshots older than N days.
func (db *DB) PurgeOldWifiSnapshots(days int) (int64, error) {
	result, err := db.conn.Exec(`
		DELETE FROM wifi_snapshots WHERE timestamp < datetime('now', ?)`,
		fmt.Sprintf("-%d days", days),
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// RecomputeBaselines recalculates baselines for all probe type/target combos.
func (db *DB) RecomputeBaselines() {
	rows, err := db.conn.Query(`
		SELECT DISTINCT probe_type, target FROM probe_results
		WHERE probe_type IN ('ping', 'gateway', 'dns', 'tcp') AND success = 1`)
	if err != nil {
		return
	}
	defer rows.Close()

	var pairs []struct{ probeType, target string }
	for rows.Next() {
		var pt, t string
		rows.Scan(&pt, &t)
		pairs = append(pairs, struct{ probeType, target string }{pt, t})
	}

	for _, p := range pairs {
		baseline, err := db.ComputeBaseline(p.probeType, p.target, 7)
		if err != nil {
			continue
		}
		db.StoreBaseline(baseline)
	}
}
