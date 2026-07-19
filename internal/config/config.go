package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Probe settings
	ProbeInterval    time.Duration `json:"probe_interval"`
	ProbeTimeout     time.Duration `json:"probe_timeout"`
	PingTargets      []string      `json:"ping_targets"`
	DNSResolvers     []string      `json:"dns_resolvers"`
	DNSTestDomains   []string      `json:"dns_test_domains"`
	TCPFallbackPorts []int         `json:"tcp_fallback_ports"`

	// Speed test settings
	SpeedTestInterval time.Duration `json:"speed_test_interval"`
	SpeedTestEnabled  bool          `json:"speed_test_enabled"`

	// Diagnosis settings
	DiagnosisWindow    time.Duration `json:"diagnosis_window"`
	ConfirmationCount  int           `json:"confirmation_count"`
	BaselineWindowDays int           `json:"baseline_window_days"`

	// Storage
	DBPath string `json:"db_path"`

	// Notifications
	NotificationsEnabled bool `json:"notifications_enabled"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		ProbeInterval: 5 * time.Second,
		ProbeTimeout:  3 * time.Second,
		PingTargets: []string{
			"8.8.8.8",       // Google DNS
			"1.1.1.1",       // Cloudflare DNS
			"208.67.222.222", // OpenDNS
		},
		DNSResolvers: []string{
			"",       // system default resolver
			"8.8.8.8",
			"1.1.1.1",
		},
		DNSTestDomains: []string{
			"google.com",
			"cloudflare.com",
			"amazon.com",
		},
		TCPFallbackPorts: []int{443},

		SpeedTestInterval: 1 * time.Hour,
		SpeedTestEnabled:  true,

		DiagnosisWindow:    30 * time.Second,
		ConfirmationCount:  3,
		BaselineWindowDays: 7,

		DBPath: defaultDBPath(),

		NotificationsEnabled: true,
	}
}

// Load reads config from a JSON file, falling back to defaults for missing fields.
func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes the config to a JSON file.
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func defaultDBPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "netpulse", "netpulse.db")
}

// ConfigDir returns the directory where netpulse stores its data.
func ConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "netpulse")
}
