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

// SetupTunRouteRules sets up TUN default route on Windows
// This is needed because sing-tun's AutoRoute doesn't work properly on Windows
func SetupTunRouteRules() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	// Add default route via TUN interface gateway with lower metric than physical interface
	// TUN gateway is 172.19.0.2, metric should be lower than physical interface (typically 35)
	cmd := exec.Command("route", "add", "0.0.0.0", "mask", "0.0.0.0", "172.19.0.2", "metric", "1")
	output, err := cmd.CombinedOutput()
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
	if runtime.GOOS != "windows" || !tunDefaultRouteAdded {
		return nil
	}

	cmd := exec.Command("route", "delete", "0.0.0.0")
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Warn("CleanupTunRouteRules: failed to delete TUN default route: %v, output: %s", err, string(output))
	} else {
		log.Info("[TUN] Removed default route via TUN")
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
		// Windows: Use PowerShell to get gateway, excluding TUN interface by InterfaceMetric
		cmd = exec.Command("powershell", "-Command",
			"(Get-NetRoute -DestinationPrefix '0.0.0.0/0' | Where-Object {$_.InterfaceMetric -gt 0} | Sort-Object InterfaceMetric | Select-Object -First 1).NextHop")
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
