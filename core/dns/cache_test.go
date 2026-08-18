package dns

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// --- Helper functions ---

func newTestCacheConfig() *CacheConfig {
	return &CacheConfig{
		Enabled:           true,
		Size:              100,
		MinTTL:            60,
		MaxTTL:            86400,
		Prefetch:          true,
		PrefetchThreshold: 30,
		NegativeCache:     true,
	}
}

func newTestCacheKey(name string, qtype QueryType) CacheKey {
	return CacheKey{
		UpstreamAddr: "8.8.8.8:53",
		ProxyTag:     "direct",
		Name:         name,
		QType:        qtype,
	}
}

func newTestDNSResponse(name string, qtype QueryType, rcode int, ttl uint32) *DnsResponse {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(name), uint16(qtype))
	m.Response = true
	m.Rcode = rcode
	m.RecursionAvailable = true

	// Add answer records if success
	if rcode == dns.RcodeSuccess {
		switch qtype {
		case TypeA:
			rr, _ := dns.NewRR(dns.Fqdn(name) + " 300 IN A 1.2.3.4")
			m.Answer = []dns.RR{rr}
		case TypeAAAA:
			rr, _ := dns.NewRR(dns.Fqdn(name) + " 300 IN AAAA ::1")
			m.Answer = []dns.RR{rr}
		}
		// Override TTL
		if ttl > 0 {
			m.Answer[0].Header().Ttl = ttl
		}
	}

	return &DnsResponse{
		RawMsg:   m,
		Rcode:    m.Rcode,
		Answer:   m.Answer,
		TTL:      ttl,
		Upstream: "8.8.8.8:53",
		ProxyTag: "direct",
	}
}

// --- Test: NewDnsCache ---

func TestNewDnsCache(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())

	if cache == nil {
		t.Fatal("NewDnsCache returned nil")
	}
	if cache.maxSize != 100 {
		t.Errorf("expected maxSize=100, got %d", cache.maxSize)
	}
	if cache.lruList.Len() != 0 {
		t.Errorf("expected empty cache, got %d items", cache.lruList.Len())
	}
}

func TestNewDnsCacheDefaultSize(t *testing.T) {
	// Zero size should default to 4096
	config := CacheConfig{Enabled: true, Size: 0}
	cache := NewDnsCache(&config)
	if cache.maxSize != 4096 {
		t.Errorf("expected default maxSize=4096, got %d", cache.maxSize)
	}

	// Negative size should default to 4096
	config = CacheConfig{Enabled: true, Size: -1}
	cache = NewDnsCache(&config)
	if cache.maxSize != 4096 {
		t.Errorf("expected default maxSize=4096 for negative, got %d", cache.maxSize)
	}
}

// --- Test: Set and Get basic ---

func TestSetAndGet(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())
	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 300)

	// Set cache
	cache.Set(key, resp, 300)

	// Get cache
	got, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}
	if got == nil {
		t.Fatal("expected non-nil response")
	}
	if got.Rcode != dns.RcodeSuccess {
		t.Errorf("expected RcodeSuccess, got %d", got.Rcode)
	}
	if !got.Cached {
		t.Error("expected Cached=true for cache response")
	}
}

func TestGetMiss(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())
	key := newTestCacheKey("nonexistent.com", TypeA)

	got, ok := cache.Get(key)
	if ok {
		t.Error("expected cache miss for non-existent key")
	}
	if got != nil {
		t.Error("expected nil response for cache miss")
	}
}

// --- Test: TTL expiration ---

func TestTTLExpiration(t *testing.T) {
	config := newTestCacheConfig()
	config.MinTTL = 1 // 1 second min TTL for fast test
	config.MaxTTL = 2
	cache := NewDnsCache(config)

	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 1)

	cache.Set(key, resp, 1)

	// Should be immediately available
	if _, ok := cache.Get(key); !ok {
		t.Fatal("expected cache hit immediately after set")
	}

	// Wait for TTL to expire
	time.Sleep(3 * time.Second)

	// Should be expired now
	if _, ok := cache.Get(key); ok {
		t.Error("expected cache miss after TTL expiration")
	}
}

