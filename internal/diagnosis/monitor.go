package diagnosis

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/amit/netpulse/internal/storage"
)

// VerdictHandler is called when a new diagnosis verdict is produced.
type VerdictHandler func(*Verdict)

// Monitor continuously evaluates network health and produces diagnoses.
// It runs on a timer, aggregates recent probe data, and runs the engine.
type Monitor struct {
	db          *storage.DB
	engine      *Engine
	aggregator  *Aggregator
	interval    time.Duration
	window      time.Duration
	handler     VerdictHandler

	mu           sync.RWMutex
	lastVerdict  *Verdict
	activeDiagID *int64 // ID of the currently active diagnosis in the DB
	running      bool
	cancel       context.CancelFunc
}

// NewMonitor creates a diagnosis monitor.
func NewMonitor(db *storage.DB, engine *Engine, interval, window time.Duration, handler VerdictHandler) *Monitor {
	return &Monitor{
		db:         db,
		engine:     engine,
		aggregator: NewAggregator(),
		interval:   interval,
		window:     window,
		handler:    handler,
	}
}

// Start begins the diagnosis monitoring loop.
func (m *Monitor) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.mu.Unlock()

	go m.loop(ctx)
	log.Printf("[diagnosis] monitor started, interval=%s window=%s", m.interval, m.window)
}

// Stop halts the diagnosis monitor.
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return
	}
	m.running = false
	m.cancel()
	log.Println("[diagnosis] monitor stopped")
}

// LastVerdict returns the most recent diagnosis verdict.
func (m *Monitor) LastVerdict() *Verdict {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastVerdict
}

func (m *Monitor) loop(ctx context.Context) {
	// Wait a bit for initial data to accumulate
	select {
	case <-time.After(m.window):
	case <-ctx.Done():
		return
	}

	// Run immediately then on interval
	m.evaluate()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.evaluate()
		}
	}
}

func (m *Monitor) evaluate() {
	// Get recent probe results
	since := time.Now().Add(-m.window)
	results, err := m.db.GetProbeResultsSince(since)
	if err != nil {
		log.Printf("[diagnosis] error fetching probe results: %v", err)
		return
	}

	// Build snapshot (no Wi-Fi data yet — will integrate when wifi package is ready)
	snap := m.aggregator.BuildSnapshot(results, nil, int(m.window.Seconds()))

	// Run diagnosis engine
	verdict := m.engine.Evaluate(snap)

	m.mu.Lock()
	prevVerdict := m.lastVerdict
	m.lastVerdict = verdict
	m.mu.Unlock()

	// Determine if state changed
	stateChanged := prevVerdict == nil ||
		prevVerdict.Category != verdict.Category ||
		prevVerdict.Severity != verdict.Severity

	if stateChanged {
		log.Printf("[diagnosis] %s [%s] %s (confidence: %.0f%%)",
			verdict.Severity, verdict.Category, verdict.Title, verdict.Confidence*100)

		// Store diagnosis if it's not healthy
		if verdict.Category != CategoryHealthy {
			m.storeDiagnosis(verdict)
		}

		// If we moved back to healthy, resolve the active diagnosis
		if verdict.Category == CategoryHealthy && m.activeDiagID != nil {
			if err := m.db.ResolveDiagnosis(*m.activeDiagID); err != nil {
				log.Printf("[diagnosis] error resolving diagnosis: %v", err)
			}
			m.mu.Lock()
			m.activeDiagID = nil
			m.mu.Unlock()
		}
	}

	// Notify handler
	if m.handler != nil && stateChanged {
		m.handler(verdict)
	}
}

func (m *Monitor) storeDiagnosis(v *Verdict) {
	d := &storage.Diagnosis{
		Timestamp:   v.Timestamp,
		Category:    string(v.Category),
		Severity:    string(v.Severity),
		Title:       v.Title,
		Description: v.Description,
		Evidence:    v.Evidence,
	}

	id, err := m.db.InsertDiagnosis(d)
	if err != nil {
		log.Printf("[diagnosis] error storing diagnosis: %v", err)
		return
	}

	m.mu.Lock()
	m.activeDiagID = &id
	m.mu.Unlock()
}
