package dns

import (
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// Helper: create a basic DnsQuery for testing.
func testQuery(name string, qtype QueryType) *DnsQuery {
	return &DnsQuery{
		Name:   name,
		QType:  qtype,
		QClass: dns.ClassINET,
	}
}

// Helper: create a basic DnsResponse for testing.
func testResponse(query *DnsQuery, rtt time.Duration, upstream string) *DnsResponse {
	return &DnsResponse{
		Query:    *query,
		Rcode:    dns.RcodeSuccess,
		Upstream: upstream,
		ProxyTag: "direct",
		RTT:      rtt,
		Cached:   false,
		TTL:      300,
	}
}

// TestNewDnsMetrics verifies that NewDnsMetrics initializes correctly.
func TestNewDnsMetrics(t *testing.T) {
	m := NewDnsMetrics()
	if m == nil {
		t.Fatal("NewDnsMetrics() returned nil")
	}

	// Verify latency buckets are initialized.
	if len(m.LatencyBuckets) != 6 {
		t.Fatalf("expected 6 latency buckets, got %d", len(m.LatencyBuckets))
	}

	// Verify all counters start at zero.
	snap := m.Snapshot()
	if snap.TotalQueries != 0 || snap.TotalResponses != 0 || snap.TotalErrors != 0 {
		t.Fatal("expected all counters to be zero")
	}
}

// TestRecordQueryAndResponse tests basic query and response recording.
func TestRecordQueryAndResponse(t *testing.T) {
	m := NewDnsMetrics()

	// Record a query.
	q := testQuery("example.com", TypeA)
	m.RecordQuery(q, "upstream-1")

	// Record a response.
	resp := testResponse(q, 10*time.Millisecond, "upstream-1")
	m.RecordResponse(resp, 10*time.Millisecond)

	// Snapshot and verify.
	snap := m.Snapshot()
	if snap.TotalQueries != 1 {
		t.Errorf("expected 1 query, got %d", snap.TotalQueries)
	}
	if snap.TotalResponses != 1 {
		t.Errorf("expected 1 response, got %d", snap.TotalResponses)
	}

	// Verify query type breakdown.
	if snap.QueriesByType["A"] != 1 {
		t.Errorf("expected 1 A query, got %d", snap.QueriesByType["A"])
	}

	// Verify upstream stats.
	um, ok := snap.QueriesByUpstream["upstream-1"]
	if !ok {
		t.Fatal("expected upstream-1 stats")
	}
	if um.Queries != 1 {
		t.Errorf("expected 1 query for upstream-1, got %d", um.Queries)
	}
	if um.Successes != 1 {
		t.Errorf("expected 1 success for upstream-1, got %d", um.Successes)
	}
	if um.AvgRTT != 10*time.Millisecond {
		t.Errorf("expected AvgRTT=10ms, got %v", um.AvgRTT)
	}
}

// TestCacheHitMiss tests cache hit/miss counting.
func TestCacheHitMiss(t *testing.T) {
	m := NewDnsMetrics()

	m.RecordCacheHit()
	m.RecordCacheHit()
	m.RecordCacheHit()
	m.RecordCacheMiss()

	snap := m.Snapshot()
	if snap.CacheHits != 3 {
		t.Errorf("expected 3 cache hits, got %d", snap.CacheHits)
	}
	if snap.CacheMisses != 1 {
		t.Errorf("expected 1 cache miss, got %d", snap.CacheMisses)
	}
}

// TestCacheEvictionAndPrefetch tests eviction and prefetch counting.
func TestCacheEvictionAndPrefetch(t *testing.T) {
	m := NewDnsMetrics()

	m.RecordCacheEviction()
	m.RecordCacheEviction()
	m.RecordCacheEviction()
	m.RecordPrefetch()

	snap := m.Snapshot()
	if snap.CacheEvictions != 3 {
		t.Errorf("expected 3 evictions, got %d", snap.CacheEvictions)
	}
	if snap.PrefetchCount != 1 {
		t.Errorf("expected 1 prefetch, got %d", snap.PrefetchCount)
	}
}

// TestUpstreamStats tests per-upstream statistics including RTT tracking.
func TestUpstreamStats(t *testing.T) {
	m := NewDnsMetrics()

	upstream := "8.8.8.8:53"

	// Record queries.
	for i := 0; i < 5; i++ {
		q := testQuery("test.com", TypeA)
		m.RecordQuery(q, upstream)
	}

	// Record responses with different RTTs.
	rtts := []time.Duration{5, 10, 15, 20, 25}
	for _, rtt := range rtts {
		q := testQuery("test.com", TypeA)
		resp := testResponse(q, rtt*time.Millisecond, upstream)
		m.RecordResponse(resp, rtt*time.Millisecond)
	}

	// Also record some errors.
	m.RecordError(upstream, dns.ErrRcode)
	m.RecordError(upstream, dns.ErrRcode)

	snap := m.Snapshot()

	// Verify upstream stats.
	um, ok := snap.QueriesByUpstream[upstream]
	if !ok {
		t.Fatal("expected upstream stats")
	}
	if um.Queries != 5 {
		t.Errorf("expected 5 queries, got %d", um.Queries)
	}
	if um.Successes != 5 {
		t.Errorf("expected 5 successes, got %d", um.Successes)
	}
	if um.Errors != 2 {
		t.Errorf("expected 2 errors, got %d", um.Errors)
	}
	if um.AvgRTT == 0 {
		t.Error("expected non-zero AvgRTT")
	}
	if um.MinRTT != 5*time.Millisecond {
		t.Errorf("expected MinRTT=5ms, got %v", um.MinRTT)
	}
	if um.MaxRTT != 25*time.Millisecond {
		t.Errorf("expected MaxRTT=25ms, got %v", um.MaxRTT)
	}

	// Verify errors by upstream.
	if snap.ErrorsByUpstream[upstream] != 2 {
		t.Errorf("expected 2 errors for upstream, got %d", snap.ErrorsByUpstream[upstream])
	}
}

// TestLatencyDistribution tests latency bucket recording.
func TestLatencyDistribution(t *testing.T) {
	m := NewDnsMetrics()

	// Record responses with various latencies.
	rtts := []time.Duration{
		500 * time.Microsecond,  // <1ms
		3 * time.Millisecond,    // 1-5ms
		7 * time.Millisecond,    // 5-10ms
		30 * time.Millisecond,   // 10-50ms
		100 * time.Millisecond,  // 50-200ms
		500 * time.Millisecond,  // 200ms+
	}

	upstream := "latency-upstream"
	for _, rtt := range rtts {
		q := testQuery("test.com", TypeA)
		resp := testResponse(q, rtt, upstream)
		m.RecordResponse(resp, rtt)
	}

	snap := m.Snapshot()
	if len(snap.LatencyBuckets) != 6 {
		t.Fatalf("expected 6 latency buckets, got %d", len(snap.LatencyBuckets))
	}

	// Verify each bucket has exactly 1 entry.
	expected := map[string]int64{
		"<1ms":   1,
		"1-5ms":  1,
		"5-10ms": 1,
		"10-50ms": 1,
		"50-200ms": 1,
		"200ms+": 1,
	}
	for _, lb := range snap.LatencyBuckets {
		expect, ok := expected[lb.Label]
		if !ok {
			t.Errorf("unexpected bucket label: %s", lb.Label)
			continue
		}
		if lb.Count != expect {
			t.Errorf("bucket %s: expected %d, got %d", lb.Label, expect, lb.Count)
		}
	}
}

// TestRuleHit tests rule hit counting.
func TestRuleHit(t *testing.T) {
	m := NewDnsMetrics()

	m.RecordRuleHit("rule-1")
	m.RecordRuleHit("rule-1")
	m.RecordRuleHit("rule-2")
	m.RecordRuleHit("") // Should be ignored.

	snap := m.Snapshot()
	if snap.HitsByRule["rule-1"] != 2 {
		t.Errorf("expected 2 hits for rule-1, got %d", snap.HitsByRule["rule-1"])
	}
	if snap.HitsByRule["rule-2"] != 1 {
		t.Errorf("expected 1 hit for rule-2, got %d", snap.HitsByRule["rule-2"])
	}
	if len(snap.HitsByRule) != 2 {
		t.Errorf("expected 2 rules in map, got %d", len(snap.HitsByRule))
	}
}

// TestActionRouting tests action routing counting.
func TestActionRouting(t *testing.T) {
	m := NewDnsMetrics()

	m.RecordAction("route")
	m.RecordAction("route")
	m.RecordAction("reject")
	m.RecordAction("bypass")
	m.RecordAction("bypass")
	m.RecordAction("bypass")

	snap := m.Snapshot()
	if snap.RoutedQueries != 2 {
		t.Errorf("expected 2 routed, got %d", snap.RoutedQueries)
	}
	if snap.RejectedQueries != 1 {
		t.Errorf("expected 1 rejected, got %d", snap.RejectedQueries)
	}
	if snap.BypassedQueries != 3 {
		t.Errorf("expected 3 bypassed, got %d", snap.BypassedQueries)
	}
}

// TestConcurrentSafety tests that metrics are safe for concurrent access.
func TestConcurrentSafety(t *testing.T) {
	m := NewDnsMetrics()

	var wg sync.WaitGroup
	n := 100

	// Concurrently record queries from multiple goroutines.
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			q := testQuery("concurrent.test", TypeA)
			m.RecordQuery(q, "upstream-concurrent")
			m.RecordCacheHit()
			m.RecordAction("route")
			m.RecordRuleHit("rule-concurrent")
		}(i)
	}

	wg.Wait()

	snap := m.Snapshot()
	if snap.TotalQueries != int64(n) {
		t.Errorf("expected %d queries, got %d", n, snap.TotalQueries)
	}
	if snap.CacheHits != int64(n) {
		t.Errorf("expected %d cache hits, got %d", n, snap.CacheHits)
	}
	if snap.RoutedQueries != int64(n) {
		t.Errorf("expected %d routed, got %d", n, snap.RoutedQueries)
	}
	if snap.HitsByRule["rule-concurrent"] != int64(n) {
		t.Errorf("expected %d rule hits, got %d", n, snap.HitsByRule["rule-concurrent"])
	}
}

