package cmd

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/spf13/cobra"
)

type ResolveResult struct {
	Domain string   `json:"domain"`
	IPv4   []string `json:"ipv4,omitempty"`
	IPv6   []string `json:"ipv6,omitempty"`
}

var resolveCmd = &cobra.Command{
	Use:   "resolve <domain>",
	Short: "Resolve domain name to IP addresses (A/AAAA)",
	Long: "Forward DNS lookup:\n" +
		"  domain -> IP address(es)\n" +
		"Example:\n" +
		"  netkit resolve example.com",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		asJSON, err := cmd.Flags().GetBool("json")
		if err != nil {
			return err
		}

		domain := args[0]
		ips, err := net.LookupIP(domain)
		if err != nil {
			return fmt.Errorf("domain lookup failed for %s: %w", domain, err)
		}
		if len(ips) == 0 {
			return fmt.Errorf("no IP addresses found for %s", domain)
		}

		result := ResolveResult{Domain: domain}
		for _, ip := range ips {
			if v4 := ip.To4(); v4 != nil {
				result.IPv4 = append(result.IPv4, v4.String())
				continue
			}
			result.IPv6 = append(result.IPv6, ip.String())
		}

		if asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		cmd.Printf("Forward DNS for %s:\n", domain)
		if len(result.IPv4) > 0 {
			cmd.Println("  IPv4 (A):")
			for _, ip := range result.IPv4 {
				cmd.Printf("    - %s\n", ip)
			}
		}
		if len(result.IPv6) > 0 {
			cmd.Println("  IPv6 (AAAA):")
			for _, ip := range result.IPv6 {
				cmd.Printf("    - %s\n", ip)
			}
		}

		return nil
	},
}

func init() {
	resolveCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(resolveCmd)
}
