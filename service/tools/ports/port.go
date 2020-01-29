package ports

import (
	"github.com/cakturk/go-netstat/netstat"
	"strconv"
	"strings"
)

/*
protocol: tcp udp
loose: 宽松模式
*/
func IsPortOccupied(port string, protocol string, loose bool) (occupied bool, which string) {
	pint, _ := strconv.Atoi(port)
	p := uint16(pint)
	var tabs []netstat.SockTabEntry
	var tmp []netstat.SockTabEntry
	var err error
	couldBeOccupy := func(s *netstat.SockTabEntry) bool {
		if s.LocalAddr.Port != p {
			return false
		}
		if loose {
			switch s.State {
			case netstat.FinWait1, netstat.FinWait2, netstat.Close, netstat.CloseWait, netstat.Closing, netstat.TimeWait, netstat.LastAck:
				return false
			}
			return true
		} else {
			switch s.State {
			case netstat.FinWait1, netstat.FinWait2, netstat.Close, netstat.CloseWait, netstat.Closing, netstat.TimeWait, netstat.LastAck, netstat.Established:
				return false
			}
			return true
		}

	}
	switch strings.ToLower(protocol) {
	case "tcp":
		tabs, err = netstat.TCPSocks(couldBeOccupy)
		tmp, err = netstat.TCP6Socks(couldBeOccupy)
		tabs = append(tabs, tmp...)
	case "udp":
		tabs, err = netstat.UDPSocks(couldBeOccupy)
		tmp, err = netstat.UDP6Socks(couldBeOccupy)
		tabs = append(tabs, tmp...)
	}
	if err == nil && len(tabs) > 0 {
		if tabs[0].Process != nil {
			which = tabs[0].Process.String()
		}
		occupied = true
	}
	return
}
