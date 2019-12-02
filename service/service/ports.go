package service

import (
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"V2RayA/tools"
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
		return errors.New("端口之间不能重复，请检查")
	}
	if ports.Socks5 != p.Socks5 {
		p.Socks5 = ports.Socks5
		if ports.Socks5 != 0 {
			if o, w := tools.IsPortOccupied(strconv.Itoa(p.Socks5), "tcp"); o {
				arr := strings.Split(w, "/")
				if arr[1] != "v2ray" {
					return errors.New(fmt.Sprintf("%v端口已被%v占用，请检查", p.Socks5, w))
				}
			} else if o, w := tools.IsPortOccupied(strconv.Itoa(p.Socks5), "udp"); o {
				arr := strings.Split(w, "/")
				if arr[1] != "v2ray" {
					return errors.New(fmt.Sprintf("%v端口已被%v占用，请检查", p.Socks5, w))
				}
			}
		}
	}
	if ports.Http != p.Http {
		p.Http = ports.Http
		if ports.Http != 0 {
			if o, w := tools.IsPortOccupied(strconv.Itoa(p.Http), "tcp"); o {
				arr := strings.Split(w, "/")
				if arr[1] != "v2ray" {
					return errors.New(fmt.Sprintf("%v端口已被%v占用，请检查", p.Http, w))
				}
			}
		}
	}
	if ports.HttpWithPac != p.HttpWithPac {
		p.HttpWithPac = ports.HttpWithPac
		if ports.HttpWithPac != 0 {
			if o, w := tools.IsPortOccupied(strconv.Itoa(p.HttpWithPac), "tcp"); o {
				arr := strings.Split(w, "/")
				if arr[1] != "v2ray" {
					return errors.New(fmt.Sprintf("%v端口已被%v占用，请检查", p.HttpWithPac, w))
				}
			}
		}
	}
	err = configure.SetPorts(&p)
	if err != nil {
		return
	}
	return v2ray.UpdateV2rayWithConnectedServer()
}
