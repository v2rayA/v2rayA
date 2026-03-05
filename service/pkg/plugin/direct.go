package plugin

import (
	"context"
	"net"
	"net/netip"

	"github.com/daeuniverse/outbound/netproxy"
	"github.com/daeuniverse/outbound/protocol/direct"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

var (
	directDialer netproxy.Dialer
)

func init() {
	directDialer = direct.NewDirectDialerLaddr(netip.Addr{}, direct.Option{FullCone: true})
}

type Direct struct {
}

// Addr implements Dialer.
func (d *Direct) Addr() string {
	return ""
}

// Dial implements Dialer.
func (d *Direct) Dial(network string, addr string) (c net.Conn, err error) {
	return d.DialContext(context.Background(), network, addr)
}

func (d *Direct) DialContext(ctx context.Context, network string, addr string) (c net.Conn, err error) {
	log.Info("[%s] dialing %s", "direct", addr)
	magicNetwork := netproxy.MagicNetwork{
		Network: network,
		Mark:    ShouldSetMark(),
	}
	conn, err := directDialer.DialContext(ctx, magicNetwork.Encode(), addr)
	if err != nil {
		return nil, err
	}
	return NewFakeNetConn(conn), nil
}

// DialUDP implements Dialer.
func (d *Direct) DialUDP(network string) (pc FakeNetPacketConn, err error) {
	log.Info("[%s] dialing udp", "direct")
	magicNetwork := netproxy.MagicNetwork{
		Network: "udp",
		Mark:    ShouldSetMark(),
	}
	conn, err := directDialer.DialContext(context.TODO(), magicNetwork.Encode(), "")
	if err != nil {
		return nil, err
	}
	// directDialer returns directPacketConn which implements net.PacketConn
	// and has SyscallConn() method. Return it directly to preserve raw socket access.
	if netConn, ok := conn.(net.PacketConn); ok {
		return netConn, nil
	}
	// Fallback: wrap it if it's not a net.PacketConn
	return NewFakeNetPacketConn(conn.(netproxy.PacketConn)), nil
}

var _ Dialer = &Direct{}
