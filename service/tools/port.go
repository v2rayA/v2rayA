package tools

import (
	"github.com/cakturk/go-netstat/netstat"
	"strconv"
	"strings"
)

/*
protocol: tcp tcp6 udp udp6
*/
func IsPortOccupied(port string, protocol string) (occupied bool, which string) {
	pint, _ := strconv.Atoi(port)
	p := uint16(pint)
	var tabs []netstat.SockTabEntry
	var tmp []netstat.SockTabEntry
	var err error
	switch strings.ToLower(protocol) {
	case "tcp":
		tabs, err = netstat.TCPSocks(func(s *netstat.SockTabEntry) bool {
			return s.LocalAddr.Port == p
		})
		tmp, err = netstat.TCP6Socks(func(s *netstat.SockTabEntry) bool {
			return s.LocalAddr.Port == p
		})
		tabs = append(tabs, tmp...)
	case "udp":
		tabs, err = netstat.UDPSocks(func(s *netstat.SockTabEntry) bool {
			return s.LocalAddr.Port == p
		})
		tmp, err = netstat.UDP6Socks(func(s *netstat.SockTabEntry) bool {
			return s.LocalAddr.Port == p
		})
		tabs = append(tabs, tmp...)
	}
	if err == nil && len(tabs) > 0 {
		which = tabs[0].Process.String()
		occupied = true
	}
	return
}
