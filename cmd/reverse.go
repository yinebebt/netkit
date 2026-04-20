package cmd

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"
)

var reverseCmd = &cobra.Command{
	Use:   "reverse <ip>",
	Short: "Resolve IP address to hostnames (PTR / reverse DNS)",
	Long: "Reverse DNS lookup (PTR):\n" +
		"  IP -> hostname(s)\n" +
		"Example:\n" +
		"  netkit reverse 8.8.8.8",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ipStr := args[0]
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return fmt.Errorf("invalid IP address: %s", ipStr)
		}

		names, err := net.LookupAddr(ipStr)
		if err != nil {
			return fmt.Errorf("reverse lookup failed for %s: %w", ipStr, err)
		}
		if len(names) == 0 {
			return fmt.Errorf("no PTR records found for %s", ipStr)
		}

		cmd.Printf("Reverse DNS (PTR) for %s:\n", ipStr)
		for _, name := range names {
			cmd.Printf("  - %s\n", name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reverseCmd)
}
