-- NetPulse SQLite Schema

-- Raw probe results
CREATE TABLE IF NOT EXISTS probe_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL DEFAULT (datetime('now')),
    probe_type TEXT NOT NULL,         -- 'ping', 'dns', 'tcp', 'gateway', 'wifi'
    target TEXT NOT NULL,             -- IP, hostname, or interface
    success INTEGER NOT NULL,         -- 1 = success, 0 = failure
    latency_ms REAL,                  -- round-trip time in milliseconds
    jitter_ms REAL,                   -- jitter if available
    packet_loss REAL,                 -- 0.0 to 1.0
    network_id TEXT DEFAULT '',       -- SSID or connection name
    extra_json TEXT,                  -- additional probe-specific data (JSON)
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_probe_results_timestamp ON probe_results(timestamp);
CREATE INDEX IF NOT EXISTS idx_probe_results_type ON probe_results(probe_type, timestamp);

-- Diagnosis events
CREATE TABLE IF NOT EXISTS diagnoses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL DEFAULT (datetime('now')),
    category TEXT NOT NULL,           -- 'gateway', 'isp', 'dns', 'wifi', 'throughput'
    severity TEXT NOT NULL,           -- 'info', 'warning', 'critical'
    title TEXT NOT NULL,              -- short human-readable title
    description TEXT NOT NULL,        -- plain-language explanation
    evidence_json TEXT NOT NULL,      -- JSON array of evidence items
    resolved INTEGER NOT NULL DEFAULT 0,
    resolved_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_diagnoses_timestamp ON diagnoses(timestamp);
CREATE INDEX IF NOT EXISTS idx_diagnoses_category ON diagnoses(category, timestamp);
CREATE INDEX IF NOT EXISTS idx_diagnoses_resolved ON diagnoses(resolved, timestamp);

-- Speed test results
CREATE TABLE IF NOT EXISTS speed_tests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL DEFAULT (datetime('now')),
    download_mbps REAL,
    upload_mbps REAL,
    latency_ms REAL,
    jitter_ms REAL,
    server TEXT,
    triggered_by TEXT NOT NULL,       -- 'scheduled', 'manual'
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_speed_tests_timestamp ON speed_tests(timestamp);

-- Wi-Fi snapshots
CREATE TABLE IF NOT EXISTS wifi_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL DEFAULT (datetime('now')),
    interface TEXT NOT NULL,
    ssid TEXT,
    bssid TEXT,
    frequency_mhz INTEGER,
    channel INTEGER,
    signal_dbm INTEGER,              -- RSSI in dBm
    noise_dbm INTEGER,               -- noise floor
    link_speed_mbps REAL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_wifi_snapshots_timestamp ON wifi_snapshots(timestamp);

-- Baseline statistics (rolling aggregates for comparison)
CREATE TABLE IF NOT EXISTS baselines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    probe_type TEXT NOT NULL,
    target TEXT NOT NULL,
    period_start DATETIME NOT NULL,
    period_end DATETIME NOT NULL,
    p50_latency_ms REAL,
    p95_latency_ms REAL,
    avg_latency_ms REAL,
    packet_loss_rate REAL,
    sample_count INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(probe_type, target, period_start)
);

CREATE INDEX IF NOT EXISTS idx_baselines_lookup ON baselines(probe_type, target, period_end);

-- Network change events
CREATE TABLE IF NOT EXISTS network_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL DEFAULT (datetime('now')),
    reason TEXT NOT NULL,              -- 'ssid_change', 'interface_change', 'gateway_change', 'disconnected', 'reconnected'
    prev_ssid TEXT,
    prev_type TEXT,
    prev_interface TEXT,
    prev_gateway TEXT,
    curr_ssid TEXT,
    curr_type TEXT,
    curr_interface TEXT,
    curr_gateway TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_network_events_timestamp ON network_events(timestamp);
