package iptables

import "V2RayA/common/cmds"

type spoofingFilter struct{ iptablesSetter }

var SpoofingFilter spoofingFilter

func (f *spoofingFilter) GetSetupCommands() SetupCommands {
	commands := `
# 建链
iptables -N SF_IN
iptables -N SF_FWD
iptables -A INPUT -j SF_IN
iptables -A FORWARD -j SF_FWD

#input chain
iptables -A SF_IN -p udp --sport 53 -m u32 --u32 "0&0x0F000000=0x05000000 && 22&0xFFFF@16=0x01010101,0xffffffff,0x4a7d7f66,0x4a7d9b66,0x4a7d2766,0x4a7d2771,0xd155e58a,0x042442b2,0x0807c62d,0x253d369e" -j DROP
iptables -A SF_IN -p udp --sport 53 -m u32 --u32 "0&0x0F000000=0x05000000 && 22&0xFFFF@16=0x2e52ae44,0x3b1803ad,0x402158a1,0x4021632f,0x4042a3fb,0x4168cafc,0x41a0db71,0x422dfced,0x480ecd68,0x480ecd63" -j DROP
iptables -A SF_IN -p udp --sport 53 -m u32 --u32 "0&0x0F000000=0x05000000 && 22&0xFFFF@16=0x4e10310f,0x5d2e0859,0x80797e8b,0x9f6a794b,0xa9840d67,0xc043c606,0xca6a0102,0xcab50755,0xcba1e6ab,0xcb620741" -j DROP
iptables -A SF_IN -p udp --sport 53 -m u32 --u32 "0&0x0F000000=0x05000000 && 22&0xFFFF@16=0xcf0c5862,0xd0381f2b,0xd1913632,0xd1dc1eae,0xd1244921,0xd35e4293,0xd5a9fb23,0xd8ddbcb6,0xd8eab30d,0xf3b9bb03" -j DROP
iptables -A SF_IN -p udp --sport 53 -m u32 --u32 "0&0x0F000000=0x05000000 && 22&0xFFFF@16=0xf3b9bb27,0x1759053c,0x25d06f78,0x31027b38,0x364c8701,0x4d04075c,0x76053106,0xbc050460,0xbda31105,0xc504040c" -j DROP
iptables -A SF_IN -p udp --sport 53 -m u32 --u32 "0&0x0F000000=0x05000000 && 22&0xFFFF@16=0xf9812e30,0xfd9d0ea5" -j DROP
 
#forward chain
iptables -A SF_FWD -p udp --sport 53 -m u32 --u32 "0&0x0F000000=0x05000000 && 22&0xFFFF@16=0x01010101,0xffffffff,0x4a7d7f66,0x4a7d9b66,0x4a7d2766,0x4a7d2771,0xd155e58a,0x042442b2,0x0807c62d,0x253d369e" -j DROP
iptables -A SF_FWD -p udp --sport 53 -m u32 --u32 "0&0x0F000000=0x05000000 && 22&0xFFFF@16=0x2e52ae44,0x3b1803ad,0x402158a1,0x4021632f,0x4042a3fb,0x4168cafc,0x41a0db71,0x422dfced,0x480ecd68,0x480ecd63" -j DROP
iptables -A SF_FWD -p udp --sport 53 -m u32 --u32 "0&0x0F000000=0x05000000 && 22&0xFFFF@16=0x4e10310f,0x5d2e0859,0x80797e8b,0x9f6a794b,0xa9840d67,0xc043c606,0xca6a0102,0xcab50755,0xcba1e6ab,0xcb620741" -j DROP
iptables -A SF_FWD -p udp --sport 53 -m u32 --u32 "0&0x0F000000=0x05000000 && 22&0xFFFF@16=0xcf0c5862,0xd0381f2b,0xd1913632,0xd1dc1eae,0xd1244921,0xd35e4293,0xd5a9fb23,0xd8ddbcb6,0xd8eab30d,0xf3b9bb03" -j DROP
iptables -A SF_FWD -p udp --sport 53 -m u32 --u32 "0&0x0F000000=0x05000000 && 22&0xFFFF@16=0xf3b9bb27,0x1759053c,0x25d06f78,0x31027b38,0x364c8701,0x4d04075c,0x76053106,0xbc050460,0xbda31105,0xc504040c" -j DROP
iptables -A SF_FWD -p udp --sport 53 -m u32 --u32 "0&0x0F000000=0x05000000 && 22&0xFFFF@16=0xf9812e30,0xfd9d0ea5" -j DROP
`
	if cmds.IsCommandValid("sysctl") {
		commands += `
#禁用ipv6
sysctl -w net.ipv6.conf.all.disable_ipv6=1
sysctl -w net.ipv6.conf.default.disable_ipv6=1
	`
	}
	return SetupCommands(commands)
}

func (f *spoofingFilter) GetCleanCommands() CleanCommands {
	commands := `
iptables -F SF_IN
iptables -D INPUT -j SF_IN
iptables -X SF_IN
iptables -F SF_FWD
iptables -D FORWARD -j SF_FWD
iptables -X SF_FWD
`
	if cmds.IsCommandValid("sysctl") {
		commands += `
sysctl -w net.ipv6.conf.all.disable_ipv6=0
sysctl -w net.ipv6.conf.default.disable_ipv6=0
`
	}
	return CleanCommands(commands)
}
