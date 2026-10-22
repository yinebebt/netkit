package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// NetkitIP relevant info about the IP address.
type NetkitIP struct {
	IP      string
	Version string
	Scope   string
	Class   string
	Geo     *IPGeoInfo
}

// IPGeoInfo geolocation information.
type IPGeoInfo struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	IP         string `json:"ip"`
	Continent  string `json:"continent"`
	Country    string `json:"country"`
	Region     string `json:"region"`
	Connection struct {
		Org string `json:"org"`
		ISP string `json:"isp"`
	} `json:"connection"`
}

func fetchGeolocation(ip string) (*IPGeoInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://ipwho.is/%s", ip), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ipwho.is returned status %d", resp.StatusCode)
	}

	var geo IPGeoInfo
	if err := json.NewDecoder(resp.Body).Decode(&geo); err != nil {
		return nil, err
	}

	if !geo.Success {
		if geo.Message == "" {
			geo.Message = "unknown error"
		}
		return nil, fmt.Errorf("geolocation lookup failed: %s", geo.Message)
	}

	return &geo, nil
}

func ipv4Class(ip net.IP) string {
	// net.IP uses a 16-byte representation. For IPv4-mapped IPv6 addresses
	// (RFC 4291), the actual IPv4 address occupies the final 4 bytes.
	// Index 12 corresponds to the first octet of the IPv4 address.
	firstOctet := ip[12]
	switch {
	case firstOctet < 128:
		return "Class A"
	case firstOctet < 192:
		return "Class B"
	case firstOctet < 224:
		return "Class C"
	case firstOctet < 240:
		return "Class D (Multicast)"
	default:
		return "Class E (Reserved)"
	}
}

var infoCmd = &cobra.Command{
	Use:   "info <ip>",
	Short: "Show information about an IP address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		asJSON, err := cmd.Flags().GetBool("json")
		if err != nil {
			return err
		}
		noGeo, err := cmd.Flags().GetBool("no-geo")
		if err != nil {
			return err
		}

		ipStr := args[0]
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return fmt.Errorf("invalid IP address: %s", ipStr)
		}

		nk := NetkitIP{
			IP: ipStr,
		}

		if ip.To4() != nil {
			nk.Version = "IPv4"
			nk.Class = ipv4Class(ip)
		} else {
			nk.Version = "IPv6"
		}

		switch {
		case ip.IsLoopback():
			nk.Scope = "Loopback"
		case ip.IsPrivate():
			nk.Scope = "Private"
		case ip.IsMulticast():
			nk.Scope = "Multicast"
		case ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast():
			nk.Scope = "Link-Local"
		case ip.IsUnspecified():
			nk.Scope = "Unspecified"
		default:
			nk.Scope = "Public"
			if !noGeo {
				geo, geoErr := fetchGeolocation(ipStr)
				if geoErr == nil {
					nk.Geo = geo
				} else {
					cmd.PrintErrf("Warning: could not fetch geolocation info: %v\n", geoErr)
				}
			}
		}

		if asJSON {
			return printIPInfoJSON(cmd, nk)
		}

		printIPInfo(cmd, nk)
		return nil
	},
}

func init() {
	infoCmd.Flags().Bool("json", false, "Output as JSON")
	infoCmd.Flags().Bool("no-geo", false, "Skip geolocation lookup for public IPs")
	rootCmd.AddCommand(infoCmd)
}

func printIPInfo(cmd *cobra.Command, nk NetkitIP) {
	cmd.Println("IP Address:", nk.IP)
	cmd.Println("  Version:", nk.Version)
	if nk.Version == "IPv4" {
		cmd.Println("  Class:", nk.Class)
	}
	cmd.Println("  Scope:", nk.Scope)

	if nk.Scope == "Public" && nk.Geo != nil {
		cmd.Println("  Geo Info:")
		cmd.Printf("    Continent: %s\n", nk.Geo.Continent)
		cmd.Printf("    Country:   %s\n", nk.Geo.Country)
		cmd.Printf("    Region:    %s\n", nk.Geo.Region)
		cmd.Printf("    ISP:       %s\n", nk.Geo.Connection.ISP)
		cmd.Printf("    Org:       %s\n", nk.Geo.Connection.Org)
	}
}

func printIPInfoJSON(cmd *cobra.Command, nk NetkitIP) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(nk)
}
