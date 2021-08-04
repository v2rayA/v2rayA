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
	{"114.114.115.115:53", "tcp"}, //114
	{"114.114.115.115:53", "udp"},
	{"119.29.29.29:53", "udp"}, //dnspod
	{"119.29.29.29:53", "tcp"},
	{"223.6.6.6:53", "tcp"}, //alidns
	{"223.6.6.6:53", "udp"},
	{"180.76.76.76:53", "udp"},   //baidudns
	{"208.67.222.222:53", "tcp"}, //opendns
	{"208.67.222.222:53", "udp"},
	{"1.2.4.8:53", "udp"}, //cnnic
}

func init() {
	rand.Seed(time.Now().UnixNano())
	DefaultResolver = &net.Resolver{
		PreferGo:     true,
		StrictErrors: false,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			server := dnsServers[rand.Intn(len(dnsServers))]
			address = server.addr
			network = server.network
			return d.DialContext(ctx, network, address)
		},
	}
}

func LookupHost(host string) (addrs []string, err error) {
	return DefaultResolver.LookupHost(context.Background(), host)
}
