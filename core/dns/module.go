package dns

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/netip"
	"sync"
	"syscall"
	"time"

	"github.com/miekg/dns"
)

// DnsModule is the top-level lifecycle manager for the DNS module.
// It coordinates the listener, router, upstream manager, cache, and statistics.
type DnsModule struct {
	config      *DnsModuleConfig
	listener    *DnsListener
	handler     *DnsHandler
	router      *Router
	upstreamMgr *UpstreamManager
	cache       *DnsCache // DNS 缓存模块（LRU + TTL 双淘汰）
	stats       DnsStats
	metrics     *DnsMetrics     // 指标收集器（新增）
	diagnostics *Diagnostics    // 诊断模块（新增）
	dispatcher  RouteDispatcher // xray-core 内部路由调度器（Start 前设置）
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	healthy     bool
}

// NewDnsModule creates a new DnsModule with the given configuration.
func NewDnsModule(config *DnsModuleConfig) *DnsModule {
	if config == nil {
		config = DefaultDnsModuleConfig()
	}
	return &DnsModule{
		config: config,
		stats: DnsStats{
			RoutedQueries: make(map[string]int64),
			StartedAt:     time.Now(),
		},
		metrics:     NewDnsMetrics(),
		diagnostics: nil, // 在 Start 中创建，因为需要引用 module 自身
	}
}