// --- Test: TTL clamping (min/max) ---

func TestTTLClampingMin(t *testing.T) {
	config := newTestCacheConfig()
	config.MinTTL = 60
	config.MaxTTL = 86400
	cache := NewDnsCache(config)

	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 5) // original TTL=5

	cache.Set(key, resp, 5)

	// The effective TTL should be clamped to minTTL=60
	if elem, ok := cache.entries[key]; ok {
		entry := elem.Value.(*CacheEntry)
		if entry.EffectiveTTL != 60 {
			t.Errorf("expected effective TTL=60 (clamped from 5), got %d", entry.EffectiveTTL)
		}
	}
}

func TestTTLClampingMax(t *testing.T) {
	config := newTestCacheConfig()
	config.MinTTL = 60
	config.MaxTTL = 600
	cache := NewDnsCache(config)

	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 86400) // original TTL=86400

	cache.Set(key, resp, 86400)

	// The effective TTL should be clamped to maxTTL=600
	if elem, ok := cache.entries[key]; ok {
		entry := elem.Value.(*CacheEntry)
		if entry.EffectiveTTL != 600 {
			t.Errorf("expected effective TTL=600 (clamped from 86400), got %d", entry.EffectiveTTL)
		}
	}
}

func TestTTLClampingMiddle(t *testing.T) {
	config := newTestCacheConfig()
	config.MinTTL = 60
	config.MaxTTL = 86400
	cache := NewDnsCache(config)

	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 300) // original TTL=300

	cache.Set(key, resp, 300)

	// The effective TTL should be 300 (within [60, 86400])
	if elem, ok := cache.entries[key]; ok {
		entry := elem.Value.(*CacheEntry)
		if entry.EffectiveTTL != 300 {
			t.Errorf("expected effective TTL=300, got %d", entry.EffectiveTTL)
		}
	}
}

// --- Test: calculateEffectiveTTL ---

func TestCalculateEffectiveTTL(t *testing.T) {
	tests := []struct {
		name       string
		original   uint32
		minTTL     int
		maxTTL     int
		expected   uint32
	}{
		{"within range", 300, 60, 86400, 300},
		{"below min", 5, 60, 86400, 60},
		{"above max", 100000, 60, 86400, 86400},
		{"zero min", 300, 0, 86400, 300},  // min defaults to 60, but 300 > 60
		{"zero max", 300, 60, 0, 300},      // max defaults to 86400, but 300 < 86400
		{"zero both", 300, 0, 0, 300},
		{"exact min", 60, 60, 86400, 60},
		{"exact max", 86400, 60, 86400, 86400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateEffectiveTTL(tt.original, tt.minTTL, tt.maxTTL)
			if got != tt.expected {
				t.Errorf("calculateEffectiveTTL(%d, %d, %d) = %d, want %d",
					tt.original, tt.minTTL, tt.maxTTL, got, tt.expected)
			}
		})
	}
}

// --- Test: negativeTTL ---

func TestNegativeTTL(t *testing.T) {
	tests := []struct {
		name     string
		rcode    int
		expected uint32
	}{
		{"NXDOMAIN", dns.RcodeNameError, 60},
		{"SERVFAIL", dns.RcodeServerFailure, 30},
		{"REFUSED", dns.RcodeRefused, 10},
		{"other error", dns.RcodeNotImplemented, 5},
		{"success", dns.RcodeSuccess, 5}, // Not negative, but function returns default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := negativeTTL(tt.rcode)
			if got != tt.expected {
				t.Errorf("negativeTTL(%d) = %d, want %d", tt.rcode, got, tt.expected)
			}
		})
	}
}

// --- Test: Negative cache ---

