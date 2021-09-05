package socks5

import (
	"github.com/v2rayA/v2rayA/pkg/plugin/socks"
	"net"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// PktConn .
type PktConn struct {
	net.PacketConn

	writeAddr net.Addr // write to and read from addr

	tgtAddr   socks.Addr
	tgtHeader bool

	ctrlConn net.Conn // tcp control conn
}

// NewPktConn returns a PktConn
func NewPktConn(c net.PacketConn, writeAddr net.Addr, tgtAddr socks.Addr, tgtHeader bool, ctrlConn net.Conn) *PktConn {
	pc := &PktConn{
		PacketConn: c,
		writeAddr:  writeAddr,
		tgtAddr:    tgtAddr,
		tgtHeader:  tgtHeader,
		ctrlConn:   ctrlConn}

	if ctrlConn != nil {
		go func() {
			buf := make([]byte, 1)
			for {
				_, err := ctrlConn.Read(buf)
				if err, ok := err.(net.Error); ok && err.Timeout() {
					continue
				}
				log.Trace("[socks5] dialudp udp associate end")
				return
			}
		}()
	}

	return pc
}

// ReadFrom overrides the original function from net.PacketConn
func (pc *PktConn) ReadFrom(b []byte) (int, net.Addr, error) {
	if !pc.tgtHeader {
		return pc.PacketConn.ReadFrom(b)
	}

	buf := make([]byte, len(b))
	n, raddr, err := pc.PacketConn.ReadFrom(buf)
	if err != nil {
		return n, raddr, err
	}

	// https://tools.ietf.org/html/rfc1928#section-7
	// +----+------+------+----------+----------+----------+
	// |RSV | FRAG | ATYP | DST.ADDR | DST.PORT |   DATA   |
	// +----+------+------+----------+----------+----------+
	// | 2  |  1   |  1   | Variable |    2     | Variable |
	// +----+------+------+----------+----------+----------+
	tgtAddr := socks.SplitAddr(buf[3:])
	copy(b, buf[3+len(tgtAddr):])

	//test
	if pc.writeAddr == nil {
		pc.writeAddr = raddr
	}

	if pc.tgtAddr == nil {
		pc.tgtAddr = tgtAddr
	}

	return n - len(tgtAddr) - 3, raddr, err
}

// WriteTo overrides the original function from net.PacketConn
func (pc *PktConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	if !pc.tgtHeader {
		return pc.PacketConn.WriteTo(b, addr)
	}

	buf := append([]byte{0, 0, 0}, pc.tgtAddr...)
	buf = append(buf, b[:]...)
	return pc.PacketConn.WriteTo(buf, pc.writeAddr)
}

// Close .
func (pc *PktConn) Close() error {
	if pc.ctrlConn != nil {
		pc.ctrlConn.Close()
	}

	return pc.PacketConn.Close()
}
