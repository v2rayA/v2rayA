package v2ray

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/netTools/netstat"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray/where"
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

// killOrphanCore terminates a core process that was reparented to init after
// the v2rayA that started it died. It reports whether the port is free again.
func killOrphanCore(p *netstat.Process) bool {
	if runtime.GOOS == "windows" || p.PPID != "1" {
		return false
	}
	binPath, err := where.GetV2rayBinPath()
	if err != nil || filepath.Base(binPath) != p.Name {
		return false
	}
	pid, err := strconv.Atoi(p.PID)
	if err != nil {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	log.Warn("killing the orphaned %v (pid %v) that still holds the port", p.Name, p.PID)
	if err := proc.Kill(); err != nil {
		log.Warn("cannot kill the orphaned %v (pid %v): %v", p.Name, p.PID, err)
		return false
	}
	// Not our child, so Wait cannot reap it; give the kernel a moment to drop
	// the socket before the caller looks again.
	_, _ = proc.Wait()
	time.Sleep(200 * time.Millisecond)
	return true
}

var OccupiedErr = fmt.Errorf("is already in use")

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
			// A core that outlived the v2rayA which started it — the OOM
			// killer takes v2rayA first, init adopts the core and it keeps the
			// ports, so every later start failed until someone killed it by
			// hand. It is ours to clean up.
			if killOrphanCore(p) {
				continue
			}
			occupiedErr := fmt.Errorf("port %d %w by %v (pid %v)", s.LocalAddress.Port, OccupiedErr, p.Name, p.PID)
			codedErr := common.Coded("PORT_OCCUPIED", occupiedErr, map[string]interface{}{"port": s.LocalAddress.Port})
			if configure.GetSettingNotNil().PortSharing {
				// want to listen 0.0.0.0, which conflicts with all IPs
				return codedErr
			}
			if s.LocalAddress.IP.IsUnspecified() {
				return codedErr
			}
			if s.LocalAddress.IP.IsLoopback() {
				return codedErr
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
