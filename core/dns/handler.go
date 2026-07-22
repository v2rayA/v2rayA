package dns

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DnsHandler processes incoming DNS queries and returns responses.
// It uses the Router to determine which upstream to query, and the
// UpstreamManager to send queries to upstream DNS servers.
type DnsHandler struct {
	timeout time.Duration
	module  *DnsModule
}

// prefetchAsync 异步执行预取，后台刷新缓存条目
func (h *DnsHandler) prefetchAsync(ctx context.Context, upstream *UpstreamInstance,
	upstreamMgr *UpstreamManager, query *DnsQuery, key CacheKey) {

	if upstream == nil || upstreamMgr == nil {
		h.module.cache.SetPrefetching(key, false)
		return
	}

	// 记录预取指标
	if h.module.metrics != nil {
		h.module.metrics.RecordPrefetch()
	}

	log.Printf("[dns] prefetch: %s %s via %s (tag=%s)",
		dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr, upstream.ProxyTag)

	resp, err := upstreamMgr.Exchange(upstream, query)
	if err != nil {
		log.Printf("[dns] prefetch error for %s %s: %v",
			dns.Type(uint16(query.QType)).String(), query.Name, err)
		h.module.cache.RecordPrefetchFailure(key)
		return
	}

	if resp == nil {
		h.module.cache.RecordPrefetchFailure(key)
		return
	}

	// 预取成功，更新缓存
	h.module.cache.Set(key, resp, resp.TTL)
	h.module.cache.RecordPrefetchSuccess(key)

	log.Printf("[dns] prefetch completed: %s %s via %s",
		dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr)
}

// buildCacheKey 构建完整的缓存键（路由后已知上游和代理渠道）
func buildCacheKey(query *DnsQuery, upstream *UpstreamInstance) CacheKey {
	proxyTag := "direct"
	if upstream != nil && upstream.ProxyTag != "" {
		proxyTag = upstream.ProxyTag
	}
	upstreamAddr := ""
	if upstream != nil {
		upstreamAddr = upstream.Addr
	}
	return CacheKey{
		UpstreamAddr: upstreamAddr,
		ProxyTag:     proxyTag,
		Name:         query.Name,
		QType:        query.QType,
	}
}

// NewDnsHandler creates a new DnsHandler with the given timeout and module reference.
func NewDnsHandler(timeout time.Duration, module *DnsModule) *DnsHandler {
	return &DnsHandler{
		timeout: timeout,
		module:  module,
	}
}

// HandleQuery processes a DNS query by routing it through the router
// and sending it to the appropriate upstream server.
//
// Processing flow:
//  1. Route the query using the Router.
//  2. Based on the action (route/reject/bypass), take appropriate action.
//  3. For "route": send the query to the matched upstream via UpstreamManager.
//  4. For "reject": return a REFUSED response.
//  5. For "bypass": send the query directly (bypass routing).
//  6. Track statistics.
func (h *DnsHandler) HandleQuery(ctx context.Context, query *DnsQuery) (*DnsResponse, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("dns handler: context cancelled: %w", ctx.Err())
	default:
	}

	// Log the incoming query (structured, concise).
	clientIP := "unknown"
	if query.ClientIP != nil {
		clientIP = query.ClientIP.String()
	}
	log.Printf("[dns] query: %s %s from %s", QTypeToString(query.QType), query.Name, clientIP)

	// Get module references.
	module := h.module
	if module == nil {
		// Fallback: return stub response if no module configured.
		log.Printf("[dns] no module configured, returning stub response for %s", query.Name)
		return h.buildStubResponse(query), nil
	}

	router := module.GetRouter()
	upstreamMgr := module.GetUpstreamManager()

	// Track total queries (backward compatible stats).
	module.stats.mu.Lock()
	module.stats.TotalQueries++
	module.stats.mu.Unlock()

	// Step 1: Route the query.
	var result *RouteResult
	if router != nil {
		result = router.Route(query)
	} else {
		// No router — use default behavior.
		result = &RouteResult{
			Action: "route",
			Policy: "single",
		}
	}

	// Step 2: Handle based on action.
	switch result.Action {
	case "reject":
		return h.handleReject(query, module, result)

	case "bypass":
		return h.handleBypass(query, module, result, upstreamMgr)

	default: // "route"
		return h.handleRoute(ctx, query, module, result, upstreamMgr)
	}
}

