package v2ray

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"time"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"

	"github.com/v2rayA/v2rayA/kernel/iptables"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	tunSupported = true
	// tunBindsEgress says the core binds its sockets to the physical
	// interface (no socket mark here).
	tunBindsEgress = true
	tunDeviceName  = "v2rayA"
)

// Two /1 routes cover every destination while the physical default route
// stays in place; the core's sockets are bound to the physical interface,
// so they never follow them.
var (
	tunHalfRoutes4 = []netip.Prefix{netip.MustParsePrefix("0.0.0.0/1"), netip.MustParsePrefix("128.0.0.0/1")}
	tunHalfRoutes6 = []netip.Prefix{netip.MustParsePrefix("::/1"), netip.MustParsePrefix("8000::/1")}
)

var tunLUID winipcfg.LUID

func tunInstalled() bool { return tunLUID != 0 }

// tunIPv6Enabled follows the same ipv6-support setting the Linux rules do.
func tunIPv6Enabled() bool { return iptables.IsIPv6Supported() }

func tunDevicePresent() bool {
	_, err := net.InterfaceByName(tunDeviceName)
	return err == nil
}

// tunNodeRoutes holds the /32 and /128 routes installed on the physical
// interface for the proxy nodes, for removal on stop.
var tunNodeRoutes []struct {
	luid   winipcfg.LUID
	prefix netip.Prefix
	gw     netip.Addr
}

// tunEgressInterface is the interface the lowest-metric IPv4 default route
// points at, ignoring the TUN adapter should it still exist.
func tunEgressInterface() string {
	rows, err := winipcfg.GetIPForwardTable2(windows.AF_INET)
	if err != nil {
		log.Warn("tun: cannot read the route table: %v", err)
		return ""
	}
	var best *winipcfg.MibIPforwardRow2
	for i := range rows {
		r := &rows[i]
		if r.DestinationPrefix.PrefixLength != 0 {
			continue
		}
		if iface, err := r.InterfaceLUID.Interface(); err == nil && iface.Alias() == tunDeviceName {
			continue
		}
		if best == nil || r.Metric < best.Metric {
			best = r
		}
	}
	if best == nil {
		return ""
	}
	iface, err := net.InterfaceByIndex(int(best.InterfaceIndex))
	if err != nil {
		return ""
	}
	return iface.Name
}

// physicalDefault returns the best default route of a family, excluding the
// TUN adapter, for host routes to the proxy nodes.
func physicalDefault(family winipcfg.AddressFamily) (winipcfg.LUID, netip.Addr, bool) {
	rows, err := winipcfg.GetIPForwardTable2(family)
	if err != nil {
		return 0, netip.Addr{}, false
	}
	var best *winipcfg.MibIPforwardRow2
	for i := range rows {
		r := &rows[i]
		if r.DestinationPrefix.PrefixLength != 0 {
			continue
		}
		if iface, err := r.InterfaceLUID.Interface(); err == nil && iface.Alias() == tunDeviceName {
			continue
		}
		if best == nil || r.Metric < best.Metric {
			best = r
		}
	}
	if best == nil {
		return 0, netip.Addr{}, false
	}
	return best.InterfaceLUID, best.NextHop.Addr(), true
}

