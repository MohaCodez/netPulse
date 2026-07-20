package scanner

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Device represents a discovered device on the local network.
type Device struct {
	IP        string    `json:"ip"`
	MAC       string    `json:"mac"`
	Vendor    string    `json:"vendor"`
	Hostname  string    `json:"hostname"`
	Interface string    `json:"interface"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	IsGateway bool      `json:"is_gateway"`
	IsLocal   bool      `json:"is_local"`
}

// Scanner discovers devices on the local network.
type Scanner struct {
	mu      sync.RWMutex
	devices map[string]*Device // keyed by MAC
}

// NewScanner creates a network scanner.
func NewScanner() *Scanner {
	return &Scanner{
		devices: make(map[string]*Device),
	}
}

// Scan performs a network scan and returns all discovered devices.
func (s *Scanner) Scan(ctx context.Context) ([]*Device, error) {
	// Get local interface info
	localIP, subnet, iface, err := getLocalNetwork()
	if err != nil {
		return nil, fmt.Errorf("get local network: %w", err)
	}

	// Get gateway
	gateway := getGateway()

	// Ping sweep to populate ARP table
	s.pingSweep(ctx, subnet)

	// Wait a moment for ARP entries to populate
	time.Sleep(500 * time.Millisecond)

	// Read ARP table
	arpEntries, err := readARPTable()
	if err != nil {
		return nil, fmt.Errorf("read arp table: %w", err)
	}

	now := time.Now()

	s.mu.Lock()
	// Update known devices
	for _, entry := range arpEntries {
		mac := strings.ToUpper(entry.MAC)
		if mac == "00:00:00:00:00:00" || mac == "" {
			continue
		}

		if dev, exists := s.devices[mac]; exists {
			dev.LastSeen = now
			dev.IP = entry.IP // IP might change (DHCP)
		} else {
			s.devices[mac] = &Device{
				IP:        entry.IP,
				MAC:       mac,
				Vendor:    LookupVendor(mac),
				Hostname:  resolveHostname(entry.IP),
				Interface: iface,
				FirstSeen: now,
				LastSeen:  now,
				IsGateway: entry.IP == gateway,
				IsLocal:   entry.IP == localIP,
			}
		}
	}

	// Add self
	localMAC := getLocalMAC(iface)
	if localMAC != "" {
		mac := strings.ToUpper(localMAC)
		if _, exists := s.devices[mac]; !exists {
			hostname, _ := os.Hostname()
			s.devices[mac] = &Device{
				IP:        localIP,
				MAC:       mac,
				Vendor:    LookupVendor(mac),
				Hostname:  hostname,
				Interface: iface,
				FirstSeen: now,
				LastSeen:  now,
				IsLocal:   true,
			}
		} else {
			s.devices[mac].LastSeen = now
		}
	}

	// Build result
	var result []*Device
	for _, dev := range s.devices {
		result = append(result, dev)
	}
	s.mu.Unlock()

	return result, nil
}

// GetDevices returns currently known devices without rescanning.
func (s *Scanner) GetDevices() []*Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Device
	for _, dev := range s.devices {
		result = append(result, dev)
	}
	return result
}

// pingSweep sends ICMP pings to all IPs in the subnet to populate ARP cache.
func (s *Scanner) pingSweep(ctx context.Context, subnet string) {
	// Parse CIDR
	ip, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return
	}

	// Calculate all IPs in subnet
	var ips []string
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		ips = append(ips, ip.String())
	}

	// Skip network and broadcast
	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	// Ping all in parallel (just to populate ARP, don't care about responses)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 50) // limit concurrent pings

	for _, target := range ips {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			conn, err := net.DialTimeout("udp", ip+":1", 200*time.Millisecond)
			if err == nil {
				conn.Close()
			}
		}(target)
	}
	wg.Wait()
}

type arpEntry struct {
	IP  string
	MAC string
}

func readARPTable() ([]arpEntry, error) {
	data, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return nil, err
	}

	var entries []arpEntry
	lines := strings.Split(string(data), "\n")
	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		ip := fields[0]
		mac := fields[3]
		if mac == "00:00:00:00:00:00" {
			continue
		}
		entries = append(entries, arpEntry{IP: ip, MAC: mac})
	}
	return entries, nil
}

func getLocalNetwork() (localIP, subnet, iface string, err error) {
	// Find the default interface with an IP
	out, err := exec.Command("ip", "-4", "route", "show", "default").Output()
	if err != nil {
		return "", "", "", err
	}

	// Parse: "default via 192.168.1.1 dev wlp9s0 ..."
	fields := strings.Fields(string(out))
	for i, f := range fields {
		if f == "dev" && i+1 < len(fields) {
			iface = fields[i+1]
			break
		}
	}

	if iface == "" {
		return "", "", "", fmt.Errorf("no default interface found")
	}

	// Get IP and subnet of that interface
	out, err = exec.Command("ip", "-4", "addr", "show", iface).Output()
	if err != nil {
		return "", "", "", err
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "inet ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				cidr := parts[1] // e.g., "192.168.1.8/24"
				ip, ipNet, err := net.ParseCIDR(cidr)
				if err == nil {
					localIP = ip.String()
					subnet = ipNet.String()
				}
			}
			break
		}
	}

	if localIP == "" {
		return "", "", "", fmt.Errorf("no IP found on %s", iface)
	}

	return localIP, subnet, iface, nil
}

func getGateway() string {
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(out))
	for i, f := range fields {
		if f == "via" && i+1 < len(fields) {
			return fields[i+1]
		}
	}
	return ""
}

func getLocalMAC(iface string) string {
	data, err := os.ReadFile(fmt.Sprintf("/sys/class/net/%s/address", iface))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func resolveHostname(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
