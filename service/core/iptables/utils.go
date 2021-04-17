package iptables

import (
	"github.com/v2rayA/v2rayA/common/cmds"
	"net"
	"os"
	"strconv"
	"strings"
)

func IPNet2CIDR(ipnet *net.IPNet) string {
	ones, _ := ipnet.Mask.Size()
	return ipnet.IP.String() + "/" + strconv.Itoa(ones)
}

func GetLocalCIDR() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var cidrs []string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			cidrs = append(cidrs, IPNet2CIDR(ipnet))
		}
	}
	return cidrs, nil
}

func IsIPv6Supported() bool {
	b, err := os.ReadFile("/proc/sys/net/ipv6/conf/default/disable_ipv6")
	if err != nil {
		return false
	}
	if strings.TrimSpace(string(b)) == "1"{
		return false
	}
	return cmds.IsCommandValid("ip6tables")
}
