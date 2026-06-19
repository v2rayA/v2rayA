package service

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	autoSwitchInterval        = 5 * time.Minute
	autoSwitchConfirmInterval = 5 * time.Second
	autoSwitchProbeTimeout    = 8 * time.Second
	autoSwitchTestURL         = "https://chatgpt.com/"
)

var autoSwitchMu sync.Mutex

func StartAutoSwitchTicker() {
	go func() {
		timer := time.NewTimer(autoSwitchInterval)
		defer timer.Stop()
		for range timer.C {
			AutoSwitchOnce()
			timer.Reset(autoSwitchInterval)
		}
	}()
}

func AutoSwitchOnce() {
	if !autoSwitchMu.TryLock() {
		return
	}
	defer autoSwitchMu.Unlock()

	if !v2ray.ProcessManager.Running() || configure.GetConnectedServers() == nil {
		return
	}
	if currentProxyReachable() {
		return
	}
	log.Warn("[AutoSwitch] current proxy cannot reach %s, confirming twice", autoSwitchTestURL)
	for i := 0; i < 2; i++ {
		time.Sleep(autoSwitchConfirmInterval)
		if currentProxyReachable() {
			log.Info("[AutoSwitch] current proxy recovered")
			return
		}
	}
	if err := switchToReachableServer(); err != nil {
		log.Warn("[AutoSwitch] %v", err)
	}
}

func currentProxyReachable() bool {
	c, err := httpClient.GetHttpClientWithv2rayAProxy()
	if err != nil {
		log.Debug("[AutoSwitch] GetHttpClientWithv2rayAProxy: %v", err)
		return false
	}
	defer c.CloseIdleConnections()
	c.Timeout = autoSwitchProbeTimeout
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return http.ErrUseLastResponse
		}
		return nil
	}
	req, err := http.NewRequest(http.MethodGet, autoSwitchTestURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", "curl/7.70.0")
	resp, err := c.Do(req)
	if err != nil {
		log.Debug("[AutoSwitch] probe failed: %v", err)
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 500
}

func switchToReachableServer() error {
	connected := configure.GetConnectedServers()
	if connected == nil || connected.Len() == 0 {
		return nil
	}
	outbound := connected.Get()[0].Outbound
	if outbound == "" {
		outbound = "proxy"
	}
	backup := configure.GetConnectedServersByOutbound(outbound)
	candidates := allServerWhiches(outbound)
	if len(candidates) == 0 {
		return fmt.Errorf("no candidate servers")
	}

	log.Info("[AutoSwitch] testing %d candidate servers", len(candidates))
	tested, err := TestHttpLatency(candidates, autoSwitchProbeTimeout, 16, false, "")
	if err != nil {
		return fmt.Errorf("test candidate servers: %w", err)
	}
	for _, candidate := range tested {
		if !latencyAvailable(candidate.Latency) {
			continue
		}
		isCurrent := false
		if backup != nil {
			for _, old := range backup.Get() {
				if old.EqualTo(*candidate) {
					isCurrent = true
					break
				}
			}
		}
		if isCurrent {
			continue
		}
		log.Info("[AutoSwitch] trying server: type=%s sub=%d id=%d outbound=%s latency=%s", candidate.TYPE, candidate.Sub, candidate.ID, outbound, candidate.Latency)
		if err := replaceOutbound(outbound, candidate, backup); err != nil {
			log.Warn("[AutoSwitch] switch failed: %v", err)
			continue
		}
		time.Sleep(autoSwitchConfirmInterval)
		if currentProxyReachable() {
			log.Info("[AutoSwitch] switched to server: type=%s sub=%d id=%d outbound=%s", candidate.TYPE, candidate.Sub, candidate.ID, outbound)
			return nil
		}
		log.Warn("[AutoSwitch] switched server cannot reach %s, trying next", autoSwitchTestURL)
	}
	if backup != nil {
		_ = configure.OverwriteConnects(backup)
		_ = v2ray.UpdateV2RayConfig()
	}
	return fmt.Errorf("no reachable candidate server found")
}

func allServerWhiches(outbound string) []*configure.Which {
	var whiches []*configure.Which
	servers := configure.GetServers()
	for i := range servers {
		whiches = append(whiches, &configure.Which{
			TYPE:     configure.ServerType,
			ID:       i + 1,
			Outbound: outbound,
		})
	}
	subscriptions := configure.GetSubscriptions()
	for sub := range subscriptions {
		for i := range subscriptions[sub].Servers {
			whiches = append(whiches, &configure.Which{
				TYPE:     configure.SubscriptionServerType,
				ID:       i + 1,
				Sub:      sub,
				Outbound: outbound,
			})
		}
	}
	return whiches
}

func latencyAvailable(latency string) bool {
	return strings.HasSuffix(latency, "ms")
}

func replaceOutbound(outbound string, which *configure.Which, backup *configure.Whiches) (err error) {
	if outbound == "" {
		outbound = "proxy"
	}
	which.Outbound = outbound
	defer func() {
		if err != nil && backup != nil && v2ray.ProcessManager.Running() {
			_ = configure.OverwriteConnects(backup)
			_ = v2ray.UpdateV2RayConfig()
		}
	}()
	if err = configure.ClearConnects(outbound); err != nil {
		return err
	}
	if err = checkSupport([]*configure.Which{which}); err != nil {
		return err
	}
	if err = configure.AddConnect(*which); err != nil {
		return err
	}
	if v2ray.ProcessManager.Running() {
		return v2ray.UpdateV2RayConfig()
	}
	return nil
}
