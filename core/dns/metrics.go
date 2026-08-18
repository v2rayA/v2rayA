package dns

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// DnsMetrics DNS 模块指标收集器
// 使用 atomic 操作和 sync.Map 确保并发安全
type DnsMetrics struct {
	// 查询总量
	TotalQueries     atomic.Int64
	TotalResponses   atomic.Int64
	TotalErrors      atomic.Int64
	TotalTimeouts    atomic.Int64

	// 按记录类型统计
	QueriesByType sync.Map // map[QueryType]int64

	// 按上游统计
	QueriesByUpstream   sync.Map // map[string]*UpstreamMetrics
	ErrorsByUpstream    sync.Map // map[string]int64

	// 按规则统计
	HitsByRule sync.Map // map[string]int64 — ruleID → 命中次数

	// 缓存统计
	CacheHits      atomic.Int64
	CacheMisses    atomic.Int64
	CacheEvictions atomic.Int64
	PrefetchCount  atomic.Int64

	// 延迟分布
	LatencyBuckets []LatencyBucket

	// Action 统计
	RoutedQueries   atomic.Int64
	RejectedQueries atomic.Int64
	BypassedQueries atomic.Int64

	mu        sync.Mutex
	startedAt time.Time
}

// UpstreamMetrics 上游服务器的详细指标
type UpstreamMetrics struct {
	Queries   int64
	Successes int64
	Errors    int64
	AvgRTT    time.Duration
	MinRTT    time.Duration
	MaxRTT    time.Duration
	LastQuery time.Time
	mu        sync.Mutex
}

// LatencyBucket 延迟分布区间
type LatencyBucket struct {
	Label string // "<1ms", "1-5ms", "5-10ms", "10-50ms", "50-200ms", "200ms+"
	Count int64
}

// MetricsSnapshot 指标快照，用于一次性读取所有指标
type MetricsSnapshot struct {
	TotalQueries     int64
	TotalResponses   int64
	TotalErrors      int64
	TotalTimeouts    int64
	QueriesByType    map[string]int64
	QueriesByUpstream map[string]*UpstreamMetrics
	ErrorsByUpstream map[string]int64
	HitsByRule       map[string]int64
	CacheHits        int64
	CacheMisses      int64
	CacheEvictions   int64
	PrefetchCount    int64
	LatencyBuckets   []LatencyBucket
	RoutedQueries    int64
	RejectedQueries  int64
	BypassedQueries  int64
	Uptime           string
}

// NewDnsMetrics 创建新的指标收集器
func NewDnsMetrics() *DnsMetrics {
	return &DnsMetrics{
		startedAt: time.Now(),
		LatencyBuckets: []LatencyBucket{
			{Label: "<1ms", Count: 0},
			{Label: "1-5ms", Count: 0},
			{Label: "5-10ms", Count: 0},
			{Label: "10-50ms", Count: 0},
			{Label: "50-200ms", Count: 0},
			{Label: "200ms+", Count: 0},
		},
	}
}

// getOrCreateUpstreamMetrics 获取或创建上游指标
func (m *DnsMetrics) getOrCreateUpstreamMetrics(upstream string) *UpstreamMetrics {
	actual, _ := m.QueriesByUpstream.LoadOrStore(upstream, &UpstreamMetrics{
		MinRTT: time.Duration(1<<63 - 1), // MaxInt64 作为初始值
	})
	return actual.(*UpstreamMetrics)
}

// getOrInitCounter 获取或初始化一个 atomic.Int64 计数器
func getOrInitCounter(m *sync.Map, key interface{}) *atomic.Int64 {
	v, _ := m.LoadOrStore(key, &atomic.Int64{})
	return v.(*atomic.Int64)
}

// RecordQuery 记录一次查询
func (m *DnsMetrics) RecordQuery(query *DnsQuery, upstream string) {
	m.TotalQueries.Add(1)

	// 按类型统计（使用 atomic.Int64 指针避免竞态）
	counter := getOrInitCounter(&m.QueriesByType, query.QType)
	counter.Add(1)

	// 按上游统计
	if upstream != "" {
		um := m.getOrCreateUpstreamMetrics(upstream)
		um.mu.Lock()
		um.Queries++
		um.LastQuery = time.Now()
		um.mu.Unlock()
	}
}

