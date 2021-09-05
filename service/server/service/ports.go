package service

import (
	"fmt"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
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
	if ports.HttpWithPac != 0 {
		set[ports.HttpWithPac] = struct{}{}
		cnt++
	}
	if ports.VlessGrpc != 0 {
		set[ports.VlessGrpc] = struct{}{}
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
	if ports.HttpWithPac != origin.HttpWithPac {
		origin.HttpWithPac = ports.HttpWithPac
		if origin.HttpWithPac != 0 {
			detectSyntax = append(detectSyntax, strconv.Itoa(origin.HttpWithPac)+":tcp")
		}
	}
	if ports.VlessGrpc != origin.VlessGrpc {
		origin.VlessGrpc = ports.VlessGrpc
		if origin.VlessGrpc != 0 {
			detectSyntax = append(detectSyntax, strconv.Itoa(origin.VlessGrpc)+":tcp")
		}
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