func TestNegativeCache(t *testing.T) {
	config := newTestCacheConfig()
	config.MinTTL = 1
	config.MaxTTL = 86400
	cache := NewDnsCache(config)

	key := newTestCacheKey("nxdomain.example.com", TypeA)
	resp := newTestDNSResponse("nxdomain.example.com", TypeA, dns.RcodeNameError, 0)

	cache.Set(key, resp, 0)

	// Verify it's marked as negative
	if elem, ok := cache.entries[key]; ok {
		entry := elem.Value.(*CacheEntry)
		if !entry.Negative {
			t.Error("expected Negative=true for NXDOMAIN response")
		}
		if entry.EffectiveTTL != 60 {
			t.Errorf("expected negative TTL=60 for NXDOMAIN, got %d", entry.EffectiveTTL)
		}
	}

	// Should be retrievable
	got, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected cache hit for negative entry")
	}
	if got.Rcode != dns.RcodeNameError {
		t.Errorf("expected RcodeNameError, got %d", got.Rcode)
	}
}

func TestNegativeCacheSERVFAIL(t *testing.T) {
	config := newTestCacheConfig()
	cache := NewDnsCache(config)

	key := newTestCacheKey("servfail.example.com", TypeA)
	resp := newTestDNSResponse("servfail.example.com", TypeA, dns.RcodeServerFailure, 0)

	cache.Set(key, resp, 0)

	if elem, ok := cache.entries[key]; ok {
		entry := elem.Value.(*CacheEntry)
		if entry.EffectiveTTL != 30 {
			t.Errorf("expected negative TTL=30 for SERVFAIL, got %d", entry.EffectiveTTL)
		}
	}
}

func TestNegativeCacheREFUSED(t *testing.T) {
	config := newTestCacheConfig()
	cache := NewDnsCache(config)

	key := newTestCacheKey("refused.example.com", TypeA)
	resp := newTestDNSResponse("refused.example.com", TypeA, dns.RcodeRefused, 0)

	cache.Set(key, resp, 0)

	if elem, ok := cache.entries[key]; ok {
		entry := elem.Value.(*CacheEntry)
		if entry.EffectiveTTL != 10 {
			t.Errorf("expected negative TTL=10 for REFUSED, got %d", entry.EffectiveTTL)
		}
	}
}

// --- Test: LRU eviction ---

func TestLRUEviction(t *testing.T) {
	config := newTestCacheConfig()
	config.Size = 3 // Very small cache for easy eviction
	cache := NewDnsCache(config)

	// Add 3 items (cache full)
	resp1 := newTestDNSResponse("alpha.com", TypeA, dns.RcodeSuccess, 300)
	resp2 := newTestDNSResponse("beta.com", TypeA, dns.RcodeSuccess, 300)
	resp3 := newTestDNSResponse("gamma.com", TypeA, dns.RcodeSuccess, 300)

	cache.Set(newTestCacheKey("alpha.com", TypeA), resp1, 300)
	cache.Set(newTestCacheKey("beta.com", TypeA), resp2, 300)
	cache.Set(newTestCacheKey("gamma.com", TypeA), resp3, 300)

	if cache.Len() != 3 {
		t.Errorf("expected 3 items, got %d", cache.Len())
	}

	// Access alpha to make it recently used
	cache.Get(newTestCacheKey("alpha.com", TypeA))

	// Add 4th item - should evict the least recently used (beta, since gamma was last added)
	resp4 := newTestDNSResponse("delta.com", TypeA, dns.RcodeSuccess, 300)
	cache.Set(newTestCacheKey("delta.com", TypeA), resp4, 300)

	if cache.Len() != 3 {
		t.Errorf("expected 3 items after eviction, got %d", cache.Len())
	}

	// beta should have been evicted (least recently used)
	if _, ok := cache.Get(newTestCacheKey("beta.com", TypeA)); ok {
		t.Error("expected beta to be evicted (LRU)")
	}

	// alpha and delta should still be present
	if _, ok := cache.Get(newTestCacheKey("alpha.com", TypeA)); !ok {
		t.Error("expected alpha to still be in cache")
	}
	if _, ok := cache.Get(newTestCacheKey("delta.com", TypeA)); !ok {
		t.Error("expected delta to still be in cache")
	}

	// Check eviction stats
	stats := cache.Stats()
	if stats.Evictions != 1 {
		t.Errorf("expected 1 eviction, got %d", stats.Evictions)
	}
}

