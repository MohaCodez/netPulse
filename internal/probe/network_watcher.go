package probe

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// NetworkInfo represents the current active network connection.
type NetworkInfo struct {
	Interface string
	SSID      string
	Type      string // "wifi", "ethernet", "cellular", "unknown"
	Gateway   string
}

// NetworkChangeEvent is emitted when the active network changes.
type NetworkChangeEvent struct {
	Timestamp time.Time
	Previous  *NetworkInfo
	Current   *NetworkInfo
	Reason    string // "ssid_change", "interface_change", "gateway_change", "disconnected", "reconnected"
}

// NetworkChangeHandler is called when a network change is detected.
type NetworkChangeHandler func(NetworkChangeEvent)

// NetworkWatcher monitors for active network changes.
type NetworkWatcher struct {
	interval time.Duration
	handler  NetworkChangeHandler

	mu      sync.RWMutex
	current *NetworkInfo
	running bool
	cancel  context.CancelFunc
}

// NewNetworkWatcher creates a watcher that polls for network changes.
func NewNetworkWatcher(interval time.Duration, handler NetworkChangeHandler) *NetworkWatcher {
	return &NetworkWatcher{
		interval: interval,
		handler:  handler,
	}
}

// Start begins watching for network changes.
func (w *NetworkWatcher) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.mu.Unlock()

	go w.loop(ctx)
	log.Println("[network] watcher started")
}

// Stop halts the network watcher.
func (w *NetworkWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.running {
		return
	}
	w.running = false
	w.cancel()
	log.Println("[network] watcher stopped")
}

// Current returns the current network info.
func (w *NetworkWatcher) Current() *NetworkInfo {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.current
}

func (w *NetworkWatcher) loop(ctx context.Context) {
	// Get initial state
	info := detectNetwork(ctx)
	w.mu.Lock()
	w.current = info
	w.mu.Unlock()

	if info != nil {
		log.Printf("[network] initial: %s (%s) via %s, gw=%s", info.SSID, info.Type, info.Interface, info.Gateway)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.check(ctx)
		}
	}
}

func (w *NetworkWatcher) check(ctx context.Context) {
	newInfo := detectNetwork(ctx)

	w.mu.RLock()
	prev := w.current
	w.mu.RUnlock()

	event := detectChange(prev, newInfo)
	if event == nil {
		return
	}

	event.Timestamp = time.Now()
	event.Previous = prev
	event.Current = newInfo

	w.mu.Lock()
	w.current = newInfo
	w.mu.Unlock()

	log.Printf("[network] change detected: %s | %s → %s",
		event.Reason, formatNetwork(prev), formatNetwork(newInfo))

	if w.handler != nil {
		w.handler(*event)
	}
}

func detectNetwork(ctx context.Context) *NetworkInfo {
	info := &NetworkInfo{}

	// Get active connection via nmcli
	out, err := exec.CommandContext(ctx, "nmcli", "-t", "-f", "TYPE,DEVICE,NAME", "connection", "show", "--active").Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			parts := strings.SplitN(line, ":", 3)
			if len(parts) < 3 {
				continue
			}
			connType := parts[0]
			device := parts[1]
			name := parts[2]

			// Skip loopback and bridge
			if connType == "loopback" || connType == "bridge" {
				continue
			}

			switch {
			case strings.Contains(connType, "wireless") || strings.Contains(connType, "wifi"):
				info.Type = "wifi"
				info.Interface = device
				info.SSID = name
			case strings.Contains(connType, "ethernet"):
				info.Type = "ethernet"
				info.Interface = device
				info.SSID = name
			case strings.Contains(connType, "gsm") || strings.Contains(connType, "cdma"):
				info.Type = "cellular"
				info.Interface = device
				info.SSID = name
			default:
				if info.Type == "" {
					info.Type = "unknown"
					info.Interface = device
					info.SSID = name
				}
			}

			// Prefer wifi/cellular over ethernet for primary display
			if info.Type == "wifi" || info.Type == "cellular" {
				break
			}
		}
	}

	// Get gateway
	gateway, _ := detectGateway()
	info.Gateway = gateway

	if info.Interface == "" {
		return nil // no active connection
	}

	return info
}

func detectChange(prev, curr *NetworkInfo) *NetworkChangeEvent {
	if prev == nil && curr == nil {
		return nil
	}

	if prev == nil && curr != nil {
		return &NetworkChangeEvent{Reason: "reconnected"}
	}

	if prev != nil && curr == nil {
		return &NetworkChangeEvent{Reason: "disconnected"}
	}

	if prev.SSID != curr.SSID {
		return &NetworkChangeEvent{Reason: "ssid_change"}
	}

	if prev.Interface != curr.Interface {
		return &NetworkChangeEvent{Reason: "interface_change"}
	}

	if prev.Gateway != curr.Gateway {
		return &NetworkChangeEvent{Reason: "gateway_change"}
	}

	return nil
}

func formatNetwork(info *NetworkInfo) string {
	if info == nil {
		return "disconnected"
	}
	return info.SSID + " (" + info.Type + "/" + info.Interface + ")"
}
