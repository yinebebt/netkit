package cmd

import (
	"testing"
)

func TestCalcCIDR(t *testing.T) {
	tests := []struct {
		input, network, netmask, broadcast, firstHost, lastHost, total, usable string
	}{
		{"192.168.1.0/24", "192.168.1.0/24", "255.255.255.0", "192.168.1.255", "192.168.1.1", "192.168.1.254", "256", "254"},
		{"10.1.2.3/8", "10.0.0.0/8", "255.0.0.0", "10.255.255.255", "10.0.0.1", "10.255.255.254", "16777216", "16777214"},
		{"172.16.5.4/30", "172.16.5.4/30", "255.255.255.252", "172.16.5.7", "172.16.5.5", "172.16.5.6", "4", "2"},
		{"0.0.0.0/0", "0.0.0.0/0", "0.0.0.0", "255.255.255.255", "0.0.0.1", "255.255.255.254", "4294967296", "4294967294"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := calcCIDR(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			check := func(field, got, want string) {
				t.Helper()
				if got != want {
					t.Errorf("%s = %q, want %q", field, got, want)
				}
			}
			check("Network", got.Network, tt.network)
			check("Netmask", got.Netmask, tt.netmask)
			check("Broadcast", got.Broadcast, tt.broadcast)
			check("FirstHost", got.FirstHost, tt.firstHost)
			check("LastHost", got.LastHost, tt.lastHost)
			check("Total", got.Total.String(), tt.total)
			check("Usable", got.Usable.String(), tt.usable)
		})
	}
}

func TestCalcCIDRInvalid(t *testing.T) {
	for _, input := range []string{"", "10.0.0.0", "10.0.0.0/33"} {
		t.Run(input, func(t *testing.T) {
			if _, err := calcCIDR(input); err == nil {
				t.Errorf("expected error for %q", input)
			}
		})
	}
}
