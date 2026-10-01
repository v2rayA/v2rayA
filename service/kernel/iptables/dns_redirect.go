package iptables

import "fmt"

// NftDNSRedirect installs NAT independently of the transparent proxy's mangle
// hooks. A DNS_MARK accept verdict does not translate the destination port.
// Incoming DNS uses the transparent proxy's existing prerouting chain in
// TPROXY/Redirect mode, or the core's DNS relay when routed through the TUN.
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
}
`, family, port)
	return Setter{PreFunc: func() error {
		// A previous process may have died before removing its table.
		_ = executeCommands("nft delete table inet v2raya_dns 2>/dev/null || true", false)
		if err := executeCommandWithInput("nft", []string{"-f", "-"}, rules); err != nil {
			_ = executeCommands("nft delete table inet v2raya_dns 2>/dev/null || true", false)
			return err
		}
		return nil
	}}
}
