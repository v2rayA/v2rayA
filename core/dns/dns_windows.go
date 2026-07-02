package dns

import (
	"net/netip"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func GetSystemDNS() (servers []netip.AddrPort) {
	aas, err := adapterAddresses()
	if err != nil {
		return defaultDNS
	}
	for _, aa := range aas {
		for dns := aa.FirstDnsServerAddress; dns != nil; dns = dns.Next {
			if aa.OperStatus != windows.IfOperStatusUp {
				continue
			}
			sa, err := dns.Address.Sockaddr.Sockaddr()
			if err != nil {
				continue
			}
			var addr netip.Addr
			switch sa := sa.(type) {
			case *syscall.SockaddrInet4:
				addr = netip.AddrFrom4(sa.Addr)
			case *syscall.SockaddrInet6:
				if sa.Addr[0] == 0xfe && sa.Addr[1] == 0xc0 {
					continue
				}
				addr = netip.AddrFrom16(sa.Addr)
			default:
				continue
			}
			servers = append(servers, netip.AddrPortFrom(addr, 53))
		}
	}
	return
}

func adapterAddresses() ([]*windows.IpAdapterAddresses, error) {
	var b []byte
	l := uint32(15000)
	for {
		b = make([]byte, l)
		err := windows.GetAdaptersAddresses(syscall.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_PREFIX, 0, (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])), &l)
		if err == nil {
			if l == 0 {
				return nil, nil
			}
			break
		}
		if err.(syscall.Errno) != syscall.ERROR_BUFFER_OVERFLOW {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
		if l <= uint32(len(b)) {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
	}
	var aas []*windows.IpAdapterAddresses
	for aa := (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])); aa != nil; aa = aa.Next {
		aas = append(aas, aa)
	}
	return aas, nil
}
