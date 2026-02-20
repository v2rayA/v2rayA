package service

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-leo/slicex"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
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

// AutoIncrease tries to increase ports automatically when ports are disabled. Only supports servers.
func AutoIncrease() {
	log.Info("AutoIncrease: start to auto increase vmess ports")
	// Get connected configures
	css := configure.GetConnectedServers()
	if css.Len() == 0 {
		log.Info("AutoIncrease: GetConnectedServers empty")
		return
	}

	server := configure.GetServers()

	b, e := json.Marshal(css)
	log.Info("AutoIncrease: css:", string(b), e)
	s, e := json.Marshal(server)
	log.Info("AutoIncrease: server", string(s), e)

	// Get servers from connected configures which are vmess protocol and have ports seted
	srvWt := make([]*configure.Which, 0, css.Len())
	for _, cs := range css.Get() {
		if cs.TYPE != configure.ServerType {
			continue
		}
		if cs.ID <= 0 || cs.ID > len(server) {
			log.Info("AutoIncrease: ID ", cs.ID, len(server))
			continue
		}
		ind := cs.ID - 1
		if server[ind].ServerObj.GetProtocol() != "vmess" {
			log.Info("AutoIncrease: Protocol ", server[ind].ServerObj.GetProtocol(), "vmess")
			continue
		}

		v2, err := serverObj.ParseVmessURL(server[ind].ServerObj.ExportToURL())
		if err != nil {
			log.Error("AutoIncrease: parse vmess url error: %v", err)
			continue
		}

		if v2.Ports == "" {
			continue
		}
		srvWt = append(srvWt, cs)
	}
	if len(srvWt) == 0 {
		log.Info("AutoIncrease: srvWt empty")
		return
	}

	wt, err := Ping(srvWt, 5*time.Second)
	if err != nil {
		log.Error("AutoIncrease: %v", err)
		return
	}

	for _, w := range wt {
		if w.Latency != "TIMEOUT" && w.Latency != "" {
			continue
		}
		srv := server[w.ID-1]
		v2, err := serverObj.ParseVmessURL(srv.ServerObj.ExportToURL())
		if err != nil {
			log.Error("AutoIncrease: parse vmess url error: %v", err)
			continue
		}
		ports, err := parsePorts(v2.Ports)
		if err != nil {
			log.Error("AutoIncrease: parse vmess ports error: %v", err)
			continue
		}
		if p, _ := strconv.Atoi(v2.Port); p >= ports[1] {
			v2.Port = strconv.Itoa(ports[0])
		} else {
			v2.Port = strconv.Itoa(p + 1)
		}
		log.Info("update server port to ", v2.Port)

		err = Import(v2.ExportToURL(), w)
		if err != nil {
			log.Error("AutoIncrease: import vmess url error: %v", err)
			continue
		}
	}
}

func parsePorts(portsStr string) (ports []int, err error) {
	if portsStr == "" {
		return
	}
	s := strings.Split(portsStr, "-")
	if len(s) != 2 {
		return nil, fmt.Errorf("invalid ports syntax")
	}
	start, err := strconv.Atoi(s[0])
	if err != nil {
		return nil, fmt.Errorf("invalid ports syntax")
	}
	end, err := strconv.Atoi(s[1])
	if err != nil {
		return nil, fmt.Errorf("invalid ports syntax")
	}
	if start > end || start <= 0 || end > 65535 {
		return nil, fmt.Errorf("invalid ports syntax")
	}
	ports = make([]int, 2)
	ports[0] = start
	ports[1] = end
	return ports, nil
}
