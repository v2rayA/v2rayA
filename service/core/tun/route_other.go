//go:build !linux
// +build !linux

package tun

import (
	"fmt"
	"net/netip"
	"os/exec"
	"runtime"
	"strings"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

var excludedRoutes []netip.Prefix
var tunDefaultRouteAdded bool
var loopbackRouteAdded bool
var loopback6RouteAdded bool

// SetupTunRouteRules sets up TUN default route on Windows
// This is needed because sing-tun's AutoRoute doesn't work properly on Windows
func SetupTunRouteRules() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	// Add explicit IPv4 loopback route
	cmd := exec.Command("route", "add", "127.0.0.0", "mask", "255.0.0.0", "127.0.0.1", "metric", "0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(output), "对象已存在") && !strings.Contains(string(output), "already exists") {
			log.Warn("SetupTunRouteRules: failed to add loopback route: %v, output: %s", err, string(output))
		} else {
			log.Info("[TUN] Loopback route already exists")
		}
	} else {
		loopbackRouteAdded = true
		log.Info("[TUN] Added loopback route 127.0.0.0/8 via 127.0.0.1 metric 0")
	}

	// Add explicit IPv6 loopback route (::1/128) using netsh
	cmd6 := exec.Command("netsh", "interface", "ipv6", "add", "route", "::1/128", "interface=loopback", "nexthop=::1", "metric=0")
	output6, err6 := cmd6.CombinedOutput()
	if err6 != nil {
		if !strings.Contains(string(output6), "对象已存在") && !strings.Contains(string(output6), "Element already exists") {
			log.Warn("SetupTunRouteRules: failed to add IPv6 loopback route: %v, output: %s", err6, string(output6))
		} else {
			log.Info("[TUN] IPv6 loopback route already exists")
		}
	} else {
		loopback6RouteAdded = true
		log.Info("[TUN] Added IPv6 loopback route ::1/128 via ::1 metric 0")
	}

	// Add default route via TUN interface gateway with lower metric than physical interface
	// TUN gateway is 172.19.0.2, metric should be lower than physical interface (typically 35)
	cmd = exec.Command("route", "add", "0.0.0.0", "mask", "0.0.0.0", "172.19.0.2", "metric", "1")
	output, err = cmd.CombinedOutput()
	if err != nil {
		// Check if route already exists
		if !strings.Contains(string(output), "对象已存在") && !strings.Contains(string(output), "already exists") {
			log.Warn("SetupTunRouteRules: failed to add default route via TUN: %v, output: %s", err, string(output))
			return err
		}
	}
	tunDefaultRouteAdded = true
	log.Info("[TUN] Added default route 0.0.0.0/0 via 172.19.0.2 metric 1")
	return nil
}

// CleanupTunRouteRules removes TUN default route on Windows
func CleanupTunRouteRules() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	// Remove IPv4 loopback route if it was added
	if loopbackRouteAdded {
		cmd := exec.Command("route", "delete", "127.0.0.0", "mask", "255.0.0.0")
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Warn("CleanupTunRouteRules: failed to delete loopback route: %v, output: %s", err, string(output))
		} else {
			log.Info("[TUN] Removed loopback route 127.0.0.0/8")
		}
		loopbackRouteAdded = false
	}
	// Remove IPv6 loopback route if it was added
	if loopback6RouteAdded {
		cmd6 := exec.Command("netsh", "interface", "ipv6", "delete", "route", "::1/128")
		if output6, err6 := cmd6.CombinedOutput(); err6 != nil {
			log.Warn("CleanupTunRouteRules: failed to delete IPv6 loopback route: %v, output: %s", err6, string(output6))
		} else {
			log.Info("[TUN] Removed IPv6 loopback route ::1/128")
		}
		loopback6RouteAdded = false
	}

	// Remove default route if it was added
	if !tunDefaultRouteAdded {
		return nil
	}

	// Specify the gateway to only delete the TUN route, not the physical interface route
	cmd := exec.Command("route", "delete", "0.0.0.0", "mask", "0.0.0.0", "172.19.0.2")
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Warn("CleanupTunRouteRules: failed to delete TUN default route: %v, output: %s", err, string(output))
	} else {
		log.Info("[TUN] Removed default route via TUN (172.19.0.2)")
	}
	tunDefaultRouteAdded = false
	return nil
}

