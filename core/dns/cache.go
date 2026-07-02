package dns

import (
	"container/list"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// CacheKey 缓存键 - 四元组确保隔离
type CacheKey struct {
	UpstreamAddr string    // 上游地址
	ProxyTag     string    // 代理渠道标签
	Name         string    // 查询域名（小写）
	QType        QueryType // 查询类型
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Key               CacheKey
	Response          *DnsResponse
	OriginalTTL       uint32    // 上游返回的原始 TTL
	EffectiveTTL      uint32    // 钳制后的生效 TTL
	StoredAt          time.Time // 存储时间
	ExpiresAt         time.Time // 过期时间
	HitCount          int64     // 命中次数
	Prefetching       bool      // 是否正在预取
	Negative          bool      // 是否为负缓存
	PrefetchFailCount int       // 预取失败次数
	PrefetchBackoff   time.Duration // 当前预取退避时长
	LastPrefetchFail  time.Time // 上次预取失败时间
}

// DnsCache DNS 缓存（LRU + TTL 双淘汰）
type DnsCache struct {
	config  *CacheConfig
	entries map[CacheKey]*list.Element
	lruList *list.List // LRU 顺序链表
	mu      sync.RWMutex
	maxSize int
	stats   CacheStats
}

// CacheStats 缓存统计
type CacheStats struct {
	Size        int
	Hits        int64
	Misses      int64
	Evictions   int64
	PrefetchCnt int64
	HitRate     float64
}

// NewDnsCache 创建新的 DNS 缓存实例
func NewDnsCache(config *CacheConfig) *DnsCache {
	maxSize := config.Size
	if maxSize <= 0 {
		maxSize = 4096
	}
	return &DnsCache{
		config:  config,
		entries: make(map[CacheKey]*list.Element),
		lruList: list.New(),
		maxSize: maxSize,
	}
}

// calculateEffectiveTTL 计算生效 TTL（参照 mihomo 策略）
// 公式: effectiveTTL = min(max(minTTL, originalTTL), maxTTL)
func calculateEffectiveTTL(originalTTL uint32, minTTL, maxTTL int) uint32 {
	if minTTL <= 0 {
		minTTL = 60 // 默认最小 60 秒
	}
	if maxTTL <= 0 {
		maxTTL = 86400 // 默认最大 86400 秒（24小时）
	}

	effective := originalTTL
	if effective < uint32(minTTL) {
		effective = uint32(minTTL)
	}
	if effective > uint32(maxTTL) {
		effective = uint32(maxTTL)
	}
	return effective
}

// negativeTTL 负缓存 TTL（秒）
// NXDOMAIN (3) → 60s, SERVFAIL (2) → 30s, REFUSED (5) → 10s, 其他 → 5s
func negativeTTL(rcode int) uint32 {
	switch rcode {
	case dns.RcodeNameError: // NXDOMAIN
		return 60
	case dns.RcodeServerFailure: // SERVFAIL
		return 30
	case dns.RcodeRefused: // REFUSED
		return 10
	default:
		return 5
	}
}

// prefetchBackoff 计算预取退避时间
// 3 次失败后，退避时间逐步增加 5s→30s→120s
func prefetchBackoff(failCount int) time.Duration {
	switch failCount {
	case 0:
		return 0
	case 1:
		return 5 * time.Second
	case 2:
		return 30 * time.Second
	default:
		return 120 * time.Second
	}
}

// Get 获取缓存条目
// 返回缓存的响应和是否命中。如果条目不存在或已过期，返回 nil, false。
func (c *DnsCache) Get(key CacheKey) (*DnsResponse, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.entries[key]
	if !ok {
		c.stats.Misses++
		c.updateHitRate()
		return nil, false
	}

	entry, ok := elem.Value.(*CacheEntry)
	if !ok {
		c.stats.Misses++
		c.updateHitRate()
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(entry.ExpiresAt) {
		c.removeElement(elem)
		c.stats.Misses++
		c.updateHitRate()
		return nil, false
	}

	// 移动到链表头部（最近使用）
	c.lruList.MoveToFront(elem)

	entry.HitCount++
	c.stats.Hits++
	c.updateHitRate()

	// 标记响应来自缓存
	resp := *entry.Response
	resp.Cached = true
	return &resp, true
}

// Set 设置缓存条目
// 如果键已存在则更新；否则添加新条目。超过最大容量时淘汰最久未使用的条目。
func (c *DnsCache) Set(key CacheKey, resp *DnsResponse, originalTTL uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// 如果键已存在，更新旧条目
	if elem, ok := c.entries[key]; ok {
		c.lruList.MoveToFront(elem)
		entry := elem.Value.(*CacheEntry)

		// 重置预取状态
		entry.Prefetching = false

		// 如果是成功响应，重置预取失败计数
		if resp != nil && resp.Rcode == dns.RcodeSuccess {
			entry.PrefetchFailCount = 0
			entry.PrefetchBackoff = 0
		}

		// 计算生效 TTL
		entry.Response = resp
		entry.OriginalTTL = originalTTL
		entry.StoredAt = now
		entry.EffectiveTTL = computeEffectiveTTL(resp, originalTTL, c.config)
		entry.ExpiresAt = now.Add(time.Duration(entry.EffectiveTTL) * time.Second)
		entry.Negative = resp != nil && resp.Rcode != dns.RcodeSuccess

		c.stats.Size = c.lruList.Len()
		c.updateHitRate()
		return
	}

	// 计算生效 TTL
	effectiveTTL := computeEffectiveTTL(resp, originalTTL, c.config)
	negative := resp != nil && resp.Rcode != dns.RcodeSuccess

	entry := &CacheEntry{
		Key:          key,
		Response:     resp,
		OriginalTTL:  originalTTL,
		EffectiveTTL: effectiveTTL,
		StoredAt:     now,
		ExpiresAt:    now.Add(time.Duration(effectiveTTL) * time.Second),
		Negative:     negative,
	}

	elem := c.lruList.PushFront(entry)
	c.entries[key] = elem

	// LRU 淘汰：超过最大条目数时淘汰最久未使用的条目
	if c.lruList.Len() > c.maxSize {
		back := c.lruList.Back()
		if back != nil {
			c.removeElement(back)
			c.stats.Evictions++
		}
	}

	c.stats.Size = c.lruList.Len()
	c.updateHitRate()
}

// computeEffectiveTTL 根据响应类型计算生效 TTL
// 成功响应使用 min/max 钳制；负响应使用独立的负缓存 TTL
func computeEffectiveTTL(resp *DnsResponse, originalTTL uint32, config *CacheConfig) uint32 {
	if resp != nil && resp.Rcode != dns.RcodeSuccess {
		return negativeTTL(resp.Rcode)
	}
	return calculateEffectiveTTL(originalTTL, config.MinTTL, config.MaxTTL)
}

// Remove 从缓存中移除指定键的条目
func (c *DnsCache) Remove(key CacheKey) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.entries[key]; ok {
		c.removeElement(elem)
		c.stats.Size = c.lruList.Len()
		c.updateHitRate()
	}
}

// Clear 清空所有缓存条目
func (c *DnsCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[CacheKey]*list.Element)
	c.lruList.Init()
	c.stats.Size = 0
	c.updateHitRate()
}

