package v2ray

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/v2rayA/v2rayA/common/netTools/netstat"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func (t *Template) checkDuplicatedTags() error {
	inboundTagsSet := make(map[string]interface{})
	for _, in := range t.Inbounds {
		tag := in.Tag
		if _, exists := inboundTagsSet[tag]; exists {
			return fmt.Errorf("duplicated inbound tag: %v", tag)
		} else {
			inboundTagsSet[tag] = nil
		}
	}
	outboundTagsSet := make(map[string]interface{})
	for _, out := range t.Outbounds {
		tag := out.Tag
		if _, exists := outboundTagsSet[tag]; exists {
			return fmt.Errorf("duplicated outbound tag: %v", tag)
		} else {
			outboundTagsSet[tag] = nil
		}
	}
	return nil
}

func (t *Template) checkDuplicatedInboundSockets() error {
	inboundSocketSet := make(map[string]interface{})
	for _, in := range t.Inbounds {
		if in.Listen == "" {
			// https://www.v2fly.org/config/inbounds.html#inboundobject
			in.Listen = "0.0.0.0"
		}
		socket := net.JoinHostPort(in.Listen, strconv.Itoa(in.Port))
		if _, exists := inboundSocketSet[socket]; exists {
			return fmt.Errorf("duplicated inbound listening address: %v", socket)
		} else {
			inboundSocketSet[socket] = nil
		}
	}
	return nil
}

var OccupiedErr = fmt.Errorf("port is occupied")

func PortOccupied(syntax []string) (err error) {
	occupied, sockets, err := ports.IsPortOccupied(syntax)
	if err != nil {
		if errors.Is(err, netstat.ErrorNotSupportOSErr) {
			log.Trace("PortOccupied: %v", err)
			return nil
		}
		return
	}
	if occupied {
		if err = netstat.FillProcesses(sockets); err != nil {
			if errors.Is(err, netstat.ErrorNotSupportOSErr) {
				log.Warn("cannot judge port occupation: %v", err)
				return nil
			}
			return fmt.Errorf("failed to check if port is occupied: %w", err)
		}
		for _, s := range sockets {
			p := s.Proc
			if p == nil {
				continue
			}
			if ownPID := strconv.Itoa(os.Getpid()); p.PPID == ownPID ||
				p.PID == ownPID {
				continue
			}
			occupiedErr := fmt.Errorf("%w by %v(%v): %v", OccupiedErr, p.Name, p.PID, s.LocalAddress.Port)
			if configure.GetSettingNotNil().PortSharing {
				// want to listen 0.0.0.0, which conflicts with all IPs
				return occupiedErr
			}
			if s.LocalAddress.IP.IsUnspecified() {
				return occupiedErr
			}
			if s.LocalAddress.IP.IsLoopback() {
				return occupiedErr
			}
		}
	}
	return nil
}

func (t *Template) CheckInboundPortsOccupied() (err error) {
	var st []string
	for _, in := range t.Inbounds {
		switch strings.ToLower(in.Protocol) {
		case "http", "vmess", "vless", "trojan":
			st = append(st, strconv.Itoa(in.Port)+":tcp")
		case "dokodemo-door":
			if strings.HasPrefix(in.Tag, "dns-in") {
				// checked before
				continue
			} else if in.Settings != nil && in.Settings.Network != "" {
				st = append(st, strconv.Itoa(in.Port)+":"+in.Settings.Network)
			} else {
				st = append(st, strconv.Itoa(in.Port)+":tcp,udp")
			}
		default:
			st = append(st, strconv.Itoa(in.Port)+":tcp,udp")
		}
	}
	return PortOccupied(st)
}