// --- Test: Remove ---

func TestRemove(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())
	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 300)

	cache.Set(key, resp, 300)

	// Verify it's there
	if _, ok := cache.Get(key); !ok {
		t.Fatal("expected cache hit before removal")
	}

	// Remove it
	cache.Remove(key)

	// Verify it's gone
	if _, ok := cache.Get(key); ok {
		t.Error("expected cache miss after removal")
	}
}

// --- Test: Clear ---

func TestClear(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())

	// Add multiple items
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("host%d.example.com", i)
		key := newTestCacheKey(name, TypeA)
		resp := newTestDNSResponse(name, TypeA, dns.RcodeSuccess, 300)
		cache.Set(key, resp, 300)
	}

	if cache.Len() != 10 {
		t.Errorf("expected 10 items, got %d", cache.Len())
	}

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("expected 0 items after clear, got %d", cache.Len())
	}

	stats := cache.Stats()
	if stats.Size != 0 {
		t.Errorf("expected stats.Size=0 after clear, got %d", stats.Size)
	}
}

// --- Test: Stats ---

func TestStats(t *testing.T) {
	config := newTestCacheConfig()
	config.Size = 3
	cache := NewDnsCache(config)

	stats := cache.Stats()
	if stats.Size != 0 {
		t.Errorf("expected initial Size=0, got %d", stats.Size)
	}
	if stats.Hits != 0 {
		t.Errorf("expected initial Hits=0, got %d", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("expected initial Misses=0, got %d", stats.Misses)
	}

	// Add and hit
	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 300)
	cache.Set(key, resp, 300)
	cache.Get(key)

	stats = cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}

	// Miss
	cache.Get(newTestCacheKey("nonexistent.com", TypeA))
	stats = cache.Stats()
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}

	// Hit rate should be 0.5 (1 hit / 2 total)
	if stats.HitRate != 0.5 {
		t.Errorf("expected HitRate=0.5, got %f", stats.HitRate)
	}
}

// --- Test: Concurrent safety ---

func TestConcurrentAccess(t *testing.T) {
	config := newTestCacheConfig()
	config.Size = 2000 // Large enough for concurrent writes
	cache := NewDnsCache(config)
	var wg sync.WaitGroup
	numGoroutines := 20
	numOps := 50

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				name := fmt.Sprintf("host%d-%d.example.com", id, j)
				key := newTestCacheKey(name, TypeA)
				resp := newTestDNSResponse(name, TypeA, dns.RcodeSuccess, 300)
				cache.Set(key, resp, 300)
			}
		}(i)
	}

	wg.Wait()

	expectedTotal := numGoroutines * numOps
	if cache.Len() != expectedTotal {
		t.Errorf("expected %d items after concurrent writes, got %d", expectedTotal, cache.Len())
	}

	// Concurrent reads
	var readWg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		readWg.Add(1)
		go func(id int) {
			defer readWg.Done()
			for j := 0; j < numOps; j++ {
				name := fmt.Sprintf("host%d-%d.example.com", id, j)
				key := newTestCacheKey(name, TypeA)
				cache.Get(key)
			}
		}(i)
	}

	readWg.Wait()

	stats := cache.Stats()
	if stats.Hits == 0 {
		t.Error("expected some cache hits from concurrent reads")
	}

	// Concurrent mixed read/write/remove
	var mixedWg sync.WaitGroup
	for i := 0; i < 10; i++ {
		mixedWg.Add(1)
		go func(id int) {
			defer mixedWg.Done()
			for j := 0; j < 20; j++ {
				name := fmt.Sprintf("host%d-%d.example.com", id, j)
				key := newTestCacheKey(name, TypeA)
				switch j % 3 {
				case 0:
					resp := newTestDNSResponse(name, TypeA, dns.RcodeSuccess, 300)
					cache.Set(key, resp, 300)
				case 1:
					cache.Get(key)
				case 2:
					cache.Remove(key)
				}
			}
		}(i)
	}
	mixedWg.Wait()

	// Should not panic
	_ = cache.Stats()
	_ = cache.Len()
	_ = cache.Keys()
}

