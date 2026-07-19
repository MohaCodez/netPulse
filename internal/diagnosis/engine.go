package diagnosis

import (
	"fmt"
	"time"

	"github.com/amit/netpulse/internal/storage"
)

// Engine is the root-cause diagnosis engine.
// It evaluates a ProbeSnapshot using a layered decision tree:
//
//	Layer 1: Gateway reachable?        → if no: local network / router issue
//	Layer 2: External targets reachable? → if no: ISP / upstream issue
//	Layer 3: DNS resolving?            → if no: DNS issue (system vs upstream)
//	Layer 4: Wi-Fi signal adequate?    → if weak: Wi-Fi issue
//	Layer 5: Latency/jitter/loss       → throughput degradation
//
// Each layer produces evidence and only fires if higher-priority layers pass.
type Engine struct {
	thresholds *Thresholds
}

// NewEngine creates a diagnosis engine with the given thresholds.
func NewEngine(t *Thresholds) *Engine {
	if t == nil {
		t = DefaultThresholds()
	}
	return &Engine{thresholds: t}
}

// GetThresholds returns the current thresholds.
func (e *Engine) GetThresholds() *Thresholds {
	return e.thresholds
}

// SetThresholds updates the thresholds.
func (e *Engine) SetThresholds(t *Thresholds) {
	e.thresholds = t
}

// Evaluate runs the full decision tree against a probe snapshot and returns a verdict.
func (e *Engine) Evaluate(snap *ProbeSnapshot) *Verdict {
	now := time.Now()

	// Not enough data to diagnose
	if snap.TotalProbes < e.thresholds.MinProbesForDiagnosis {
		return &Verdict{
			Category:    CategoryHealthy,
			Severity:    SeverityInfo,
			Title:       "Insufficient data",
			Description: "Not enough probe results yet to make a diagnosis.",
			Confidence:  0.0,
			Timestamp:   now,
		}
	}

	// Layer 1: Gateway check
	if v := e.checkGateway(snap, now); v != nil {
		return v
	}

	// Layer 2: ISP / external connectivity
	if v := e.checkISP(snap, now); v != nil {
		return v
	}

	// Layer 3: DNS
	if v := e.checkDNS(snap, now); v != nil {
		return v
	}

	// Layer 4: Wi-Fi signal
	if v := e.checkWifi(snap, now); v != nil {
		return v
	}

	// Layer 5: Throughput degradation (high latency/jitter/loss but everything reachable)
	if v := e.checkThroughput(snap, now); v != nil {
		return v
	}

	// All clear
	return &Verdict{
		Category:    CategoryHealthy,
		Severity:    SeverityInfo,
		Title:       "Network healthy",
		Description: "All probes are passing within normal parameters.",
		Confidence:  1.0,
		Timestamp:   now,
		Evidence: []storage.Evidence{
			{
				Type:        "summary",
				Description: "All layers passed",
				Value:       fmt.Sprintf("gateway=✓ external=%d/%d dns=✓ latency=%.0fms", snap.ExternalTargetsUp, snap.ExternalTargetsTotal, snap.ExternalLatencyMs),
				Timestamp:   now,
			},
		},
	}
}