// TestSnapshotConsistency tests that Snapshot returns consistent data.
func TestSnapshotConsistency(t *testing.T) {
	m := NewDnsMetrics()

	// Record some data.
	q := testQuery("snapshot.test", TypeAAAA)
	m.RecordQuery(q, "upstream-snap")
	resp := testResponse(q, 50*time.Millisecond, "upstream-snap")
	m.RecordResponse(resp, 50*time.Millisecond)
	m.RecordError("upstream-snap", dns.ErrRcode)
	m.RecordCacheHit()
	m.RecordCacheMiss()
	m.RecordRuleHit("rule-snap")
	m.RecordAction("route")

	// Take snapshot.
	snap := m.Snapshot()

	// Verify total consistency.
	if snap.TotalQueries != snap.QueriesByUpstream["upstream-snap"].Queries {
		t.Errorf("total queries (%d) != upstream queries (%d)",
			snap.TotalQueries, snap.QueriesByUpstream["upstream-snap"].Queries)
	}
	if snap.TotalErrors != snap.ErrorsByUpstream["upstream-snap"] {
		t.Errorf("total errors (%d) != upstream errors (%d)",
			snap.TotalErrors, snap.ErrorsByUpstream["upstream-snap"])
	}

	// Verify query type exists.
	if snap.QueriesByType["AAAA"] != 1 {
		t.Errorf("expected 1 AAAA query, got %d", snap.QueriesByType["AAAA"])
	}
}

