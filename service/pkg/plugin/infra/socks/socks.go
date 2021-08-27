package socks

import (
	"errors"
	"io"
	"net"
	"strconv"
)

// SOCKS auth type
const (
	AuthNone     = 0
	AuthPassword = 2
)

// SOCKS request commands as defined in RFC 1928 section 4
const (
	CmdError        byte = 0
	CmdConnect      byte = 1
	CmdBind         byte = 2
	CmdUDPAssociate byte = 3
)

// SOCKS address types as defined in RFC 1928 section 5
const (
	ATypIP4    = 1
	ATypDomain = 3
	ATypIP6    = 4
)

// MaxAddrLen is the maximum size of SOCKS address in bytes
const MaxAddrLen = 1 + 1 + 255 + 2

// Errors are socks5 errors
var Errors = []error{
	errors.New(""),
	errors.New("general failure"),
	errors.New("connection forbidden"),
	errors.New("network unreachable"),
	errors.New("host unreachable"),
	errors.New("connection refused"),
	errors.New("TTL expired"),
	errors.New("command not supported"),
	errors.New("address type not supported"),
	errors.New("socks5UDPAssociate"),
}

// Addr represents a SOCKS address as defined in RFC 1928 section 5.
type Addr []byte

// String serializes SOCKS address a to string form.
func (a Addr) String() string {
	var host, port string

	switch a[0] { // address type
	case ATypDomain:
		host = string(a[2 : 2+int(a[1])])
		port = strconv.Itoa((int(a[2+int(a[1])]) << 8) | int(a[2+int(a[1])+1]))
	case ATypIP4:
		host = net.IP(a[1 : 1+net.IPv4len]).String()
		port = strconv.Itoa((int(a[1+net.IPv4len]) << 8) | int(a[1+net.IPv4len+1]))
	case ATypIP6:
		host = net.IP(a[1 : 1+net.IPv6len]).String()
		port = strconv.Itoa((int(a[1+net.IPv6len]) << 8) | int(a[1+net.IPv6len+1]))
	}

	return net.JoinHostPort(host, port)
}

// ReadAddrBuf reads just enough bytes from r to get a valid Addr.
func ReadAddrBuf(r io.Reader, b []byte) (Addr, error) {
	if len(b) < MaxAddrLen {
		return nil, io.ErrShortBuffer
	}
	_, err := io.ReadFull(r, b[:1]) // read 1st byte for address type
	if err != nil {
		return nil, err
	}

	switch b[0] {
	case ATypDomain:
		_, err = io.ReadFull(r, b[1:2]) // read 2nd byte for domain length
		if err != nil {
			return nil, err
		}
		_, err = io.ReadFull(r, b[2:2+int(b[1])+2])
		return b[:1+1+int(b[1])+2], err
	case ATypIP4:
		_, err = io.ReadFull(r, b[1:1+net.IPv4len+2])
		return b[:1+net.IPv4len+2], err
	case ATypIP6:
		_, err = io.ReadFull(r, b[1:1+net.IPv6len+2])
		return b[:1+net.IPv6len+2], err
	}

	return nil, Errors[8]
}

// ReadAddr reads just enough bytes from r to get a valid Addr.
func ReadAddr(r io.Reader) (Addr, error) {
	return ReadAddrBuf(r, make([]byte, MaxAddrLen))
}

// SplitAddr slices a SOCKS address from beginning of b. Returns nil if failed.
func SplitAddr(b []byte) Addr {
	addrLen := 1
	if len(b) < addrLen {
		return nil
	}

	switch b[0] {
	case ATypDomain:
		if len(b) < 2 {
			return nil
		}
		addrLen = 1 + 1 + int(b[1]) + 2
	case ATypIP4:
		addrLen = 1 + net.IPv4len + 2
	case ATypIP6:
		addrLen = 1 + net.IPv6len + 2
	default:
		return nil
	}

	if len(b) < addrLen {
		return nil
	}

	return b[:addrLen]
}

// ParseAddr parses the address in string s. Returns nil if failed.
func ParseAddr(s string) Addr {
	var addr Addr
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			addr = make([]byte, 1+net.IPv4len+2)
			addr[0] = ATypIP4
			copy(addr[1:], ip4)
		} else {
			addr = make([]byte, 1+net.IPv6len+2)
			addr[0] = ATypIP6
			copy(addr[1:], ip)
		}
	} else {
		if len(host) > 255 {
			return nil
		}
		addr = make([]byte, 1+1+len(host)+2)
		addr[0] = ATypDomain
		addr[1] = byte(len(host))
		copy(addr[2:], host)
	}

	portnum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil
	}

	addr[len(addr)-2], addr[len(addr)-1] = byte(portnum>>8), byte(portnum)

	return addr
}
