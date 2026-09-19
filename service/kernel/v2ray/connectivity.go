package v2ray

import (
	"net"
	"runtime"
	"strconv"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	connectivityCheckInterval = 15 * time.Second
	connectivityCheckTimeout  = 5 * time.Second
	// connectivityBackoffBase is the initial retry delay after the first consecutive
	// probe failure (30 s).  Subsequent failures double the delay up to
	// connectivityBackoffMax to avoid a rapid stop→start oscillation loop.
	connectivityBackoffBase = 30 * time.Second
	// connectivityBackoffMax caps the exponential backoff delay.
	connectivityBackoffMax = 120 * time.Second
)

// connectivityProbeAddr returns the core's API listener, which every mode
// has. A successful TCP dial to loopback says the core process is alive;
// it says nothing about the TUN itself. A failing TUN device is reported
// by the core's inbound and does not stop the core.
func connectivityProbeAddr(t *Template) string {
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(t.ApiPort))
}

// connectivityStartupDelay is the initial wait before the first connectivity
// probe.  On Windows it is larger to accommodate wintun driver initialisation.
var connectivityStartupDelay time.Duration

func init() {
	connectivityStartupDelay = 5 * time.Second
	if runtime.GOOS == "windows" {
		connectivityStartupDelay = 15 * time.Second
	}
}

// probePhysicalConnectivity checks whether the core's API port is reachable.
// A TCP dial to loopback never routes through the TUN device.
func probePhysicalConnectivity(t *Template) bool {
	conn, err := net.DialTimeout("tcp", connectivityProbeAddr(t), connectivityCheckTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// tunDataPathDead reports a tun-mode device that went away while the core
// process lives: the inbound closes its device when the pumps fail, and
// nothing recreates it. Every routed packet is being dropped, and no later
// probe can succeed, so this is a stop, not a pause.
func tunDataPathDead(t *Template) bool {
	return t.Setting != nil && t.Setting.TransparentType == configure.TransparentTun && !tunDeviceAlive()
}

// connectivityBackoffDelay returns the delay to wait after consecutiveFailures
// consecutive probe failures, using exponential backoff capped at
// connectivityBackoffMax.
func connectivityBackoffDelay(consecutiveFailures int) time.Duration {
	delay := connectivityBackoffBase
	for i := 1; i < consecutiveFailures; i++ {
		delay *= 2
		if delay >= connectivityBackoffMax {
			return connectivityBackoffMax
		}
	}
	return delay
}

func (m *CoreProcessManager) startConnectivityMonitor(t *Template) {
	if t == nil || t.Setting == nil || !IsTransparentOn(t.Setting) || t.Setting.TransparentType == configure.TransparentSystemProxy {
		return
	}

	m.mu.Lock()
	if m.testing || m.p == nil {
		m.mu.Unlock()
		return
	}
	m.stopConnectivityMonitorLocked()
	stopCh := make(chan struct{})
	m.connectivityStop = stopCh
	m.mu.Unlock()

	go m.connectivityLoop(stopCh, t)
}

func (m *CoreProcessManager) connectivityLoop(stopCh chan struct{}, t *Template) {
	// Wait for the transparent proxy to fully initialize before
	// the first connectivity check.  Probing too early on Windows can yield a
	// false network-unavailable result while the wintun driver is still setting
	// up routes, which would tear the routes down prematurely and leave the frontend
	// stuck on "检测中" (Checking).
	select {
	case <-stopCh:
		return
	case <-time.After(connectivityStartupDelay):
	}

	failureCount := 0
	for {
		stop, healthy := m.syncConnectivityState(t)
		if stop {
			return
		}

		var delay time.Duration
		if !healthy {
			failureCount++
			delay = connectivityBackoffDelay(failureCount)
		} else {
			failureCount = 0
			delay = connectivityCheckInterval
		}

		select {
		case <-stopCh:
			return
		case <-time.After(delay):
		}
	}
}

// syncConnectivityState checks and updates the transparent proxy connectivity
// state.  It returns (stop, healthy): stop=true means the monitor goroutine
// should exit; healthy=true means the local SOCKS5 probe succeeded.
func (m *CoreProcessManager) syncConnectivityState(t *Template) (stop bool, healthy bool) {
	if t == nil || t.Setting == nil {
		return true, false
	}

	m.mu.Lock()
	if m.p == nil || m.testing {
		m.mu.Unlock()
		return true, false
	}
	paused := m.networkPaused
	setting := t.Setting
	m.mu.Unlock()

	if setting.TransparentType == configure.TransparentSystemProxy || !IsTransparentOn(setting) {
		return false, true
	}

	dead := tunDataPathDead(t)
	if dead || !probePhysicalConnectivity(t) {
		if paused {
			return false, false
		}
		deleteTransparentProxyRulesKeepSystemProxy()
		m.mu.Lock()
		if m.p != nil && !m.testing {
			m.networkPaused = true
			m.mu.Unlock()
			ApiFeed.ProductMessage("running_state", map[string]interface{}{"running": false, "networkPaused": true})
			if dead {
				// Resuming would wait for a device that is not coming
				// back and rerun the whole setup, hooks included, on
				// every check. The user starts the proxy again.
				log.Warn("tun: the device is gone while the core is running; the transparent proxy is paused until it is started again")
				return true, false
			}
			log.Info("connectivity lost, transparent proxy paused")
			return false, false
		}
		m.mu.Unlock()
		return true, false
	}

	if !paused {
		return false, true
	}
	if err := m.CheckAndSetupTransparentProxy(false, setting, t); err != nil {
		log.Warn("failed to resume transparent proxy after network recovery: %v", err)
		return false, true
	}
	m.mu.Lock()
	if m.p != nil && !m.testing {
		m.networkPaused = false
		m.mu.Unlock()
		ApiFeed.ProductMessage("running_state", map[string]interface{}{"running": true, "networkPaused": false})
		log.Info("connectivity restored, transparent proxy resumed")
		return false, true
	}
	m.mu.Unlock()
	return true, true
}