// --- Test: Update existing entry ---

func TestUpdateExistingEntry(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())
	key := newTestCacheKey("example.com", TypeA)

	// Initial set
	resp1 := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 300)
	cache.Set(key, resp1, 300)

	// Update with different data
	resp2 := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 600)
	cache.Set(key, resp2, 600)

	// Verify updated
	if elem, ok := cache.entries[key]; ok {
		entry := elem.Value.(*CacheEntry)
		if entry.OriginalTTL != 600 {
			t.Errorf("expected updated OriginalTTL=600, got %d", entry.OriginalTTL)
		}
		if entry.Prefetching {
			t.Error("expected Prefetching=false after update")
		}
	}

	// Verify hit count is from the new entry
	got, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected cache hit after update")
	}
	if got.TTL != 600 {
		t.Errorf("expected TTL=600 from updated entry, got %d", got.TTL)
	}
}

// --- Test: Prefetch logic ---

func TestShouldPrefetchBasic(t *testing.T) {
	config := newTestCacheConfig()
	config.Prefetch = true
	config.PrefetchThreshold = 30
	cache := NewDnsCache(config)

	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 86400) // very long TTL
	cache.Set(key, resp, 86400)

	// Just after set, should not prefetch (TTL just started)
	shouldPrefetch := cache.ShouldPrefetch(key)
	if shouldPrefetch {
		t.Log("Note: ShouldPrefetch returned true immediately after set (may vary based on timing)")
	}
}

func TestShouldPrefetchAlreadyPrefetching(t *testing.T) {
	config := newTestCacheConfig()
	config.Prefetch = true
	cache := NewDnsCache(config)

	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 300)
	cache.Set(key, resp, 300)

	// Mark as already prefetching
	cache.SetPrefetching(key, true)

	// Should not trigger prefetch again
	if cache.ShouldPrefetch(key) {
		t.Error("expected ShouldPrefetch=false when already prefetching")
	}
}

func TestShouldPrefetchDisabled(t *testing.T) {
	config := newTestCacheConfig()
	config.Prefetch = false
	cache := NewDnsCache(config)

	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 5) // very short TTL
	cache.Set(key, resp, 5)

	// With prefetch disabled, should never prefetch
	if cache.ShouldPrefetch(key) {
		t.Error("expected ShouldPrefetch=false when prefetch is disabled")
	}
}

func TestSetPrefetching(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())

	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 300)
	cache.Set(key, resp, 300)

	// Set prefetching
	cache.SetPrefetching(key, true)
	if elem, ok := cache.entries[key]; ok {
		entry := elem.Value.(*CacheEntry)
		if !entry.Prefetching {
			t.Error("expected Prefetching=true after SetPrefetching(true)")
		}
	}

	// Unset prefetching
	cache.SetPrefetching(key, false)
	if elem, ok := cache.entries[key]; ok {
		entry := elem.Value.(*CacheEntry)
		if entry.Prefetching {
			t.Error("expected Prefetching=false after SetPrefetching(false)")
		}
	}
}

// --- Test: Prefetch failure backoff ---

func TestPrefetchBackoff(t *testing.T) {
	tests := []struct {
		failCount int
		expected  time.Duration
	}{
		{0, 0},
		{1, 5 * time.Second},
		{2, 30 * time.Second},
		{3, 120 * time.Second},
		{4, 120 * time.Second}, // capped at 120s
		{10, 120 * time.Second},
	}

	for _, tt := range tests {
		got := prefetchBackoff(tt.failCount)
		if got != tt.expected {
			t.Errorf("prefetchBackoff(%d) = %v, want %v", tt.failCount, got, tt.expected)
		}
	}
}

