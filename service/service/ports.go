package service

import (
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	ports2 "V2RayA/tools/ports"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	if ports.Socks5 != p.Socks5 {
		p.Socks5 = ports.Socks5
		if ports.Socks5 != 0 {
			if o, w := ports2.IsPortOccupied(strconv.Itoa(p.Socks5), "tcp", true); o {
				arr := strings.Split(w, "/")
				if arr[1] != "v2ray" {
					return errors.New(fmt.Sprintf("port %v is occupied by %v", p.Socks5, w))
				}
			} else if o, w := ports2.IsPortOccupied(strconv.Itoa(p.Socks5), "udp", true); o {
				arr := strings.Split(w, "/")
				if arr[1] != "v2ray" {
					return errors.New(fmt.Sprintf("port %v is occupied by %v", p.Socks5, w))
				}
			}
		}
	}
	if ports.Http != p.Http {
		p.Http = ports.Http
		if ports.Http != 0 {
			if o, w := ports2.IsPortOccupied(strconv.Itoa(p.Http), "tcp", true); o {
				arr := strings.Split(w, "/")
				if arr[1] != "v2ray" {
					return errors.New(fmt.Sprintf("port %v is occupied by %v", p.Http, w))
				}
			}
		}
	}
	if ports.HttpWithPac != p.HttpWithPac {
		p.HttpWithPac = ports.HttpWithPac
		if ports.HttpWithPac != 0 {
			if o, w := ports2.IsPortOccupied(strconv.Itoa(p.HttpWithPac), "tcp", true); o {
				arr := strings.Split(w, "/")
				if arr[1] != "v2ray" {
					return errors.New(fmt.Sprintf("port %v is occupied by %v", p.HttpWithPac, w))
				}
			}
		}
	}
	err = configure.SetPorts(&p)
	if err != nil {
		return
	}
	return v2ray.UpdateV2RayConfig(nil)
}
