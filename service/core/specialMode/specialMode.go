package specialMode

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/db/configure"
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
	return setting.Transparent != configure.TransparentClose &&
		setting.AntiPollution != configure.AntipollutionClosed && (
		setting.TransparentType == configure.TransparentRedirect) ||
		setting.SpecialMode == configure.SpecialModeFakeDns
}

var couldListen struct {
	err        error
	lastUpdate time.Time
	mu         sync.Mutex
}

func CouldLocalDnsListen() (err error) {
	// cache for 3 seconds
	couldListen.mu.Lock()
	defer couldListen.mu.Unlock()
	if time.Since(couldListen.lastUpdate) < 3*time.Second {
		return couldListen.err
	}
	defer func() {
		couldListen.lastUpdate = time.Now()
		couldListen.err = err
	}()
	occupied, socket, err := ports.IsPortOccupied([]string{"53:udp"})
	if err != nil {
		return fmt.Errorf("%w: %v", DnsPortCheckFailedErr, err)
	}

	if occupied {
		p, err := socket.Process()
		if err != nil {
			return fmt.Errorf("%w: %v. please try again later", DnsPortCheckFailedErr, err)
		}
		if p.PPID == strconv.Itoa(os.Getpid()) {
			return nil
		}
		return fmt.Errorf("%w: %v(%v)", DnsPortOccupied, p.Name, p.PID)
	}
	return nil
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