// RecordResponse 记录一次响应
func (m *DnsMetrics) RecordResponse(resp *DnsResponse, rtt time.Duration) {
	m.TotalResponses.Add(1)

	// 记录延迟分布
	m.recordLatency(rtt)

	// 按上游记录 RTT
	upstream := resp.Upstream
	if upstream != "" {
		um := m.getOrCreateUpstreamMetrics(upstream)
		um.mu.Lock()
		um.Successes++
		if rtt > 0 {
			// 更新平均 RTT（指数移动平均）
			if um.AvgRTT == 0 {
				um.AvgRTT = rtt
			} else {
				um.AvgRTT = (um.AvgRTT*2 + rtt) / 3
			}
			// 更新最小 RTT
			if rtt < um.MinRTT {
				um.MinRTT = rtt
			}
			// 更新最大 RTT
			if rtt > um.MaxRTT {
				um.MaxRTT = rtt
			}
		}
		um.mu.Unlock()
	}
}

// RecordError 记录一次错误
func (m *DnsMetrics) RecordError(upstream string, err error) {
	m.TotalErrors.Add(1)

	if upstream != "" {
		// 按上游统计错误（使用 atomic.Int64 指针）
		counter := getOrInitCounter(&m.ErrorsByUpstream, upstream)
		counter.Add(1)

		um := m.getOrCreateUpstreamMetrics(upstream)
		um.mu.Lock()
		um.Errors++
		um.mu.Unlock()
	}
}

// RecordTimeout 记录一次超时
func (m *DnsMetrics) RecordTimeout(upstream string) {
	m.TotalTimeouts.Add(1)
	m.RecordError(upstream, fmt.Errorf("timeout"))
}

// RecordCacheHit 记录缓存命中
func (m *DnsMetrics) RecordCacheHit() {
	m.CacheHits.Add(1)
}

// RecordCacheMiss 记录缓存未命中
func (m *DnsMetrics) RecordCacheMiss() {
	m.CacheMisses.Add(1)
}

// RecordCacheEviction 记录缓存淘汰
func (m *DnsMetrics) RecordCacheEviction() {
	m.CacheEvictions.Add(1)
}

// RecordPrefetch 记录预取
func (m *DnsMetrics) RecordPrefetch() {
	m.PrefetchCount.Add(1)
}

// RecordRuleHit 记录规则命中
func (m *DnsMetrics) RecordRuleHit(ruleID string) {
	if ruleID == "" {
		return
	}
	counter := getOrInitCounter(&m.HitsByRule, ruleID)
	counter.Add(1)
}

// RecordAction 记录分流动作
func (m *DnsMetrics) RecordAction(action string) {
	switch action {
	case "route":
		m.RoutedQueries.Add(1)
	case "reject":
		m.RejectedQueries.Add(1)
	case "bypass":
		m.BypassedQueries.Add(1)
	}
}

// recordLatency 记录延迟到对应区间
func (m *DnsMetrics) recordLatency(rtt time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ms := rtt.Milliseconds()
	switch {
	case ms < 1:
		m.LatencyBuckets[0].Count++
	case ms <= 5:
		m.LatencyBuckets[1].Count++
	case ms <= 10:
		m.LatencyBuckets[2].Count++
	case ms <= 50:
		m.LatencyBuckets[3].Count++
	case ms <= 200:
		m.LatencyBuckets[4].Count++
	default:
		m.LatencyBuckets[5].Count++
	}
}

// Snapshot 返回当前指标的原子快照
func (m *DnsMetrics) Snapshot() *MetricsSnapshot {
	s := &MetricsSnapshot{
		TotalQueries:     m.TotalQueries.Load(),
		TotalResponses:   m.TotalResponses.Load(),
		TotalErrors:      m.TotalErrors.Load(),
		TotalTimeouts:    m.TotalTimeouts.Load(),
		CacheHits:        m.CacheHits.Load(),
		CacheMisses:      m.CacheMisses.Load(),
		CacheEvictions:   m.CacheEvictions.Load(),
		PrefetchCount:    m.PrefetchCount.Load(),
		RoutedQueries:    m.RoutedQueries.Load(),
		RejectedQueries:  m.RejectedQueries.Load(),
		BypassedQueries:  m.BypassedQueries.Load(),
		Uptime:           time.Since(m.startedAt).Round(time.Second).String(),
		QueriesByType:    make(map[string]int64),
		QueriesByUpstream: make(map[string]*UpstreamMetrics),
		ErrorsByUpstream: make(map[string]int64),
		HitsByRule:       make(map[string]int64),
	}

	// 拷贝按类型统计（atomic.Int64 指针）
	m.QueriesByType.Range(func(key, val interface{}) bool {
		qt := key.(QueryType)
		counter := val.(*atomic.Int64)
		s.QueriesByType[QTypeToString(qt)] = counter.Load()
		return true
	})

	// 拷贝按上游统计（使用指针避免复制 sync.Mutex）
	m.QueriesByUpstream.Range(func(key, val interface{}) bool {
		upstream := key.(string)
		um := val.(*UpstreamMetrics)
		s.QueriesByUpstream[upstream] = um
		return true
	})

	// 拷贝按上游错误统计（atomic.Int64 指针）
	m.ErrorsByUpstream.Range(func(key, val interface{}) bool {
		counter := val.(*atomic.Int64)
		s.ErrorsByUpstream[key.(string)] = counter.Load()
		return true
	})

	// 拷贝按规则命中统计（atomic.Int64 指针）
	m.HitsByRule.Range(func(key, val interface{}) bool {
		counter := val.(*atomic.Int64)
		s.HitsByRule[key.(string)] = counter.Load()
		return true
	})

	// 拷贝延迟分布
	m.mu.Lock()
	s.LatencyBuckets = make([]LatencyBucket, len(m.LatencyBuckets))
	copy(s.LatencyBuckets, m.LatencyBuckets)
	m.mu.Unlock()

	return s
}