// Start initializes and starts all DNS module components.
// It validates the configuration, creates the handler, router, upstream manager,
// and listener, and begins serving DNS queries on the configured address.
func (m *DnsModule) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.healthy {
		return fmt.Errorf("dns module: already started")
	}

	if err := m.config.Validate(); err != nil {
		return fmt.Errorf("dns module: invalid config: %w", err)
	}

	m.ctx, m.cancel = context.WithCancel(context.Background())

	// Initialize diagnostics (requires module reference).
	m.diagnostics = NewDiagnostics(m)

	// Initialize statistics.
	m.stats = DnsStats{
		RoutedQueries: make(map[string]int64),
		StartedAt:     time.Now(),
	}

	// Reset metrics (fresh start).
	m.metrics.Reset()

	// 解析 bootstrap domain — 当 DNS 上游地址是域名时，用系统 DNS 解析出 IP。
	// 避免 DNS 模块自身依赖 DNS 的循环依赖。
	if len(m.config.Bootstrap) > 0 {
		log.Printf("[dns module] bootstrap: %d domains to resolve: %v", len(m.config.Bootstrap), m.config.Bootstrap)
		resolved := m.resolveBootstrap(m.config.Bootstrap)
		log.Printf("[dns module] bootstrap: resolved %d/%d domains", len(resolved), len(m.config.Bootstrap))
		for domain, ip := range resolved {
			log.Printf("[dns module] bootstrap: %s → %s", domain, ip)
		}
		for i, upstream := range m.config.Upstreams {
			if upstream.Bootstrap {
				continue
			}
			before := upstream.Addr
			// 从上游地址中提取主机名
			host, port, err := net.SplitHostPort(upstream.Addr)
			if err != nil {
				// 格式不符合 host:port，尝试直接匹配
				if ip, ok := resolved[upstream.Addr]; ok {
					m.config.Upstreams[i].Addr = ip + ":53"
					log.Printf("[dns module] bootstrap: upstream %s → %s (was %s)", upstream.ID, m.config.Upstreams[i].Addr, before)
				} else {
					log.Printf("[dns module] bootstrap: upstream %s addr=%s not resolved, keeping as-is", upstream.ID, upstream.Addr)
				}
				continue
			}
			if net.ParseIP(host) == nil {
				if ip, ok := resolved[host]; ok {
					m.config.Upstreams[i].Addr = net.JoinHostPort(ip, port)
					log.Printf("[dns module] bootstrap: upstream %s → %s (was %s)", upstream.ID, m.config.Upstreams[i].Addr, before)
				} else {
					log.Printf("[dns module] bootstrap: upstream %s host=%s not resolved, keeping as-is", upstream.ID, host)
				}
			}
		}
	} else {
		log.Printf("[dns module] bootstrap: no domains to resolve (list is empty)")
	}

	// 为所有缺失端口的上游地址补充默认端口
	for i, upstream := range m.config.Upstreams {
		if _, _, err := net.SplitHostPort(upstream.Addr); err != nil {
			// 没有端口号，根据协议补充
			port := "53"
			switch upstream.Protocol {
			case "tcp-tls", "tls":
				port = "853"
			}
			m.config.Upstreams[i].Addr = net.JoinHostPort(upstream.Addr, port)
			log.Printf("[dns module] added default port %s to upstream %s (protocol=%s)", port, upstream.Addr, upstream.Protocol)
		}
	}

	// 打印最终上游列表
	log.Printf("[dns module] upstream list after bootstrap:")
	for _, u := range m.config.Upstreams {
		log.Printf("[dns module]   id=%s addr=%s protocol=%s proxy_tag=%s", u.ID, u.Addr, u.Protocol, u.ProxyTag)
	}

	// Create the UpstreamManager.
	m.upstreamMgr = NewUpstreamManager(m.config.Upstreams)

	// Set xray-core internal dispatcher (if configured before Start).
	if m.dispatcher != nil {
		m.upstreamMgr.SetDispatcher(m.dispatcher)
	}

	// Set proxy address resolver from ProxyMap (if configured).
	if len(m.config.ProxyMap) > 0 {
		proxyMap := m.config.ProxyMap
		m.upstreamMgr.SetProxyAddrResolver(func(proxyTag string) string {
			if addr, ok := proxyMap[proxyTag]; ok {
				return addr
			}
			return ""
		})
	}

	// Initialize cache if enabled.
	if m.config.Cache.Enabled {
		m.cache = NewDnsCache(&m.config.Cache)
		log.Printf("[dns module] cache initialized: size=%d, minTTL=%d, maxTTL=%d, prefetch=%v",
			m.config.Cache.Size, m.config.Cache.MinTTL, m.config.Cache.MaxTTL, m.config.Cache.Prefetch)
	} else {
		m.cache = nil
		log.Printf("[dns module] cache disabled")
	}

	// Build rules from config.
	rules := make([]*DnsRule, len(m.config.Rules))
	for i, rc := range m.config.Rules {
		rules[i] = &DnsRule{
			ID:           rc.ID,
			Domain:       rc.Domain,
			DomainSuffix: rc.DomainSuffix,
			DomainRegex:  rc.DomainRegex,
			IP:           rc.IP,
			QueryType:    rc.QueryType,
			ClientIP:     rc.ClientIP,
			Upstream:     rc.Upstream,
			Action:       rc.Action,
			Policy:       rc.Policy,
		}
	}

	// Determine default upstream: the LAST non-bootstrap upstream.
	// Convention: the first upstreams have domain-specific rules,
	// and the LAST upstream (with empty domain) is the fallback.
	defaultUpstream := ""
	for _, u := range m.config.Upstreams {
		if !u.Bootstrap {
			defaultUpstream = u.ID
			if defaultUpstream == "" {
				defaultUpstream = u.Addr
			}
			// Continue iterating — the LAST non-bootstrap upstream wins.
		}
	}
	if defaultUpstream == "" && len(m.config.Upstreams) > 0 {
		u := m.config.Upstreams[len(m.config.Upstreams)-1]
		defaultUpstream = u.ID
		if defaultUpstream == "" {
			defaultUpstream = u.Addr
		}
	}

	// Create the Router.
	router, err := NewRouter(rules, m.config.Upstreams, defaultUpstream)
	if err != nil {
		m.cancel()
		return fmt.Errorf("dns module: failed to create router: %w", err)
	}
	m.router = router

	// Create the DNS handler with the configured timeout.
	timeout := time.Duration(m.config.Listener.Timeout) * time.Second
	m.handler = NewDnsHandler(timeout, m)

	// Create the DNS listener.
	m.listener = NewDnsListener(&m.config.Listener, m.handler)

	// Start the listener.
	log.Printf("[dns module] starting DNS listener on %s", m.config.Listener.ListenAddr)
	if err := m.listener.Start(); err != nil {
		m.cancel()
		return fmt.Errorf("dns module: failed to start listener: %w", err)
	}

	m.healthy = true
	log.Printf("[dns module] started successfully: %s", m.config.String())
	return nil
}

