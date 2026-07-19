package probe

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ResultHandler is called with each probe result.
type ResultHandler func(Result)

// Runner manages probe scheduling and execution.
type Runner struct {
	probes   []Probe
	interval time.Duration
	handler  ResultHandler
	mu       sync.RWMutex
	running  bool
	cancel   context.CancelFunc
}

// NewRunner creates a probe runner with the given interval and result handler.
func NewRunner(interval time.Duration, handler ResultHandler) *Runner {
	return &Runner{
		interval: interval,
		handler:  handler,
	}
}

// AddProbe registers a probe to be run on each tick.
func (r *Runner) AddProbe(p Probe) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.probes = append(r.probes, p)
}

// Start begins the probe loop. It runs all probes in parallel on each tick.
func (r *Runner) Start() {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return
	}
	r.running = true
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.mu.Unlock()

	go r.loop(ctx)
	log.Printf("[probe] runner started with %d probes, interval=%s", len(r.probes), r.interval)
}

// Stop halts the probe loop.
func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.running {
		return
	}
	r.running = false
	r.cancel()
	log.Println("[probe] runner stopped")
}

// IsRunning returns whether the runner is active.
func (r *Runner) IsRunning() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.running
}

func (r *Runner) loop(ctx context.Context) {
	// Run immediately on start
	r.runAll(ctx)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.runAll(ctx)
		}
	}
}

func (r *Runner) runAll(ctx context.Context) {
	r.mu.RLock()
	probes := make([]Probe, len(r.probes))
	copy(probes, r.probes)
	r.mu.RUnlock()

	var wg sync.WaitGroup
	for _, p := range probes {
		wg.Add(1)
		go func(probe Probe) {
			defer wg.Done()
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("[probe] panic in %s probe: %v", probe.Type(), rec)
				}
			}()

			result := probe.Execute(ctx)
			if r.handler != nil {
				r.handler(result)
			}
		}(p)
	}
	wg.Wait()
}

// Status returns a summary of the runner state.
func (r *Runner) Status() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	state := "stopped"
	if r.running {
		state = "running"
	}
	return fmt.Sprintf("runner: %s, probes: %d, interval: %s", state, len(r.probes), r.interval)
}