// TestReset tests that Reset clears all metrics.
func TestReset(t *testing.T) {
	m := NewDnsMetrics()

	// Record some data.
	q := testQuery("reset.test", TypeA)
	m.RecordQuery(q, "upstream-reset")
	resp := testResponse(q, 10*time.Millisecond, "upstream-reset")
	m.RecordResponse(resp, 10*time.Millisecond)
	m.RecordCacheHit()
	m.RecordCacheMiss()
	m.RecordError("upstream-reset", dns.ErrRcode)
	m.RecordRuleHit("rule-reset")
	m.RecordAction("route")

	// Reset.
	m.Reset()

	// Verify all counters are zero.
	snap := m.Snapshot()
	if snap.TotalQueries != 0 {
		t.Errorf("expected 0 queries after reset, got %d", snap.TotalQueries)
	}
	if snap.TotalResponses != 0 {
		t.Errorf("expected 0 responses after reset, got %d", snap.TotalResponses)
	}
	if snap.TotalErrors != 0 {
		t.Errorf("expected 0 errors after reset, got %d", snap.TotalErrors)
	}
	if snap.CacheHits != 0 {
		t.Errorf("expected 0 cache hits after reset, got %d", snap.CacheHits)
	}
	if snap.CacheMisses != 0 {
		t.Errorf("expected 0 cache misses after reset, got %d", snap.CacheMisses)
	}
	if snap.RoutedQueries != 0 {
		t.Errorf("expected 0 routed after reset, got %d", snap.RoutedQueries)
	}
	if len(snap.QueriesByUpstream) != 0 {
		t.Errorf("expected empty upstream map after reset, got %d entries", len(snap.QueriesByUpstream))
	}
	if len(snap.HitsByRule) != 0 {
		t.Errorf("expected empty rule hits after reset, got %d entries", len(snap.HitsByRule))
	}
}