// Layer 1: Gateway
func (e *Engine) checkGateway(snap *ProbeSnapshot, now time.Time) *Verdict {
	if snap.GatewayReachable && snap.GatewayPacketLoss < e.thresholds.LossWarning {
		return nil // gateway is fine, continue to next layer
	}

	var evidence []storage.Evidence

	if !snap.GatewayReachable {
		// Gateway completely unreachable
		evidence = append(evidence, storage.Evidence{
			Type:        "probe_failure",
			Description: "Default gateway is unreachable",
			Value:       "Gateway ping: 100% packet loss",
			Timestamp:   now,
		})

		// Differentiate: is it our NIC or the router?
		if snap.WifiSignalDBm != nil && *snap.WifiSignalDBm > -50 {
			evidence = append(evidence, storage.Evidence{
				Type:        "context",
				Description: "Wi-Fi signal is strong, suggesting router/modem issue rather than wireless",
				Value:       fmt.Sprintf("Signal: %d dBm (strong)", *snap.WifiSignalDBm),
				Timestamp:   now,
			})
		}

		return &Verdict{
			Category:    CategoryGateway,
			Severity:    SeverityCritical,
			Title:       "Gateway unreachable — local network down",
			Description: "Your router/gateway is not responding. This means your device cannot reach the local network. Check if your router is powered on and your network cable/Wi-Fi connection is active.",
			Evidence:    evidence,
			Confidence:  0.95,
			Timestamp:   now,
		}
	}

	// Gateway reachable but lossy
	evidence = append(evidence, storage.Evidence{
		Type:        "metric",
		Description: "High packet loss to gateway",
		Value:       fmt.Sprintf("%.0f%% packet loss, %.0fms latency", snap.GatewayPacketLoss*100, snap.GatewayLatencyMs),
		Timestamp:   now,
	})

	severity := SeverityWarning
	if snap.GatewayPacketLoss >= e.thresholds.LossCritical {
		severity = SeverityCritical
	}

	return &Verdict{
		Category:    CategoryGateway,
		Severity:    severity,
		Title:       "Gateway connection unstable",
		Description: fmt.Sprintf("Your router is reachable but dropping %.0f%% of packets. This usually indicates a failing network cable, Wi-Fi interference, or router under heavy load.", snap.GatewayPacketLoss*100),
		Evidence:    evidence,
		Confidence:  0.85,
		Timestamp:   now,
	}
}

// Layer 2: ISP / external connectivity
func (e *Engine) checkISP(snap *ProbeSnapshot, now time.Time) *Verdict {
	if snap.ExternalReachable && snap.ExternalPacketLoss < e.thresholds.LossWarning {
		return nil // external connectivity is fine
	}

	var evidence []storage.Evidence

	// Gateway works but external doesn't → ISP problem
	evidence = append(evidence, storage.Evidence{
		Type:        "differential",
		Description: "Gateway is reachable but external targets are not",
		Value:       fmt.Sprintf("Gateway: ✓ (%.0fms) | External: %d/%d targets up", snap.GatewayLatencyMs, snap.ExternalTargetsUp, snap.ExternalTargetsTotal),
		Timestamp:   now,
	})

	if !snap.ExternalReachable {
		// Check TCP fallback — if TCP works but ICMP doesn't, it's not a real outage
		if snap.TCPReachable {
			evidence = append(evidence, storage.Evidence{
				Type:        "context",
				Description: "TCP connections succeed — ICMP may be rate-limited by ISP",
				Value:       fmt.Sprintf("TCP latency: %.0fms (working)", snap.TCPLatencyMs),
				Timestamp:   now,
			})
			return &Verdict{
				Category:    CategoryISP,
				Severity:    SeverityInfo,
				Title:       "ICMP blocked but internet working",
				Description: "Ping probes are failing but TCP connections succeed. Your ISP or an upstream router is likely rate-limiting ICMP (ping) packets. Internet connectivity is fine.",
				Evidence:    evidence,
				Confidence:  0.9,
				Timestamp:   now,
			}
		}

		return &Verdict{
			Category:    CategoryISP,
			Severity:    SeverityCritical,
			Title:       "Internet down — ISP issue",
			Description: "Your router is reachable but nothing beyond it responds. This is an ISP outage or your modem has lost its upstream connection. Check your modem's status lights or contact your ISP.",
			Evidence:    evidence,
			Confidence:  0.9,
			Timestamp:   now,
		}
	}

	// Partial connectivity — some targets up, high loss
	evidence = append(evidence, storage.Evidence{
		Type:        "metric",
		Description: "Partial external connectivity with high packet loss",
		Value:       fmt.Sprintf("%.0f%% loss to external targets, %d/%d reachable", snap.ExternalPacketLoss*100, snap.ExternalTargetsUp, snap.ExternalTargetsTotal),
		Timestamp:   now,
	})

	severity := SeverityWarning
	if snap.ExternalPacketLoss >= e.thresholds.LossCritical {
		severity = SeverityCritical
	}

	return &Verdict{
		Category:    CategoryISP,
		Severity:    severity,
		Title:       "ISP connection degraded",
		Description: fmt.Sprintf("Your internet connection is partially working but losing %.0f%% of packets to external servers. This typically indicates ISP congestion or a routing problem upstream of your network.", snap.ExternalPacketLoss*100),
		Evidence:    evidence,
		Confidence:  0.8,
		Timestamp:   now,
	}
}

