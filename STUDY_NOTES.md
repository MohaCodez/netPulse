# NetPulse — Code Study Checklist (Interview Prep)

Goal: be able to explain *every* major decision in this codebase, not just what it does but *why* it's built that way. Organized by how likely you are to be asked, and how deep you need to go.

---

## TIER 1 — MAJOR CONCEPTS (must explain fluently, whiteboard-ready)

These are the things an interviewer will definitely probe. You should be able to explain each without looking at code.

### The core problem and why it's hard

**60-second pitch:**
"When someone says 'my internet is slow,' there are at least 5 distinct failure points: your Wi-Fi link to the router, the router itself, your ISP's upstream connection, DNS resolution, or pure throughput degradation under load. Every existing tool — ping, traceroute, speed test — tests ONE of these. None of them do differential diagnosis: 'your gateway is fine, DNS is fine, but your ISP is dropping packets between your router and their backbone.' That's what NetPulse does — it runs all probes simultaneously and uses a layered decision tree to isolate which layer is actually broken."

**Why naive tools fail:**
- `ping 8.8.8.8` tells you nothing if it fails — is it your Wi-Fi? Your router? Your ISP? DNS? All look the same: "timeout."
- Speed test sites only measure throughput — they can't tell you *why* it's slow (signal? congestion? DNS delay before the test even starts?)
- Uptime monitors (UptimeRobot, Pingdom) test from *outside* your network — they can't see your local Wi-Fi problems at all

**The insight:** You need probes at *every layer simultaneously*, then compare results differentially. If gateway ping fails but Wi-Fi signal is strong → router is dead. If gateway works but external doesn't → ISP problem. This differential approach is what makes the diagnosis engine useful.

---

### The 5-layer diagnosis decision tree

**File:** `internal/diagnosis/engine.go`

**Order and rationale:**
```
Layer 1: Gateway    → Is your router reachable?
Layer 2: ISP        → Is the internet reachable beyond your router?
Layer 3: DNS        → Are name servers responding?
Layer 4: Wi-Fi      → Is signal strength adequate?
Layer 5: Throughput  → Is performance degraded vs your baseline?
```

**Why this order matters:** Each layer rules out one variable before the next runs. If the gateway is unreachable, there's no point checking ISP connectivity (of course external pings will fail too). If the ISP is down, DNS will also fail — but the root cause is ISP, not DNS. The tree prevents misdiagnosis by testing from most-local to most-remote.

**Deep dive — Layer 2 (ISP) with the ICMP trick:**

The ISP layer checks: "Gateway works, but can we reach external IPs?"

Critical subtlety: many ISPs rate-limit or block ICMP (ping) packets. If all pings to 8.8.8.8 fail, that LOOKS like an outage — but it might just be ICMP filtering. So the engine checks: "Do TCP connections to port 443 succeed?" If TCP works but ICMP doesn't:
- Verdict: `SeverityInfo` — "ICMP blocked but internet working"
- Evidence: "TCP latency: 40ms (working)"

This prevents a false alarm that would trigger on a naive ping-only monitor. It's a real-world networking gotcha that most monitoring tools get wrong.

**Deep dive — Layer 3 (DNS differential):**

DNS layer doesn't just check "can you resolve names?" — it checks *which resolvers work*:
- System resolver fails, but 1.1.1.1 and 8.8.8.8 work → "Your ISP's DNS is down. Switch to Cloudflare/Google to fix immediately."
- ALL resolvers fail → "DNS completely broken. Possible UDP:53 filtering."

This differential tells the user exactly what to do (change DNS settings) rather than just "DNS is broken."

**Confidence and Evidence:**

Every verdict carries:
- `Confidence float64` (0.0–1.0) — how sure are we? Higher with more data points, lower near thresholds.
- `Evidence []Evidence` — array of supporting data points. Each has `Type` (metric, differential, correlation, context), `Description`, `Value`, and `Timestamp`.

Why evidence matters: "Your internet is down" is useless. "Your internet is down because: gateway responds in 3ms (local link fine), but 0/3 external targets respond over 30 seconds, and TCP:443 also fails (confirmed not just ICMP blocking)" — that's actionable.

