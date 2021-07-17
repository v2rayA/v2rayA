package specialMode

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/global"
	"os"
	"strconv"
)

var (
	DnsPortCheckFailedErr = fmt.Errorf("failed to check dns port occupation")
	DnsPortOccupied       = fmt.Errorf("dns port 53 is occupied")
)

func ShouldLocalDnsListen() bool {
	setting := configure.GetSettingNotNil()
	return setting.Transparent != configure.TransparentClose && !global.SupportTproxy
}
func CouldLocalDnsListen() error {
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
