package service

import (
	ports2 "V2RayA/common/ports"
	"V2RayA/core/v2ray"
	"V2RayA/persistence/configure"
	"errors"
	"fmt"
	"strconv"
)

func GetPortsDefault() configure.Ports {
	p := configure.GetPorts()
	if p == nil {
		p = new(configure.Ports)
		p.Socks5 = 20170
		p.Http = 20171
		p.HttpWithPac = 20172
	}
	return *p
}

func SetPorts(ports *configure.Ports) (err error) {
	p := GetPortsDefault()
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
	if ports.HttpWithPac != 0 {
		set[ports.HttpWithPac] = struct{}{}
		cnt++
	}
	if cnt > len(set) {
		return errors.New("ports duplicate. check it")
	}
	detectSyntax := make([]string, 0, 3)
	if ports.Socks5 != p.Socks5 {
		p.Socks5 = ports.Socks5
		if p.Socks5 != 0 {
			detectSyntax = append(detectSyntax, strconv.Itoa(p.Socks5)+":tcp,udp")
		}
	}
	if ports.Http != p.Http {
		p.Http = ports.Http
		if p.Http != 0 {
			detectSyntax = append(detectSyntax, strconv.Itoa(p.Http)+":tcp")
		}
	}
	if ports.HttpWithPac != p.HttpWithPac {
		p.HttpWithPac = ports.HttpWithPac
		if p.HttpWithPac != 0 {
			detectSyntax = append(detectSyntax, strconv.Itoa(p.HttpWithPac)+":tcp")
		}
	}
	if o, v := ports2.IsPortOccupied(detectSyntax); o {
		process, err := v.Process()
		if err == nil && process.Name != "v2ray" {
			return errors.New(fmt.Sprintf("port %v is occupied by %v", v.LocalAddress.Port, process.Name))
		}
	}
	err = configure.SetPorts(&p)
	if err != nil {
		return
	}
	return v2ray.UpdateV2RayConfig(nil)
}
