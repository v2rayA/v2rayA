package netTools

import (
	"net"
)

var intranet4 = []string{
	"0.0.0.0/32",
	"10.0.0.0/8",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"224.0.0.0/4",
	"240.0.0.0/4",
}

var jokernet4 = []string{
	"0.0.0.0/8",
	"127.0.0.0/8",
	"240.0.0.0/4",
}

var intranet6 = []string{
	"::/128",
	"::1/128",
	"64:ff9b::/96",
	"100::/64",
	"2001::/32",
	"2001:20::/28",
	"2001:db8::/32",
	"2002::/16",
	"fc00::/7",
	"fe80::/10",
	"ff00::/8",
}
var jokernet6 = []string{
	"::/128",
	"::1/128",
	"fc00::/7",
	"ff00::/8",
}

type IPNets struct {
	nets []*net.IPNet
}

var (
	ipnetsIntranet4 *IPNets
	ipnetsJokernet4 *IPNets
	ipnetsIntranet6 *IPNets
	ipnetsJokernet6 *IPNets
)

func NewIPNets(cidrs []string) (*IPNets, error) {
	n := new(IPNets)
	for _, cidr := range cidrs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		n.nets = append(n.nets, ipnet)
	}
	return n, nil
}

func (n *IPNets) Match(ip net.IP) bool {
	for _, n := range n.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func init() {
	ipnetsIntranet4, _ = NewIPNets(intranet4)
	ipnetsJokernet4, _ = NewIPNets(jokernet4)
	ipnetsIntranet6, _ = NewIPNets(intranet6)
	ipnetsJokernet6, _ = NewIPNets(jokernet6)
}

func IsIntranet4(ipv4 *[4]byte) bool {
	return ipnetsIntranet4.Match(net.IP(ipv4[:]))
}

func IsJokernet4(ipv4 *[4]byte) bool {
	return ipnetsJokernet4.Match(net.IP(ipv4[:]))
}

func IsIntranet6(ipv6 *[16]byte) bool {
	v6 := net.IP(ipv6[:])
	return ipnetsIntranet6.Match(v6)
}

func IsJokernet6(ipv6 *[16]byte) bool {
	v6 := net.IP(ipv6[:])
	return ipnetsJokernet6.Match(v6)
}
