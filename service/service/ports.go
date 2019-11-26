package service

import (
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"V2RayA/tools"
	"errors"
	"fmt"
	"strconv"
)

func GetPorts() *configure.Ports {
	p := configure.GetPorts()
	if p.Socks5 == 0 {
		p.Socks5 = 20170
	}
	if p.Http == 0 {
		p.Http = 20171
	}
	if p.HttpWithPac == 0 {
		p.HttpWithPac = 20172
	}
	return p
}

func SetPorts(ports *configure.Ports) (err error) {
	p := configure.GetPorts()
	if ports.HttpWithPac == ports.Http || ports.HttpWithPac == ports.Socks5 || ports.Http == ports.Socks5 {
		return errors.New("端口之间不能重复，请检查")
	}
	if ports.Socks5 != 0 && ports.Socks5 != p.Socks5 {
		p.Socks5 = ports.Socks5
		if o, w := tools.IsPortOccupied(strconv.Itoa(p.Socks5)); o {
			return errors.New(fmt.Sprintf("%v端口已被%v占用，请检查", p.Socks5, w))
		}
	}
	if ports.Http != 0 && ports.Http != p.Http {
		p.Http = ports.Http
		if o, w := tools.IsPortOccupied(strconv.Itoa(p.Http)); o {
			return errors.New(fmt.Sprintf("%v端口已被%v占用，请检查", p.Http, w))
		}
	}
	if ports.HttpWithPac != 0 && ports.HttpWithPac != p.HttpWithPac {
		p.HttpWithPac = ports.HttpWithPac
		if o, w := tools.IsPortOccupied(strconv.Itoa(p.HttpWithPac)); o {
			return errors.New(fmt.Sprintf("%v端口已被%v占用，请检查", p.HttpWithPac, w))
		}
	}
	err = configure.SetPorts(p)
	if err != nil {
		return
	}
	return v2ray.UpdateV2rayWithConnectedServer()
}