// Stats 返回缓存统计信息的快照
func (c *DnsCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.stats
}

// removeElement 从 LRU 链表和索引中移除元素
func (c *DnsCache) removeElement(elem *list.Element) {
	entry, ok := elem.Value.(*CacheEntry)
	if ok {
		delete(c.entries, entry.Key)
	}
	c.lruList.Remove(elem)
}

// updateHitRate 更新缓存命中率
func (c *DnsCache) updateHitRate() {
	total := c.stats.Hits + c.stats.Misses
	if total > 0 {
		c.stats.HitRate = float64(c.stats.Hits) / float64(total)
	} else {
		c.stats.HitRate = 0
	}
}

// ShouldPrefetch 判断是否需要预取
// 策略：TTL 剩余低于 10% 或低于 PrefetchThreshold 秒时触发
func (c *DnsCache) ShouldPrefetch(key CacheKey) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elem, ok := c.entries[key]
	if !ok {
		return false
	}

	entry, ok := elem.Value.(*CacheEntry)
	if !ok {
		return false
	}

	return shouldPrefetch(entry, c.config)
}

// shouldPrefetch 判断条目是否需要预取（内部方法，不持有锁）
func shouldPrefetch(entry *CacheEntry, config *CacheConfig) bool {
	if !config.Prefetch {
		return false
	}

	if entry.Prefetching {
		return false // 已在预取中，避免重复
	}

	// 预取失败退避检查
	if entry.PrefetchFailCount >= 3 {
		// 超过 3 次失败，不再自动触发预取
		return false
	}

	// 检查退避期是否已过
	backoff := prefetchBackoff(entry.PrefetchFailCount)
	if backoff > 0 && time.Since(entry.LastPrefetchFail) < backoff {
		return false
	}

	remaining := time.Until(entry.ExpiresAt)
	totalTTL := entry.ExpiresAt.Sub(entry.StoredAt)

	// TTL 剩余低于 10%
	if totalTTL > 0 && remaining < totalTTL/10 {
		return true
	}

	// TTL 剩余低于阈值（默认 30 秒）
	threshold := config.PrefetchThreshold
	if threshold <= 0 {
		threshold = 30
	}
	if remaining < time.Duration(threshold)*time.Second {
		return true
	}

	return false
}

// SetPrefetching 设置条目的预取状态
func (c *DnsCache) SetPrefetching(key CacheKey, status bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.entries[key]
	if !ok {
		return
	}

	entry, ok := elem.Value.(*CacheEntry)
	if !ok {
		return
	}

	entry.Prefetching = status
	if status {
		c.stats.PrefetchCnt++
	}
}

// RecordPrefetchFailure 记录预取失败并更新退避状态
func (c *DnsCache) RecordPrefetchFailure(key CacheKey) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.entries[key]
	if !ok {
		return
	}

	entry, ok := elem.Value.(*CacheEntry)
	if !ok {
		return
	}

	entry.PrefetchFailCount++
	entry.LastPrefetchFail = time.Now()
	entry.Prefetching = false
	entry.PrefetchBackoff = prefetchBackoff(entry.PrefetchFailCount)
}

// RecordPrefetchSuccess 记录预取成功并重置退避状态
func (c *DnsCache) RecordPrefetchSuccess(key CacheKey) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.entries[key]
	if !ok {
		return
	}

	entry, ok := elem.Value.(*CacheEntry)
	if !ok {
		return
	}

	entry.PrefetchFailCount = 0
	entry.PrefetchBackoff = 0
	entry.Prefetching = false
}

// Len 返回当前缓存条目数
func (c *DnsCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lruList.Len()
}

// Keys 返回所有缓存键的列表（用于调试和监控）
func (c *DnsCache) Keys() []CacheKey {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]CacheKey, 0, len(c.entries))
	for k := range c.entries {
		keys = append(keys, k)
	}
	return keys
}
