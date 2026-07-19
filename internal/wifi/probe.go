package wifi

import (
	"context"
	"time"

	"github.com/amit/netpulse/internal/probe"
)

// Probe implements the probe.Probe interface for Wi-Fi stats collection.
type Probe struct {
	collector *Collector
	timeout   time.Duration
}

// NewProbe creates a Wi-Fi probe that collects signal stats on each tick.
func NewProbe(iface string, timeout time.Duration) *Probe {
	return &Probe{
		collector: NewCollector(iface),
		timeout:   timeout,
	}
}

func (p *Probe) Type() string { return "wifi" }

func (p *Probe) Execute(ctx context.Context) probe.Result {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	r := probe.Result{
		Type:      "wifi",
		Target:    "local",
		Timestamp: time.Now(),
	}

	stats, err := p.collector.Collect(ctx)
	if err != nil {
		r.Error = err
		r.Success = false
		return r
	}

	if stats == nil {
		// No Wi-Fi interface found — not an error, just not applicable
		r.Success = true
		r.Extra = map[string]interface{}{
			"available": false,
		}
		return r
	}

	r.Success = true
	r.Target = stats.Interface
	r.Extra = map[string]interface{}{
		"available":      true,
		"interface":      stats.Interface,
		"ssid":           stats.SSID,
		"bssid":          stats.BSSID,
		"frequency_mhz":  stats.FrequencyMHz,
		"channel":        stats.Channel,
		"signal_dbm":     stats.SignalDBm,
		"noise_dbm":      stats.NoiseDBm,
		"link_speed_mbps": stats.LinkSpeedMbps,
	}

	return r
}
