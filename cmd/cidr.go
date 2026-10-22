package cmd

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/netip"

	"github.com/spf13/cobra"
)

// CIDRInfo describes the address range covered by a prefix.
// Classless Inter-domain Routing: https://www.rfc-editor.org/info/rfc4632/
type CIDRInfo struct {
	Input     string   `json:"input"`
	Network   string   `json:"network"`
	Version   string   `json:"version"`
	PrefixLen int      `json:"prefix_len"`
	Netmask   string   `json:"netmask,omitempty"`
	Wildcard  string   `json:"wildcard,omitempty"`
	Broadcast string   `json:"broadcast,omitempty"`
	FirstHost string   `json:"first_host"`
	LastHost  string   `json:"last_host"`
	Total     *big.Int `json:"total_addresses"`
	Usable    *big.Int `json:"usable_hosts"`
}

func calcCIDR(input string) (CIDRInfo, error) {
	p, err := netip.ParsePrefix(input)
	if err != nil {
		return CIDRInfo{}, fmt.Errorf("invalid CIDR: %s", input)
	}
	if p.Addr().Is4In6() {
		return CIDRInfo{}, fmt.Errorf("IPv4-mapped IPv6 prefixes are not supported: %s", input)
	}
	p = p.Masked()

	first := p.Addr()
	last := lastAddr(p)
	hostBits := first.BitLen() - p.Bits()
	total := new(big.Int).Lsh(big.NewInt(1), uint(hostBits))

	info := CIDRInfo{
		Input:     input,
		Network:   p.String(),
		PrefixLen: p.Bits(),
		FirstHost: first.String(),
		LastHost:  last.String(),
		Total:     total,
		Usable:    new(big.Int).Set(total),
	}

	if first.Is6() {
		info.Version = "IPv6"
		return info, nil
	}

	info.Version = "IPv4"
	mask := net.CIDRMask(p.Bits(), 32)
	info.Netmask = net.IP(mask).String()
	wildcard := make(net.IP, 4)
	for i := range mask {
		wildcard[i] = ^mask[i]
	}
	info.Wildcard = wildcard.String()

	// /31 (RFC 3021) and /32 have no network or broadcast address to reserve.
	if hostBits >= 2 {
		info.Broadcast = last.String()
		info.FirstHost = first.Next().String()
		info.LastHost = last.Prev().String()
		info.Usable.Sub(total, big.NewInt(2))
	}
	return info, nil
}

// lastAddr returns the highest address in a masked prefix.
func lastAddr(p netip.Prefix) netip.Addr {
	b := p.Addr().As16()
	offset := 128 - p.Addr().BitLen()
	for i := offset + p.Bits(); i < 128; i++ {
		b[i/8] |= 1 << (7 - i%8)
	}
	a := netip.AddrFrom16(b)
	if p.Addr().Is4() {
		return a.Unmap()
	}
	return a
}

var cidrCmd = &cobra.Command{
	Use:   "cidr <prefix>",
	Short: "Calculate the address range of a CIDR block",
	Long: "Show network, netmask, broadcast, host range and size of a CIDR block.\n" +
		"Host bits in the input are cleared (10.1.2.3/24 -> 10.1.2.0/24).\n" +
		"Examples:\n" +
		"  netkit cidr 192.168.1.0/24\n" +
		"  netkit cidr 2001:db8::/48 --json",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		asJSON, err := cmd.Flags().GetBool("json")
		if err != nil {
			return err
		}

		info, err := calcCIDR(args[0])
		if err != nil {
			return err
		}

		if asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(info)
		}

		cmd.Println("Network:", info.Network)
		cmd.Println("  Version:   ", info.Version)
		if info.Version == "IPv4" {
			cmd.Println("  Netmask:   ", info.Netmask)
			cmd.Println("  Wildcard:  ", info.Wildcard)
			if info.Broadcast != "" {
				cmd.Println("  Broadcast: ", info.Broadcast)
			}
		}
		cmd.Printf("  Host range: %s - %s\n", info.FirstHost, info.LastHost)
		cmd.Println("  Addresses: ", info.Total)
		cmd.Println("  Usable:    ", info.Usable)
		return nil
	},
}

func init() {
	cidrCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(cidrCmd)
}
