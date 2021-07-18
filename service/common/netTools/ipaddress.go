package netTools

import (
	"bytes"
	"github.com/v2rayA/v2rayA/infra/dataStructure/trie"
	"net"
	"strconv"
	"strings"
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
	"::ffff:0:0/96",
	"::ffff:0:0:0/96",
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
	trieIntranet4   *trie.Trie
	trieJokernet4   *trie.Trie
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
	trieIntranet4 = New4(intranet4)
	trieJokernet4 = New4(jokernet4)
	ipnetsIntranet6, _ = NewIPNets(intranet6)
	ipnetsJokernet6, _ = NewIPNets(jokernet6)
}

func New4(CIDRs []string) *trie.Trie {
	dict := make([]string, 0, len(CIDRs))
	for _, CIDR := range CIDRs {
		grp := strings.SplitN(CIDR, "/", 2)
		l, _ := strconv.Atoi(grp[1])
		arr := strings.Split(grp[0], ".")
		var builder strings.Builder
		for _, sec := range arr {
			itg, _ := strconv.Atoi(sec)
			tmp := strconv.FormatInt(int64(itg), 2)
			builder.WriteString(strings.Repeat("0", 8-len(tmp)))
			builder.WriteString(tmp)
			if builder.Len() >= l {
				break
			}
		}
		dict = append(dict, builder.String()[:l])
	}
	return trie.New(dict)
}

func ipv4ToBin(ipv4 *[4]byte) string {
	var buff = new(bytes.Buffer)
	for _, b := range ipv4 {
		tmp := strconv.FormatInt(int64(b), 2)
		buff.WriteString(strings.Repeat("0", 8-len(tmp)) + tmp)
	}
	return buff.String()
}

func IsIntranet4(ipv4 *[4]byte) bool {
	return trieIntranet4.Match(ipv4ToBin(ipv4)) != ""
}

func IsJokernet4(ipv4 *[4]byte) bool {
	return trieJokernet4.Match(ipv4ToBin(ipv4)) != ""
}

func IsIntranet6(ipv6 *[16]byte) bool {
	v6 := net.IP(ipv6[:])
	return ipnetsIntranet6.Match(v6)
}

func IsJokernet6(ipv6 *[16]byte) bool {
	v6 := net.IP(ipv6[:])
	return ipnetsJokernet6.Match(v6)
}
