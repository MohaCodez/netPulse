package diagnosis

import (
	"github.com/amit/netpulse/internal/storage"
)

// Aggregator builds a ProbeSnapshot from raw probe results within a time window.
type Aggregator struct{}

// NewAggregator creates a new probe result aggregator.
func NewAggregator() *Aggregator {
	return &Aggregator{}
}

// BuildSnapshot aggregates recent probe results into a snapshot for diagnosis.
func (a *Aggregator) BuildSnapshot(results []storage.ProbeResult, wifiSnap *storage.WifiSnapshot, windowSeconds int) *ProbeSnapshot {
	snap := &ProbeSnapshot{
		WindowSeconds: windowSeconds,
	}

	// Categorize results
	var gatewayResults, externalResults, dnsResults, tcpResults []storage.ProbeResult

	for _, r := range results {
		snap.TotalProbes++
		if !r.Success {
			snap.FailedProbes++
		}

		switch r.ProbeType {
		case "gateway":
			gatewayResults = append(gatewayResults, r)
		case "ping":
			externalResults = append(externalResults, r)
		case "dns":
			dnsResults = append(dnsResults, r)
		case "tcp":
			tcpResults = append(tcpResults, r)
		}
	}

	// Aggregate gateway metrics
	a.aggregateGateway(snap, gatewayResults)

	// Aggregate external ping metrics
	a.aggregateExternal(snap, externalResults)

	// Aggregate DNS metrics
	a.aggregateDNS(snap, dnsResults)

	// Aggregate TCP metrics
	a.aggregateTCP(snap, tcpResults)

	// Wi-Fi snapshot
	if wifiSnap != nil {
		snap.WifiSignalDBm = &wifiSnap.SignalDBm
		if wifiSnap.NoiseDBm != 0 {
			snap.WifiNoiseDBm = &wifiSnap.NoiseDBm
		}
		if wifiSnap.Channel != 0 {
			snap.WifiChannel = &wifiSnap.Channel
		}
		if wifiSnap.LinkSpeedMbps != 0 {
			speed := wifiSnap.LinkSpeedMbps
			snap.WifiLinkSpeed = &speed
		}
	}

	return snap
}

func (a *Aggregator) aggregateGateway(snap *ProbeSnapshot, results []storage.ProbeResult) {
	if len(results) == 0 {
		return
	}

	successCount := 0
	var totalLatency, totalJitter, totalLoss float64

	for _, r := range results {
		if r.Success {
			successCount++
			totalLatency += r.LatencyMs
			totalJitter += r.JitterMs
		}
		totalLoss += r.PacketLoss
	}

	snap.GatewayReachable = successCount > 0
	if successCount > 0 {
		snap.GatewayLatencyMs = totalLatency / float64(successCount)
		snap.GatewayJitterMs = totalJitter / float64(successCount)
	}
	snap.GatewayPacketLoss = totalLoss / float64(len(results))
}

func (a *Aggregator) aggregateExternal(snap *ProbeSnapshot, results []storage.ProbeResult) {
	if len(results) == 0 {
		return
	}

	// Track unique targets and their reachability
	targetSuccess := make(map[string]bool)
	successCount := 0
	var totalLatency, totalJitter, totalLoss float64

	for _, r := range results {
		if r.Success {
			successCount++
			totalLatency += r.LatencyMs
			totalJitter += r.JitterMs
			targetSuccess[r.Target] = true
		} else {
			if _, exists := targetSuccess[r.Target]; !exists {
				targetSuccess[r.Target] = false
			}
		}
		totalLoss += r.PacketLoss
	}

	snap.ExternalReachable = successCount > 0
	if successCount > 0 {
		snap.ExternalLatencyMs = totalLatency / float64(successCount)
		snap.ExternalJitterMs = totalJitter / float64(successCount)
	}
	snap.ExternalPacketLoss = totalLoss / float64(len(results))

	// Count distinct targets
	snap.ExternalTargetsTotal = len(targetSuccess)
	for _, up := range targetSuccess {
		if up {
			snap.ExternalTargetsUp++
		}
	}
}

func (a *Aggregator) aggregateDNS(snap *ProbeSnapshot, results []storage.ProbeResult) {
	if len(results) == 0 {
		return
	}

	successCount := 0
	var totalLatency float64
	systemOK := false
	alternatesOK := false

	for _, r := range results {
		if r.Success {
			successCount++
			totalLatency += r.LatencyMs
		}

		// Check resolver type from extra data
		if r.Extra != nil {
			if resolver, ok := r.Extra["resolver"]; ok {
				resolverStr, _ := resolver.(string)
				if r.Success {
					if resolverStr == "system" || resolverStr == "" {
						systemOK = true
					} else {
						alternatesOK = true
					}
				}
			}
			// Handle multi-resolver results
			if resolverResults, ok := r.Extra["resolver_results"]; ok {
				if rrSlice, ok := resolverResults.([]interface{}); ok {
					for _, rr := range rrSlice {
						if rrMap, ok := rr.(map[string]interface{}); ok {
							resolver, _ := rrMap["resolver"].(string)
							success, _ := rrMap["success"].(bool)
							if success {
								if resolver == "" {
									systemOK = true
								} else {
									alternatesOK = true
								}
							}
						}
					}
				}
			}
		}
	}

	snap.DNSResolving = successCount > 0
	if successCount > 0 {
		snap.DNSLatencyMs = totalLatency / float64(successCount)
	}
	snap.DNSFailRate = 1.0 - float64(successCount)/float64(len(results))
	snap.DNSSystemOK = systemOK
	snap.DNSAlternatesOK = alternatesOK
}

func (a *Aggregator) aggregateTCP(snap *ProbeSnapshot, results []storage.ProbeResult) {
	if len(results) == 0 {
		return
	}

	successCount := 0
	var totalLatency float64

	for _, r := range results {
		if r.Success {
			successCount++
			totalLatency += r.LatencyMs
		}
	}

	snap.TCPReachable = successCount > 0
	if successCount > 0 {
		snap.TCPLatencyMs = totalLatency / float64(successCount)
	}
}