---

### Why a decision tree instead of ML/scoring model

**Interview answer:**

"I chose a rule-based decision tree for three reasons:

1. **Explainability** — Every verdict has a clear 'because' chain. An ML model might say 'ISP issue, 73% confidence' but can't show you *why*. The tree can point to exact evidence.

2. **Debuggability** — When the diagnosis is wrong (and it will be), I can trace exactly which condition fired incorrectly and fix it. With a trained model, debugging is much harder.

3. **Testability** — The engine is pure functions on a `ProbeSnapshot` struct. I have 32 unit tests covering every layer, every edge case, every priority ordering. That's possible because the logic is deterministic.

The tradeoff: a weighted scoring model handles *correlated multi-cause failures* better. Example: signal is -72 dBm (borderline) AND gateway loss is 5% (borderline) — individually neither crosses a threshold, but together they indicate Wi-Fi trouble. The tree handles some of these with correlation checks (gateway loss + weak signal → strengthen Wi-Fi verdict), but a full scoring model would be more elegant.

For v1, correctness and explainability matter more than handling every edge case. You can always layer scoring on top of a working tree; you can't easily add explainability to a black-box model."

---

### Concurrency model

**File:** `internal/probe/runner.go`

**How probes run:**
```go
// Each tick, all probes run simultaneously:
for _, p := range probes {
    wg.Add(1)
    go func(probe Probe) {
        defer wg.Done()
        defer func() { recover() }()  // panic isolation
        result := probe.Execute(ctx)
        handler(result)
    }(p)
}
wg.Wait()
```

**Key design decisions:**

1. **Goroutine per probe + WaitGroup:** All 9 probes run in parallel because they're I/O-bound (waiting for network responses). Sequential execution would make a 5-second tick take 45+ seconds.

2. **`sync.RWMutex` on probe list:** The runner can have probes added (`AddProbe`) while the loop is running. `RLock` during iteration, `Lock` for mutations. This allows dynamic probe registration without stopping the runner.

3. **`context.Context` for shutdown:** The runner's `Start()` creates a context with `cancel`. When `Stop()` is called, cancel fires, and:
   - The ticker loop exits
   - Any in-flight probe that respects context (like DNS resolver, TCP dialer) gets cancelled
   - This is cleaner than a `running bool` because it propagates to all child operations automatically

4. **Panic recovery per goroutine:** If one probe panics (e.g., `pro-bing` library bug, nil pointer from malformed network data), it's caught by `recover()` and logged. The other 8 probes continue running. This is essential for a long-running monitor — one bad probe result must never crash the entire application.

**Why this matters in an interview:** This demonstrates understanding of production Go patterns: goroutine lifecycle management, graceful shutdown, failure isolation, and proper mutex usage.

---

### Data model / SQLite schema

**File:** `internal/storage/schema.sql`

**6 tables:**
| Table | Purpose |
|-------|---------|
| `probe_results` | Every individual probe measurement (timestamp, type, target, success, latency, jitter, loss, network_id) |
| `diagnoses` | Diagnosis events with evidence (timestamp, category, severity, title, description, evidence_json, resolved) |
| `speed_tests` | Speed test results (download, upload, latency, jitter, server, triggered_by) |
| `wifi_snapshots` | Wi-Fi measurements (signal, channel, band, link speed, SSID, BSSID) |
| `baselines` | Precomputed rolling statistics (p50/p95 latency, avg, packet loss per probe type/target/period) |
| `network_events` | Network change detections (SSID change, interface change, gateway change, disconnect/reconnect) |

**Key design decisions:**

1. **`evidence_json` and `extra_json` as JSON blobs:** Different probe types produce different extra data (ping has min/max RTT; DNS has resolver list; Wi-Fi has channel/band). Rather than 20 nullable columns or a complex normalized schema, a JSON blob captures heterogeneous data flexibly. Tradeoff: can't query inside these efficiently with SQL, but we rarely need to — most queries filter by timestamp and probe_type, which are indexed.

