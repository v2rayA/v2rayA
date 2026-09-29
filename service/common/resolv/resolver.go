package resolv

import (
	"context"
	"math/rand"
	"net"
	"time"
)

var defaultResolver *net.Resolver
var systemResolver *net.Resolver

// PreferredServers lists the user's direct DNS upstreams (host:port), tried
// before the built-in public ones; the service sets it.
var PreferredServers = func() []string { return nil }

var dnsServers = []struct {
	addr    string
	network string
}{
	{"119.29.29.29:53", "udp"}, //dnspod
	{"119.29.29.29:53", "tcp"},
	{"223.6.6.6:53", "tcp"}, //alidns
	{"223.6.6.6:53", "udp"},
	{"180.76.76.76:53", "udp"},   //baidudns
	{"208.67.222.222:53", "tcp"}, //opendns
	{"208.67.222.222:53", "udp"},
}

func init() {
	rand.Seed(time.Now().UnixNano())
	defaultResolver, systemResolver = newResolvers(&net.Dialer{Timeout: time.Second})
}

func newResolvers(dialer *net.Dialer) (fallback, system *net.Resolver) {
	// DNS upstream addresses must not recurse through the probe's resolver.
	dnsDialer := *dialer
	dnsDialer.Resolver = nil
	fallback = &net.Resolver{
		PreferGo:     true,
		StrictErrors: false,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			if preferred := PreferredServers(); len(preferred) > 0 {
				return dnsDialer.DialContext(ctx, network, preferred[rand.Intn(len(preferred))])
			}
			server := dnsServers[rand.Intn(len(dnsServers))]
			address = server.addr
			network = server.network
			return dnsDialer.DialContext(ctx, network, address)
		},
	}
	system = &net.Resolver{
		PreferGo:     true,
		StrictErrors: false,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dnsDialer.DialContext(ctx, network, address)
		},
	}
	return fallback, system
}

func LookupHost(host string) (addrs []string, err error) {
	return lookupHost(host, systemResolver, defaultResolver)
}

// DirectResolver uses the configured direct upstreams, or the built-in
// public servers, with the caller's socket policy.
func DirectResolver(dialer *net.Dialer) *net.Resolver {
	fallback, _ := newResolvers(dialer)
	return fallback
}

// LookupHostWithDialer preserves the system-DNS recheck and fallback while
// applying the caller's socket policy to every DNS connection.
func LookupHostWithDialer(host string, dialer *net.Dialer) ([]string, error) {
	fallback, system := newResolvers(dialer)
	return lookupHost(host, system, fallback)
}

func lookupHost(host string, system, fallback *net.Resolver) (addrs []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	addrs, err = system.LookupHost(ctx, host)
	lookupAgain := len(addrs) == 0 || err != nil
	if !lookupAgain {
		for _, addr := range addrs {
			if ip := net.ParseIP(addr); ip != nil && (ip.IsLoopback() || ip.IsUnspecified()) {
				lookupAgain = true
				break
			}
		}
	}
	if lookupAgain {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return fallback.LookupHost(ctx, host)
	}
	return addrs, err
}