// Reset 重置所有指标
func (m *DnsMetrics) Reset() {
	m.TotalQueries.Store(0)
	m.TotalResponses.Store(0)
	m.TotalErrors.Store(0)
	m.TotalTimeouts.Store(0)
	m.CacheHits.Store(0)
	m.CacheMisses.Store(0)
	m.CacheEvictions.Store(0)
	m.PrefetchCount.Store(0)
	m.RoutedQueries.Store(0)
	m.RejectedQueries.Store(0)
	m.BypassedQueries.Store(0)

	m.QueriesByType.Range(func(key, _ interface{}) bool {
		m.QueriesByType.Delete(key)
		return true
	})
	m.QueriesByUpstream.Range(func(key, _ interface{}) bool {
		m.QueriesByUpstream.Delete(key)
		return true
	})
	m.ErrorsByUpstream.Range(func(key, _ interface{}) bool {
		m.ErrorsByUpstream.Delete(key)
		return true
	})
	m.HitsByRule.Range(func(key, _ interface{}) bool {
		m.HitsByRule.Delete(key)
		return true
	})

	m.mu.Lock()
	for i := range m.LatencyBuckets {
		m.LatencyBuckets[i].Count = 0
	}
	m.mu.Unlock()

	m.startedAt = time.Now()
}

// String 返回人类可读的指标概要
func (m *DnsMetrics) String() string {
	s := m.Snapshot()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("📊 DNS Metrics (uptime: %s)\n", s.Uptime))
	b.WriteString(fmt.Sprintf("   Queries: %d total, %d responses, %d errors, %d timeouts\n",
		s.TotalQueries, s.TotalResponses, s.TotalErrors, s.TotalTimeouts))
	b.WriteString(fmt.Sprintf("   Cache: %d hits, %d misses, %d evictions, %d prefetches\n",
		s.CacheHits, s.CacheMisses, s.CacheEvictions, s.PrefetchCount))
	b.WriteString(fmt.Sprintf("   Actions: %d routed, %d rejected, %d bypassed\n",
		s.RoutedQueries, s.RejectedQueries, s.BypassedQueries))

	if len(s.QueriesByType) > 0 {
		b.WriteString("   By Type:\n")
		for qt, count := range s.QueriesByType {
			b.WriteString(fmt.Sprintf("     %s: %d\n", qt, count))
		}
	}
	if len(s.QueriesByUpstream) > 0 {
		b.WriteString("   By Upstream:\n")
		for upstream, um := range s.QueriesByUpstream {
			b.WriteString(fmt.Sprintf("     %s: queries=%d successes=%d errors=%d avgRTT=%s\n",
				upstream, um.Queries, um.Successes, um.Errors, um.AvgRTT))
		}
	}
	if len(s.HitsByRule) > 0 {
		b.WriteString("   Rule Hits:\n")
		for ruleID, count := range s.HitsByRule {
			b.WriteString(fmt.Sprintf("     %s: %d\n", ruleID, count))
		}
	}
	if len(s.LatencyBuckets) > 0 {
		b.WriteString("   Latency Distribution:\n")
		for _, lb := range s.LatencyBuckets {
			if lb.Count > 0 {
				b.WriteString(fmt.Sprintf("     %s: %d\n", lb.Label, lb.Count))
			}
		}
	}
	return b.String()
}
