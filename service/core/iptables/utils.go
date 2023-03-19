package iptables

import (
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/cmds"
	"github.com/v2rayA/v2rayA/conf"
	"golang.org/x/net/nettest"
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

func IsNft() bool {
	if _, isNft := Redirect.(*nftRedirect); isNft {
		return true
	}
	return false
}

func IsIPv6Supported() bool {
	switch conf.GetEnvironmentConfig().IPV6Support {
	case "on":
		return true
	case "off":
		return false
	default:
	}
	if common.IsDocker() {
		return false
	}
	if !nettest.SupportsIPv6() {
		return false
	}
	if IsNft() {
		return true
	}
	return cmds.IsCommandValid("ip6tables") || cmds.IsCommandValid("ip6tables-nft")
}

func IsNftablesSupported() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	switch conf.GetEnvironmentConfig().NftablesSupport {
	// Warning:
	// This is an experimental feature for nftables support.
	// The default value is "off" for now but may be changed to "auto" in the future
	case "on":
		return true
	case "off":
		return false
	default:
	}
	if common.IsDocker() {
		return false
	}
	if !cmds.IsCommandValid("nft") {
		// No nft.
		return false
	}
	out, err := exec.Command("iptables", "--version").Output()
	if err != nil {
		// No iptables.
		return true
	}
	return strings.Contains(string(out), "nf_tables")
}
