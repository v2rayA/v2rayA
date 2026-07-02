package dns

import (
	"fmt"
	"strings"
	"time"
)

// Diagnostics DNS 诊断信息收集器
type Diagnostics struct {
	module *DnsModule
}

// DebugInfo 调试信息快照
type DebugInfo struct {
	Uptime          string
	ListenerStatus  string
	CacheStats      CacheStats
	MetricsSnapshot *MetricsSnapshot
	UpstreamStatus  []UpstreamStatus
	RuleStatus      []RuleStatus
}

// UpstreamStatus 上游服务器状态
type UpstreamStatus struct {
	ID       string
	Addr     string
	ProxyTag string
	Healthy  bool
	Queries  int64
	Errors   int64
	AvgRTT   string
	LastSeen string
}

// RuleStatus 规则状态
type RuleStatus struct {
	ID      string
	Matches int64
	Action  string
	Active  bool
}

// NewDiagnostics 创建新的诊断信息收集器
func NewDiagnostics(module *DnsModule) *Diagnostics {
	return &Diagnostics{
		module: module,
	}
}

// GetDebugInfo 收集完整的调试信息
func (d *Diagnostics) GetDebugInfo() *DebugInfo {
	info := &DebugInfo{}

	// 运行时间
	info.Uptime = time.Since(d.module.stats.StartedAt).Round(time.Second).String()

	// 监听状态
	listenerStatus := "stopped"
	if d.module.listener != nil && d.module.listener.Healthy() {
		listenerStatus = fmt.Sprintf("running on %s", d.module.config.Listener.ListenAddr)
	}
	info.ListenerStatus = listenerStatus

	// 缓存统计
	if d.module.cache != nil {
		info.CacheStats = d.module.cache.Stats()
	}

	// 指标快照
	if d.module.metrics != nil {
		info.MetricsSnapshot = d.module.metrics.Snapshot()
	}

	// 上游状态
	upstreamMgr := d.module.GetUpstreamManager()
	if upstreamMgr != nil {
		upstreams := upstreamMgr.ListUpstreams()
		info.UpstreamStatus = make([]UpstreamStatus, 0, len(upstreams))

		// 获取指标快照用于填充上游统计
		var snap *MetricsSnapshot
		if d.module.metrics != nil {
			snap = d.module.metrics.Snapshot()
		}

		for _, u := range upstreams {
			status := UpstreamStatus{
				ID:       u.ID,
				Addr:     u.Addr,
				ProxyTag: u.ProxyTag,
				Healthy:  u.Healthy,
			}

			// 从指标快照中获取上游统计
			if snap != nil {
				// 尝试用 ID 查找
				if um, ok := snap.QueriesByUpstream[u.ID]; ok {
					status.Queries = um.Queries
					status.Errors = um.Errors
					status.AvgRTT = um.AvgRTT.Round(time.Millisecond).String()
					if !um.LastQuery.IsZero() {
						status.LastSeen = um.LastQuery.Format("15:04:05")
					}
				} else if um, ok := snap.QueriesByUpstream[u.Addr]; ok {
					status.Queries = um.Queries
					status.Errors = um.Errors
					status.AvgRTT = um.AvgRTT.Round(time.Millisecond).String()
					if !um.LastQuery.IsZero() {
						status.LastSeen = um.LastQuery.Format("15:04:05")
					}
				}
			}

			if status.LastSeen == "" {
				status.LastSeen = "never"
			}
			if status.AvgRTT == "" {
				status.AvgRTT = "N/A"
			}

			info.UpstreamStatus = append(info.UpstreamStatus, status)
		}
	}

	// 规则状态
	router := d.module.GetRouter()
	if router != nil {
		rules := router.ListRules()
		info.RuleStatus = make([]RuleStatus, 0, len(rules))

		// 获取规则命中次数
		ruleHits := make(map[string]int64)
		if d.module.metrics != nil {
			snap := d.module.metrics.Snapshot()
			for k, v := range snap.HitsByRule {
				ruleHits[k] = v
			}
		}

		for _, rule := range rules {
			status := RuleStatus{
				ID:     rule.ID,
				Action: rule.Action,
				Active: true,
				Matches: ruleHits[rule.ID],
			}
			info.RuleStatus = append(info.RuleStatus, status)
		}
	}

	return info
}