// handleRoute sends the query to the matched upstream server.
// 查询流程（方案2：先路由再查缓存）：
//
//  1. Route the query to determine upstream + proxy channel.
//  2. Build full CacheKey (with upstream + proxy tag).
//  3. Check cache. If hit: return cached response, optionally trigger prefetch.
//  4. If miss: send query to upstream, store result in cache, return response.
func (h *DnsHandler) handleRoute(ctx context.Context, query *DnsQuery, module *DnsModule,
	result *RouteResult, upstreamMgr *UpstreamManager) (*DnsResponse, error) {

	upstreamID := result.UpstreamID
	var upstream *UpstreamInstance

	if upstreamMgr != nil && upstreamID != "" {
		var ok bool
		upstream, ok = upstreamMgr.GetUpstream(upstreamID)
		if !ok {
			log.Printf("[dns] upstream %q not found, using default", upstreamID)
		}
	}

	// If no upstream found, or no router matched, use first available upstream.
	if upstream == nil && upstreamMgr != nil {
		// Try to find a non-bootstrap upstream.
		for _, u := range module.config.Upstreams {
			if !u.Bootstrap {
				upstream, _ = upstreamMgr.GetUpstream(u.ID)
				if upstream == nil {
					upstream, _ = upstreamMgr.GetUpstream(u.Addr)
				}
				break
			}
		}
		// Fallback: use any upstream.
		if upstream == nil && len(module.config.Upstreams) > 0 {
			u := module.config.Upstreams[0]
			upstream, _ = upstreamMgr.GetUpstream(u.ID)
			if upstream == nil {
				upstream, _ = upstreamMgr.GetUpstream(u.Addr)
			}
		}
	}

	if upstream == nil || upstreamMgr == nil {
		log.Printf("[dns] no upstream available for %s %s, returning stub", QTypeToString(query.QType), query.Name)
		return h.buildStubResponse(query), nil
	}

	// Build full CacheKey after routing (upstream + proxy tag known).
	cacheKey := buildCacheKey(query, upstream)

	// Record route metrics.
	if module.metrics != nil {
		module.metrics.RecordQuery(query, upstreamID)
		module.metrics.RecordAction("route")
		if result.RuleID != "" {
			module.metrics.RecordRuleHit(result.RuleID)
		}
	}

	// Log route decision (structured, concise).
	proxyInfo := upstream.ProxyTag
	if proxyInfo == "" {
		proxyInfo = "direct"
	}
	ruleInfo := ""
	if result.RuleID != "" {
		ruleInfo = fmt.Sprintf(" rule(%s)", result.RuleID)
	}
	log.Printf("[dns] route: %s %s → %s(%s)%s action(route)",
		QTypeToString(query.QType), query.Name, upstreamID, proxyInfo, ruleInfo)

	// Check cache (if enabled).
	cache := module.cache
	if cache != nil {
		if cached, ok := cache.Get(cacheKey); ok {
			// Log cache hit.
			_ = cached // DnsResponse, used below
			log.Printf("[dns] cache-hit: %s %s",
				QTypeToString(query.QType), query.Name)

			// Track cache hit (backward compatible stats).
			module.stats.mu.Lock()
			module.stats.CacheHits++
			module.stats.mu.Unlock()

			// Track cache hit (metrics).
			if module.metrics != nil {
				module.metrics.RecordCacheHit()
			}

			// Check if prefetch is needed (async refresh).
			if cache.ShouldPrefetch(cacheKey) {
				cache.SetPrefetching(cacheKey, true)
				go h.prefetchAsync(ctx, upstream, upstreamMgr, query, cacheKey)
			}

			return cached, nil
		}

		// Log cache miss.
		log.Printf("[dns] cache-miss: %s %s", QTypeToString(query.QType), query.Name)

		// Track cache miss (backward compatible stats).
		module.stats.mu.Lock()
		module.stats.CacheMisses++
		module.stats.mu.Unlock()

		// Track cache miss (metrics).
		if module.metrics != nil {
			module.metrics.RecordCacheMiss()
		}
	}

	// Send query to upstream (cache miss or cache disabled).
	start := time.Now()
	resp, err := upstreamMgr.Exchange(upstream, query)
	rtt := time.Since(start)

	if err != nil {
		log.Printf("[dns] error: %s %s: %v", upstream.Addr, query.Name, err)

		// Track error in stats.
		module.stats.mu.Lock()
		module.stats.RoutedQueries[upstreamID]++
		module.stats.mu.Unlock()

		// Track error in metrics.
		if module.metrics != nil {
			module.metrics.RecordError(upstreamID, err)
		}

		return nil, fmt.Errorf("dns handler: upstream error: %w", err)
	}

	if resp == nil {
		log.Printf("[dns] empty response for %s %s", QTypeToString(query.QType), query.Name)
		return h.buildStubResponse(query), nil
	}

	// Log successful upstream response.
	log.Printf("[dns] upstream: %s via %s (rtt=%dms)",
		upstream.Addr, proxyInfo, rtt.Milliseconds())

	// Store response in cache (if enabled and not a bootstrap query).
	if cache != nil && !query.IsBootstrap {
		cache.Set(cacheKey, resp, resp.TTL)
	}

	// Track statistics.
	module.stats.mu.Lock()
	module.stats.RoutedQueries[upstreamID]++
	module.stats.AvgRTT = updateAvgRTT(module.stats.AvgRTT, rtt, module.stats.TotalQueries)
	module.stats.mu.Unlock()

	// Record response metrics.
	if module.metrics != nil {
		module.metrics.RecordResponse(resp, rtt)
	}

	return resp, nil
}

