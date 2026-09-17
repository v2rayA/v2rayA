package service

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/go-leo/slicex"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
)

func GetPorts() configure.Ports {
	p := configure.GetPortsNotNil()
	return *p
}

func SetPorts(ports *configure.Ports) (err error) {
	origin := GetPorts()
	set := map[int]struct{}{}
	for _, port := range []int{
		ports.Socks5,
		ports.Http,
		ports.Socks5WithPac,
		ports.HttpWithPac,
		ports.Vmess,
		ports.Api.Port,
	} {
		if port < 0 || port > 65535 {
			return common.Coded("INVALID_PORT", fmt.Errorf("port %d is outside valid range 0-65535", port), map[string]interface{}{"port": port})
		}
		if port == 0 {
			continue
		}
		if _, ok := set[port]; ok {
			return common.Coded("PORT_DUPLICATE", fmt.Errorf("port %d is assigned to more than one inbound; each port can be used once", port), map[string]interface{}{"port": port})
		}
		set[port] = struct{}{}
	}
	detectSyntax := make([]string, 0)
	if ports.Socks5 != origin.Socks5 {
		origin.Socks5 = ports.Socks5
		if origin.Socks5 != 0 {
			detectSyntax = append(detectSyntax, strconv.Itoa(origin.Socks5)+":tcp,udp")
		}
	}
	if ports.Http != origin.Http {
		origin.Http = ports.Http
		if origin.Http != 0 {
			detectSyntax = append(detectSyntax, strconv.Itoa(origin.Http)+":tcp")
		}
	}
	if ports.Socks5WithPac != origin.Socks5WithPac {
		origin.Socks5WithPac = ports.Socks5WithPac
		if origin.Socks5WithPac != 0 {
			detectSyntax = append(detectSyntax, strconv.Itoa(origin.Socks5WithPac)+":tcp,udp")
		}
	}
	if ports.HttpWithPac != origin.HttpWithPac {
		origin.HttpWithPac = ports.HttpWithPac
		if origin.HttpWithPac != 0 {
			detectSyntax = append(detectSyntax, strconv.Itoa(origin.HttpWithPac)+":tcp")
		}
	}
	if ports.Vmess != origin.Vmess {
		origin.Vmess = ports.Vmess
		if origin.Vmess != 0 {
			detectSyntax = append(detectSyntax, strconv.Itoa(origin.Vmess)+":tcp")
		}
	}
	if ports.Api.Port != origin.Api.Port || !reflect.DeepEqual(ports.Api.Services, origin.Api.Services) {
		origin.Api = ports.Api
		if origin.Api.Port != 0 {
			detectSyntax = append(detectSyntax, strconv.Itoa(origin.Api.Port)+":tcp")
		}
		// logger service is required
		origin.Api.Services = slicex.Uniq(append(origin.Api.Services, "LoggerService"))
	}
	if err = v2ray.PortOccupied(detectSyntax); err != nil {
		return err
	}
	if err = configure.SetPorts(&origin); err != nil {
		return err
	}
	if v2ray.ProcessManager.Running() {
		err = v2ray.UpdateV2RayConfig()
	}
	return
}
