package probe

import (
	"fmt"
	"os"
	"strings"
)

// CheckPingPermissions verifies that unprivileged ICMP ping is allowed.
// On Linux, this requires /proc/sys/net/ipv4/ping_group_range to include the current user's GID.
func CheckPingPermissions() error {
	data, err := os.ReadFile("/proc/sys/net/ipv4/ping_group_range")
	if err != nil {
		return fmt.Errorf("cannot read ping_group_range: %w (ping probes may fail)", err)
	}

	fields := strings.Fields(strings.TrimSpace(string(data)))
	if len(fields) < 2 {
		return fmt.Errorf("unexpected ping_group_range format: %s", string(data))
	}

	// If range is "1 0" or similar where min > max, ping is disabled
	if fields[0] == "1" && fields[1] == "0" {
		return fmt.Errorf(
			"unprivileged ICMP ping is disabled on this system.\n" +
				"Fix with: sudo sysctl -w net.ipv4.ping_group_range=\"0 2147483647\"\n" +
				"To persist: echo 'net.ipv4.ping_group_range = 0 2147483647' | sudo tee /etc/sysctl.d/99-ping.conf")
	}

	return nil
}
