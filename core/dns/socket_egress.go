package dns

import (
	"net"
	"net/netip"
	"strings"
	"sync/atomic"
)

// egressInterface is the interface the module's own upstream sockets are
// bound to on platforms without SO_MARK. With a TUN holding the default
// route, an unbound socket would send the module's upstream queries into
// the TUN, where the core's port-53 interception would hand them straight
// back to this module. Empty means unbound.
var egressInterface atomic.Pointer[string]

// SetEgressInterface records the physical interface for later dialers; the
// module calls it from Start with its configuration.
func SetEgressInterface(name string) {
	egressInterface.Store(&name)
}

// egressInterfaceIndex resolves the configured interface, or returns 0 and
// false when none is configured or it does not exist.
func egressInterfaceIndex() (int, bool) {
	p := egressInterface.Load()
	if p == nil || *p == "" {
		return 0, false
	}
	iface, err := net.InterfaceByName(*p)
	if err != nil {
		return 0, false
	}
	return iface.Index, true
}

// isIPv6Address decides which family's interface-binding option a dial to
// address needs, from the address itself when it parses and from the
// network name otherwise.
func isIPv6Address(network, address string) bool {
	if ap, err := netip.ParseAddrPort(address); err == nil {
		return ap.Addr().Unmap().Is6()
	}
	if host, _, err := net.SplitHostPort(address); err == nil {
		if a, err := netip.ParseAddr(host); err == nil {
			return a.Unmap().Is6()
		}
	}
	return strings.HasSuffix(network, "6")
}
