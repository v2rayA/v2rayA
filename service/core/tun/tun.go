package tun

import "net/netip"

type Stack string

const (
	StackGvisor = Stack("gvisor")
	StackSystem = Stack("system")
)

type Tun interface {
	Start(stack Stack) error
	Close() error
	AddDomainWhitelist(domain string)
	AddIPWhitelist(addr netip.Addr)
	SetFakeIP(enabled bool)
	SetIPv6(enabled bool)
}

var Default = NewSingTun()