// SetupExcludeRoutes adds static routes for server addresses to bypass TUN
// This is critical for Windows/macOS where fwmark is not available
func SetupExcludeRoutes(excludeAddrs []netip.Prefix) error {
	if len(excludeAddrs) == 0 {
		return nil
	}

	excludedRoutes = excludeAddrs
	gateway, err := getDefaultGateway()
	if err != nil {
		log.Warn("SetupExcludeRoutes: failed to get default gateway: %v", err)
		return err
	}

	for _, prefix := range excludeAddrs {
		addr := prefix.Addr()
		var cmd string
		var cmdArgs []string

		if runtime.GOOS == "windows" {
			// Windows: route add <ip> mask <mask> <gateway>
			if addr.Is4() {
				mask := netmaskFromPrefix(prefix)
				cmd = "route"
				cmdArgs = []string{"add", addr.String(), "mask", mask, gateway, "metric", "1"}
			} else {
				// IPv6 on Windows using netsh
				cmd = "netsh"
				cmdArgs = []string{"interface", "ipv6", "add", "route", prefix.String(), "nexthop=" + gateway, "metric=1"}
			}
		} else {
			// macOS/BSD: route add <ip> <gateway>
			if addr.Is4() {
				cmd = "route"
				cmdArgs = []string{"add", addr.String(), gateway}
			} else {
				cmd = "route"
				cmdArgs = []string{"add", "-inet6", addr.String(), gateway}
			}
		}

		execCmd := exec.Command(cmd, cmdArgs...)
		if output, err := execCmd.CombinedOutput(); err != nil {
			log.Warn("SetupExcludeRoutes: failed to add route for %s: %v, output: %s", addr.String(), err, string(output))
		} else {
			log.Info("Added route for server %s via %s", addr.String(), gateway)
		}
	}

	return nil
}

// CleanupExcludeRoutes removes static routes added for server addresses
func CleanupExcludeRoutes() error {
	if len(excludedRoutes) == 0 {
		return nil
	}

	for _, prefix := range excludedRoutes {
		addr := prefix.Addr()
		var cmd string
		var cmdArgs []string

		if runtime.GOOS == "windows" {
			if addr.Is4() {
				cmd = "route"
				cmdArgs = []string{"delete", addr.String()}
			} else {
				cmd = "netsh"
				cmdArgs = []string{"interface", "ipv6", "delete", "route", prefix.String()}
			}
		} else {
			// macOS/BSD
			if addr.Is4() {
				cmd = "route"
				cmdArgs = []string{"delete", addr.String()}
			} else {
				cmd = "route"
				cmdArgs = []string{"delete", "-inet6", addr.String()}
			}
		}

		execCmd := exec.Command(cmd, cmdArgs...)
		if output, err := execCmd.CombinedOutput(); err != nil {
			log.Warn("CleanupExcludeRoutes: failed to delete route for %s: %v, output: %s", addr.String(), err, string(output))
		}
	}

	excludedRoutes = nil
	return nil
}

// getDefaultGateway retrieves the default gateway IP address
func getDefaultGateway() (string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// Windows: Get the gateway, excluding TUN gateway (172.19.0.2)
		// Sort by InterfaceMetric to prefer the physical interface
		cmd = exec.Command("powershell", "-Command",
			"(Get-NetRoute -DestinationPrefix '0.0.0.0/0' | Where-Object {$_.NextHop -ne '172.19.0.2' -and $_.NextHop -ne '0.0.0.0'} | Sort-Object InterfaceMetric | Select-Object -First 1).NextHop")
	} else {
		// macOS/BSD: route -n get default
		cmd = exec.Command("sh", "-c", "route -n get default | grep gateway | awk '{print $2}'")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get gateway: %w, output: %s", err, string(output))
	}

	gateway := trimOutput(string(output))

	if gateway == "" {
		return "", fmt.Errorf("empty gateway")
	}

	return gateway, nil
}

