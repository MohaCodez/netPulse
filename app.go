package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/amit/netpulse/internal/config"
	"github.com/amit/netpulse/internal/diagnosis"
	"github.com/amit/netpulse/internal/notifier"
	"github.com/amit/netpulse/internal/probe"
	"github.com/amit/netpulse/internal/scanner"
	"github.com/amit/netpulse/internal/speedtest"
	"github.com/amit/netpulse/internal/storage"
	"github.com/amit/netpulse/internal/wifi"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("NetPulse starting...")

	// Load configuration
	cfgPath := config.ConfigDir() + "/config.json"
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Open database
	db, err := storage.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Create the app API (bound to frontend)
	app := NewApp(cfg, db)

	// Run Wails
	err = wails.Run(&options.App{
		Title:            "NetPulse",
		Width:            1200,
		Height:           800,
		MinWidth:         800,
		MinHeight:        600,
		HideWindowOnClose: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatalf("Error running app: %v", err)
	}
}

// App is the main application struct bound to the frontend.
type App struct {
	ctx            context.Context
	cfg            *config.Config
	db             *storage.DB
	writer         *storage.Writer
	runner         *probe.Runner
	diagMonitor    *diagnosis.Monitor
	diagEngine     *diagnosis.Engine
	speedRunner    *speedtest.Runner
	notify         *notifier.Notifier
	networkWatcher *probe.NetworkWatcher
	netScanner     *scanner.Scanner
}

// NewApp creates a new App instance.
func NewApp(cfg *config.Config, db *storage.DB) *App {
	return &App{
		cfg: cfg,
		db:  db,
	}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Set up batched writer
	a.writer = storage.NewWriter(a.db, 256)

	// Set up notifications
	a.notify = notifier.NewNotifier(a.cfg.NotificationsEnabled, 60*time.Second)

	// Set up probe runner
	a.runner = probe.NewRunner(a.cfg.ProbeInterval, a.handleProbeResult)
	a.runner.AddProbe(probe.NewGatewayProbe(a.cfg.ProbeTimeout))
	for _, target := range a.cfg.PingTargets {
		a.runner.AddProbe(probe.NewPingProbe(target, a.cfg.ProbeTimeout))
	}
	for _, domain := range a.cfg.DNSTestDomains {
		a.runner.AddProbe(probe.NewDNSMultiProbe(domain, a.cfg.DNSResolvers, a.cfg.ProbeTimeout))
	}
	for _, target := range a.cfg.PingTargets[:1] {
		for _, port := range a.cfg.TCPFallbackPorts {
			a.runner.AddProbe(probe.NewTCPProbe(target, port, a.cfg.ProbeTimeout))
		}
	}
	a.runner.AddProbe(wifi.NewProbe("", a.cfg.ProbeTimeout))
	a.runner.Start()

	// Set up diagnosis
	a.diagEngine = diagnosis.NewEngine(nil)
	a.diagMonitor = diagnosis.NewMonitor(a.db, a.diagEngine, 10*time.Second, a.cfg.DiagnosisWindow, a.handleDiagnosis)
	a.diagMonitor.Start()

	// Set up speed test
	if a.cfg.SpeedTestEnabled {
		a.speedRunner = speedtest.NewRunner(nil, a.cfg.SpeedTestInterval, a.handleSpeedTest)
		a.speedRunner.Start()
	}

	// Set up network change watcher
	a.networkWatcher = probe.NewNetworkWatcher(3*time.Second, a.handleNetworkChange)
	a.networkWatcher.Start()

	// Set up network scanner
	a.netScanner = scanner.NewScanner()

	// Set up periodic maintenance (purge old data, recompute baselines)
	maintenance := storage.NewMaintenance(a.db, 1*time.Hour)
	maintenance.Start()

	log.Println("All systems started")
}

// shutdown is called when the app is closing.
func (a *App) shutdown(ctx context.Context) {
	log.Println("Shutting down...")
	if a.networkWatcher != nil {
		a.networkWatcher.Stop()
	}
	if a.diagMonitor != nil {
		a.diagMonitor.Stop()
	}
	if a.runner != nil {
		a.runner.Stop()
	}
	if a.speedRunner != nil {
		a.speedRunner.Stop()
	}
	if a.writer != nil {
		a.writer.Stop()
	}
	if a.db != nil {
		a.db.Close()
	}
	log.Println("NetPulse stopped.")
}

