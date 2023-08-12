package plugin

import (
	"net"

	"github.com/daeuniverse/softwind/netproxy"
	"github.com/daeuniverse/softwind/protocol/direct"
)

type Direct struct {
	d netproxy.Dialer
}

// Addr implements Dialer.
func (d *Direct) Addr() string {
	return ""
}

// Dial implements Dialer.
func (d *Direct) Dial(network string, addr string) (c net.Conn, err error) {
	conn, err := direct.FullconeDirect.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &netproxy.FakeNetConn{
		Conn:  conn,
		LAddr: nil,
		RAddr: nil,
	}, nil
}

// DialUDP implements Dialer.
func (d *Direct) DialUDP(network string, addr string) (pc net.PacketConn, writeTo net.Addr, err error) {
	conn, err := direct.FullconeDirect.Dial("udp", addr)
	if err != nil {
		return nil, nil, err
	}
	return &netproxy.FakeNetPacketConn{
		PacketConn: conn.(netproxy.PacketConn),
		LAddr:      nil,
		RAddr:      nil,
	}, nil, nil
}

var _ Dialer = &Direct{}
