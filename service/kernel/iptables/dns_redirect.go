package iptables

import "fmt"

// NftDNSRedirect installs NAT independently of the transparent proxy's mangle
// hooks. A DNS_MARK accept verdict does not translate the destination port.
func NftDNSRedirect(port string, ipv6 bool) Setter {
	family := "ipv4"
	if ipv6 {
		family = "{ ipv4, ipv6 }"
	}
	rules := fmt.Sprintf(`table inet v2raya_dns {
    chain output {
        type nat hook output priority dstnat - 1; policy accept;
        meta mark & 0x80 == 0x80 return
        meta nfproto %s meta l4proto { tcp, udp } th dport 53 redirect to :%s
    }
    chain prerouting {
        type nat hook prerouting priority dstnat - 1; policy accept;
        meta mark & 0x80 == 0x80 return
        meta nfproto %s meta l4proto { tcp, udp } th dport 53 redirect to :%s
    }
}
`, family, port, family, port)
	return Setter{PreFunc: func() error {
		return executeCommandWithInput("nft", []string{"-f", "-"}, rules)
	}}
}
