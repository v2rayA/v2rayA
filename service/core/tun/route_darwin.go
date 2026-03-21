//go:build darwin
// +build darwin

package tun

import (
	"fmt"
	"net/netip"
	"os/exec"
	"strings"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

var excludedRoutes []netip.Prefix

// SetupTunRouteRules is a no-op on macOS.
//
// macOS does not support fwmark policy routing. sing-tun's AutoRoute
// will automatically direct traffic to the TUN interface via routing table priorities.
func SetupTunRouteRules() error {
	return nil
}

// CleanupTunRouteRules is a no-op on macOS.
func CleanupTunRouteRules() error {
	return nil
}

// setTunRouteAutoMode is a no-op on macOS.
func setTunRouteAutoMode(_ bool) {}

// DynAddExcludeRoute is a no-op on macOS (sing-tun's AutoRoute handles exclusions).
func DynAddExcludeRoute(_ netip.Addr) {}

// SetupExcludeRoutes adds "bypass TUN" static host routes for proxy server addresses on macOS.
//
// macOS lacks fwmark, so we must explicitly add routes via the physical gateway for each
// server IP to prevent proxy traffic from being captured by the TUN again.
func SetupExcludeRoutes(addrs []netip.Prefix) error {
	if len(addrs) == 0 {
		return nil
	}
	excludedRoutes = addrs

	gw, err := getDefaultGateway()
	if err != nil {
		log.Warn("[TUN][macOS] Failed to get default gateway: %v", err)
		return err
	}

	for _, prefix := range addrs {
		addr := prefix.Addr()
		var out []byte
		if addr.Is4() {
			// macOS route add -host <ip> <gw> (host route, exact match for single IP)
			out, err = exec.Command("route", "add", "-host", addr.String(), gw).CombinedOutput()
		} else {
			// macOS route add -inet6 <prefix> <gw>
			out, err = exec.Command("route", "add", "-inet6", prefix.String(), gw).CombinedOutput()
		}
		if err != nil {
			s := string(out)
			// Ignore "route already exists" errors
			if !strings.Contains(s, "File exists") && !strings.Contains(s, "already exists") {
				log.Warn("[TUN][macOS] Failed to add exclude route %s: %v, output: %s", addr, err, s)
			} else {
				log.Info("[TUN][macOS] Exclude route %s already exists", addr)
			}
		} else {
			log.Info("[TUN][macOS] Added exclude route %s -> %s", addr, gw)
		}
	}
	return nil
}

// CleanupExcludeRoutes deletes all static routes added by SetupExcludeRoutes.
func CleanupExcludeRoutes() error {
	for _, prefix := range excludedRoutes {
		addr := prefix.Addr()
		var out []byte
		var err error
		if addr.Is4() {
			out, err = exec.Command("route", "delete", "-host", addr.String()).CombinedOutput()
		} else {
			out, err = exec.Command("route", "delete", "-inet6", prefix.String()).CombinedOutput()
		}
		if err != nil {
			log.Warn("[TUN][macOS] Failed to delete exclude route %s: %v, output: %s", addr, err, string(out))
		}
	}
	excludedRoutes = nil
	return nil
}

// SetupTunDNS sets DNS servers for the primary network service via networksetup on macOS.
//
// If the primary network service cannot be found, it is skipped (sing-tun will complete
// DNS interception at the TUN interface level via the DNSServers option, so system-level
// configuration may not be strictly necessary).
func SetupTunDNS(dnsServers []netip.Addr, _ string) error {
	if len(dnsServers) == 0 {
		return nil
	}

	svc, err := getPrimaryNetworkService()
	if err != nil {
		log.Warn("[TUN][macOS] SetupTunDNS: Failed to get primary network service (%v), skipping DNS setting", err)
		return nil
	}

	var addrs []string
	for _, dns := range dnsServers {
		if dns.Is4() { // Prefer IPv4 DNS
			addrs = append(addrs, dns.String())
		}
	}
	if len(addrs) == 0 {
		return nil
	}

	args := append([]string{"-setdnsservers", svc}, addrs...)
	out, err := exec.Command("networksetup", args...).CombinedOutput()
	if err != nil {
		log.Warn("[TUN][macOS] Failed to set DNS: %v, output: %s", err, string(out))
		return err
	}
	log.Info("[TUN][macOS] DNS for network service '%s' has been set to: %s", svc, strings.Join(addrs, ", "))
	return nil
}

// CleanupTunDNS restores the primary network service's DNS to automatic acquisition (DHCP/Empty).
func CleanupTunDNS(_ string) error {
	svc, err := getPrimaryNetworkService()
	if err != nil {
		return nil
	}
	out, err := exec.Command("networksetup", "-setdnsservers", svc, "Empty").CombinedOutput()
	if err != nil {
		log.Warn("[TUN][macOS] Failed to reset DNS: %v, output: %s", err, string(out))
	} else {
		log.Info("[TUN][macOS] DNS for network service '%s' has been reset to automatic", svc)
	}
	return nil
}

// getDefaultGateway retrieves the default IPv4 gateway via `route -n get default` on macOS.
func getDefaultGateway() (string, error) {
	out, err := exec.Command("sh", "-c", "route -n get default 2>/dev/null | awk '/gateway/{print $2}'").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get default gateway: %w, output: %s", err, string(out))
	}
	gw := strings.TrimSpace(string(out))
	if gw == "" {
		return "", fmt.Errorf("default gateway is empty (possible no network connection)")
	}
	return gw, nil
}

// getPrimaryNetworkService infers the primary network service name from the current default route's network interface.
func getPrimaryNetworkService() (string, error) {
	// 1. Get the interface name used by the current default route (e.g., en0)
	ifOut, err := exec.Command("sh", "-c", "route -n get default 2>/dev/null | awk '/interface/{print $2}'").CombinedOutput()
	if err != nil || strings.TrimSpace(string(ifOut)) == "" {
		return "", fmt.Errorf("failed to get default interface name")
	}
	ifName := strings.TrimSpace(string(ifOut))

	// 2. Iterate through the networksetup service list to find the service corresponding to that interface
	svcOut, err := exec.Command("networksetup", "-listallhardwareports").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("networksetup -listallhardwareports failed: %w", err)
	}

	// Output format:
	// Hardware Port: Wi-Fi
	// Device: en0
	// Ethernet Address: ...
	lines := strings.Split(string(svcOut), "\n")
	var lastSvc string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Hardware Port:") {
			lastSvc = strings.TrimSpace(strings.TrimPrefix(line, "Hardware Port:"))
		}
		if strings.HasPrefix(line, "Device:") {
			dev := strings.TrimSpace(strings.TrimPrefix(line, "Device:"))
			if dev == ifName {
				return lastSvc, nil
			}
		}
	}
	return "", fmt.Errorf("no network service found for interface %s", ifName)
}
