package resolv

import (
	"context"
	"math/rand"
	"net"
	"time"
)

var DefaultResolver *net.Resolver

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
	dialer := net.Dialer{Timeout: 1000 * time.Millisecond}
	DefaultResolver = &net.Resolver{
		PreferGo:     false,
		StrictErrors: false,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			server := dnsServers[rand.Intn(len(dnsServers))]
			address = server.addr
			network = server.network
			return dialer.DialContext(ctx, network, address)
		},
	}
}

func LookupHost(host string) (addrs []string, err error) {
	addrs, err = net.LookupHost(host)
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
		return DefaultResolver.LookupHost(context.Background(), host)
	}
	return addrs, err
}
