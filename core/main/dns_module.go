// Package main provides DNS module integration for v2raya-core.
// The DNS module starts before xray-core and listens on the configured port (default 52353).
// v2rayA (the service layer) writes DNS module configuration into the xray JSON config file
// under the "dns_module" key; v2raya-core reads it here at startup.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/v2rayA/v2raya-core/dns"
	"github.com/xtls/xray-core/common/cmdarg"
	xcore "github.com/xtls/xray-core/core"
)

// dnsModuleManager manages the lifecycle of the DNS module within v2raya-core.
// It reads the "dns_module" configuration from xray's JSON config files and
// starts/stops the DNS listener independently of xray-core.
type dnsModuleManager struct {
	module *dns.DnsModule
	mu     sync.Mutex
}

var dnsMgr dnsModuleManager

// startDnsModuleWithXray reads the dns_module config, creates the DNS module,
// connects it to xray-core's internal routing dispatcher, and starts it.
// DNS starts AFTER xray so it can use xray's routing for internal dispatching
// (like xray-core's traditional DNS module).
func startDnsModuleWithXray(configFiles cmdarg.Arg, srv *xcore.Instance) error {
	cfg, err := extractDnsModuleConfig(configFiles)
	if err != nil {
		return fmt.Errorf("dns module: extract config: %w", err)
	}
	if cfg == nil {
		log.Println("[dns module] no dns_module config found, skipping")
		return nil
	}

	// Create the xray dispatcher adapter for internal routing.
	adapter, err := newXrayDispatcherAdapter(srv)
	if err != nil {
		return fmt.Errorf("dns module: create dispatcher: %w", err)
	}

	module := dns.NewDnsModule(cfg)

	// 在 Start 之前设置 xray-core 路由调度器。
	module.SetDispatcher(adapter)

	if err := module.Start(); err != nil {
		return fmt.Errorf("dns module: start failed: %w", err)
	}

	// Start 后设置 geosite 解析器（Router 在 Start 中创建）。
	// geositeResolver 使用 geosite.dat 展开分流规则中的 geosite: 标签，
	// 使 geosite:cn → 具体域名列表，实现真正的分流。
	if router := module.GetRouter(); router != nil {
		router.SetGeositeResolver(geositeResolver)
		// 重新加载规则以应用 geosite 解析器
		router.Reload(module.Config())
		log.Println("[dns module] geosite resolver set, rules reloaded with geosite expansion")
	}

	dnsMgr.mu.Lock()
	dnsMgr.module = module
	dnsMgr.mu.Unlock()

	log.Printf("[dns module] started successfully, listening on %s", cfg.Listener.ListenAddr)
	return nil
}

// stopDnsModule gracefully stops the DNS module.
func stopDnsModule() {
	dnsMgr.mu.Lock()
	module := dnsMgr.module
	dnsMgr.module = nil
	dnsMgr.mu.Unlock()

	if module != nil {
		log.Println("[dns module] stopping...")
		if err := module.Stop(); err != nil {
			log.Printf("[dns module] stop error: %v", err)
		}
		log.Println("[dns module] stopped")
	}
}

// extractDnsModuleConfig reads all config files, merges them, and extracts the
// "dns_module" configuration. Returns nil if the key is not present.
func extractDnsModuleConfig(configFiles cmdarg.Arg) (*dns.DnsModuleConfig, error) {
	if len(configFiles) == 0 {
		return nil, nil
	}

	// Merge all config files into a single JSON object.
	merged := make(map[string]interface{})
	for _, configFile := range configFiles {
		if configFile == "stdin:" {
			continue // skip stdin, unlikely to contain dns_module
		}
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("read config %q: %w", configFile, err)
		}

		var cfg map[string]interface{}
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Printf("[dns module] warning: cannot parse config %q as JSON: %v", configFile, err)
			continue
		}

		// Merge into the combined config.
		for k, v := range cfg {
			merged[k] = v
		}
	}

	// Check for dns_module key.
	raw, ok := merged["dns_module"]
	if !ok {
		return nil, nil
	}

	// Marshal the dns_module value back to JSON for unmarshaling into DnsModuleConfig.
	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshal dns_module config: %w", err)
	}

	var cfg dns.DnsModuleConfig
	if err := json.Unmarshal(jsonBytes, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal dns_module config: %w", err)
	}

	// Apply defaults where needed.
	if cfg.Listener.ListenAddr == "" {
		cfg.Listener.ListenAddr = "0.0.0.0:52353"
	}
	if cfg.Listener.Timeout <= 0 {
		cfg.Listener.Timeout = 5
	}
	if cfg.Cache.Size <= 0 {
		cfg.Cache.Size = 4096
	}
	if cfg.Cache.MinTTL <= 0 {
		cfg.Cache.MinTTL = 60
	}
	if cfg.Cache.MaxTTL <= 0 {
		cfg.Cache.MaxTTL = 86400
	}

	return &cfg, nil
}

// waitForDnsModuleReady blocks until the DNS module is healthy or a timeout expires.
// This allows the transparent proxy setup (iptables rules) to wait for DNS readiness.
func waitForDnsModuleReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		dnsMgr.mu.Lock()
		module := dnsMgr.module
		dnsMgr.mu.Unlock()

		if module != nil && module.Healthy() {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("dns module: not ready within %v", timeout)
}

// DnsModuleReady returns true if the DNS module is running and healthy.
// This is used by external health checks (e.g., from v2rayA service via API).
func DnsModuleReady() bool {
	dnsMgr.mu.Lock()
	module := dnsMgr.module
	dnsMgr.mu.Unlock()
	return module != nil && module.Healthy()
}
