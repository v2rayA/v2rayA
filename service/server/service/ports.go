package service

import (
	"fmt"
	"github.com/go-leo/slicex"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"reflect"
	"strconv"
)

func GetPorts() configure.Ports {
	p := configure.GetPortsNotNil()
	return *p
}

func SetPorts(ports *configure.Ports) (err error) {
	origin := GetPorts()
	set := map[int]struct{}{}
	cnt := 0
	if ports.Socks5 != 0 {
		set[ports.Socks5] = struct{}{}
		cnt++
	}
	if ports.Http != 0 {
		set[ports.Http] = struct{}{}
		cnt++
	}
	if ports.Socks5WithPac != 0 {
		set[ports.Socks5WithPac] = struct{}{}
		cnt++
	}
	if ports.HttpWithPac != 0 {
		set[ports.HttpWithPac] = struct{}{}
		cnt++
	}
	if ports.Vmess != 0 {
		set[ports.Vmess] = struct{}{}
		cnt++
	}
	if cnt > len(set) {
		return fmt.Errorf("ports duplicate. check it")
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
