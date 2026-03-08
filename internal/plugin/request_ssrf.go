package plugin

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

// blockedCIDRs contains CIDR ranges that are blocked for outbound requests.
// Initialized once in init(). Covers RFC 1918 private, link-local, loopback,
// CGN, IETF assignments, benchmarking, and IPv6 equivalents.
var blockedCIDRs []*net.IPNet

// loopbackCIDRs tracks which blockedCIDRs entries are loopback ranges.
// These are skipped when AllowLocalhost is true.
var loopbackCIDRs = map[string]bool{
	"127.0.0.0/8": true,
	"::1/128":     true,
}

func init() {
	cidrs := []string{
		"127.0.0.0/8",    // loopback (allowed when AllowLocalhost=true)
		"10.0.0.0/8",     // RFC 1918 private
		"172.16.0.0/12",  // RFC 1918 private
		"192.168.0.0/16", // RFC 1918 private
		"169.254.0.0/16", // link-local (includes cloud metadata 169.254.169.254)
		"fc00::/7",       // IPv6 unique local
		"fe80::/10",      // IPv6 link-local
		"::1/128",        // IPv6 loopback (allowed when AllowLocalhost=true)
		"0.0.0.0/8",      // "this" network
		"100.64.0.0/10",  // shared address space (CGN)
		"192.0.0.0/24",   // IETF protocol assignments
		"198.18.0.0/15",  // benchmarking
	}

	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Sprintf("invalid CIDR in blockedCIDRs: %s", cidr))
		}
		blockedCIDRs = append(blockedCIDRs, ipNet)
	}
}

// isBlockedIP returns true if the IP falls within any blocked CIDR range.
// When allowLocalhost is true, 127.0.0.0/8 and ::1/128 are permitted.
func isBlockedIP(ip net.IP, allowLocalhost bool) bool {
	for _, cidr := range blockedCIDRs {
		if cidr.Contains(ip) {
			if allowLocalhost && loopbackCIDRs[cidr.String()] {
				continue
			}
			return true
		}
	}
	return false
}

// ssrfSafeDialer returns a net.Dialer with a Control hook that blocks
// connections to private/reserved IP ranges. The check runs after DNS
// resolution (in the Control callback), so DNS rebinding attacks that
// resolve a public domain to a private IP are caught.
func ssrfSafeDialer(cfg RequestEngineConfig) *net.Dialer {
	return &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return fmt.Errorf("failed to parse address %q: %w", address, err)
			}
			ip := net.ParseIP(host)
			if ip == nil {
				return fmt.Errorf("invalid IP in resolved address: %q", host)
			}
			if isBlockedIP(ip, cfg.AllowLocalhost) {
				return fmt.Errorf("request to private/reserved IP blocked: %s", ip)
			}
			return nil
		},
	}
}