// netmaskFromPrefix converts a prefix length to netmask string (IPv4 only)
func netmaskFromPrefix(prefix netip.Prefix) string {
	bits := prefix.Bits()
	switch bits {
	case 32:
		return "255.255.255.255"
	case 31:
		return "255.255.255.254"
	case 30:
		return "255.255.255.252"
	case 29:
		return "255.255.255.248"
	case 28:
		return "255.255.255.240"
	case 27:
		return "255.255.255.224"
	case 26:
		return "255.255.255.192"
	case 25:
		return "255.255.255.128"
	case 24:
		return "255.255.255.0"
	case 16:
		return "255.255.0.0"
	case 8:
		return "255.0.0.0"
	default:
		// Generic calculation for other cases
		mask := ^uint32(0) << (32 - bits)
		return fmt.Sprintf("%d.%d.%d.%d",
			byte(mask>>24), byte(mask>>16), byte(mask>>8), byte(mask))
	}
}

// trimOutput removes whitespace from command output
func trimOutput(s string) string {
	return strings.TrimSpace(s)
}

// SetupTunDNS sets DNS servers for TUN interface on Windows
func SetupTunDNS(dnsServers []netip.Addr, tunName string) error {
	if runtime.GOOS != "windows" || len(dnsServers) == 0 {
		return nil
	}

	// Get the actual interface name (may have index suffix like v2raya-tun0)
	interfaceName, err := getTunInterfaceName(tunName)
	if err != nil {
		log.Warn("SetupTunDNS: failed to get interface name: %v", err)
		return err
	}

	// Build DNS server list
	var dnsIPv4, dnsIPv6 []string
	for _, dns := range dnsServers {
		if dns.Is4() {
			dnsIPv4 = append(dnsIPv4, dns.String())
		} else if dns.Is6() {
			dnsIPv6 = append(dnsIPv6, dns.String())
		}
	}

	// Set IPv4 DNS servers
	if len(dnsIPv4) > 0 {
		dnsListIPv4 := strings.Join(dnsIPv4, ",")
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("Set-DnsClientServerAddress -InterfaceAlias '%s' -ServerAddresses %s", interfaceName, dnsListIPv4))
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Warn("SetupTunDNS: failed to set IPv4 DNS: %v, output: %s", err, string(output))
			return err
		}
		log.Info("[TUN] Set interface '%s' IPv4 DNS servers: %s", interfaceName, dnsListIPv4)
	}

	// Set IPv6 DNS servers
	if len(dnsIPv6) > 0 {
		dnsListIPv6 := strings.Join(dnsIPv6, ",")
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("Set-DnsClientServerAddress -InterfaceAlias '%s' -ServerAddresses %s", interfaceName, dnsListIPv6))
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Warn("SetupTunDNS: failed to set IPv6 DNS: %v, output: %s", err, string(output))
		}
		log.Info("[TUN] Set interface '%s' IPv6 DNS servers: %s", interfaceName, dnsListIPv6)
	}

	return nil
}

// CleanupTunDNS resets DNS servers for TUN interface on Windows
func CleanupTunDNS(tunName string) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	// Get the actual interface name
	interfaceName, err := getTunInterfaceName(tunName)
	if err != nil {
		// Interface may already be deleted, this is not an error
		return nil
	}

	// Reset to automatic DNS (DHCP)
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Set-DnsClientServerAddress -InterfaceAlias '%s' -ResetServerAddresses", interfaceName))
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Warn("CleanupTunDNS: failed to reset DNS: %v, output: %s", err, string(output))
	}
	log.Info("[TUN] Reset interface '%s' DNS servers", interfaceName)
	return nil
}

// getTunInterfaceName gets the actual TUN interface name on Windows
// The interface name may have an index suffix (e.g., v2raya-tun0)
func getTunInterfaceName(baseName string) (string, error) {
	if runtime.GOOS != "windows" {
		return baseName, nil
	}

	// Try to find interface by partial name match
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("(Get-NetAdapter | Where-Object {$_.Name -like '*%s*' -and $_.Status -eq 'Up'} | Select-Object -First 1).Name", baseName))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get interface name: %w, output: %s", err, string(output))
	}

	interfaceName := trimOutput(string(output))
	if interfaceName == "" {
		return "", fmt.Errorf("interface not found for base name: %s", baseName)
	}

	return interfaceName, nil
}