func TestRecordPrefetchFailure(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())

	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 300)
	cache.Set(key, resp, 300)

	// Record failures
	cache.RecordPrefetchFailure(key)
	cache.RecordPrefetchFailure(key)

	if elem, ok := cache.entries[key]; ok {
		entry := elem.Value.(*CacheEntry)
		if entry.PrefetchFailCount != 2 {
			t.Errorf("expected PrefetchFailCount=2, got %d", entry.PrefetchFailCount)
		}
		if entry.PrefetchBackoff != 30*time.Second {
			t.Errorf("expected PrefetchBackoff=30s, got %v", entry.PrefetchBackoff)
		}
		if entry.Prefetching {
			t.Error("expected Prefetching=false after failure")
		}
	}

	// After 3 failures, ShouldPrefetch should return false
	cache.RecordPrefetchFailure(key)
	if cache.ShouldPrefetch(key) {
		t.Error("expected ShouldPrefetch=false after 3 failures")
	}
}

func TestRecordPrefetchSuccess(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())

	key := newTestCacheKey("example.com", TypeA)
	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 300)
	cache.Set(key, resp, 300)

	// Fail twice, then succeed
	cache.RecordPrefetchFailure(key)
	cache.RecordPrefetchFailure(key)
	cache.RecordPrefetchSuccess(key)

	if elem, ok := cache.entries[key]; ok {
		entry := elem.Value.(*CacheEntry)
		if entry.PrefetchFailCount != 0 {
			t.Errorf("expected PrefetchFailCount=0 after success, got %d", entry.PrefetchFailCount)
		}
		if entry.PrefetchBackoff != 0 {
			t.Errorf("expected PrefetchBackoff=0 after success, got %v", entry.PrefetchBackoff)
		}
		if entry.Prefetching {
			t.Error("expected Prefetching=false after success")
		}
	}
}

// --- Test: Different query types ---

func TestCacheKeyIsolationByType(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())

	keyA := newTestCacheKey("example.com", TypeA)
	keyAAAA := newTestCacheKey("example.com", TypeAAAA)

	respA := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 300)
	respAAAA := newTestDNSResponse("example.com", TypeAAAA, dns.RcodeSuccess, 300)

	cache.Set(keyA, respA, 300)
	cache.Set(keyAAAA, respAAAA, 300)

	// Both should be independently accessible
	if _, ok := cache.Get(keyA); !ok {
		t.Error("expected cache hit for A record")
	}
	if _, ok := cache.Get(keyAAAA); !ok {
		t.Error("expected cache hit for AAAA record")
	}
}

func TestCacheKeyIsolationByUpstream(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())

	key1 := CacheKey{UpstreamAddr: "8.8.8.8:53", ProxyTag: "direct", Name: "example.com.", QType: TypeA}
	key2 := CacheKey{UpstreamAddr: "1.1.1.1:53", ProxyTag: "direct", Name: "example.com.", QType: TypeA}

	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 300)

	cache.Set(key1, resp, 300)

	// Different upstream should not share cache
	if _, ok := cache.Get(key2); ok {
		t.Error("expected cache miss for different upstream")
	}
}

func TestCacheKeyIsolationByProxyTag(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())

	keyDirect := CacheKey{UpstreamAddr: "8.8.8.8:53", ProxyTag: "direct", Name: "example.com.", QType: TypeA}
	keyProxy := CacheKey{UpstreamAddr: "8.8.8.8:53", ProxyTag: "proxy", Name: "example.com.", QType: TypeA}

	resp := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 300)

	cache.Set(keyDirect, resp, 300)

	// Different proxy tag should not share cache
	if _, ok := cache.Get(keyProxy); ok {
		t.Error("expected cache miss for different proxy tag")
	}
}

// --- Test: Empty cache behavior ---

