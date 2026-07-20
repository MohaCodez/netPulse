package diagnosis

import (
	"testing"
	"time"

	"github.com/amit/netpulse/internal/storage"
)

// --- Helper to create a healthy baseline snapshot ---
func healthySnapshot() *ProbeSnapshot {
	return &ProbeSnapshot{
		GatewayReachable:     true,
		GatewayLatencyMs:     3.0,
		GatewayPacketLoss:    0.0,
		GatewayJitterMs:      1.0,
		ExternalReachable:    true,
		ExternalLatencyMs:    35.0,
		ExternalPacketLoss:   0.0,
		ExternalJitterMs:     5.0,
		ExternalTargetsUp:    3,
		ExternalTargetsTotal: 3,
		DNSResolving:         true,
		DNSLatencyMs:         20.0,
		DNSFailRate:          0.0,
		DNSSystemOK:          true,
		DNSAlternatesOK:      true,
		TCPReachable:         true,
		TCPLatencyMs:         30.0,
		LatencyVsBaseline:    1.0,
		TotalProbes:          20,
		FailedProbes:         0,
		WindowSeconds:        30,
	}
}

func TestEngine_HealthyNetwork(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryHealthy {
		t.Errorf("expected CategoryHealthy, got %s: %s", verdict.Category, verdict.Title)
	}
	if verdict.Confidence != 1.0 {
		t.Errorf("expected confidence 1.0, got %f", verdict.Confidence)
	}
}

func TestEngine_InsufficientData(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.TotalProbes = 1 // below MinProbesForDiagnosis (3)

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryHealthy {
		t.Errorf("expected CategoryHealthy for insufficient data, got %s", verdict.Category)
	}
	if verdict.Confidence != 0.0 {
		t.Errorf("expected confidence 0.0, got %f", verdict.Confidence)
	}
}

// --- Layer 1: Gateway ---

func TestEngine_GatewayUnreachable(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.GatewayReachable = false
	snap.GatewayPacketLoss = 1.0

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryGateway {
		t.Errorf("expected CategoryGateway, got %s: %s", verdict.Category, verdict.Title)
	}
	if verdict.Severity != SeverityCritical {
		t.Errorf("expected SeverityCritical, got %s", verdict.Severity)
	}
	if len(verdict.Evidence) == 0 {
		t.Error("expected evidence, got none")
	}
}

func TestEngine_GatewayHighLoss(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.GatewayPacketLoss = 0.3 // 30% loss

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryGateway {
		t.Errorf("expected CategoryGateway, got %s: %s", verdict.Category, verdict.Title)
	}
	if verdict.Severity != SeverityWarning {
		t.Errorf("expected SeverityWarning, got %s", verdict.Severity)
	}
}

func TestEngine_GatewayCriticalLoss(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.GatewayPacketLoss = 0.6 // 60% loss

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryGateway {
		t.Errorf("expected CategoryGateway, got %s", verdict.Category)
	}
	if verdict.Severity != SeverityCritical {
		t.Errorf("expected SeverityCritical, got %s", verdict.Severity)
	}
}

// --- Layer 2: ISP ---

func TestEngine_ISPDown(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.ExternalReachable = false
	snap.ExternalPacketLoss = 1.0
	snap.ExternalTargetsUp = 0
	snap.TCPReachable = false

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryISP {
		t.Errorf("expected CategoryISP, got %s: %s", verdict.Category, verdict.Title)
	}
	if verdict.Severity != SeverityCritical {
		t.Errorf("expected SeverityCritical, got %s", verdict.Severity)
	}
}

func TestEngine_ISPDown_ICMPBlocked(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.ExternalReachable = false
	snap.ExternalPacketLoss = 1.0
	snap.ExternalTargetsUp = 0
	snap.TCPReachable = true // TCP works — ICMP just blocked
	snap.TCPLatencyMs = 40.0

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryISP {
		t.Errorf("expected CategoryISP, got %s: %s", verdict.Category, verdict.Title)
	}
	if verdict.Severity != SeverityInfo {
		t.Errorf("expected SeverityInfo (ICMP blocked, not a real outage), got %s", verdict.Severity)
	}
}

