package iptables

import (
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/cmds"
	"github.com/v2rayA/v2rayA/global"
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
	if global.GetEnvironmentConfig().ForceIPV6On {
		return true
	}
	if common.IsInDocker() {
		return false
	}
	if b, err := os.ReadFile("/proc/sys/net/ipv6/conf/default/disable_ipv6"); err != nil || strings.TrimSpace(string(b)) == "1" {
		return false
	}
	if b, err := os.ReadFile("/proc/sys/net/ipv6/conf/all/disable_ipv6"); err != nil || strings.TrimSpace(string(b)) == "1" {
		return false
	}
	return cmds.IsCommandValid("ip6tables")
}