func tunRoutesUp(_ *Template, nodeIPs []string) error {
	// Without a physical default route the core's sockets were not bound
	// to an interface, and the /1 routes would send the core's own
	// connections back into its device, each one spawning the next.
	// Refusing here leaves the device up and unrouted, which is harmless.
	if _, _, ok := physicalDefault(windows.AF_INET); !ok {
		return errors.New("tun: no IPv4 default route; connect to a network and start the transparent proxy again")
	}
	// Host routes for the proxy nodes go on the physical interface before
	// the TUN routes exist, so every process reaches the nodes directly.
	tunNodeRoutes = nil
	for _, ip := range nodeIPs {
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			continue
		}
		family := winipcfg.AddressFamily(windows.AF_INET)
		bits := 32
		if addr.Is6() {
			family, bits = windows.AF_INET6, 128
		}
		luid, gw, ok := physicalDefault(family)
		if !ok {
			continue
		}
		prefix := netip.PrefixFrom(addr, bits)
		if err := luid.AddRoute(prefix, gw, 0); err != nil {
			log.Warn("tun: node route %s: %v", prefix, err)
			continue
		}
		tunNodeRoutes = append(tunNodeRoutes, struct {
			luid   winipcfg.LUID
			prefix netip.Prefix
			gw     netip.Addr
		}{luid, prefix, gw})
	}
	var iface *net.Interface
	deadline := time.Now().Add(5 * time.Second)
	for {
		var err error
		if iface, err = net.InterfaceByName(tunDeviceName); err == nil {
			break
		}
		if time.Now().After(deadline) {
			tunRoutesDown()
			return fmt.Errorf("adapter %s did not appear; the core did not start its TUN inbound", tunDeviceName)
		}
		time.Sleep(100 * time.Millisecond)
	}
	luid, err := winipcfg.LUIDFromIndex(uint32(iface.Index))
	if err != nil {
		tunRoutesDown()
		return fmt.Errorf("tun: adapter LUID: %w", err)
	}
	tunLUID = luid
	gw4, gw6 := netip.MustParseAddr(tunGateway4), netip.MustParseAddr(tunGateway6)
	for _, p := range tunHalfRoutes4 {
		if err := luid.AddRoute(p, gw4, 0); err != nil {
			tunRoutesDown()
			return fmt.Errorf("tun: add route %s: %w", p, err)
		}
	}
	// IPv6 follows the ipv6-support setting, like the inbound's address:
	// with it off the device has no IPv6 address and routing IPv6 into it
	// would black-hole every IPv6 connection.
	if tunIPv6Enabled() {
		for _, p := range tunHalfRoutes6 {
			if err := luid.AddRoute(p, gw6, 0); err != nil {
				log.Warn("tun: add route %s: %v", p, err)
			}
		}
	}
	// The system resolver asks the TUN's gateway address, whose queries the
	// core intercepts; the lowest interface metric makes it the first
	// resolver tried.
	if err := luid.SetDNS(windows.AF_INET, []netip.Addr{gw4}, nil); err != nil {
		tunRoutesDown()
		return fmt.Errorf("tun: set DNS: %w", err)
	}
	if tunIPv6Enabled() {
		if err := luid.SetDNS(windows.AF_INET6, []netip.Addr{gw6}, nil); err != nil {
			log.Warn("tun: set IPv6 DNS: %v", err)
		}
	}
	for _, family := range []winipcfg.AddressFamily{windows.AF_INET, windows.AF_INET6} {
		ipif, err := luid.IPInterface(family)
		if err != nil {
			continue
		}
		ipif.UseAutomaticMetric = false
		ipif.Metric = 0
		if err := ipif.Set(); err != nil {
			log.Warn("tun: interface metric: %v", err)
		}
	}
	log.Info("tun: routes and DNS installed for %s", tunDeviceName)
	return nil
}

func tunRoutesDown() {
	for _, r := range tunNodeRoutes {
		_ = r.luid.DeleteRoute(r.prefix, r.gw)
	}
	tunNodeRoutes = nil
	if tunLUID == 0 {
		return
	}
	gw4, gw6 := netip.MustParseAddr(tunGateway4), netip.MustParseAddr(tunGateway6)
	for _, p := range tunHalfRoutes4 {
		_ = tunLUID.DeleteRoute(p, gw4)
	}
	for _, p := range tunHalfRoutes6 {
		_ = tunLUID.DeleteRoute(p, gw6)
	}
	_ = tunLUID.FlushDNS(windows.AF_INET)
	_ = tunLUID.FlushDNS(windows.AF_INET6)
	tunLUID = 0
}

// tunCleanupResidual has nothing to do: the adapter, and with it every
// route and DNS setting on it, goes away with the core process.
func tunCleanupResidual() {}