// Layer 3: DNS
func (e *Engine) checkDNS(snap *ProbeSnapshot, now time.Time) *Verdict {
	if snap.DNSResolving && snap.DNSFailRate < e.thresholds.LossWarning {
		return nil // DNS is fine
	}

	var evidence []storage.Evidence

	if !snap.DNSResolving {
		// DNS completely broken — differentiate system vs upstream
		if !snap.DNSSystemOK && snap.DNSAlternatesOK {
			evidence = append(evidence, storage.Evidence{
				Type:        "differential",
				Description: "System DNS resolver failing but alternate resolvers work",
				Value:       "System resolver: ✗ | Cloudflare (1.1.1.1): ✓ | Google (8.8.8.8): ✓",
				Timestamp:   now,
			})
			return &Verdict{
				Category:    CategoryDNS,
				Severity:    SeverityWarning,
				Title:       "System DNS resolver failing",
				Description: "Your configured DNS resolver is not responding, but alternative resolvers (like Cloudflare and Google) work fine. Your ISP's DNS server may be down. Switching to 1.1.1.1 or 8.8.8.8 in your network settings would fix this immediately.",
				Evidence:    evidence,
				Confidence:  0.9,
				Timestamp:   now,
			}
		}

		if !snap.DNSSystemOK && !snap.DNSAlternatesOK {
			evidence = append(evidence, storage.Evidence{
				Type:        "probe_failure",
				Description: "All DNS resolvers failing",
				Value:       "System resolver: ✗ | Alternates: ✗",
				Timestamp:   now,
			})
			// If pings work but DNS doesn't on ALL resolvers, likely a UDP filtering issue
			evidence = append(evidence, storage.Evidence{
				Type:        "context",
				Description: "External pings work but DNS (UDP:53) fails everywhere — possible UDP filtering",
				Value:       fmt.Sprintf("Ping: ✓ (%.0fms) | DNS: ✗ all resolvers", snap.ExternalLatencyMs),
				Timestamp:   now,
			})
			return &Verdict{
				Category:    CategoryDNS,
				Severity:    SeverityCritical,
				Title:       "DNS completely broken",
				Description: "No DNS resolver is responding. External servers are reachable by IP, but name resolution is failing. This could be a firewall blocking UDP port 53, or a widespread DNS outage.",
				Evidence:    evidence,
				Confidence:  0.85,
				Timestamp:   now,
			}
		}
	}

	// DNS partially working but slow/lossy
	evidence = append(evidence, storage.Evidence{
		Type:        "metric",
		Description: "DNS resolution degraded",
		Value:       fmt.Sprintf("%.0f%% failure rate, avg resolution: %.0fms", snap.DNSFailRate*100, snap.DNSLatencyMs),
		Timestamp:   now,
	})

	severity := SeverityWarning
	title := "DNS resolution slow"
	desc := fmt.Sprintf("DNS queries are succeeding but %.0f%% are failing and average resolution time is %.0fms (normally under 50ms). This causes web pages to load slowly even when connectivity is fine.", snap.DNSFailRate*100, snap.DNSLatencyMs)

	if snap.DNSLatencyMs > 500 {
		severity = SeverityCritical
		title = "DNS severely degraded"
	}

	return &Verdict{
		Category:    CategoryDNS,
		Severity:    severity,
		Title:       title,
		Description: desc,
		Evidence:    evidence,
		Confidence:  0.8,
		Timestamp:   now,
	}
}

