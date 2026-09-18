// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/netip"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

// From bsd/sys/socketvar.h: the kinds of the generic records that
// net.inet.*.pcblist_n returns.
const (
	xsoSocket = 0x001
	xsoInpcb  = 0x010
)

// platformOwnerOf reads the kernel's PCB list through sysctl, which carries
// the last PID to use each socket, then resolves the executable path with
// proc_pidpath. The list (bsd/netinet/in_pcblist.c get_pcblist_n) is an
// xinpgen header followed, per socket, by xgen records — xinpcb_n,
// xsocket_n, buffers, stats, xtcpcb_n — each starting with its length and
// kind and padded to 8 bytes. The inpcb record holds the local port (+18),
// the address family flag (+44) and the local address (+64 for IPv6, +76
// for IPv4); the socket record that follows holds so_last_pid (+68).
func platformOwnerOf(proto string, src netip.AddrPort) ([]owner, error) {
	var name string
	switch proto {
	case "tcp":
		name = "net.inet.tcp.pcblist_n"
	case "udp":
		name = "net.inet.udp.pcblist_n"
	default:
		return nil, fmt.Errorf("unknown protocol %q", proto)
	}
	value, err := syscall.Sysctl(name)
	if err != nil {
		return nil, fmt.Errorf("sysctl %s: %w", name, err)
	}
	buf := []byte(value)
	if len(buf) < 24 {
		return nil, nil
	}
	addr := src.Addr().Unmap()
	var wildcard uint32
	var haveWildcard bool
	// exact/wild say whether the inpcb just seen matched, so that the socket
	// record after it yields the pid.
	exact, wild := false, false
	for i := int(binary.LittleEndian.Uint32(buf[0:4])); i+8 <= len(buf); {
		ln := int(binary.LittleEndian.Uint32(buf[i : i+4]))
		kind := binary.LittleEndian.Uint32(buf[i+4 : i+8])
		if ln < 8 || i+ln > len(buf) {
			break
		}
		switch {
		case kind == xsoInpcb && ln >= 80:
			exact, wild = false, false
			if binary.BigEndian.Uint16(buf[i+18:i+20]) != src.Port() {
				break
			}
			flag := buf[i+44] // inp_vflag: 0x1 IPv4, 0x2 IPv6, both for a dual-stack socket
			var local netip.Addr
			switch {
			case flag&0x2 != 0:
				// An AF_INET6 socket keeps an IPv4 peer as a v4-mapped
				// address; Unmap makes it comparable with the flow's.
				local = netip.AddrFrom16([16]byte(buf[i+64 : i+80])).Unmap()
			case flag&0x1 != 0:
				local = netip.AddrFrom4([4]byte(buf[i+76 : i+80]))
			}
			if local.IsValid() && local.Is4() != addr.Is4() {
				local = netip.Addr{}
			}
			exact = local.IsValid() && local == addr
			wild = local.IsValid() && local.IsUnspecified()
		case kind == xsoSocket && ln >= 72 && (exact || wild):
			pid := binary.LittleEndian.Uint32(buf[i+68 : i+72]) // so_last_pid
			if exact {
				return []owner{{pid: pid, name: processName(pid)}}, nil
			}
			if !haveWildcard {
				wildcard, haveWildcard = pid, true
			}
			exact, wild = false, false
		}
		i += (ln + 7) &^ 7
	}
	if haveWildcard {
		return []owner{{pid: wildcard, name: processName(wildcard)}}, nil
	}
	return nil, nil
}

// processName reads the executable path from kern.procargs2, whose payload
// is argc followed by the NUL-terminated path; the 16-byte p_comm from
// kern.proc.pid is the fallback for processes that refuse it.
func processName(pid uint32) string {
	if raw, err := unix.SysctlRaw("kern.procargs2", int(pid)); err == nil && len(raw) > 4 {
		path := raw[4:]
		if n := bytes.IndexByte(path, 0); n > 0 {
			return filepath.Base(string(path[:n]))
		}
	}
	if kp, err := unix.SysctlKinfoProc("kern.proc.pid", int(pid)); err == nil {
		return unix.ByteSliceToString(kp.Proc.P_comm[:])
	}
	return ""
}
