package dns

import (
	"context"
	"net"
)

// RouteDispatcher is the interface for dispatching DNS queries through
// v2raya-core's internal routing. It mirrors the subset of xray-core's
// routing.Dispatcher that the DNS module needs, avoiding a direct import
// of xray-core packages in the DNS module.
//
// xray-core's default.Dispatcher implements this interface.
// The DNS module uses this to route queries through the appropriate outbound
// (e.g., "proxy" outbound) based on the proxy tag, just like xray-core's
// built-in DNS module does via session.ContextWithInbound.
type RouteDispatcher interface {
	// Dispatch creates a connection to the given destination through
	// xray-core's routing engine. The proxyTag controls the routing
	// decision (e.g., "proxy" routes through proxy outbound,
	// "direct" routes directly). The implementation sets
	// session.ContextWithInbound internally, like xray-core's DNS module.
	// Returns a net.Conn for DNS-over-TCP exchange.
	Dispatch(ctx context.Context, network, addr, proxyTag string) (net.Conn, error)
}
