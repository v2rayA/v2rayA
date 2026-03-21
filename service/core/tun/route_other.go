//go:build !linux && !windows && !darwin
// +build !linux,!windows,!darwin

package tun

import (
	"fmt"
	"net/netip"
	"os/exec"
	"strings"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

var excludedRoutes []netip.Prefix

// SetupTunRouteRules is a no-operation on FreeBSD/OpenBSD.
// These platforms do not currently support policy routing rule configuration.
func SetupTunRouteRules() error {
	return nil
}

// CleanupTunRouteRules is a no-operation on FreeBSD/OpenBSD.
func CleanupTunRouteRules() error {
	return nil
}

// setTunRouteAutoMode is a no-operation on BSD and similar platforms.
func setTunRouteAutoMode(_ bool) {}

// DynAddExcludeRoute is a no-operation on BSD and similar platforms.
func DynAddExcludeRoute(_ netip.Addr) {}

// SetupExcludeRoutes adds "bypass TUN" static routes for proxy servers on FreeBSD/OpenBSD.
// BSD system route command syntax is largely identical to macOS.
func SetupExcludeRoutes(addrs []netip.Prefix) error {
	if len(addrs) == 0 {
		return nil
	}
	excludedRoutes = addrs

	gw, err := getDefaultGateway()
	if err != nil {
		log.Warn("[TUN][BSD] Failed to get default gateway: %v", err)
		return err
	}

	for _, prefix := range addrs {
		addr := prefix.Addr()
		var out []byte
		if addr.Is4() {
			out, err = exec.Command("route", "add", "-host", addr.String(), gw).CombinedOutput()
		} else {
			out, err = exec.Command("route", "add", "-inet6", prefix.String(), gw).CombinedOutput()
		}
		if err != nil {
			s := string(out)
			if !strings.Contains(s, "File exists") && !strings.Contains(s, "already exists") {
				log.Warn("[TUN][BSD] Failed to add exclude route %s: %v, output: %s", addr, err, s)
			}
		} else {
			log.Info("[TUN][BSD] Added exclude route %s -> %s", addr, gw)
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
			log.Warn("[TUN][BSD] Failed to delete exclude route %s: %v, output: %s", addr, err, string(out))
		}
	}
	excludedRoutes = nil
	return nil
}

// SetupTunDNS does not currently support automatic DNS configuration on FreeBSD/OpenBSD and is a no-operation.
func SetupTunDNS(_ []netip.Addr, _ string) error {
	return nil
}

// CleanupTunDNS is a no-operation on FreeBSD/OpenBSD.
func CleanupTunDNS(_ string) error {
	return nil
}

// getDefaultGateway retrieves the default IPv4 gateway for BSD systems via `netstat -rn`.
func getDefaultGateway() (string, error) {
	// FreeBSD/OpenBSD: the "default" destination line in netstat -rn
	out, err := exec.Command("sh", "-c", "netstat -rn 2>/dev/null | awk '/^default/{print $2; exit}'").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get default gateway: %w, output: %s", err, string(out))
	}
	gw := strings.TrimSpace(string(out))
	if gw == "" {
		return "", fmt.Errorf("default gateway is empty (possibly no network connection)")
	}
	return gw, nil
}