// Layer 4: Wi-Fi
func (e *Engine) checkWifi(snap *ProbeSnapshot, now time.Time) *Verdict {
	if snap.WifiSignalDBm == nil {
		return nil // no Wi-Fi data available, skip this layer
	}

	signal := *snap.WifiSignalDBm
	if signal > e.thresholds.WifiSignalWarning {
		return nil // signal is fine
	}

	var evidence []storage.Evidence

	evidence = append(evidence, storage.Evidence{
		Type:        "metric",
		Description: "Wi-Fi signal strength",
		Value:       fmt.Sprintf("%d dBm", signal),
		Timestamp:   now,
	})

	if snap.WifiChannel != nil {
		evidence = append(evidence, storage.Evidence{
			Type:        "context",
			Description: "Wi-Fi channel",
			Value:       fmt.Sprintf("Channel %d", *snap.WifiChannel),
			Timestamp:   now,
		})
	}

	if snap.WifiNoiseDBm != nil {
		snr := signal - *snap.WifiNoiseDBm
		evidence = append(evidence, storage.Evidence{
			Type:        "metric",
			Description: "Signal-to-noise ratio",
			Value:       fmt.Sprintf("%d dB (signal: %d dBm, noise: %d dBm)", snr, signal, *snap.WifiNoiseDBm),
			Timestamp:   now,
		})
	}

	severity := SeverityWarning
	title := "Weak Wi-Fi signal"
	desc := fmt.Sprintf("Your Wi-Fi signal is %d dBm, which is below the recommended -70 dBm threshold. This can cause packet loss and speed drops. Try moving closer to your router or reducing interference from other devices.", signal)

	if signal <= e.thresholds.WifiSignalCritical {
		severity = SeverityCritical
		title = "Very weak Wi-Fi signal"
		desc = fmt.Sprintf("Your Wi-Fi signal is critically weak at %d dBm. At this level, connections will be unreliable with frequent drops. You need to move closer to your router, remove physical obstructions, or consider a Wi-Fi extender.", signal)
	}

	// If gateway loss is high AND signal is weak, strengthen the Wi-Fi verdict
	if snap.GatewayPacketLoss > 0.05 {
		evidence = append(evidence, storage.Evidence{
			Type:        "correlation",
			Description: "Gateway packet loss correlates with weak signal",
			Value:       fmt.Sprintf("Gateway loss: %.0f%% (likely caused by weak Wi-Fi)", snap.GatewayPacketLoss*100),
			Timestamp:   now,
		})
	}

	return &Verdict{
		Category:    CategoryWifi,
		Severity:    severity,
		Title:       title,
		Description: desc,
		Evidence:    evidence,
		Confidence:  0.75,
		Timestamp:   now,
	}
}

// Layer 5: Throughput degradation
func (e *Engine) checkThroughput(snap *ProbeSnapshot, now time.Time) *Verdict {
	// Everything is reachable but quality is degraded
	highLatency := snap.ExternalLatencyMs > e.thresholds.LatencyWarningMs ||
		snap.LatencyVsBaseline > e.thresholds.LatencyMultiplierWarning
	highJitter := snap.ExternalJitterMs > e.thresholds.JitterWarningMs
	someLoss := snap.ExternalPacketLoss > 0.02 // >2% loss is notable

	if !highLatency && !highJitter && !someLoss {
		return nil // throughput is fine
	}

	var evidence []storage.Evidence

	if highLatency {
		evidence = append(evidence, storage.Evidence{
			Type:        "metric",
			Description: "Elevated latency",
			Value:       fmt.Sprintf("%.0fms (%.1fx baseline)", snap.ExternalLatencyMs, snap.LatencyVsBaseline),
			Timestamp:   now,
		})
	}

	if highJitter {
		evidence = append(evidence, storage.Evidence{
			Type:        "metric",
			Description: "High jitter (latency variance)",
			Value:       fmt.Sprintf("%.0fms jitter", snap.ExternalJitterMs),
			Timestamp:   now,
		})
	}

	if someLoss {
		evidence = append(evidence, storage.Evidence{
			Type:        "metric",
			Description: "Packet loss detected",
			Value:       fmt.Sprintf("%.1f%% packet loss", snap.ExternalPacketLoss*100),
			Timestamp:   now,
		})
	}

	severity := SeverityWarning
	if snap.ExternalLatencyMs > e.thresholds.LatencyCriticalMs ||
		snap.LatencyVsBaseline > e.thresholds.LatencyMultiplierCritical ||
		snap.ExternalJitterMs > e.thresholds.JitterCriticalMs {
		severity = SeverityCritical
	}

	title := "Network performance degraded"
	desc := fmt.Sprintf("Your connection is working but performance is degraded: %.0fms latency (%.1fx your normal), %.0fms jitter, %.1f%% packet loss. This could be ISP congestion, network saturation from other devices, or a background download consuming bandwidth.",
		snap.ExternalLatencyMs, snap.LatencyVsBaseline, snap.ExternalJitterMs, snap.ExternalPacketLoss*100)

	return &Verdict{
		Category:    CategoryThroughput,
		Severity:    severity,
		Title:       title,
		Description: desc,
		Evidence:    evidence,
		Confidence:  0.7,
		Timestamp:   now,
	}
}
