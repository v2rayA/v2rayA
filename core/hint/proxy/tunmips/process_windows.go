// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	iphlpapi                = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetExtendedTcpTable = iphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable = iphlpapi.NewProc("GetExtendedUdpTable")
)

// From iprtrmib.h / udpmib.h.
const (
	tcpTableOwnerPidAll = 5
	udpTableOwnerPid    = 1
)

// platformOwnerOf reads the owner-PID tables iphlpapi maintains. Ports in
// the rows are in network byte order in the low 16 bits of a DWORD.
func platformOwnerOf(proto string, src netip.AddrPort) ([]owner, error) {
	addr := src.Addr().Unmap()
	family := uint32(windows.AF_INET)
	if addr.Is6() {
		family = windows.AF_INET6
	}
	var (
		table []byte
		err   error
	)
	switch proto {
	case "tcp":
		table, err = extendedTable(procGetExtendedTcpTable, family, tcpTableOwnerPidAll)
	case "udp":
		table, err = extendedTable(procGetExtendedUdpTable, family, udpTableOwnerPid)
	default:
		return nil, fmt.Errorf("unknown protocol %q", proto)
	}
	if err != nil {
		return nil, err
	}
	pid, ok := findOwnerPid(table, proto, addr, src.Port())
	if !ok {
		return nil, nil
	}
	return []owner{{pid: pid, name: processName(pid)}}, nil
}

func extendedTable(proc *windows.LazyProc, family uint32, class uint32) ([]byte, error) {
	var size uint32
	for i := 0; i < 8; i++ {
		buf := make([]byte, size)
		var p unsafe.Pointer
		if size > 0 {
			p = unsafe.Pointer(&buf[0])
		}
		r, _, _ := proc.Call(uintptr(p), uintptr(unsafe.Pointer(&size)), 0, uintptr(family), uintptr(class), 0)
		switch windows.Errno(r) {
		case 0:
			return buf, nil
		case windows.ERROR_INSUFFICIENT_BUFFER:
			continue
		default:
			return nil, fmt.Errorf("%s: %w", proc.Name, windows.Errno(r))
		}
	}
	return nil, fmt.Errorf("%s: table keeps growing", proc.Name)
}

// findOwnerPid walks the rows. Layouts (all little-endian DWORDs):
//
//	MIB_TCPROW_OWNER_PID   state, localAddr, localPort, remoteAddr, remotePort, pid          (24 bytes)
//	MIB_TCP6ROW_OWNER_PID  localAddr[16], localScope, localPort, remoteAddr[16], remoteScope, remotePort, state, pid (56 bytes)
//	MIB_UDPROW_OWNER_PID   localAddr, localPort, pid                                          (12 bytes)
//	MIB_UDP6ROW_OWNER_PID  localAddr[16], localScope, localPort, pid                          (28 bytes)
func findOwnerPid(table []byte, proto string, addr netip.Addr, port uint16) (uint32, bool) {
	if len(table) < 4 {
		return 0, false
	}
	n := int(binary.LittleEndian.Uint32(table))
	rows := table[4:]
	var rowSize, addrOff, portOff, pidOff int
	switch {
	case proto == "tcp" && addr.Is4():
		rowSize, addrOff, portOff, pidOff = 24, 4, 8, 20
	case proto == "tcp":
		rowSize, addrOff, portOff, pidOff = 56, 0, 20, 52
	case addr.Is4():
		rowSize, addrOff, portOff, pidOff = 12, 0, 4, 8
	default:
		rowSize, addrOff, portOff, pidOff = 28, 0, 20, 24
	}
	var wildcard uint32
	var haveWildcard bool
	for i := 0; i < n && (i+1)*rowSize <= len(rows); i++ {
		row := rows[i*rowSize : (i+1)*rowSize]
		rowPort := binary.BigEndian.Uint16(row[portOff : portOff+2])
		if rowPort != port {
			continue
		}
		var local netip.Addr
		if addr.Is4() {
			local = netip.AddrFrom4([4]byte(row[addrOff : addrOff+4]))
		} else {
			local = netip.AddrFrom16([16]byte(row[addrOff : addrOff+16]))
		}
		pid := binary.LittleEndian.Uint32(row[pidOff : pidOff+4])
		if local == addr {
			return pid, true
		}
		if local.IsUnspecified() && !haveWildcard {
			wildcard, haveWildcard = pid, true
		}
	}
	return wildcard, haveWildcard
}

func processName(pid uint32) string {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)
	buf := make([]uint16, windows.MAX_PATH)
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(h, 0, &buf[0], &size); err != nil {
		return ""
	}
	return filepath.Base(windows.UTF16ToString(buf[:size]))
}
