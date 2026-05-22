package v2ray

import (
	"net"
	"runtime"
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
	// connectivitySocksAddr is the local SOCKS5 forward address used for probing.
	// It matches the "transparent" inbound port added by setInbound for TransparentTun
	// (tinytunSocksPort = 52345).  Using a local TCP dial avoids routing the probe
	// through the TUN device itself, which could produce false negatives while
	// wintun routes are being (re-)configured on Windows.
	connectivitySocksAddr = "127.0.0.1:52345"
)

// connectivityStartupDelay is the initial wait before the first connectivity
// probe.  On Windows it is larger to accommodate wintun driver initialisation.
var connectivityStartupDelay time.Duration

func init() {
	connectivityStartupDelay = 5 * time.Second
	if runtime.GOOS == "windows" {
		connectivityStartupDelay = 15 * time.Second
	}
}

// probePhysicalConnectivity checks whether the local SOCKS5 forward port is
// reachable.  A successful TCP dial to the loopback address confirms that the
// v2ray/xray inbound is up without routing the probe through the TUN device.
func probePhysicalConnectivity() bool {
	conn, err := net.DialTimeout("tcp", connectivitySocksAddr, connectivityCheckTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
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
	// Wait for the transparent proxy (e.g. TinyTun) to fully initialize before
	// the first connectivity check.  Probing too early on Windows can yield a
	// false network-unavailable result while the wintun driver is still setting
	// up routes, which would stop TinyTun prematurely and leave the frontend
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

	if !probePhysicalConnectivity() {
		if paused {
			return false, false
		}
		deleteTransparentProxyRulesKeepSystemProxy()
		m.mu.Lock()
		if m.p != nil && !m.testing {
			m.networkPaused = true
			m.mu.Unlock()
			ApiFeed.ProductMessage("running_state", map[string]interface{}{"running": false, "networkPaused": true})
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
