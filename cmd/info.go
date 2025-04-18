package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"net"
	"net/http"
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
	resp, err := http.Get(fmt.Sprintf("https://ipwho.is/%s", ip))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var geo IPGeoInfo
	if err := json.NewDecoder(resp.Body).Decode(&geo); err != nil {
		return nil, err
	}

	if geo.IP == "" {
		return nil, fmt.Errorf("no data for IP %s", ip)
	}

	return &geo, nil
}

func iPv4Class(ip net.IP) string {
	firstOctet := ip[12] // since net.IP size is extended to 16byte.
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
	Run: func(cmd *cobra.Command, args []string) {
		ipStr := args[0]
		ip := net.ParseIP(ipStr)
		if ip == nil {
			fmt.Println("Invalid IP address.")
			return
		}

		nk := NetkitIP{
			IP: ipStr,
		}

		if ip.To4() != nil {
			nk.Version = "IPv4"
			nk.Class = iPv4Class(ip)
		} else {
			nk.Version = "IPv6"
		}

		if ip.IsPrivate() {
			nk.Scope = "Private"
		} else {
			nk.Scope = "Public"
			geo, err := fetchGeolocation(ipStr)
			if err == nil {
				nk.Geo = geo
			}
		}

		printIPInfo(nk)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func printIPInfo(nk NetkitIP) {
	fmt.Println("IP Address:", nk.IP)
	fmt.Println("  Version:", nk.Version)
	if nk.Version == "IPv4" {
		fmt.Println("  Class:", nk.Class)
	}
	fmt.Println("  Scope:", nk.Scope)

	if nk.Scope == "Public" && nk.Geo != nil {
		fmt.Println("  Geo Info:")
		fmt.Printf("    Continent: %s\n", nk.Geo.Continent)
		fmt.Printf("    Country:   %s\n", nk.Geo.Country)
		fmt.Printf("    Region:    %s\n", nk.Geo.Region)
		fmt.Printf("    ISP:       %s\n", nk.Geo.Connection.ISP)
		fmt.Printf("    Org:       %s\n", nk.Geo.Connection.Org)
	}
}
