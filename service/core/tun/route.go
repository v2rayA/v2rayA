//go:build linux
// +build linux

package tun

import (
	"net/netip"

	"github.com/v2rayA/v2rayA/common/cmds"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	// TUN_ROUTE_TABLE is a dedicated routing table for TUN interface
	TUN_ROUTE_TABLE = 2026
	// FWMARK is the mark used by v2ray/xray outbound and plugins
	FWMARK = 0x80
)

// SetupTunRouteRules sets up policy routing rules to exclude marked traffic from TUN
// This ensures that traffic from v2ray/xray cores and plugins bypass the TUN interface
func SetupTunRouteRules() error {
	commands := []string{
		// IPv4: Prioritize fwmark 0x80 traffic to use main routing table
		// This prevents v2ray/xray/plugin traffic from being captured by TUN
		"ip rule add fwmark 0x80 table main pref 100 2>/dev/null || true",

		// IPv6: Same for IPv6 traffic
		"ip -6 rule add fwmark 0x80 table main pref 100 2>/dev/null || true",
	}

	for _, cmd := range commands {
		if err := cmds.ExecCommands(cmd, false); err != nil {
			log.Warn("SetupTunRouteRules: failed to execute '%s': %v", cmd, err)
		}
	}

	log.Info("TUN route rules for fwmark 0x80 traffic established")
	return nil
}

// CleanupTunRouteRules removes the policy routing rules added for TUN
func CleanupTunRouteRules() error {
	commands := []string{
		// Remove IPv4 rule
		"ip rule del fwmark 0x80 table main pref 100 2>/dev/null || true",

		// Remove IPv6 rule
		"ip -6 rule del fwmark 0x80 table main pref 100 2>/dev/null || true",
	}

	for _, cmd := range commands {
		if err := cmds.ExecCommands(cmd, false); err != nil {
			log.Warn("CleanupTunRouteRules: failed to execute '%s': %v", cmd, err)
		}
	}

	log.Info("TUN route rules for fwmark 0x80 traffic removed")
	return nil
}

// SetupExcludeRoutes is a no-op on Linux since fwmark handles exclusion
func SetupExcludeRoutes(excludeAddrs []netip.Prefix) error {
	// On Linux, we use fwmark policy routing instead of static routes
	return nil
}

// CleanupExcludeRoutes is a no-op on Linux
func CleanupExcludeRoutes() error {
	// On Linux, we use fwmark policy routing instead of static routes
	return nil
}