func TestEngine_ISPDegraded(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.ExternalPacketLoss = 0.25 // 25% loss
	snap.ExternalTargetsUp = 2
	snap.ExternalTargetsTotal = 3

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryISP {
		t.Errorf("expected CategoryISP, got %s: %s", verdict.Category, verdict.Title)
	}
	if verdict.Severity != SeverityWarning {
		t.Errorf("expected SeverityWarning, got %s", verdict.Severity)
	}
}

// --- Layer 3: DNS ---

func TestEngine_DNSSystemFailing_AlternatesOK(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.DNSResolving = false
	snap.DNSSystemOK = false
	snap.DNSAlternatesOK = true

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryDNS {
		t.Errorf("expected CategoryDNS, got %s: %s", verdict.Category, verdict.Title)
	}
	if verdict.Severity != SeverityWarning {
		t.Errorf("expected SeverityWarning, got %s", verdict.Severity)
	}
}

func TestEngine_DNSAllFailing(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.DNSResolving = false
	snap.DNSSystemOK = false
	snap.DNSAlternatesOK = false

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryDNS {
		t.Errorf("expected CategoryDNS, got %s: %s", verdict.Category, verdict.Title)
	}
	if verdict.Severity != SeverityCritical {
		t.Errorf("expected SeverityCritical, got %s", verdict.Severity)
	}
}

func TestEngine_DNSSlow(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.DNSResolving = true
	snap.DNSFailRate = 0.15 // 15% fail rate
	snap.DNSLatencyMs = 200.0

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryDNS {
		t.Errorf("expected CategoryDNS, got %s: %s", verdict.Category, verdict.Title)
	}
}

// --- Layer 4: Wi-Fi ---

func TestEngine_WifiWeakSignal(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	signal := -75
	snap.WifiSignalDBm = &signal

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryWifi {
		t.Errorf("expected CategoryWifi, got %s: %s", verdict.Category, verdict.Title)
	}
	if verdict.Severity != SeverityWarning {
		t.Errorf("expected SeverityWarning, got %s", verdict.Severity)
	}
}

func TestEngine_WifiCriticalSignal(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	signal := -85
	snap.WifiSignalDBm = &signal

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryWifi {
		t.Errorf("expected CategoryWifi, got %s: %s", verdict.Category, verdict.Title)
	}
	if verdict.Severity != SeverityCritical {
		t.Errorf("expected SeverityCritical, got %s", verdict.Severity)
	}
}

func TestEngine_WifiGoodSignal_NoAlarm(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	signal := -55
	snap.WifiSignalDBm = &signal

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryHealthy {
		t.Errorf("expected CategoryHealthy (good signal), got %s: %s", verdict.Category, verdict.Title)
	}
}

func TestEngine_WifiNil_Skipped(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.WifiSignalDBm = nil // no wifi data

	verdict := engine.Evaluate(snap)

	if verdict.Category == CategoryWifi {
		t.Error("should not diagnose wifi when no data available")
	}
}

// --- Layer 5: Throughput ---

func TestEngine_HighLatency(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.ExternalLatencyMs = 200.0
	snap.LatencyVsBaseline = 3.0

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryThroughput {
		t.Errorf("expected CategoryThroughput, got %s: %s", verdict.Category, verdict.Title)
	}
	if verdict.Severity != SeverityWarning {
		t.Errorf("expected SeverityWarning, got %s", verdict.Severity)
	}
}

func TestEngine_CriticalLatency(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.ExternalLatencyMs = 600.0
	snap.LatencyVsBaseline = 6.0

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryThroughput {
		t.Errorf("expected CategoryThroughput, got %s: %s", verdict.Category, verdict.Title)
	}
	if verdict.Severity != SeverityCritical {
		t.Errorf("expected SeverityCritical, got %s", verdict.Severity)
	}
}

func TestEngine_HighJitter(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.ExternalJitterMs = 80.0

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryThroughput {
		t.Errorf("expected CategoryThroughput, got %s: %s", verdict.Category, verdict.Title)
	}
}

func TestEngine_MildLoss(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.ExternalPacketLoss = 0.05 // 5% — not enough to trigger ISP layer but triggers throughput

	verdict := engine.Evaluate(snap)

	if verdict.Category != CategoryThroughput {
		t.Errorf("expected CategoryThroughput, got %s: %s", verdict.Category, verdict.Title)
	}
}

