package plugin

import (
	"fmt"
	"net"
	"net/netip"
	"syscall"

	"github.com/daeuniverse/softwind/netproxy"
)

type PacketConnConverter struct {
	net.PacketConn
	Target   string
	cacheTgt net.Addr
}

// Read implements netproxy.PacketConn.
func (pc *PacketConnConverter) Read(b []byte) (n int, err error) {
	n, _, err = pc.PacketConn.ReadFrom(b)
	return
}

// Write implements netproxy.PacketConn.
func (pc *PacketConnConverter) Write(b []byte) (n int, err error) {
	if pc.cacheTgt == nil {
		addr, err := net.ResolveUDPAddr("udp", pc.Target)
		if err != nil {
			return 0, err
		}
		pc.cacheTgt = addr
	}
	return pc.PacketConn.WriteTo(b, pc.cacheTgt)
}

// ReadFrom implements netproxy.PacketConn.
func (pc *PacketConnConverter) ReadFrom(p []byte) (n int, addr netip.AddrPort, err error) {
	n, _addr, err := pc.PacketConn.ReadFrom(p)
	if err != nil {
		return 0, netip.AddrPort{}, err
	}
	return n, _addr.(*net.UDPAddr).AddrPort(), nil
}

// WriteTo implements netproxy.PacketConn.
func (pc *PacketConnConverter) WriteTo(p []byte, _addr string) (n int, err error) {
	addr, err := net.ResolveUDPAddr("udp", _addr)
	if err != nil {
		return 0, err
	}
	return pc.PacketConn.WriteTo(p, addr)
}

func (pc *PacketConnConverter) SetWriteBuffer(size int) error {
	c, ok := pc.PacketConn.(interface{ SetWriteBuffer(int) error })
	if !ok {
		return fmt.Errorf("connection doesn't allow setting of send buffer size. Not a *net.UDPConn?: %T", pc.PacketConn)
	}
	return c.SetWriteBuffer(size)
}

func (pc *PacketConnConverter) SetReadBuffer(size int) error {
	c, ok := pc.PacketConn.(interface{ SetReadBuffer(int) error })
	if !ok {
		return fmt.Errorf("connection doesn't allow setting of send buffer size. Not a *net.UDPConn?: %T", pc.PacketConn)
	}
	return c.SetReadBuffer(size)
}

func (pc *PacketConnConverter) SyscallConn() (syscall.RawConn, error) {
	c, ok := pc.PacketConn.(interface {
		SyscallConn() (syscall.RawConn, error)
	})
	if !ok {
		return nil, fmt.Errorf("connection doesn't allow to get Syscall.RawConn. Not a *net.UDPConn?: %T", pc.PacketConn)
	}
	return c.SyscallConn()
}

// ReadMsgUDP implements quic.OOBCapablePacketConn.
func (conn *PacketConnConverter) ReadMsgUDP(b []byte, oob []byte) (n int, oobn int, flags int, addr *net.UDPAddr, err error) {
	c, ok := conn.PacketConn.(interface {
		ReadMsgUDP(b []byte, oob []byte) (n int, oobn int, flags int, addr *net.UDPAddr, err error)
	})
	if !ok {
		return 0, 0, 0, nil, fmt.Errorf("connection doesn't allow to get ReadMsgUDP. Not a *net.UDPConn? : %T", conn.PacketConn)
	}
	return c.ReadMsgUDP(b, oob)
}

// WriteMsgUDP implements quic.OOBCapablePacketConn.
func (conn *PacketConnConverter) WriteMsgUDP(b []byte, oob []byte, addr *net.UDPAddr) (n int, oobn int, err error) {
	c, ok := conn.PacketConn.(interface {
		WriteMsgUDP(b []byte, oob []byte, addr *net.UDPAddr) (n int, oobn int, err error)
	})
	if !ok {
		return 0, 0, fmt.Errorf("connection doesn't allow to get WriteMsgUDP. Not a *net.UDPConn? : %T", conn.PacketConn)
	}
	return c.WriteMsgUDP(b, oob, addr)
}

var _ netproxy.PacketConn = &PacketConnConverter{}
var _ interface {
	SyscallConn() (syscall.RawConn, error)
	SetReadBuffer(int) error
	ReadMsgUDP(b, oob []byte) (n, oobn, flags int, addr *net.UDPAddr, err error)
	WriteMsgUDP(b, oob []byte, addr *net.UDPAddr) (n, oobn int, err error)
} = &PacketConnConverter{}