// resolveBootstrap resolves a list of domain names using the system DNS.
// Returns a map of domain → first resolved IP address.
// This breaks the circular dependency: DNS upstream addresses that are domain names
// need DNS resolution themselves, so we use the system resolver (/etc/resolv.conf).
func (m *DnsModule) resolveBootstrap(domains []string) map[string]string {
	resolved := make(map[string]string, len(domains))
	if len(domains) == 0 {
		return resolved
	}

	// 收集可用的 bootstrap DNS 服务器（优先级从高到低）：
	// 1. 配置中保存的系统 DNS（劫持发生前由 v2rayA 读取）
	// 2. /etc/resolv.conf 实时读取
	// 3. 硬编码的著名公共 DNS（兜底）
	sysServers := m.collectBootstrapDns()
	if len(sysServers) == 0 {
		log.Printf("[dns module] warning: no DNS servers available for bootstrap resolution")
		return resolved
	}

	log.Printf("[dns module] resolving %d bootstrap domains using %d DNS servers", len(domains), len(sysServers))
	for _, s := range sysServers {
		log.Printf("[dns module]   bootstrap DNS: %s", s.String())
	}

	for _, domain := range domains {
		if domain == "" {
			continue
		}
		if net.ParseIP(domain) != nil {
			resolved[domain] = domain
			continue
		}
		// 对每个 DNS 服务器尝试解析
		for _, srv := range sysServers {
			addr := srv.String()
			m := new(dns.Msg)
			m.SetQuestion(dns.Fqdn(domain), dns.TypeA)
			m.RecursionDesired = true

			// Use SO_MARK=0x80 to bypass iptables DNS redirect loop.
			// Without this mark, the bootstrap query to the system DNS's port 53
			// would be intercepted by iptables and redirected back to our own
			// DNS module (which hasn't finished starting yet), causing a timeout.
			bootstrapMarkFd := func(network, address string, c syscall.RawConn) error {
				return c.Control(func(fd uintptr) {
					_ = setSocketMark(fd)
				})
			}
			client := &dns.Client{
				Net:     "udp",
				Timeout: 3 * time.Second,
				Dialer: &net.Dialer{
					Timeout: 3 * time.Second,
					Control: bootstrapMarkFd,
				},
			}
			resp, _, err := client.Exchange(m, addr)
			if err != nil {
				log.Printf("[dns module] bootstrap: %s via %s failed: %v", domain, addr, err)
				continue
			}
			if resp == nil || len(resp.Answer) == 0 {
				continue
			}
			for _, ans := range resp.Answer {
				if a, ok := ans.(*dns.A); ok {
					resolved[domain] = a.A.String()
					log.Printf("[dns module] bootstrap resolved %s → %s via %s", domain, a.A.String(), addr)
					break
				}
			}
			if _, ok := resolved[domain]; ok {
				break // 已解析成功
			}
		}
		if _, ok := resolved[domain]; !ok {
			log.Printf("[dns module] warning: bootstrap resolution failed for %s (all system DNS exhausted)", domain)
		}
	}

	return resolved
}

// wellKnownPublicDns 是硬编码的著名公共 DNS 服务器列表，作为 bootstrap 解析的兜底。
// 当系统 DNS 不可用（被劫持/不可达）时，使用这些知名服务器来解析上游域名。
// 同时包含 IPv4 和 IPv6 地址，确保在各种网络环境下都能工作。
var wellKnownPublicDns = []string{
	// 阿里 DNS
	"223.5.5.5:53",
	"223.6.6.6:53",
	"[2400:3200::1]:53",
	"[2400:3200:baba::1]:53",
	// 腾讯 DNS
	"119.29.29.29:53",
	// Google DNS
	"8.8.8.8:53",
	"8.8.4.4:53",
	"[2001:4860:4860::8888]:53",
	"[2001:4860:4860::8844]:53",
	// Cloudflare DNS
	"1.1.1.1:53",
	"1.0.0.1:53",
	"[2606:4700:4700::1111]:53",
	"[2606:4700:4700::1001]:53",
}

