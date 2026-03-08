package plugin

import (
	"net"
	"testing"
)

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		name           string
		ip             string
		allowLocalhost bool
		wantBlocked    bool
	}{
		// Loopback
		{name: "loopback 127.0.0.1 blocked", ip: "127.0.0.1", wantBlocked: true},
		{name: "loopback 127.0.0.1 allowed with flag", ip: "127.0.0.1", allowLocalhost: true, wantBlocked: false},
		{name: "loopback 127.99.0.1 blocked", ip: "127.99.0.1", wantBlocked: true},
		{name: "loopback 127.99.0.1 allowed with flag", ip: "127.99.0.1", allowLocalhost: true, wantBlocked: false},

		// IPv6 loopback
		{name: "ipv6 loopback blocked", ip: "::1", wantBlocked: true},
		{name: "ipv6 loopback allowed with flag", ip: "::1", allowLocalhost: true, wantBlocked: false},

		// RFC 1918 private ranges
		{name: "10.0.0.0/8", ip: "10.0.0.1", wantBlocked: true},
		{name: "10.255.255.255", ip: "10.255.255.255", wantBlocked: true},
		{name: "172.16.0.0/12", ip: "172.16.0.1", wantBlocked: true},
		{name: "172.31.255.255", ip: "172.31.255.255", wantBlocked: true},
		{name: "192.168.0.0/16", ip: "192.168.1.1", wantBlocked: true},
		{name: "192.168.255.255", ip: "192.168.255.255", wantBlocked: true},

		// RFC 1918 not affected by AllowLocalhost
		{name: "10.x with localhost flag still blocked", ip: "10.0.0.1", allowLocalhost: true, wantBlocked: true},
		{name: "192.168.x with localhost flag still blocked", ip: "192.168.1.1", allowLocalhost: true, wantBlocked: true},

		// Link-local (includes cloud metadata 169.254.169.254)
		{name: "link-local", ip: "169.254.0.1", wantBlocked: true},
		{name: "cloud metadata", ip: "169.254.169.254", wantBlocked: true},

		// IPv6 unique local and link-local
		{name: "ipv6 unique local", ip: "fd00::1", wantBlocked: true},
		{name: "ipv6 link-local", ip: "fe80::1", wantBlocked: true},

		// Other reserved ranges
		{name: "this network 0.0.0.0/8", ip: "0.0.0.1", wantBlocked: true},
		{name: "CGN 100.64.0.0/10", ip: "100.64.0.1", wantBlocked: true},
		{name: "IETF 192.0.0.0/24", ip: "192.0.0.1", wantBlocked: true},
		{name: "benchmarking 198.18.0.0/15", ip: "198.18.0.1", wantBlocked: true},

		// Public IPs should be allowed
		{name: "public 8.8.8.8", ip: "8.8.8.8", wantBlocked: false},
		{name: "public 1.1.1.1", ip: "1.1.1.1", wantBlocked: false},
		{name: "public 93.184.216.34", ip: "93.184.216.34", wantBlocked: false},
		{name: "public ipv6 2001:db8::1", ip: "2001:db8::1", wantBlocked: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			if ip == nil {
				t.Fatalf("failed to parse test IP: %s", tc.ip)
			}
			got := isBlockedIP(ip, tc.allowLocalhost)
			if got != tc.wantBlocked {
				t.Errorf("isBlockedIP(%s, allowLocalhost=%v) = %v, want %v", tc.ip, tc.allowLocalhost, got, tc.wantBlocked)
			}
		})
	}
}

func TestBlockedCIDRsInitialized(t *testing.T) {
	if len(blockedCIDRs) != 12 {
		t.Errorf("expected 12 blocked CIDR ranges, got %d", len(blockedCIDRs))
	}
}

func TestSSRFSafeDialer(t *testing.T) {
	cfg := RequestEngineConfig{AllowLocalhost: false}
	dialer := ssrfSafeDialer(cfg)

	if dialer.Timeout == 0 {
		t.Error("dialer timeout should be non-zero")
	}
	if dialer.Control == nil {
		t.Error("dialer Control hook should be set")
	}
}