2. **Timestamps indexed on every table:** This is a time-series workload. Nearly every query is "give me data from the last N minutes/hours." Without timestamp indexes, these queries would table-scan. With indexes, they're O(log n).

3. **Precomputed baselines:** The diagnosis engine needs to compare "current latency" to "your normal latency." Computing p50/p95 over 7 days of data on every 10-second diagnosis tick would be expensive. Instead, baselines are computed periodically (hourly) and stored. The engine just reads the latest baseline row — O(1) lookup.

4. **WAL mode:** SQLite's Write-Ahead Logging allows concurrent reads while writes happen. Critical for us: the diagnosis monitor reads while the probe runner writes. Without WAL, readers would block on writers (or vice versa).

---

### Overall architecture / why this stack

**Go for the agent:**
- Goroutines for parallel probes (9 concurrent network operations every 5s)
- Single static binary — no runtime dependencies, easy to install
- Low memory footprint for an always-on background process
- Strong stdlib for networking (net, net/http, context)

**Wails instead of Electron:**
- Electron bundles a full Chromium + Node.js runtime (~200MB+). Wails uses the system's native webview (~0MB overhead).
- Go backend is bound directly to JS frontend — no HTTP server, no IPC serialization, just function calls
- Same developer experience (React + TypeScript frontend) without the bloat

**SQLite instead of Postgres:**
- Zero configuration — user installs the app and it works. No "please install and configure a database server first."
- Embedded — the database is a single file in `~/.config/netpulse/`
- Perfectly adequate for single-node time-series at this scale (~17K rows/day, purged at 7 days)
- WAL mode handles our concurrent read/write pattern

**What would change for multi-node/coordinator:**
- SQLite → PostgreSQL or TimescaleDB (need concurrent access from multiple agents)
- Single binary → agent + coordinator architecture (agents push to coordinator via gRPC or HTTP)
- Local UI → web dashboard (coordinator serves the UI, not each agent)
- Add authentication between agents and coordinator

---

## TIER 2 — MINOR CONCEPTS (should know solidly, but less likely to be the focus)

### Probe types and implementation

**`ping.go`** — Uses `prometheus-community/pro-bing` library (Go implementation of ICMP ping). Runs in unprivileged mode (UDP-based, no root required). Sends 3 packets per probe, computes avg/stddev (jitter)/loss.

**`dns.go`** — Creates custom `net.Resolver` with specific resolver address. Measures time for `LookupHost`. `DNSMultiProbe` tests same domain against multiple resolvers in one shot for differential analysis.

**`gateway_linux.go`** — Reads `/proc/net/route`, finds the default route (destination `00000000`), parses the hex gateway IP. Platform-specific (`_linux.go` build tag). The hex is little-endian on x86, hence the `binary.LittleEndian` conversion.

**`network_watcher.go`** — Polls `nmcli connection show --active` every 3 seconds. Compares current SSID/interface/gateway to previous. Emits events on change. Stores transitions in `network_events` table.

### Wi-Fi stats collection

**`internal/wifi/collector.go`**

Collection priority: `nmcli` → `iwconfig` → `/proc/net/wireless`

- **nmcli** (NetworkManager CLI): richest data — SSID, BSSID, channel, frequency, signal %, link rate. Available on most modern Linux distros.
- **iwconfig**: older tool, still common. Gives ESSID, frequency, bit rate, signal level in dBm.
- **/proc/net/wireless**: last resort, only gives link quality + signal level as numbers.

**Why shell out to tools instead of a Go library:** No portable Go library reliably reads RSSI/channel/noise across Linux distros. The kernel interfaces vary (nl80211, WEXT), and wrapping them in Go would be more code than the entire wifi package. CLI tools abstract this — pragmatic tradeoff.

**Limitation:** Linux-only. macOS would need CoreWLAN (via cgo or `airport` CLI), Windows needs netsh/WMI. Not implemented for v1.

### Speed test module

**`internal/speedtest/speedtest.go`**