// --- Layer priority: higher layers take precedence ---

func TestEngine_GatewayOverridesISP(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.GatewayReachable = false
	snap.GatewayPacketLoss = 1.0
	snap.ExternalReachable = false // both down
	snap.ExternalPacketLoss = 1.0

	verdict := engine.Evaluate(snap)

	// Gateway should fire first, not ISP
	if verdict.Category != CategoryGateway {
		t.Errorf("gateway issue should take precedence over ISP, got %s", verdict.Category)
	}
}

func TestEngine_ISPOverridesDNS(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.ExternalReachable = false
	snap.ExternalPacketLoss = 1.0
	snap.ExternalTargetsUp = 0
	snap.TCPReachable = false
	snap.DNSResolving = false // DNS also down
	snap.DNSSystemOK = false

	verdict := engine.Evaluate(snap)

	// ISP should fire, not DNS (DNS failure is a symptom of ISP being down)
	if verdict.Category != CategoryISP {
		t.Errorf("ISP issue should take precedence over DNS, got %s", verdict.Category)
	}
}

// --- Evidence quality ---

func TestEngine_EvidencePresent(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()
	snap.GatewayReachable = false
	snap.GatewayPacketLoss = 1.0

	verdict := engine.Evaluate(snap)

	if len(verdict.Evidence) == 0 {
		t.Error("diagnosis should include evidence")
	}
	for _, e := range verdict.Evidence {
		if e.Type == "" || e.Description == "" {
			t.Errorf("evidence has empty fields: %+v", e)
		}
		if e.Timestamp.IsZero() {
			t.Error("evidence timestamp should not be zero")
		}
	}
}

func TestEngine_VerdictTimestamp(t *testing.T) {
	engine := NewEngine(nil)
	snap := healthySnapshot()

	before := time.Now()
	verdict := engine.Evaluate(snap)
	after := time.Now()

	if verdict.Timestamp.Before(before) || verdict.Timestamp.After(after) {
		t.Errorf("verdict timestamp %v should be between %v and %v", verdict.Timestamp, before, after)
	}
}

// --- Custom thresholds ---

func TestEngine_CustomThresholds(t *testing.T) {
	thresholds := DefaultThresholds()
	thresholds.LatencyWarningMs = 50.0 // very strict
	engine := NewEngine(thresholds)

	snap := healthySnapshot()
	snap.ExternalLatencyMs = 60.0 // would be fine with defaults, triggers with custom
	snap.LatencyVsBaseline = 1.5

	verdict := engine.Evaluate(snap)

	// Should not trigger with default (150ms) but this is a borderline case
	// The throughput layer checks both absolute AND relative thresholds
	if snap.ExternalLatencyMs > thresholds.LatencyWarningMs {
		if verdict.Category != CategoryThroughput {
			t.Errorf("custom threshold should trigger throughput, got %s", verdict.Category)
		}
	}
}

// --- Confidence scoring ---

func TestEngine_ConfidenceRange(t *testing.T) {
	engine := NewEngine(nil)

	scenarios := []struct {
		name string
		snap *ProbeSnapshot
	}{
		{"healthy", healthySnapshot()},
		{"gateway_down", func() *ProbeSnapshot {
			s := healthySnapshot()
			s.GatewayReachable = false
			s.GatewayPacketLoss = 1.0
			return s
		}()},
		{"weak_wifi", func() *ProbeSnapshot {
			s := healthySnapshot()
			signal := -75
			s.WifiSignalDBm = &signal
			return s
		}()},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			verdict := engine.Evaluate(sc.snap)
			if verdict.Confidence < 0 || verdict.Confidence > 1.0 {
				t.Errorf("confidence %f out of range [0,1]", verdict.Confidence)
			}
		})
	}
}

// Ensure the Evidence struct matches storage package
func TestEngine_EvidenceTypeCompatibility(t *testing.T) {
	e := storage.Evidence{
		Type:        "metric",
		Description: "test",
		Value:       "123",
		Timestamp:   time.Now(),
	}
	if e.Type == "" {
		t.Error("evidence type should not be empty")
	}
}
