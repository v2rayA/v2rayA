package dns

import "net/netip"

var defaultDNS = []netip.AddrPort{netip.MustParseAddrPort("127.0.0.1:53"), netip.MustParseAddrPort("[::1]:53")}