- **Download:** 6 parallel HTTP GET requests to Ubuntu/kernel.org mirrors, reads for 10 seconds, measures total bytes → Mbps.
- **Upload:** 3 parallel POST requests with 100KB random chunks to Cloudflare's `__up` endpoint. Smaller chunks avoid rate limiting.
- **Latency:** 5 HEAD requests to the download server, average + stddev.
- **Scheduling:** Runs 30s after app start, then every 1 hour. Also triggerable on-demand from UI.
- **Diagnosis interaction:** Sets `diagMonitor.SetSpeedTesting(true)` during test so the diagnosis engine ignores the latency spike caused by saturating the link.

### Notifier

**`internal/notifier/notifier.go`**

Wraps `notify-send` (Linux desktop notification daemon). Key detail: **cooldown timer** — minimum 60 seconds between notifications. Without this, a flapping network issue (connects/disconnects every 5 seconds) would spam 12 notifications per minute.

### Export

**`internal/export/export.go`**

Supports CSV and JSON for: probe results, diagnoses, speed tests. The frontend also has `ExportFullReport()` which bundles all tables into one JSON with a timestamp. Used for the "Download Report & Clear" workflow.

### Aggregator / Monitor

**`internal/diagnosis/aggregator.go`** — Takes raw `[]ProbeResult` and a `*WifiSnapshot`, categorizes by probe type, computes averages/rates, and outputs a `ProbeSnapshot` struct that the engine consumes.

**`internal/diagnosis/monitor.go`** — Runs on a timer (10s). Fetches last 30s of probe results, gets latest Wi-Fi snapshot, gets baseline for latency comparison, builds snapshot, calls `engine.Evaluate()`, stores non-healthy verdicts, resolves when state returns to healthy. Emits events to frontend via Wails runtime.

Key concept: **`diagnosis_window` (30s)** — Why not diagnose on every single probe result? Because a single dropped packet (0.003% loss rate over an hour) would trigger a diagnosis. The 30-second window provides enough data points (6 probe ticks × 9 probes = 54 results) to distinguish real problems from blips. `ConfirmationCount` (3 consecutive failures minimum) adds further debouncing.

### Local AI integration

**`ollama.go`** + **`AIChat.tsx`**