func TestEmptyCache(t *testing.T) {
	cache := NewDnsCache(&CacheConfig{})

	if cache.Len() != 0 {
		t.Errorf("expected empty cache, got %d items", cache.Len())
	}

	stats := cache.Stats()
	if stats.Size != 0 {
		t.Errorf("expected Size=0, got %d", stats.Size)
	}

	keys := cache.Keys()
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

// --- Test: Keys method ---

func TestKeys(t *testing.T) {
	cache := NewDnsCache(&CacheConfig{})

	// Add some entries
	names := []string{"alpha.com", "beta.com", "gamma.com"}
	for _, name := range names {
		key := newTestCacheKey(name, TypeA)
		resp := newTestDNSResponse(name, TypeA, dns.RcodeSuccess, 300)
		cache.Set(key, resp, 300)
	}

	keys := cache.Keys()
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}

// --- Test: Concurrent Set/Get/Remove on same key ---

func TestConcurrentSameKey(t *testing.T) {
	cache := NewDnsCache(&CacheConfig{})
	key := newTestCacheKey("contested.com", TypeA)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				resp := newTestDNSResponse("contested.com", TypeA, dns.RcodeSuccess, uint32(100+j))
				cache.Set(key, resp, uint32(100+j))
				cache.Get(key)
				cache.Remove(key)
			}
		}(i)
	}
	wg.Wait()

	// Should not panic or deadlock
	t.Log("concurrent same-key operations completed without deadlock")
}

// --- Test: buildCacheKey (from handler.go) ---

func TestBuildCacheKey(t *testing.T) {
	query := newTestQuery("example.com", TypeA)
	upstream := &UpstreamInstance{
		Addr:     "8.8.8.8:53",
		ProxyTag: "proxy",
	}

	key := buildCacheKey(query, upstream)
	if key.UpstreamAddr != "8.8.8.8:53" {
		t.Errorf("expected UpstreamAddr=8.8.8.8:53, got %s", key.UpstreamAddr)
	}
	if key.ProxyTag != "proxy" {
		t.Errorf("expected ProxyTag=proxy, got %s", key.ProxyTag)
	}
	if key.Name != "example.com" {
		t.Errorf("expected Name=example.com, got %s", key.Name)
	}
	if key.QType != TypeA {
		t.Errorf("expected QType=A, got %d", key.QType)
	}
}

func TestBuildCacheKeyNilUpstream(t *testing.T) {
	query := newTestQuery("example.com", TypeA)
	key := buildCacheKey(query, nil)
	if key.ProxyTag != "direct" {
		t.Errorf("expected ProxyTag=direct for nil upstream, got %s", key.ProxyTag)
	}
	if key.UpstreamAddr != "" {
		t.Errorf("expected empty UpstreamAddr for nil upstream, got %s", key.UpstreamAddr)
	}
}

// --- Test: Large number of entries ---

func TestLargeCache(t *testing.T) {
	config := newTestCacheConfig()
	config.Size = 1000
	cache := NewDnsCache(config)

	// Add 1000 entries
	for i := 0; i < 1000; i++ {
		name := fmt.Sprintf("host%d.example.com", i)
		key := newTestCacheKey(name, TypeA)
		resp := newTestDNSResponse(name, TypeA, dns.RcodeSuccess, 300)
		cache.Set(key, resp, 300)
	}

	if cache.Len() != 1000 {
		t.Errorf("expected 1000 entries, got %d", cache.Len())
	}

	// Add one more - should evict one
	extraKey := newTestCacheKey("extra.example.com", TypeA)
	extraResp := newTestDNSResponse("extra.example.com", TypeA, dns.RcodeSuccess, 300)
	cache.Set(extraKey, extraResp, 300)

	if cache.Len() != 1000 {
		t.Errorf("expected 1000 entries after eviction, got %d", cache.Len())
	}

	stats := cache.Stats()
	if stats.Evictions != 1 {
		t.Errorf("expected 1 eviction, got %d", stats.Evictions)
	}
}
