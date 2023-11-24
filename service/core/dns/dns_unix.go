//go:build !js && !windows

package dns

import (
	"bufio"
	"net/netip"
	"os"
	"strings"
)

func GetSystemDNS() (servers []netip.AddrPort) {
	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return defaultDNS
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		f := getFields(line)
		if len(f) < 1 {
			continue
		}
		switch f[0] {
		case "nameserver":
			if len(f) > 1 {
				if addr, err := netip.ParseAddr(f[1]); err == nil {
					servers = append(servers, netip.AddrPortFrom(addr, 53))
				}
			}
		}
	}
	return
}

func countAnyByte(s string, t string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if strings.IndexByte(t, s[i]) >= 0 {
			n++
		}
	}
	return n
}

func splitAtBytes(s string, t string) []string {
	a := make([]string, 1+countAnyByte(s, t))
	n := 0
	last := 0
	for i := 0; i < len(s); i++ {
		if strings.IndexByte(t, s[i]) >= 0 {
			if last < i {
				a[n] = s[last:i]
				n++
			}
			last = i + 1
		}
	}
	if last < len(s) {
		a[n] = s[last:]
		n++
	}
	return a[0:n]
}

func getFields(s string) []string { return splitAtBytes(s, " \r\t\n") }
