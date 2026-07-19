package probe

import (
	"context"
	"fmt"
	"net"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

// PingProbe measures ICMP reachability and latency to a target.
type PingProbe struct {
	target  string
	count   int
	timeout time.Duration
}

// NewPingProbe creates a ping probe for the given target.
func NewPingProbe(target string, timeout time.Duration) *PingProbe {
	return &PingProbe{
		target:  target,
		count:   3,
		timeout: timeout,
	}
}

func (p *PingProbe) Type() string { return "ping" }

func (p *PingProbe) Execute(ctx context.Context) Result {
	r := Result{
		Type:      "ping",
		Target:    p.target,
		Timestamp: time.Now(),
	}

	pinger, err := probing.NewPinger(p.target)
	if err != nil {
		r.Error = err
		r.Success = false
		return r
	}

	pinger.Count = p.count
	pinger.Timeout = p.timeout
	pinger.SetPrivileged(false) // Use unprivileged mode (UDP) — works without root

	// Respect context cancellation
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			pinger.Stop()
		case <-done:
		}
	}()

	err = pinger.Run()
	close(done)

	if err != nil {
		r.Error = err
		r.Success = false
		return r
	}

	stats := pinger.Statistics()
	r.Success = stats.PacketsRecv > 0
	r.LatencyMs = float64(stats.AvgRtt) / float64(time.Millisecond)
	r.JitterMs = float64(stats.StdDevRtt) / float64(time.Millisecond)
	r.PacketLoss = stats.PacketLoss / 100.0 // normalize to 0.0-1.0

	r.Extra = map[string]interface{}{
		"packets_sent": stats.PacketsSent,
		"packets_recv": stats.PacketsRecv,
		"min_rtt_ms":   float64(stats.MinRtt) / float64(time.Millisecond),
		"max_rtt_ms":   float64(stats.MaxRtt) / float64(time.Millisecond),
	}

	return r
}

// GatewayProbe pings the default gateway to test local network connectivity.
type GatewayProbe struct {
	timeout time.Duration
}

// NewGatewayProbe creates a probe that targets the default gateway.
func NewGatewayProbe(timeout time.Duration) *GatewayProbe {
	return &GatewayProbe{timeout: timeout}
}

func (p *GatewayProbe) Type() string { return "gateway" }

func (p *GatewayProbe) Execute(ctx context.Context) Result {
	gateway, err := detectGateway()
	if err != nil {
		return Result{
			Type:      "gateway",
			Target:    "unknown",
			Success:   false,
			Error:     err,
			Timestamp: time.Now(),
		}
	}

	ping := NewPingProbe(gateway, p.timeout)
	result := ping.Execute(ctx)
	result.Type = "gateway"
	result.Extra["gateway_ip"] = gateway
	return result
}

// TCPProbe tests connectivity via TCP when ICMP is blocked.
type TCPProbe struct {
	target  string
	port    int
	timeout time.Duration
}

// NewTCPProbe creates a TCP connectivity probe.
func NewTCPProbe(target string, port int, timeout time.Duration) *TCPProbe {
	return &TCPProbe{
		target:  target,
		port:    port,
		timeout: timeout,
	}
}

func (p *TCPProbe) Type() string { return "tcp" }

func (p *TCPProbe) Execute(ctx context.Context) Result {
	r := Result{
		Type:      "tcp",
		Target:    p.target,
		Timestamp: time.Now(),
	}

	addr := net.JoinHostPort(p.target, itoa(p.port))
	start := time.Now()

	dialer := net.Dialer{Timeout: p.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	elapsed := time.Since(start)

	if err != nil {
		r.Error = err
		r.Success = false
		return r
	}
	conn.Close()

	r.Success = true
	r.LatencyMs = float64(elapsed) / float64(time.Millisecond)
	r.Extra = map[string]interface{}{
		"port": p.port,
	}

	return r
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