// collectBootstrapDns 收集可用的 bootstrap DNS 服务器地址。
// 优先级：
//  1. 配置中保存的系统 DNS（劫持发生前由 v2rayA 读取并写入配置）
//  2. /etc/resolv.conf 实时读取
//  3. 硬编码的著名公共 DNS（兜底，确保即使系统 DNS 不可用也能解析）
func (m *DnsModule) collectBootstrapDns() []netip.AddrPort {
	var servers []netip.AddrPort
	seen := make(map[string]bool) // 去重

	addServers := func(list []string) {
		for _, s := range list {
			if !seen[s] {
				if ap, err := netip.ParseAddrPort(s); err == nil {
					servers = append(servers, ap)
					seen[s] = true
				}
			}
		}
	}

	// 1. 配置中保存的系统 DNS（劫持发生前由 v2rayA 读取并写入配置）
	if len(m.config.BootstrapDns) > 0 {
		addServers(m.config.BootstrapDns)
	}

	// 2. /etc/resolv.conf 实时读取（劫持可能已生效，但过滤掉自身的地址避免自引用）
	sysDns := GetSystemDNS()
	for _, ap := range sysDns {
		s := ap.String()
		// 跳过自身地址（DNS 模块的监听地址），避免自引用循环
		if isOwnAddress(ap) {
			log.Printf("[dns module] bootstrap: skipping own address %s from system DNS", s)
			continue
		}
		if !seen[s] {
			servers = append(servers, ap)
			seen[s] = true
		}
	}

	// 3. 硬编码的著名公共 DNS（兜底）
	// 即使系统 DNS 全部被劫持，也能通过公共 DNS 解析上游域名
	addServers(wellKnownPublicDns)

	return servers
}

// isOwnAddress 检查地址是否是 DNS 模块自身监听的地址。
// 用于过滤 bootstrap 中的自引用地址，避免循环。
func isOwnAddress(ap netip.AddrPort) bool {
	if !ap.Addr().IsLoopback() && !ap.Addr().IsPrivate() {
		return false
	}
	// 检查是否是常见的 DNS 模块自身地址
	selfAddrs := []string{
		"127.0.0.1:53",
		"127.2.0.17:53",
		"127.0.0.1:5353",
	}
	for _, addr := range selfAddrs {
		if parsed, err := netip.ParseAddrPort(addr); err == nil {
			if parsed == ap {
				return true
			}
		}
	}
	return false
}

// Stop gracefully shuts down the DNS module and all its components.
// It stops the listener and waits for all goroutines to complete.
func (m *DnsModule) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.healthy {
		return nil // Already stopped.
	}

	log.Printf("[dns module] stopping...")

	// Stop the listener.
	if m.listener != nil {
		if err := m.listener.Stop(); err != nil {
			log.Printf("[dns module] listener stop error: %v", err)
		}
	}

	if m.cancel != nil {
		m.cancel()
	}

	m.wg.Wait()
	m.healthy = false
	log.Printf("[dns module] stopped")
	return nil
}

// Healthy returns true if the DNS module is running and healthy.
func (m *DnsModule) Healthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.healthy {
		return false
	}
	if m.listener == nil {
		return false
	}
	return m.listener.Healthy()
}

