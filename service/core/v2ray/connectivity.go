package v2ray

import (
	"net/http"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	connectivityProbeURL      = "https://connectivitycheck.gstatic.com/generate_204"
	connectivityCheckInterval  = 15 * time.Second
	connectivityCheckTimeout   = 5 * time.Second
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