package probe

import (
	"context"
	"fmt"
	"net"
	"time"
)

// DNSProbe measures DNS resolution time for a domain against a specific resolver.
type DNSProbe struct {
	domain   string
	resolver string // empty string = system default
	timeout  time.Duration
}

// NewDNSProbe creates a DNS resolution probe.
// If resolver is empty, it uses the system's default resolver.
func NewDNSProbe(domain, resolver string, timeout time.Duration) *DNSProbe {
	return &DNSProbe{
		domain:   domain,
		resolver: resolver,
		timeout:  timeout,
	}
}

func (p *DNSProbe) Type() string { return "dns" }

func (p *DNSProbe) Execute(ctx context.Context) Result {
	r := Result{
		Type:      "dns",
		Target:    p.domain,
		Timestamp: time.Now(),
	}

	var resolver *net.Resolver
	resolverLabel := "system"

	if p.resolver != "" {
		resolverLabel = p.resolver
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: p.timeout}
				return d.DialContext(ctx, "udp", net.JoinHostPort(p.resolver, "53"))
			},
		}
	} else {
		resolver = net.DefaultResolver
	}

	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	start := time.Now()
	addrs, err := resolver.LookupHost(ctx, p.domain)
	elapsed := time.Since(start)

	if err != nil {
		r.Error = err
		r.Success = false
		r.Extra = map[string]interface{}{
			"resolver": resolverLabel,
			"error":    err.Error(),
		}
		return r
	}

	r.Success = true
	r.LatencyMs = float64(elapsed) / float64(time.Millisecond)
	r.Extra = map[string]interface{}{
		"resolver":       resolverLabel,
		"resolved_addrs": addrs,
		"domain":         p.domain,
	}

	return r
}

// DNSMultiProbe tests DNS against multiple resolvers and reports differential results.
type DNSMultiProbe struct {
	domain    string
	resolvers []string
	timeout   time.Duration
}

// NewDNSMultiProbe creates a probe that tests a domain against multiple resolvers.
func NewDNSMultiProbe(domain string, resolvers []string, timeout time.Duration) *DNSMultiProbe {
	return &DNSMultiProbe{
		domain:    domain,
		resolvers: resolvers,
		timeout:   timeout,
	}
}

func (p *DNSMultiProbe) Type() string { return "dns" }

func (p *DNSMultiProbe) Execute(ctx context.Context) Result {
	r := Result{
		Type:      "dns",
		Target:    fmt.Sprintf("%s (multi-resolver)", p.domain),
		Timestamp: time.Now(),
	}

	type resolverResult struct {
		Resolver  string  `json:"resolver"`
		Success   bool    `json:"success"`
		LatencyMs float64 `json:"latency_ms"`
		Error     string  `json:"error,omitempty"`
	}

	var results []resolverResult
	successCount := 0
	var totalLatency float64

	for _, resolver := range p.resolvers {
		probe := NewDNSProbe(p.domain, resolver, p.timeout)
		res := probe.Execute(ctx)

		rr := resolverResult{
			Resolver:  resolver,
			Success:   res.Success,
			LatencyMs: res.LatencyMs,
		}
		if res.Error != nil {
			rr.Error = res.Error.Error()
		}
		if res.Success {
			successCount++
			totalLatency += res.LatencyMs
		}
		results = append(results, rr)
	}

	r.Success = successCount > 0
	if successCount > 0 {
		r.LatencyMs = totalLatency / float64(successCount)
	}
	r.PacketLoss = 1.0 - float64(successCount)/float64(len(p.resolvers))

	r.Extra = map[string]interface{}{
		"resolver_results": results,
		"domain":           p.domain,
	}

	return r
}
