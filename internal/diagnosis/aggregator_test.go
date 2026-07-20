package diagnosis

import (
	"testing"
	"time"

	"github.com/amit/netpulse/internal/storage"
)

func TestAggregator_EmptyResults(t *testing.T) {
	agg := NewAggregator()
	snap := agg.BuildSnapshot(nil, nil, 30)

	if snap.TotalProbes != 0 {
		t.Errorf("expected 0 probes, got %d", snap.TotalProbes)
	}
	if snap.GatewayReachable {
		t.Error("gateway should not be reachable with no data")
	}
}

func TestAggregator_GatewayMetrics(t *testing.T) {
	agg := NewAggregator()
	results := []storage.ProbeResult{
		{ProbeType: "gateway", Success: true, LatencyMs: 5.0, JitterMs: 1.0, PacketLoss: 0},
		{ProbeType: "gateway", Success: true, LatencyMs: 7.0, JitterMs: 2.0, PacketLoss: 0},
		{ProbeType: "gateway", Success: false, LatencyMs: 0, PacketLoss: 1.0},
	}

	snap := agg.BuildSnapshot(results, nil, 30)

	if !snap.GatewayReachable {
		t.Error("gateway should be reachable (2/3 successful)")
	}
	// Avg latency of successful probes: (5+7)/2 = 6
	if snap.GatewayLatencyMs < 5.9 || snap.GatewayLatencyMs > 6.1 {
		t.Errorf("expected ~6ms gateway latency, got %f", snap.GatewayLatencyMs)
	}
	// Avg packet loss: (0+0+1)/3 = 0.33
	expectedLoss := 1.0 / 3.0
	if snap.GatewayPacketLoss < expectedLoss-0.01 || snap.GatewayPacketLoss > expectedLoss+0.01 {
		t.Errorf("expected ~0.33 gateway loss, got %f", snap.GatewayPacketLoss)
	}
}

func TestAggregator_ExternalTargets(t *testing.T) {
	agg := NewAggregator()
	results := []storage.ProbeResult{
		{ProbeType: "ping", Target: "8.8.8.8", Success: true, LatencyMs: 30.0},
		{ProbeType: "ping", Target: "1.1.1.1", Success: true, LatencyMs: 20.0},
		{ProbeType: "ping", Target: "208.67.222.222", Success: false, LatencyMs: 0},
	}

	snap := agg.BuildSnapshot(results, nil, 30)

	if !snap.ExternalReachable {
		t.Error("external should be reachable (2/3 targets up)")
	}
	if snap.ExternalTargetsUp != 2 {
		t.Errorf("expected 2 targets up, got %d", snap.ExternalTargetsUp)
	}
	if snap.ExternalTargetsTotal != 3 {
		t.Errorf("expected 3 total targets, got %d", snap.ExternalTargetsTotal)
	}
	// Avg of successful: (30+20)/2 = 25
	if snap.ExternalLatencyMs < 24.9 || snap.ExternalLatencyMs > 25.1 {
		t.Errorf("expected ~25ms external latency, got %f", snap.ExternalLatencyMs)
	}
}

func TestAggregator_DNSMetrics(t *testing.T) {
	agg := NewAggregator()
	results := []storage.ProbeResult{
		{ProbeType: "dns", Success: true, LatencyMs: 15.0},
		{ProbeType: "dns", Success: true, LatencyMs: 25.0},
		{ProbeType: "dns", Success: false, LatencyMs: 0},
	}

	snap := agg.BuildSnapshot(results, nil, 30)

	if !snap.DNSResolving {
		t.Error("DNS should be resolving (2/3 successful)")
	}
	// Fail rate: 1/3 = 0.33
	expectedFail := 1.0 / 3.0
	if snap.DNSFailRate < expectedFail-0.01 || snap.DNSFailRate > expectedFail+0.01 {
		t.Errorf("expected ~0.33 DNS fail rate, got %f", snap.DNSFailRate)
	}
}

func TestAggregator_TCPMetrics(t *testing.T) {
	agg := NewAggregator()
	results := []storage.ProbeResult{
		{ProbeType: "tcp", Success: true, LatencyMs: 40.0},
		{ProbeType: "tcp", Success: true, LatencyMs: 50.0},
	}

	snap := agg.BuildSnapshot(results, nil, 30)

	if !snap.TCPReachable {
		t.Error("TCP should be reachable")
	}
	if snap.TCPLatencyMs < 44.9 || snap.TCPLatencyMs > 45.1 {
		t.Errorf("expected ~45ms TCP latency, got %f", snap.TCPLatencyMs)
	}
}

func TestAggregator_WifiSnapshot(t *testing.T) {
	agg := NewAggregator()
	wifiSnap := &storage.WifiSnapshot{
		Timestamp:     time.Now(),
		Interface:     "wlan0",
		SSID:          "TestNetwork",
		SignalDBm:     -65,
		NoiseDBm:      -90,
		Channel:       6,
		LinkSpeedMbps: 72.2,
	}

	snap := agg.BuildSnapshot(nil, wifiSnap, 30)

	if snap.WifiSignalDBm == nil {
		t.Fatal("wifi signal should be set")
	}
	if *snap.WifiSignalDBm != -65 {
		t.Errorf("expected signal -65, got %d", *snap.WifiSignalDBm)
	}
	if snap.WifiChannel == nil || *snap.WifiChannel != 6 {
		t.Error("wifi channel should be 6")
	}
	if snap.WifiLinkSpeed == nil || *snap.WifiLinkSpeed != 72.2 {
		t.Error("wifi link speed should be 72.2")
	}
}

func TestAggregator_MixedProbeTypes(t *testing.T) {
	agg := NewAggregator()
	results := []storage.ProbeResult{
		{ProbeType: "gateway", Success: true, LatencyMs: 3.0},
		{ProbeType: "ping", Target: "8.8.8.8", Success: true, LatencyMs: 30.0},
		{ProbeType: "dns", Success: true, LatencyMs: 15.0},
		{ProbeType: "tcp", Success: true, LatencyMs: 35.0},
		{ProbeType: "ping", Target: "1.1.1.1", Success: false},
	}

	snap := agg.BuildSnapshot(results, nil, 30)

	if snap.TotalProbes != 5 {
		t.Errorf("expected 5 total probes, got %d", snap.TotalProbes)
	}
	if snap.FailedProbes != 1 {
		t.Errorf("expected 1 failed probe, got %d", snap.FailedProbes)
	}
	if !snap.GatewayReachable {
		t.Error("gateway should be reachable")
	}
	if !snap.ExternalReachable {
		t.Error("external should be reachable")
	}
	if !snap.DNSResolving {
		t.Error("DNS should be resolving")
	}
	if !snap.TCPReachable {
		t.Error("TCP should be reachable")
	}
}