- Talks to `http://localhost:11434/api/chat` (Ollama's local API)
- Pre-checks Ollama availability (2s timeout HEAD to `/api/tags`) before sending — instant error if not running
- Injects real-time context into every prompt: current diagnosis, Wi-Fi signal/band/channel, latency averages, speed test results, recent network events
- Model is user-switchable via dropdown (queries `/api/tags` for installed models)
- **Critical point for interviews:** The AI is a *presentation layer*, not the diagnosis source. The rule engine makes the decision; the AI explains it in natural language. This is intentional — you don't want a hallucinating LLM making network health decisions.

### Wails bindings

**`app.go`** — App struct with lifecycle hooks:
- `startup(ctx)`: initializes all subsystems (writer, probes, diagnosis, speed test, network watcher, maintenance)
- `shutdown(ctx)`: stops everything in reverse order, flushes writer, closes DB

**`api.go`** — Methods on `App` that become callable from JS:
- `GetCurrentStatus()`, `GetRecentProbes(minutes)`, `GetDiagnosisHistory(limit)`, etc.
- Each returns a response struct that Wails auto-serializes to JSON

**`analytics_api.go`** — Chart data endpoints, alert rules get/set, project report aggregation, AI chat

Pattern: Go struct methods → Wails generates TypeScript bindings → React calls them like local async functions.

### Frontend structure

**Hooks pattern:**
- `useNetworkData.ts` — polls core data (status, probes, diagnoses, speed tests, uptime, wifi)
- `useAnalytics.ts` — generic hook factory for analytics endpoints (signal history, packet loss, etc.)
- `useWailsEvents.ts` — subscribes to push events from Go (reduces polling needed)

**Component categories:**
- Status: `StatusCard`, `UptimeCard`, `WifiCard`
- Charts: `LatencyChart`, `WifiSignalChart`, `PacketLossChart`, `GatewayVsExternalChart`, `DNSComparisonChart`, `JitterChart`, `BandHopChart`, `CorrelationChart`, `SpeedHistoryChart`, `Heatmap`, `DiagnosisTimelineChart`
- Interactive: `SpeedTest`, `AIChat`, `AlertConfig`, `ExportButton`
- Reference: `Theory`, `ProjectReport`

---

## TIER 3 — DETAIL-ORIENTED (know these exist, look up exact values only if asked)

### Default thresholds (`DefaultThresholds()`)
- Packet loss: warning 10%, critical 50%
- Latency: warning 150ms, critical 500ms
- Latency multiplier vs baseline: warning 2×, critical 5×
- Wi-Fi signal: warning -70 dBm, critical -80 dBm
- Jitter: warning 50ms, critical 150ms
- `MinProbesForDiagnosis`: 3 (won't diagnose with less data)

### File locations
- Config: `~/.config/netpulse/config.json`
- Database: `~/.config/netpulse/netpulse.db`
- Default probe interval: 5 seconds
- Diagnosis window: 30 seconds
- Speed test interval: 1 hour
- Baseline window: 7 days

### Dependencies
- `github.com/mattn/go-sqlite3` — CGo SQLite driver (compiles SQLite into the binary)
- `github.com/prometheus-community/pro-bing` — ICMP ping library (unprivileged mode)
- `github.com/wailsapp/wails/v2` — Desktop app framework
- `recharts` — React charting library

### Utilities
- `cmd/dbcheck/` — standalone DB inspector, prints summary stats from SQLite
- `cmd/netpulse/` — headless CLI mode (no GUI, just probes + logging + SQLite)

### CI structure
- `.github/workflows/build.yml` — two jobs: `test` (Go unit tests, no system deps needed) and `build` (installs GTK + WebKit, builds frontend, then builds Go with embedded assets)

### Known gaps (be honest about these if asked)
- No Windows/macOS Wi-Fi support (Linux-only for nmcli/iwconfig)
- Upload speed test depends on Cloudflare not rate-limiting (sometimes reports 0)
- No test coverage for storage layer or frontend components (only diagnosis engine tested)
- Alert rules don't persist to disk (lost on restart)
- Single-node only — no multi-device monitoring
- `HideWindowOnClose` (tray mode) only works in production build, not dev mode

---

## How to Use This

1. **Go tier by tier** — don't move to Tier 2 until you can explain every Tier 1 item out loud, unaided, in plain language.
2. **For each Tier 1 item, practice a 30–60 second spoken explanation** — interview answers should be concise, not a code walkthrough.
3. **Tier 3 items are lookup-only** — you don't need to memorize numbers, just know they exist and roughly what they're for.
4. **Be honest about what you understand vs. what the AI helped generate** — interviewers respect "I designed the diagnosis logic and iterated with AI assistance on implementation" far more than pretending to have hand-written every line, and it's a more defensible position if probed.

---

## Quick Answers for Common Questions

**"How do you know the diagnosis is correct?"**
→ 32 unit tests covering all 5 layers, priority ordering, evidence quality, custom thresholds, and edge cases. Tests use pure function inputs (ProbeSnapshot structs) — no network calls, no mocks, fully deterministic.

**"What happens if the app crashes mid-write?"**
→ SQLite WAL mode + integrity check on startup. If corruption is detected, it fails fast rather than producing garbage data.

**"How does it handle network changes (Wi-Fi → hotspot)?"**
→ NetworkWatcher polls nmcli every 3s, detects SSID/interface/gateway changes, logs events, sends notification. Probe results are tagged with `network_id` so you can separate per-network analytics.

**"What about resource usage?"**
→ In production build (not dev mode): ~50-100MB RAM, negligible CPU between probe ticks. SQLite DB auto-purges data older than 7 days. Single goroutine writer batches DB inserts to reduce I/O.

**"Why not just use Prometheus/Grafana?"**
→ Those are great for server monitoring but require setup, configuration, and a separate dashboard. NetPulse is for end-users who want to understand their home network without running infrastructure. Also, Prometheus doesn't do *diagnosis* — it collects metrics, but you still need to interpret them yourself.
