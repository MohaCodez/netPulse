# NetPulse

A local-first network health monitor that **explains problems, not just detects them**.

NetPulse continuously monitors your internet/Wi-Fi health, runs a layered diagnostic engine to isolate the actual root cause of issues, and shows it in plain language with the evidence behind it — all stored locally, no cloud dependency.

![Overview](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go) ![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react) ![Wails](https://img.shields.io/badge/Wails-2.13-red)

---
<img width="1369" height="994" alt="image" src="https://github.com/user-attachments/assets/bdcd6b45-7248-44d3-b5ff-baccdc50b1a1" />

<img width="1369" height="994" alt="image" src="https://github.com/user-attachments/assets/45d0d0c6-7dc2-4fb0-90fb-2c283b312306" />

<img width="1378" height="1001" alt="image" src="https://github.com/user-attachments/assets/6232a86c-f02e-4d13-a454-6a18f7f404bc" />

<img width="1378" height="1001" alt="image" src="https://github.com/user-attachments/assets/e8657a74-c334-4aa9-98f0-a41a9971028b" />

<img width="1369" height="994" alt="image" src="https://github.com/user-attachments/assets/da494a7a-395b-4e92-8a10-fa1c9774fe96" />

<img width="1369" height="994" alt="image" src="https://github.com/user-attachments/assets/e9cf69ed-b8ae-49a7-9c88-637b93104bc9" />


## Features

### Core Monitoring
- **Internet uptime monitoring** — continuous gateway + external IP reachability checks
- **Latency/jitter/packet loss** — ICMP ping + TCP fallback (for networks that rate-limit ICMP)
- **DNS health checks** — resolution time tracked per resolver (system, Google, Cloudflare)
- **Wi-Fi signal tracking** — RSSI, channel, band, link speed via nmcli/iwconfig
- **Speed tests** — periodic + on-demand download/upload via Cloudflare CDN
- **Network change detection** — detects Wi-Fi → hotspot switches, logs events
- **LAN device scanner** — discovers devices on your network via ARP + ping sweep with vendor identification
- **Network change detection** — detects Wi-Fi → hotspot switches, logs events

### Diagnosis Engine
5-layer decision tree that isolates root cause:

1. **Gateway** — is your router reachable?
2. **ISP** — is the internet reachable beyond your router?
3. **DNS** — are name servers responding? System vs alternates?
4. **Wi-Fi** — is signal strength adequate?
5. **Throughput** — is latency/jitter/loss degraded vs baseline?

Each diagnosis includes confidence score and evidence.

### Desktop Dashboard
- Live status with current diagnosis + evidence
- Uptime rings (1h / 24h / 7-day)
- Latency timeline, gateway vs external comparison
- Wi-Fi signal chart with band hop tracking
- Signal ↔ latency correlation view
- DNS resolver comparison
- Jitter analysis
- Reliability heatmap (by hour of day)
- Diagnosis timeline (24h status bar)
- Speed test history
- Configurable alert thresholds
- Export data (JSON) / clear data

### Notifications
- Desktop notifications (via `notify-send`) on detected issues
- Configurable cooldown to prevent spam
- Network switch notifications

---

## Quick Start

### Prerequisites

- **Go 1.21+**
- **Node 18+** (for frontend build)
- **Linux** with:
  - `libgtk-3-dev` and `libwebkit2gtk-4.0-dev` (for Wails)
  - `nmcli` or `iwconfig` (for Wi-Fi stats)

```bash
# Install system dependencies (Ubuntu/Debian)
sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev

# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Run (Development)

```bash
export PATH="$HOME/go/bin:$PATH"
cd netPulse
wails dev
```

### Run (Headless / CLI only)

No GUI, just monitoring + logging:

```bash
go run ./cmd/netpulse/
```

### Build (Production)

```bash
wails build
./build/bin/netpulse
```

---

## Project Structure

```
netPulse/
├── app.go                    # Wails app entry point
├── api.go                    # Frontend API bindings
├── analytics_api.go          # Analytics chart data endpoints
├── wails.json                # Wails configuration
├── cmd/
│   ├── netpulse/             # Headless CLI entry point
│   └── dbcheck/              # Database inspection utility
├── internal/
│   ├── config/               # Configuration management
│   ├── probe/                # Probe system (ping, DNS, TCP, gateway, network watcher)
│   ├── diagnosis/            # Root-cause diagnosis engine
│   ├── storage/              # SQLite storage + analytics queries
│   ├── wifi/                 # Wi-Fi stats collector (nmcli/iwconfig/proc)
│   ├── speedtest/            # Speed test runner
│   ├── notifier/             # Desktop notifications
│   └── export/               # CSV/JSON export
└── frontend/                 # React + TypeScript dashboard
    └── src/
        ├── components/       # UI components (charts, cards, config)
        └── hooks/            # Data fetching hooks
```

---

## Configuration

Config file: `~/.config/netpulse/config.json` (auto-created with defaults)

Key settings:
| Setting | Default | Description |
|---------|---------|-------------|
| `probe_interval` | 5s | How often probes run |
| `probe_timeout` | 3s | Timeout per probe |
| `ping_targets` | 8.8.8.8, 1.1.1.1, 208.67.222.222 | External ping targets |
| `dns_resolvers` | system, 8.8.8.8, 1.1.1.1 | DNS resolvers to test |
| `speed_test_interval` | 1h | Auto speed test frequency |
| `diagnosis_window` | 30s | Data window for diagnosis |
| `baseline_window_days` | 7 | Rolling baseline period |

---

## Database

SQLite at `~/.config/netpulse/netpulse.db`

Tables: `probe_results`, `diagnoses`, `speed_tests`, `wifi_snapshots`, `baselines`, `network_events`

Inspect with:
```bash
go run ./cmd/dbcheck/
```

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go |
| Storage | SQLite (WAL mode) |
| Desktop UI | Wails v2 (WebKit2GTK) |
| Frontend | React + TypeScript + Recharts |
| Notifications | notify-send (Linux) |

---

## Security & Reliability

- **Database permissions**: SQLite file created with `0600` (owner read/write only)
- **Ping permissions check**: Warns on startup if unprivileged ICMP is disabled
- **LAN scanner rate-limited**: Maximum one scan per 60 seconds to prevent network noise
- **Speed test capped**: 4 connections × 8s duration to avoid saturating shared networks
- **Write reliability**: Batched writer with retry on queue full; integrity check on DB open
- **Alert persistence**: Threshold changes saved to `~/.config/netpulse/config.json`
- **Auto-maintenance**: Old data purged after 7 days; baselines recomputed hourly

---

## License

MIT