// handleReject returns a REFUSED response for rejected queries.
func (h *DnsHandler) handleReject(query *DnsQuery, module *DnsModule, result *RouteResult) (*DnsResponse, error) {
	log.Printf("[dns] reject: %s %s (rule=%s)", QTypeToString(query.QType), query.Name, result.RuleID)

	// Track statistics.
	module.stats.mu.Lock()
	module.stats.RejectedQueries++
	module.stats.mu.Unlock()

	// Record metrics.
	if module.metrics != nil {
		module.metrics.RecordQuery(query, "")
		module.metrics.RecordAction("reject")
		if result.RuleID != "" {
			module.metrics.RecordRuleHit(result.RuleID)
		}
	}

	return h.buildRefused(query), nil
}

// handleBypass handles bypass queries — sends directly without proxy routing.
func (h *DnsHandler) handleBypass(query *DnsQuery, module *DnsModule, result *RouteResult,
	upstreamMgr *UpstreamManager) (*DnsResponse, error) {

	log.Printf("[dns] bypass: %s %s (rule=%s)", QTypeToString(query.QType), query.Name, result.RuleID)

	// Track statistics.
	module.stats.mu.Lock()
	module.stats.BypassedQueries++
	module.stats.mu.Unlock()

	// For bypass, we still need an upstream. Find a direct upstream.
	var upstream *UpstreamInstance
	if upstreamMgr != nil {
		upstreamID := result.UpstreamID
		if upstreamID != "" {
			upstream, _ = upstreamMgr.GetUpstream(upstreamID)
		}
		// If no specific upstream for bypass, use first non-bootstrap upstream.
		if upstream == nil {
			for _, u := range module.config.Upstreams {
				if !u.Bootstrap {
					upstream, _ = upstreamMgr.GetUpstream(u.ID)
					if upstream == nil {
						upstream, _ = upstreamMgr.GetUpstream(u.Addr)
					}
					break
				}
			}
		}
	}

	if upstream == nil || upstreamMgr == nil {
		return h.buildStubResponse(query), nil
	}

	// Record bypass metrics.
	if module.metrics != nil {
		module.metrics.RecordQuery(query, result.UpstreamID)
		module.metrics.RecordAction("bypass")
		if result.RuleID != "" {
			module.metrics.RecordRuleHit(result.RuleID)
		}
	}

	start := time.Now()
	resp, err := upstreamMgr.Exchange(upstream, query)
	rtt := time.Since(start)

	if err != nil {
		log.Printf("[dns] error: %s %s: %v", upstream.Addr, query.Name, err)
		if module.metrics != nil {
			module.metrics.RecordError(result.UpstreamID, err)
		}
		return nil, fmt.Errorf("dns handler: bypass upstream error: %w", err)
	}

	if resp == nil {
		return h.buildStubResponse(query), nil
	}

	// Log successful bypass response.
	log.Printf("[dns] upstream: %s:53 direct (rtt=%dms)",
		upstream.Addr, rtt.Milliseconds())

	module.stats.mu.Lock()
	module.stats.RoutedQueries[upstream.ID]++
	module.stats.AvgRTT = updateAvgRTT(module.stats.AvgRTT, rtt, module.stats.TotalQueries)
	module.stats.mu.Unlock()

	// Record response metrics.
	if module.metrics != nil {
		module.metrics.RecordResponse(resp, rtt)
	}

	return resp, nil
}

