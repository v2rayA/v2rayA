package v2ray

import (
	"errors"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// The TUN transparent proxy lives in the core: a "tun-mips" inbound opens
// the device, assigns its address and turns packets into dispatcher
// sessions. The service's part is everything the operating system must
// know — routes, policy rules, system DNS — and the settings that inbound
// is built from. The addresses are shared by both halves.
const (
	tunAddress4 = "10.0.85.2/30"
	tunGateway4 = "10.0.85.1"
	tunAddress6 = "fdfe:dcba:9876::2/126"
	tunGateway6 = "fdfe:dcba:9876::1"
	tunMTU      = 1500
)

// ErrTunUnsupported is returned where no route configuration exists.
var ErrTunUnsupported = errors.New("the built-in TUN is not supported on this platform")

// TunSupported reports whether the built-in TUN can run here; it drives
// the GUI's option and the setting's start-up check.
func TunSupported() bool { return tunSupported }

// tunDnsTarget is where the core relays intercepted DNS queries: the DNS
// module's listener, with a wildcard bind translated to loopback.
func tunDnsTarget(setting *configure.Setting) string {
	host, port, err := net.SplitHostPort(dnsModuleListenAddr(setting))
	if err != nil {
		return "127.0.0.1:52353"
	}
	if ip := net.ParseIP(host); ip == nil || ip.IsUnspecified() {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}

// dnsModulePort is the port the DNS module listens on, which the DNS
// REDIRECT and TPROXY rules must target as well.
func dnsModulePort(setting *configure.Setting) string {
	return setting.DnsModulePort()
}

// tunInboundSettings builds the seam the core's loader validates. The
// user's exclusion list is extended with v2rayA and the core themselves by
// name, and v2rayA by PID so a renamed binary still counts; the core adds
// its own PID.
func tunInboundSettings(setting *configure.Setting) *coreObj.InboundSettings {
	s := &coreObj.InboundSettings{
		Name:             tunDeviceName,
		MTU:              tunMTU,
		Address4:         tunAddress4,
		DnsTarget:        tunDnsTarget(setting),
		ExcludeProcesses: collectExcludeProcesses(setting.TunExcludeProcesses),
		SelfPids:         []uint32{uint32(os.Getpid())},
		DirectTag:        "direct",
	}
	if tunIPv6Enabled() {
		s.Address6 = tunAddress6
	}
	return s
}

// tunMu serialises install and undo: the process manager's start path and
// the connectivity monitor's pause/resume both reach here without another
// lock in common.
var tunMu sync.Mutex

// tunTeardown holds the teardown script that pairs with the setup script
// that ran, so stop runs it even after the settings changed underneath;
// empty when no setup script ran.
var (
	tunScriptRan bool
	tunTeardown  string
)

// startTunCore installs the operating-system side once the core has the
// device up: routes and rules, or the user's setup script when automatic
// routing is off.
func startTunCore(tmpl *Template) error {
	if !TunSupported() {
		return ErrTunUnsupported
	}
	tunMu.Lock()
	defer tunMu.Unlock()
	setting := configure.GetSettingNotNil()
	if !setting.TunAutoRoute {
		if err := runTunRouteScript("setup", setting.TunSetupScript); err != nil {
			log.Warn("tun setup script error: %v", err)
		}
		tunScriptRan = true
		tunTeardown = setting.TunTeardownScript
		return nil
	}
	return tunRoutesUp(tmpl, collectNodeIPs(tmpl))
}

// stopTunCore removes what startTunCore installed. It runs before the core
// is stopped so that no packet is routed into a device that is going away.
// What to undo is decided by what was installed, not by the settings: they
// may already describe the next mode.
func stopTunCore() {
	if !TunSupported() {
		return
	}
	tunMu.Lock()
	defer tunMu.Unlock()
	if tunScriptRan {
		tunScriptRan = false
		if err := runTunRouteScript("teardown", tunTeardown); err != nil {
			log.Warn("tun teardown script error: %v", err)
		}
		tunTeardown = ""
	}
	if tunInstalled() {
		tunRoutesDown()
	}
}

// tunEgressInterfaceIfTun names the physical interface for socket binding
// when the built-in TUN is the active transparent proxy; empty otherwise
// and on Linux, where the socket mark does the job.
func tunEgressInterfaceIfTun(setting *configure.Setting) string {
	if setting.TransparentType != configure.TransparentTun || !IsTransparentOn(setting) {
		return ""
	}
	iface := tunEgressInterface()
	if iface == "" && tunBindsEgress {
		log.Warn("tun: no default route to bind the core's sockets to; if the network comes up later, restart the transparent proxy")
	}
	return iface
}

// tunDeviceAlive reports whether the core's TUN device still exists. The
// inbound closes it when its pumps fail, so a missing device while tun is
// the active mode means the data path is dead although the core process
// is not; the connectivity monitor treats that as lost connectivity.
func tunDeviceAlive() bool {
	tunMu.Lock()
	defer tunMu.Unlock()
	if !tunInstalled() {
		return true
	}
	return tunDevicePresent()
}

// CleanupTunResidual removes what a previous run left behind (routes, rules,
// the macOS resolver setting). The service calls it at startup so a crash
// or kill -9 is repaired before anything else, not only when the proxy is
// started again.
func CleanupTunResidual() {
	if !TunSupported() {
		return
	}
	tunMu.Lock()
	defer tunMu.Unlock()
	tunCleanupResidual()
}

func itoa(i int) string { return strconv.Itoa(i) }