func (a *App) handleProbeResult(r probe.Result) {
	// Get current network name
	var networkID string
	if a.networkWatcher != nil {
		if info := a.networkWatcher.Current(); info != nil {
			networkID = info.SSID
		}
	}

	pr := &storage.ProbeResult{
		Timestamp:  r.Timestamp,
		ProbeType:  r.Type,
		Target:     r.Target,
		Success:    r.Success,
		LatencyMs:  r.LatencyMs,
		JitterMs:   r.JitterMs,
		PacketLoss: r.PacketLoss,
		NetworkID:  networkID,
		Extra:      r.Extra,
	}
	a.writer.Enqueue(func() {
		if err := a.db.InsertProbeResult(pr); err != nil {
			log.Printf("[store] error saving probe result: %v", err)
		}
	})

	// Push to frontend via Wails events
	wailsRuntime.EventsEmit(a.ctx, "probe:result", map[string]interface{}{
		"probe_type": r.Type,
		"target":     r.Target,
		"success":    r.Success,
		"latency_ms": r.LatencyMs,
		"jitter_ms":  r.JitterMs,
		"packet_loss": r.PacketLoss,
		"timestamp":  r.Timestamp.Format(time.RFC3339),
	})

	// Store Wi-Fi snapshots
	if r.Type == "wifi" && r.Success && r.Extra != nil {
		if avail, ok := r.Extra["available"].(bool); ok && avail {
			ws := &storage.WifiSnapshot{
				Timestamp: r.Timestamp,
				Interface: stringVal(r.Extra, "interface"),
				SSID:      stringVal(r.Extra, "ssid"),
				BSSID:     stringVal(r.Extra, "bssid"),
			}
			if v, ok := r.Extra["frequency_mhz"].(int); ok {
				ws.FrequencyMHz = v
			}
			if v, ok := r.Extra["channel"].(int); ok {
				ws.Channel = v
			}
			if v, ok := r.Extra["signal_dbm"].(int); ok {
				ws.SignalDBm = v
			}
			if v, ok := r.Extra["noise_dbm"].(int); ok {
				ws.NoiseDBm = v
			}
			if v, ok := r.Extra["link_speed_mbps"].(float64); ok {
				ws.LinkSpeedMbps = v
			}
			a.db.InsertWifiSnapshot(ws)
		}
	}
}

func (a *App) handleDiagnosis(v *diagnosis.Verdict) {
	if v.Category != diagnosis.CategoryHealthy {
		a.notify.SendDiagnosis(string(v.Severity), v.Title, v.Description)
	}
	// Push diagnosis to frontend
	wailsRuntime.EventsEmit(a.ctx, "diagnosis:update", map[string]interface{}{
		"status":      string(v.Severity),
		"category":    string(v.Category),
		"title":       v.Title,
		"description": v.Description,
		"confidence":  v.Confidence,
		"timestamp":   v.Timestamp.Format(time.RFC3339),
	})
}

func (a *App) handleSpeedTest(result *speedtest.Result) {
	a.diagMonitor.SetSpeedTesting(false) // Clear flag after test completes
	st := &storage.SpeedTestResult{
		Timestamp:    result.Timestamp,
		DownloadMbps: result.DownloadMbps,
		UploadMbps:   result.UploadMbps,
		LatencyMs:    result.LatencyMs,
		JitterMs:     result.JitterMs,
		Server:       result.Server,
		TriggeredBy:  "scheduled",
	}
	a.db.InsertSpeedTest(st)
}

func (a *App) handleNetworkChange(event probe.NetworkChangeEvent) {
	ne := &storage.NetworkEvent{
		Timestamp: event.Timestamp,
		Reason:    event.Reason,
	}
	if event.Previous != nil {
		ne.PrevSSID = event.Previous.SSID
		ne.PrevType = event.Previous.Type
		ne.PrevInterface = event.Previous.Interface
		ne.PrevGateway = event.Previous.Gateway
	}
	if event.Current != nil {
		ne.CurrSSID = event.Current.SSID
		ne.CurrType = event.Current.Type
		ne.CurrInterface = event.Current.Interface
		ne.CurrGateway = event.Current.Gateway
	}

	if err := a.db.InsertNetworkEvent(ne); err != nil {
		log.Printf("[network] error storing event: %v", err)
	}

	// Send notification
	if event.Current != nil {
		a.notify.Send(notifier.Notification{
			Title:   "Network Changed",
			Body:    fmt.Sprintf("Switched to %s (%s)", event.Current.SSID, event.Current.Type),
			Urgency: "normal",
			Icon:    "network-wireless",
		})
	} else {
		a.notify.Send(notifier.Notification{
			Title:   "Network Disconnected",
			Body:    "Lost network connection",
			Urgency: "critical",
			Icon:    "network-offline",
		})
	}
}

func stringVal(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