// HandleQueryWithTimeout wraps HandleQuery with a timeout derived from the config.
func (h *DnsHandler) HandleQueryWithTimeout(query *DnsQuery, timeout time.Duration) (*DnsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return h.HandleQuery(ctx, query)
}

// buildStubResponse creates a minimal stub response echoing the query.
func (h *DnsHandler) buildStubResponse(query *DnsQuery) *DnsResponse {
	resp := &DnsResponse{
		Query:    *query,
		Rcode:    dns.RcodeSuccess,
		Upstream: "stub",
		ProxyTag: "direct",
		RTT:      0,
		Cached:   false,
	}

	m := new(dns.Msg)
	m.SetQuestion(query.Name, uint16(query.QType))
	m.Response = true
	m.RecursionAvailable = true
	m.Rcode = dns.RcodeSuccess

	resp.RawMsg = m
	return resp
}

// buildRefused creates a REFUSED DNS response.
func (h *DnsHandler) buildRefused(query *DnsQuery) *DnsResponse {
	m := new(dns.Msg)
	m.SetQuestion(query.Name, uint16(query.QType))
	m.Response = true
	m.Rcode = dns.RcodeRefused

	return &DnsResponse{
		Query:    *query,
		RawMsg:   m,
		Rcode:    dns.RcodeRefused,
		Upstream: "",
		ProxyTag: "",
		RTT:      0,
		Cached:   false,
	}
}

// updateAvgRTT updates the average RTT using exponential moving average.
func updateAvgRTT(current time.Duration, sample time.Duration, count int64) time.Duration {
	if count <= 1 {
		return sample
	}
	// Use weighted average: new = old * (n-1)/n + sample/n
	avg := int64(current)*(count-1)/count + int64(sample)/count
	return time.Duration(avg)
}

// QueryTypeFromString parses a DNS query type string (e.g., "A", "AAAA") to QueryType.
func QueryTypeFromString(s string) (QueryType, error) {
	s = strings.TrimSpace(s)
	switch strings.ToUpper(s) {
	case "A":
		return TypeA, nil
	case "AAAA":
		return TypeAAAA, nil
	case "CNAME":
		return TypeCNAME, nil
	case "TXT":
		return TypeTXT, nil
	case "MX":
		return TypeMX, nil
	case "SRV":
		return TypeSRV, nil
	case "NS":
		return TypeNS, nil
	case "PTR":
		return TypePTR, nil
	case "SOA":
		return TypeSOA, nil
	default:
		// Try to parse as numeric value.
		return 0, fmt.Errorf("unknown query type: %s", s)
	}
}
