package wifi

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Stats holds a Wi-Fi measurement snapshot.
type Stats struct {
	Interface     string
	SSID          string
	BSSID         string
	FrequencyMHz  int
	Channel       int
	SignalDBm     int
	NoiseDBm      int
	LinkSpeedMbps float64
	Timestamp     time.Time
}

// Collector gathers Wi-Fi statistics from the system.
type Collector struct {
	iface string
}

// NewCollector creates a Wi-Fi stats collector.
func NewCollector(iface string) *Collector {
	return &Collector{iface: iface}
}

// Collect gathers current Wi-Fi stats using the best available tool.
func (c *Collector) Collect(ctx context.Context) (*Stats, error) {
	iface := c.iface
	if iface == "" {
		var err error
		iface, err = detectWifiInterface()
		if err != nil || iface == "" {
			return nil, err
		}
	}

	// Try nmcli first (richest data)
	if stats, err := collectWithNmcli(ctx, iface); err == nil && stats != nil {
		stats.Interface = iface
		stats.Timestamp = time.Now()
		return stats, nil
	}

	// Try iwconfig
	if stats, err := collectWithIwconfig(ctx, iface); err == nil && stats != nil {
		stats.Interface = iface
		stats.Timestamp = time.Now()
		return stats, nil
	}

	// Fallback to /proc/net/wireless
	stats, err := collectFromProc(iface)
	if err != nil {
		return nil, err
	}
	stats.Interface = iface
	stats.Timestamp = time.Now()
	return stats, nil
}

func detectWifiInterface() (string, error) {
	out, err := exec.Command("cat", "/proc/net/wireless").Output()
	if err != nil {
		return "", nil
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			return strings.TrimSuffix(fields[0], ":"), nil
		}
	}
	return "", nil
}

func collectWithNmcli(ctx context.Context, iface string) (*Stats, error) {
	out, err := exec.CommandContext(ctx, "nmcli", "-t", "-f",
		"active,ssid,bssid,chan,freq,signal,rate", "dev", "wifi").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "yes:") {
			continue
		}

		// Replace escaped colons in BSSID before splitting
		cleaned := strings.ReplaceAll(line, "\\:", "#")
		parts := strings.Split(cleaned, ":")
		// Expected: [yes, SSID, BSSID, chan, freq, signal, rate]
		if len(parts) < 7 {
			continue
		}

		stats := &Stats{}
		stats.SSID = parts[1]
		stats.BSSID = strings.ReplaceAll(parts[2], "#", ":")
		stats.Channel, _ = strconv.Atoi(parts[3])

		freqStr := strings.TrimSuffix(parts[4], " MHz")
		stats.FrequencyMHz, _ = strconv.Atoi(strings.TrimSpace(freqStr))

		// nmcli signal is 0-100 percentage
		signalPct, _ := strconv.Atoi(parts[5])
		stats.SignalDBm = pctToDbm(signalPct)

		rateStr := strings.TrimSuffix(parts[6], " Mbit/s")
		stats.LinkSpeedMbps, _ = strconv.ParseFloat(strings.TrimSpace(rateStr), 64)

		// Overwrite signal with actual dBm from /proc if available
		if procStats, err := collectFromProc(iface); err == nil && procStats.SignalDBm != 0 {
			stats.SignalDBm = procStats.SignalDBm
			if procStats.NoiseDBm > -200 {
				stats.NoiseDBm = procStats.NoiseDBm
			}
		}

		return stats, nil
	}

	return nil, fmt.Errorf("no active wifi in nmcli output")
}

func collectWithIwconfig(ctx context.Context, iface string) (*Stats, error) {
	out, err := exec.CommandContext(ctx, "iwconfig", iface).Output()
	if err != nil {
		return nil, err
	}

	s := string(out)
	stats := &Stats{}

	if m := regexp.MustCompile(`ESSID:"([^"]+)"`).FindStringSubmatch(s); len(m) > 1 {
		stats.SSID = m[1]
	}
	if m := regexp.MustCompile(`Access Point:\s+([0-9A-Fa-f:]+)`).FindStringSubmatch(s); len(m) > 1 {
		stats.BSSID = m[1]
	}
	if m := regexp.MustCompile(`Frequency[=:](\d+\.?\d*)\s*GHz`).FindStringSubmatch(s); len(m) > 1 {
		ghz, _ := strconv.ParseFloat(m[1], 64)
		stats.FrequencyMHz = int(ghz * 1000)
		stats.Channel = freqToChannel(stats.FrequencyMHz)
	}
	if m := regexp.MustCompile(`Bit Rate[=:](\d+\.?\d*)\s*Mb/s`).FindStringSubmatch(s); len(m) > 1 {
		stats.LinkSpeedMbps, _ = strconv.ParseFloat(m[1], 64)
	}
	if m := regexp.MustCompile(`Signal level[=:](-?\d+)\s*dBm`).FindStringSubmatch(s); len(m) > 1 {
		stats.SignalDBm, _ = strconv.Atoi(m[1])
	}
	if m := regexp.MustCompile(`Noise level[=:](-?\d+)\s*dBm`).FindStringSubmatch(s); len(m) > 1 {
		stats.NoiseDBm, _ = strconv.Atoi(m[1])
	}

	if stats.SSID == "" {
		return nil, fmt.Errorf("not connected")
	}
	return stats, nil
}

func collectFromProc(iface string) (*Stats, error) {
	out, err := exec.Command("cat", "/proc/net/wireless").Output()
	if err != nil {
		return nil, fmt.Errorf("read /proc/net/wireless: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		name := strings.TrimSuffix(fields[0], ":")
		if name != iface {
			continue
		}

		stats := &Stats{}
		if len(fields) > 3 {
			val, _ := strconv.ParseFloat(strings.TrimSuffix(fields[3], "."), 64)
			if val < 0 {
				stats.SignalDBm = int(val)
			} else {
				stats.SignalDBm = int(val) - 110
			}
		}
		if len(fields) > 4 {
			val, _ := strconv.ParseFloat(strings.TrimSuffix(fields[4], "."), 64)
			if val < 0 && val > -200 {
				stats.NoiseDBm = int(val)
			}
		}
		return stats, nil
	}

	return nil, fmt.Errorf("interface %s not found", iface)
}

func pctToDbm(pct int) int {
	return -90 + (pct * 60 / 100)
}

func freqToChannel(freqMHz int) int {
	switch {
	case freqMHz == 2484:
		return 14
	case freqMHz >= 2412 && freqMHz <= 2472:
		return (freqMHz - 2407) / 5
	case freqMHz >= 5180 && freqMHz <= 5825:
		return (freqMHz - 5000) / 5
	case freqMHz >= 5955 && freqMHz <= 7115:
		return (freqMHz - 5950) / 5
	default:
		return 0
	}
}
