package v2ray

import (
	"net/http"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	connectivityProbeURL      = "https://connectivitycheck.gstatic.com/generate_204"
	connectivityCheckInterval = 15 * time.Second
	connectivityCheckTimeout  = 5 * time.Second
	// connectivityStartupDelay gives TinyTun time to fully set up the TUN interface
	// and routing rules before the first probe.  On Windows this is especially
	// important because the wintun driver briefly disrupts all traffic while it
	// installs routes, which would otherwise cause a spurious networkPaused=true.
	connectivityStartupDelay = 5 * time.Second
)

func probePhysicalConnectivity() bool {
	client := &http.Client{Timeout: connectivityCheckTimeout}
	req, err := http.NewRequest(http.MethodGet, connectivityProbeURL, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
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

	ticker := time.NewTicker(connectivityCheckInterval)
	defer ticker.Stop()

	for {
		if m.syncConnectivityState(t) {
			return
		}
		select {
		case <-stopCh:
			return
		case <-ticker.C:
		}
	}
}

func (m *CoreProcessManager) syncConnectivityState(t *Template) bool {
	if t == nil || t.Setting == nil {
		return true
	}

	m.mu.Lock()
	if m.p == nil || m.testing {
		m.mu.Unlock()
		return true
	}
	paused := m.networkPaused
	setting := t.Setting
	m.mu.Unlock()

	if setting.TransparentType == configure.TransparentSystemProxy || !IsTransparentOn(setting) {
		return false
	}

	if !probePhysicalConnectivity() {
		if paused {
			return false
		}
		deleteTransparentProxyRulesKeepSystemProxy()
		m.mu.Lock()
		if m.p != nil && !m.testing {
			m.networkPaused = true
			m.mu.Unlock()
			ApiFeed.ProductMessage("running_state", map[string]interface{}{"running": false, "networkPaused": true})
			log.Info("connectivity lost, transparent proxy paused")
			return false
		}
		m.mu.Unlock()
		return true
	}

	if !paused {
		return false
	}
	if err := m.CheckAndSetupTransparentProxy(false, setting, t); err != nil {
		log.Warn("failed to resume transparent proxy after network recovery: %v", err)
		return false
	}
	m.mu.Lock()
	if m.p != nil && !m.testing {
		m.networkPaused = false
		m.mu.Unlock()
		ApiFeed.ProductMessage("running_state", map[string]interface{}{"running": true, "networkPaused": false})
		log.Info("connectivity restored, transparent proxy resumed")
		return false
	}
	m.mu.Unlock()
	return true
}