// Summary 返回人类可读的诊断概要
func (d *Diagnostics) Summary() string {
	info := d.GetDebugInfo()

	var b strings.Builder
	b.WriteString("=== DNS Module Diagnostics ===\n")
	b.WriteString(fmt.Sprintf("Uptime:        %s\n", info.Uptime))
	b.WriteString(fmt.Sprintf("Listener:      %s\n", info.ListenerStatus))

	// 指标概要
	if info.MetricsSnapshot != nil {
		s := info.MetricsSnapshot
		b.WriteString(fmt.Sprintf("Queries:       %d total, %d responses, %d errors, %d timeouts\n",
			s.TotalQueries, s.TotalResponses, s.TotalErrors, s.TotalTimeouts))
		b.WriteString(fmt.Sprintf("Cache:         %d hits, %d misses, %d evictions (rate: %.1f%%)\n",
			s.CacheHits, s.CacheMisses, s.CacheEvictions, calcHitRate(s.CacheHits, s.CacheMisses)))
		b.WriteString(fmt.Sprintf("Actions:       %d routed, %d rejected, %d bypassed\n",
			s.RoutedQueries, s.RejectedQueries, s.BypassedQueries))
	}

	// 上游状态
	b.WriteString("\n--- Upstreams ---\n")
	if len(info.UpstreamStatus) > 0 {
		for _, us := range info.UpstreamStatus {
			health := "✅"
			if !us.Healthy {
				health = "❌"
			}
			b.WriteString(fmt.Sprintf("  %s %s (%s) tag=%s queries=%d errors=%d avgRTT=%s last=%s\n",
				health, us.ID, us.Addr, us.ProxyTag, us.Queries, us.Errors, us.AvgRTT, us.LastSeen))
		}
	} else {
		b.WriteString("  (none configured)\n")
	}

	// 规则状态
	b.WriteString("\n--- Rules ---\n")
	if len(info.RuleStatus) > 0 {
		for _, rs := range info.RuleStatus {
			b.WriteString(fmt.Sprintf("  %s -> %s (matches: %d)\n", rs.ID, rs.Action, rs.Matches))
		}
	} else {
		b.WriteString("  (no rules configured)\n")
	}

	// 缓存详情
	b.WriteString("\n--- Cache ---\n")
	cs := info.CacheStats
	b.WriteString(fmt.Sprintf("  Size: %d entries, Hits: %d, Misses: %d, HitRate: %.1f%%\n",
		cs.Size, cs.Hits, cs.Misses, cs.HitRate*100))
	b.WriteString(fmt.Sprintf("  Evictions: %d, Prefetches: %d\n", cs.Evictions, cs.PrefetchCnt))

	// 延迟分布
	if info.MetricsSnapshot != nil && len(info.MetricsSnapshot.LatencyBuckets) > 0 {
		b.WriteString("\n--- Latency Distribution ---\n")
		for _, lb := range info.MetricsSnapshot.LatencyBuckets {
			if lb.Count > 0 {
				bar := strings.Repeat("█", int(lb.Count/10))
				if bar == "" && lb.Count > 0 {
					bar = "▏"
				}
				b.WriteString(fmt.Sprintf("  %-10s: %5d %s\n", lb.Label, lb.Count, bar))
			}
		}
	}

	return b.String()
}

// CacheDump 导出缓存内容（调试用）
// 返回当前缓存中所有的键列表
func (d *Diagnostics) CacheDump() []CacheKey {
	if d.module.cache == nil {
		return nil
	}
	return d.module.cache.Keys()
}

// calcHitRate 计算缓存命中率（百分比）
func calcHitRate(hits, misses int64) float64 {
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}