// TestTimeoutRecording tests timeout recording.
func TestTimeoutRecording(t *testing.T) {
	m := NewDnsMetrics()

	m.RecordTimeout("upstream-timeout")
	m.RecordTimeout("upstream-timeout")

	snap := m.Snapshot()
	if snap.TotalTimeouts != 2 {
		t.Errorf("expected 2 timeouts, got %d", snap.TotalTimeouts)
	}
	if snap.TotalErrors != 2 {
		t.Errorf("expected 2 errors (timeout counted as error), got %d", snap.TotalErrors)
	}
}

// TestMultipleQTypes tests recording of multiple query types.
func TestMultipleQTypes(t *testing.T) {
	m := NewDnsMetrics()

	upstream := "upstream-multi"
	types := []QueryType{TypeA, TypeAAAA, TypeCNAME, TypeTXT, TypeMX}
	for i, qt := range types {
		q := testQuery("multi.test", qt)
		m.RecordQuery(q, upstream)

		resp := testResponse(q, time.Duration(i+1)*time.Millisecond, upstream)
		m.RecordResponse(resp, time.Duration(i+1)*time.Millisecond)
	}

	snap := m.Snapshot()
	if len(snap.QueriesByType) != 5 {
		t.Errorf("expected 5 query types, got %d", len(snap.QueriesByType))
	}
	for _, qt := range types {
		name := QTypeToString(qt)
		if snap.QueriesByType[name] != 1 {
			t.Errorf("expected 1 query of type %s, got %d", name, snap.QueriesByType[name])
		}
	}
}

// TestUpstreamMetricsConcurrent tests concurrent access to upstream metrics.
func TestUpstreamMetricsConcurrent(t *testing.T) {
	m := NewDnsMetrics()

	var wg sync.WaitGroup
	n := 50

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			upstream := "concurrent-upstream"
			q := testQuery("concurrent.test", TypeA)
			m.RecordQuery(q, upstream)
			resp := testResponse(q, time.Duration(i)*time.Millisecond, upstream)
			m.RecordResponse(resp, time.Duration(i)*time.Millisecond)
		}(i)
	}

	wg.Wait()

	snap := m.Snapshot()
	um, ok := snap.QueriesByUpstream["concurrent-upstream"]
	if !ok {
		t.Fatal("expected upstream metrics")
	}
	if um.Queries != int64(n) {
		t.Errorf("expected %d queries, got %d", n, um.Queries)
	}
	if um.Successes != int64(n) {
		t.Errorf("expected %d successes, got %d", n, um.Successes)
	}
}

// TestUptimeAfterReset tests that uptime is reset after Reset.
func TestUptimeAfterReset(t *testing.T) {
	m := NewDnsMetrics()

	// Sleep enough for Round(time.Second) to show a difference.
	time.Sleep(1100 * time.Millisecond)

	snap1 := m.Snapshot()
	t.Logf("Uptime before reset: %s", snap1.Uptime)

	m.Reset()

	snap2 := m.Snapshot()
	if snap2.Uptime == snap1.Uptime {
		t.Error("expected uptime to change after reset")
	}
	t.Logf("Uptime after reset: %s", snap2.Uptime)
}
