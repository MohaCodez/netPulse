package diagnosis

import (
	"time"

	"github.com/amit/netpulse/internal/storage"
)

// Category represents the type of network issue diagnosed.
type Category string

const (
	CategoryGateway    Category = "gateway"
	CategoryISP        Category = "isp"
	CategoryDNS        Category = "dns"
	CategoryWifi       Category = "wifi"
	CategoryThroughput Category = "throughput"
	CategoryHealthy    Category = "healthy"
)

// Severity levels for diagnoses.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Verdict is the output of the diagnosis engine for a single evaluation cycle.
type Verdict struct {
	Category    Category
	Severity    Severity
	Title       string
	Description string
	Evidence    []storage.Evidence
	Confidence  float64 // 0.0 to 1.0
	Timestamp   time.Time
}

// ProbeSnapshot holds aggregated probe data for a diagnosis window.
type ProbeSnapshot struct {
	// Gateway metrics
	GatewayReachable  bool
	GatewayLatencyMs  float64
	GatewayPacketLoss float64
	GatewayJitterMs   float64

	// External ping metrics (aggregated across targets)
	ExternalReachable  bool
	ExternalLatencyMs  float64
	ExternalPacketLoss float64
	ExternalJitterMs   float64
	ExternalTargetsUp  int
	ExternalTargetsTotal int

	// DNS metrics
	DNSResolving      bool
	DNSLatencyMs      float64
	DNSFailRate       float64
	DNSSystemOK       bool // system resolver working
	DNSAlternatesOK   bool // alternate resolvers working

	// TCP fallback
	TCPReachable bool
	TCPLatencyMs float64

	// Wi-Fi metrics (may be nil if not available)
	WifiSignalDBm  *int
	WifiNoiseDBm   *int
	WifiChannel    *int
	WifiLinkSpeed  *float64

	// Baseline comparisons
	LatencyVsBaseline float64 // multiplier: 2.0 = 2x normal latency
	LossVsBaseline    float64 // multiplier

	// Raw counts for confidence
	TotalProbes   int
	FailedProbes  int
	WindowSeconds int
}

// Thresholds defines the tunable parameters for diagnosis decisions.
type Thresholds struct {
	// Packet loss thresholds
	LossWarning  float64 // packet loss rate to trigger warning (e.g., 0.1 = 10%)
	LossCritical float64 // packet loss rate to trigger critical (e.g., 0.5 = 50%)

	// Latency thresholds (absolute)
	LatencyWarningMs  float64 // latency to trigger warning
	LatencyCriticalMs float64 // latency to trigger critical

	// Latency thresholds (relative to baseline)
	LatencyMultiplierWarning  float64 // e.g., 2.0 = 2x baseline triggers warning
	LatencyMultiplierCritical float64 // e.g., 5.0 = 5x baseline triggers critical

	// Wi-Fi signal thresholds (dBm, more negative = worse)
	WifiSignalWarning  int // e.g., -70
	WifiSignalCritical int // e.g., -80

	// Jitter thresholds
	JitterWarningMs  float64
	JitterCriticalMs float64

	// Confidence: minimum number of probes in window to make a diagnosis
	MinProbesForDiagnosis int
}

// DefaultThresholds returns sensible default thresholds.
func DefaultThresholds() *Thresholds {
	return &Thresholds{
		LossWarning:  0.10,
		LossCritical: 0.50,

		LatencyWarningMs:  150.0,
		LatencyCriticalMs: 500.0,

		LatencyMultiplierWarning:  2.0,
		LatencyMultiplierCritical: 5.0,

		WifiSignalWarning:  -70,
		WifiSignalCritical: -80,

		JitterWarningMs:  50.0,
		JitterCriticalMs: 150.0,

		MinProbesForDiagnosis: 3,
	}
}
