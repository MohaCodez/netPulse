package probe

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
)

// detectGateway reads the default gateway from /proc/net/route on Linux.
func detectGateway() (string, error) {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return "", fmt.Errorf("read route table: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		// Destination == 00000000 means default route
		if fields[1] == "00000000" {
			gateway := parseHexIP(fields[2])
			if gateway != "" {
				return gateway, nil
			}
		}
	}

	return "", fmt.Errorf("no default gateway found")
}

// parseHexIP converts a hex-encoded IP from /proc/net/route to dotted notation.
// /proc/net/route stores IPs as little-endian 32-bit hex values on x86.
func parseHexIP(hexStr string) string {
	if len(hexStr) != 8 {
		return ""
	}

	b, err := hex.DecodeString(hexStr)
	if err != nil || len(b) != 4 {
		return ""
	}

	// /proc/net/route uses host byte order (little-endian on x86)
	ipInt := binary.LittleEndian.Uint32(b)
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, ipInt)

	return ip.String()
}
