package specialMode

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common/netTools/netstat"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	DnsPortCheckFailedErr = fmt.Errorf("failed to check dns port occupation")
	DnsPortOccupied       = fmt.Errorf("dns port 53 is occupied")
)

func ShouldLocalDnsListen() bool {
	setting := configure.GetSettingNotNil()
	if setting.AntiPollution == configure.AntipollutionClosed {
		return false
	}
	if setting.Transparent == configure.TransparentClose {
		return false
	}
	if setting.TransparentType == configure.TransparentTproxy {
		return false
	}
	if conf.GetEnvironmentConfig().Lite {
		return false
	}
	return true
}

var couldListen struct {
	couldListenLocalhost bool
	err                  error
	lastUpdate           time.Time
	mu                   sync.Mutex
}

func CouldLocalDnsListen() (couldListenLocalhost bool, err error) {
	// cache for 3 seconds
	couldListen.mu.Lock()
	defer couldListen.mu.Unlock()
	if time.Since(couldListen.lastUpdate) < 3*time.Second {
		return couldListen.couldListenLocalhost, couldListen.err
	}
	defer func() {
		couldListen.lastUpdate = time.Now()
		couldListen.couldListenLocalhost = couldListenLocalhost
		couldListen.err = err
	}()
	occupied, sockets, err := ports.IsPortOccupied([]string{"53:udp"})
	if err != nil {
		return false, fmt.Errorf("%w: %v", DnsPortCheckFailedErr, err)
	}
	if err = netstat.FillProcesses(sockets); err != nil {
		return false, fmt.Errorf("%w: %v. please try again later", DnsPortCheckFailedErr, err)
	}
	//NOTICE: Special local address (127.2.0.17). Do not use v2ray.PortOccupied
	var occupiedErr error
	if occupied {
		// with PortSharing on, v2ray will try listening at 0.0.0.0, which conflicts with all IPs
		for _, socket := range sockets {
			p := socket.Proc
			if p == nil {
				continue
			}
			if p.PPID == strconv.Itoa(os.Getpid()) {
				continue
			}
			occupiedErr = fmt.Errorf("%w by %v(%v)", DnsPortOccupied, p.Name, p.PID)
			if socket.LocalAddress.IP.Equal(net.ParseIP("127.2.0.17")) {
				return false, occupiedErr
			}
			if socket.LocalAddress.IP.IsUnspecified() {
				return false, occupiedErr
			}
		}
	}
	return true, occupiedErr
}

func CouldUseSpecialMode() bool {
	setting := configure.GetSettingNotNil()
	switch setting.SpecialMode {
	case configure.SpecialModeFakeDns:
		return CouldUseFakeDns()
	case configure.SpecialModeSupervisor:
		return CouldUseSupervisor()
	}
	return true
}