// Stats returns a map of current module statistics.
// This now includes rich metrics from the DnsMetrics collector.
func (m *DnsModule) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect DnsStats (backward compatibility).
	m.stats.mu.Lock()
	totalQueries := m.stats.TotalQueries
	rejected := m.stats.RejectedQueries
	bypassed := m.stats.BypassedQueries
	cacheHits := m.stats.CacheHits
	cacheMisses := m.stats.CacheMisses
	avgRTT := m.stats.AvgRTT

	// Copy routed queries map.
	routedQueries := make(map[string]int64, len(m.stats.RoutedQueries))
	for k, v := range m.stats.RoutedQueries {
		routedQueries[k] = v
	}
	m.stats.mu.Unlock()

	// Collect cache stats if available.
	var cacheSize int
	var cacheHitRate float64
	var cacheEvictions int64
	var cachePrefetchCnt int64
	if m.cache != nil {
		cs := m.cache.Stats()
		cacheSize = cs.Size
		cacheHitRate = cs.HitRate
		cacheEvictions = cs.Evictions
		cachePrefetchCnt = cs.PrefetchCnt
	}

	// Build stats map (basic stats for backward compatibility).
	stats := map[string]interface{}{
		"healthy":          m.healthy,
		"listener":         m.config.Listener.String(),
		"total_queries":    totalQueries,
		"rejected_queries": rejected,
		"bypassed_queries": bypassed,
		"cache_hits":       cacheHits,
		"cache_misses":     cacheMisses,
		"cache_size":       cacheSize,
		"cache_hit_rate":   cacheHitRate,
		"cache_evictions":  cacheEvictions,
		"cache_prefetchs":  cachePrefetchCnt,
		"avg_rtt":          avgRTT.String(),
		"routed_queries":   routedQueries,
		"num_rules":        len(m.config.Rules),
		"num_upstreams":    len(m.config.Upstreams),
	}

	// Add rich metrics from DnsMetrics if available.
	if m.metrics != nil {
		snap := m.metrics.Snapshot()
		stats["metrics_total_queries"] = snap.TotalQueries
		stats["metrics_total_responses"] = snap.TotalResponses
		stats["metrics_total_errors"] = snap.TotalErrors
		stats["metrics_total_timeouts"] = snap.TotalTimeouts
		stats["metrics_cache_hits"] = snap.CacheHits
		stats["metrics_cache_misses"] = snap.CacheMisses
		stats["metrics_cache_evictions"] = snap.CacheEvictions
		stats["metrics_prefetch_count"] = snap.PrefetchCount
		stats["metrics_routed"] = snap.RoutedQueries
		stats["metrics_rejected"] = snap.RejectedQueries
		stats["metrics_bypassed"] = snap.BypassedQueries
		stats["metrics_uptime"] = snap.Uptime

		// Query type breakdown.
		if len(snap.QueriesByType) > 0 {
			stats["queries_by_type"] = snap.QueriesByType
		}

		// Rule hits.
		if len(snap.HitsByRule) > 0 {
			stats["rule_hits"] = snap.HitsByRule
		}

		// Latency buckets.
		if len(snap.LatencyBuckets) > 0 {
			latencyMap := make(map[string]int64)
			for _, lb := range snap.LatencyBuckets {
				latencyMap[lb.Label] = lb.Count
			}
			stats["latency_distribution"] = latencyMap
		}

		// Upstream stats.
		if len(snap.QueriesByUpstream) > 0 {
			upstreamStats := make(map[string]map[string]interface{})
			for upstream, um := range snap.QueriesByUpstream {
				upstreamStats[upstream] = map[string]interface{}{
					"queries":   um.Queries,
					"successes": um.Successes,
					"errors":    um.Errors,
					"avg_rtt":   um.AvgRTT.String(),
					"min_rtt":   um.MinRTT.String(),
					"max_rtt":   um.MaxRTT.String(),
				}
			}
			stats["upstream_stats"] = upstreamStats
		}
	}

	return stats
}

// SetDispatcher sets the xray-core routing dispatcher for internal routing.
// Must be called before Start().
func (m *DnsModule) SetDispatcher(d RouteDispatcher) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dispatcher = d
}

// Config returns the module configuration.
func (m *DnsModule) Config() *DnsModuleConfig {
	return m.config
}

// GetRouter returns the router instance.
func (m *DnsModule) GetRouter() *Router {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.router
}

// GetUpstreamManager returns the upstream manager instance.
func (m *DnsModule) GetUpstreamManager() *UpstreamManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.upstreamMgr
}

// GetCache returns the DNS cache instance.
func (m *DnsModule) GetCache() *DnsCache {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cache
}

// GetMetrics returns the metrics collector.
func (m *DnsModule) GetMetrics() *DnsMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics
}

// GetDiagnostics returns the diagnostics collector.
func (m *DnsModule) GetDiagnostics() *Diagnostics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.diagnostics
}
