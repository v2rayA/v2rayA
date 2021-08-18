package service

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/global"
	"os"
	"strconv"
)

func GetPorts() configure.Ports {
	p := configure.GetPorts()
	if p == nil {
		p = new(configure.Ports)
		p.Socks5 = 20170
		p.Http = 20171
		p.HttpWithPac = 20172
		p.VlessGrpc = 0
	}
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
		return newError("ports duplicate. check it")
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
	if err = checkAndGenerateCertForVlessGrpc(); err != nil {
		return err
	}
	if err = configure.SetPorts(&origin); err != nil {
		return err
	}
	if v2ray.IsV2RayRunning() {
		err = v2ray.UpdateV2RayConfig()
	}
	return
}

func checkAndGenerateCertForVlessGrpc() (err error) {
	if config := global.GetEnvironmentConfig(); len(config.VlessGrpcInboundCertKey) < 2 {
		// no specified cert and key
		cert := "/etc/v2raya/vlessGrpc.crt"
		key := "/etc/v2raya/vlessGrpc.key"
		_, eCert := os.Stat(cert)
		_, eKey := os.Stat(cert)
		if os.IsNotExist(eCert) || os.IsNotExist(eKey) {
			// no such files, generate them
			if err := common.GenerateCertKey(cert, key, ""); err != nil {
				return fmt.Errorf("failed to generate certificates for vlessGrpc inbound: %w", err)
			}
		}
	}
	return nil
}
