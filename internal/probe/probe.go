package probe

import (
	"context"
	"time"
)

// Result holds the outcome of a single probe execution.
type Result struct {
	Type       string
	Target     string
	Success    bool
	LatencyMs  float64
	JitterMs   float64
	PacketLoss float64
	Extra      map[string]interface{}
	Timestamp  time.Time
	Error      error
}

// Probe is the interface all probe types implement.
type Probe interface {
	// Type returns the probe type identifier (e.g., "ping", "dns", "gateway").
	Type() string
	// Execute runs the probe and returns the result.
	Execute(ctx context.Context) Result
}
