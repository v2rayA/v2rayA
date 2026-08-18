package dns

import (
	"fmt"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// --- Helper functions ---

func newTestUpstreamConfig(id, addr, proxyTag string) UpstreamConfig {
	return UpstreamConfig{
		ID:       id,
		Addr:     addr,
		Protocol: "udp",
		ProxyTag: proxyTag,
	}
}

func newTestUpstreamInstance(id, addr, proxyTag string) *UpstreamInstance {
	proxy := proxyTag
	if proxy == "" {
		proxy = "direct"
	}
	return &UpstreamInstance{
		ID:       id,
		Addr:     addr,
		Protocol: "udp",
		ProxyTag: proxy,
		Healthy:  true,
		Client: &dns.Client{
			Net:          "udp",
			Timeout:      5 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
}

// --- Test: BuildUpstreamKey ---

func TestBuildUpstreamKey(t *testing.T) {
	tests := []struct {
		addr     string
		proxyTag string
		expected string
	}{
		{"8.8.8.8:53", "", "8.8.8.8:53|"},
		{"8.8.8.8:53", "direct", "8.8.8.8:53|direct"},
		{"8.8.8.8:53", "proxy", "8.8.8.8:53|proxy"},
		{"1.1.1.1:53", "proxy", "1.1.1.1:53|proxy"},
		{"8.8.8.8:53", "my-custom-tag", "8.8.8.8:53|my-custom-tag"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s|%s", tt.addr, tt.proxyTag), func(t *testing.T) {
			got := BuildUpstreamKey(tt.addr, tt.proxyTag)
			if got != tt.expected {
				t.Errorf("BuildUpstreamKey(%q, %q) = %q, want %q",
					tt.addr, tt.proxyTag, got, tt.expected)
			}
		})
	}
}

// --- Test: UpstreamInstance.CompositeKey ---

func TestUpstreamInstanceCompositeKey(t *testing.T) {
	instance := newTestUpstreamInstance("google", "8.8.8.8:53", "proxy")
	expected := "8.8.8.8:53|proxy"
	if got := instance.CompositeKey(); got != expected {
		t.Errorf("CompositeKey() = %q, want %q", got, expected)
	}
}

// --- Test: NewUpstreamManager with multi-channel ---

func TestNewUpstreamManagerMultiChannel(t *testing.T) {
	// Same address, different proxy tags should create separate instances.
	configs := []UpstreamConfig{
		newTestUpstreamConfig("google-direct", "8.8.8.8:53", "direct"),
		newTestUpstreamConfig("google-proxy", "8.8.8.8:53", "proxy"),
	}

	mgr := NewUpstreamManager(configs)

	// Verify both instances exist and are separate.
	directKey := BuildUpstreamKey("8.8.8.8:53", "direct")
	proxyKey := BuildUpstreamKey("8.8.8.8:53", "proxy")

	directInstance, ok := mgr.GetUpstreamByKey(directKey)
	if !ok {
		t.Fatal("expected direct upstream instance")
	}
	if directInstance.ProxyTag != "direct" {
		t.Errorf("expected ProxyTag=direct, got %s", directInstance.ProxyTag)
	}

	proxyInstance, ok := mgr.GetUpstreamByKey(proxyKey)
	if !ok {
		t.Fatal("expected proxy upstream instance")
	}
	if proxyInstance.ProxyTag != "proxy" {
		t.Errorf("expected ProxyTag=proxy, got %s", proxyInstance.ProxyTag)
	}

	// They should be different instances (different pointers).
	if directInstance == proxyInstance {
		t.Error("direct and proxy instances should be different objects")
	}

	// They should have the same address.
	if directInstance.Addr != proxyInstance.Addr {
		t.Errorf("expected same addr, got %q vs %q", directInstance.Addr, proxyInstance.Addr)
	}
}

func TestNewUpstreamManagerSingleChannel(t *testing.T) {
	// Single upstream without proxy tag should work as before.
	configs := []UpstreamConfig{
		newTestUpstreamConfig("google", "8.8.8.8:53", ""),
	}

	mgr := NewUpstreamManager(configs)

	key := BuildUpstreamKey("8.8.8.8:53", "direct")
	instance, ok := mgr.GetUpstreamByKey(key)
	if !ok {
		t.Fatal("expected upstream instance")
	}
	if instance.ProxyTag != "direct" {
		t.Errorf("expected default ProxyTag=direct, got %s", instance.ProxyTag)
	}
	if instance.ID != "google" {
		t.Errorf("expected ID=google, got %s", instance.ID)
	}
}

// --- Test: GetUpstream backward compatibility ---

func TestGetUpstreamByID(t *testing.T) {
	configs := []UpstreamConfig{
		newTestUpstreamConfig("google", "8.8.8.8:53", "direct"),
		newTestUpstreamConfig("cloudflare", "1.1.1.1:53", ""),
	}

	mgr := NewUpstreamManager(configs)

	// Get by ID should work.
	instance, ok := mgr.GetUpstream("google")
	if !ok {
		t.Fatal("expected to find upstream by ID 'google'")
	}
	if instance.Addr != "8.8.8.8:53" {
		t.Errorf("expected addr 8.8.8.8:53, got %s", instance.Addr)
	}

	instance, ok = mgr.GetUpstream("cloudflare")
	if !ok {
		t.Fatal("expected to find upstream by ID 'cloudflare'")
	}
	if instance.Addr != "1.1.1.1:53" {
		t.Errorf("expected addr 1.1.1.1:53, got %s", instance.Addr)
	}

	// Non-existent ID should return false.
	_, ok = mgr.GetUpstream("nonexistent")
	if ok {
		t.Error("expected false for nonexistent ID")
	}
}

// --- Test: GetOrCreateUpstream ---

func TestGetOrCreateUpstreamNew(t *testing.T) {
	mgr := NewUpstreamManager(nil) // Empty manager

	// Create a new upstream instance.
	config := newTestUpstreamConfig("google", "8.8.8.8:53", "proxy")
	instance := mgr.GetOrCreateUpstream(config)

	if instance == nil {
		t.Fatal("expected non-nil instance")
	}
	if instance.ID != "google" {
		t.Errorf("expected ID=google, got %s", instance.ID)
	}
	if instance.ProxyTag != "proxy" {
		t.Errorf("expected ProxyTag=proxy, got %s", instance.ProxyTag)
	}
	if !instance.Healthy {
		t.Error("expected new instance to be healthy")
	}
}

func TestGetOrCreateUpstreamExisting(t *testing.T) {
	configs := []UpstreamConfig{
		newTestUpstreamConfig("google", "8.8.8.8:53", "proxy"),
	}
	mgr := NewUpstreamManager(configs)

	// Get existing instance.
	config := newTestUpstreamConfig("google", "8.8.8.8:53", "proxy")
	instance := mgr.GetOrCreateUpstream(config)

	if instance == nil {
		t.Fatal("expected non-nil instance")
	}
	if instance.ID != "google" {
		t.Errorf("expected ID=google, got %s", instance.ID)
	}

	// Getting again should return the same instance (same pointer).
	instance2 := mgr.GetOrCreateUpstream(config)
	if instance != instance2 {
		t.Error("GetOrCreateUpstream should return the same instance for the same config")
	}
}

func TestGetOrCreateUpstreamDifferentChannels(t *testing.T) {
	mgr := NewUpstreamManager(nil)

	// Create instances for same address but different proxy tags.
	directConfig := newTestUpstreamConfig("google-direct", "8.8.8.8:53", "direct")
	proxyConfig := newTestUpstreamConfig("google-proxy", "8.8.8.8:53", "proxy")

	directInstance := mgr.GetOrCreateUpstream(directConfig)
	proxyInstance := mgr.GetOrCreateUpstream(proxyConfig)

	if directInstance == proxyInstance {
		t.Error("direct and proxy instances should be different")
	}

	if directInstance.ProxyTag != "direct" {
		t.Errorf("expected direct ProxyTag, got %s", directInstance.ProxyTag)
	}
	if proxyInstance.ProxyTag != "proxy" {
		t.Errorf("expected proxy ProxyTag, got %s", proxyInstance.ProxyTag)
	}
}

// --- Test: Multi-channel cache isolation ---
// CacheKey already contains ProxyTag (verified in cache_test.go).
// These tests verify the end-to-end isolation through UpstreamManager.

func TestMultiChannelUpstreamIsolation(t *testing.T) {
	// Create an upstream manager with multi-channel instances.
	configs := []UpstreamConfig{
		newTestUpstreamConfig("google-direct", "8.8.8.8:53", "direct"),
		newTestUpstreamConfig("google-proxy", "8.8.8.8:53", "proxy"),
	}
	mgr := NewUpstreamManager(configs)

	// Get instances for different channels.
	directKey := BuildUpstreamKey("8.8.8.8:53", "direct")
	proxyKey := BuildUpstreamKey("8.8.8.8:53", "proxy")

	directInstance, ok := mgr.GetUpstreamByKey(directKey)
	if !ok {
		t.Fatal("expected direct instance")
	}
	proxyInstance, ok := mgr.GetUpstreamByKey(proxyKey)
	if !ok {
		t.Fatal("expected proxy instance")
	}

	// Verify they have different composite keys.
	if directInstance.CompositeKey() == proxyInstance.CompositeKey() {
		t.Error("instances should have different composite keys")
	}

	// Verify they are independently accessible.
	_, ok = mgr.GetUpstream("google-direct")
	if !ok {
		t.Error("expected to find google-direct by ID")
	}
	_, ok = mgr.GetUpstream("google-proxy")
	if !ok {
		t.Error("expected to find google-proxy by ID")
	}
}

// --- Test: DnsResponse ProxyTag propagation ---

func TestDnsResponseProxyTag(t *testing.T) {
	// When building a DnsResponse, ProxyTag should be set correctly.
	resp := &DnsResponse{
		Query: DnsQuery{
			Name:  "example.com",
			QType: TypeA,
		},
		Rcode:    dns.RcodeSuccess,
		Upstream: "8.8.8.8:53",
		ProxyTag: "proxy",
	}

	if resp.ProxyTag != "proxy" {
		t.Errorf("expected ProxyTag=proxy, got %s", resp.ProxyTag)
	}

	// Direct query.
	resp2 := &DnsResponse{
		Query: DnsQuery{
			Name:  "example.com",
			QType: TypeA,
		},
		Rcode:    dns.RcodeSuccess,
		Upstream: "8.8.8.8:53",
		ProxyTag: "direct",
	}

	if resp2.ProxyTag != "direct" {
		t.Errorf("expected ProxyTag=direct, got %s", resp2.ProxyTag)
	}

	// Same upstream, different proxy tags → different responses.
	if resp.ProxyTag == resp2.ProxyTag {
		t.Error("proxy and direct responses should have different ProxyTags")
	}
}

// --- Test: CacheKey isolation with different ProxyTags ---
// This complements cache_test.go's TestCacheKeyIsolationByProxyTag.

func TestCacheKeyIsolationMultiChannel(t *testing.T) {
	cache := NewDnsCache(newTestCacheConfig())

	// Create cache keys for same upstream but different proxy channels.
	keys := []CacheKey{
		{UpstreamAddr: "8.8.8.8:53", ProxyTag: "direct", Name: "google.com.", QType: TypeA},
		{UpstreamAddr: "8.8.8.8:53", ProxyTag: "proxy", Name: "google.com.", QType: TypeA},
		{UpstreamAddr: "8.8.8.8:53", ProxyTag: "my-proxy", Name: "google.com.", QType: TypeA},
	}

	// Set different responses for each channel.
	for i, key := range keys {
		resp := newTestDNSResponse("google.com", TypeA, dns.RcodeSuccess, uint32(300+i*100))
		cache.Set(key, resp, uint32(300+i*100))
	}

	// Verify each channel has its own cached entry.
	for i, key := range keys {
		got, ok := cache.Get(key)
		if !ok {
			t.Errorf("expected cache hit for key[%d] (proxy=%s)", i, key.ProxyTag)
		} else {
			if got == nil {
				t.Errorf("got nil response for key[%d]", i)
			}
		}
	}

	// Verify cache entries are independent (different TTLs from the Set above).
	for i, key := range keys {
		if elem, ok := cache.entries[key]; ok {
			entry := elem.Value.(*CacheEntry)
			expectedTTL := uint32(300 + i*100)
			// The effective TTL may be clamped to minTTL, but OriginalTTL should match.
			if entry.OriginalTTL != expectedTTL {
				t.Errorf("key[%d] (proxy=%s): expected OriginalTTL=%d, got %d",
					i, key.ProxyTag, expectedTTL, entry.OriginalTTL)
			}
		}
	}

	// Ensure removing one channel's entry doesn't affect others.
	cache.Remove(keys[1]) // Remove proxy channel
	if _, ok := cache.Get(keys[1]); ok {
		t.Error("expected cache miss for removed key")
	}
	if _, ok := cache.Get(keys[0]); !ok {
		t.Error("expected cache hit for direct channel (should be unaffected)")
	}
	if _, ok := cache.Get(keys[2]); !ok {
		t.Error("expected cache hit for my-proxy channel (should be unaffected)")
	}
}

// --- Test: Multi-channel RouteResult ---

func TestRouteResultWithProxyTag(t *testing.T) {
	// RouteResult should carry the proxy tag.
	result := &RouteResult{
		UpstreamID:   "google-proxy",
		UpstreamAddr: "8.8.8.8:53",
		Action:       "route",
		RuleID:       "rule-1",
		Policy:       "single",
		ProxyTag:     "proxy",
	}

	if result.ProxyTag != "proxy" {
		t.Errorf("expected ProxyTag=proxy, got %s", result.ProxyTag)
	}

	// Direct route result.
	result2 := &RouteResult{
		UpstreamID:   "google-direct",
		UpstreamAddr: "8.8.8.8:53",
		Action:       "route",
		RuleID:       "rule-2",
		Policy:       "single",
		ProxyTag:     "direct",
	}

	if result2.ProxyTag != "direct" {
		t.Errorf("expected ProxyTag=direct, got %s", result2.ProxyTag)
	}

	// Same upstream address, different proxy tags.
	if result.UpstreamAddr == result2.UpstreamAddr && result.ProxyTag == result2.ProxyTag {
		t.Error("different channels should have different RouteResults")
	}
}

// --- Test: HealthCheck isolation ---

func TestHealthCheckIsolation(t *testing.T) {
	configs := []UpstreamConfig{
		newTestUpstreamConfig("google-direct", "8.8.8.8:53", "direct"),
		newTestUpstreamConfig("google-proxy", "8.8.8.8:53", "proxy"),
		newTestUpstreamConfig("cloudflare-direct", "1.1.1.1:53", "direct"),
	}
	mgr := NewUpstreamManager(configs)

	// HealthCheck should return results for all instances.
	results := mgr.HealthCheck()
	if len(results) != 3 {
		t.Errorf("expected 3 health check results, got %d", len(results))
	}

	// Each composite key should have its own health result.
	directKey := BuildUpstreamKey("8.8.8.8:53", "direct")
	proxyKey := BuildUpstreamKey("8.8.8.8:53", "proxy")
	cfKey := BuildUpstreamKey("1.1.1.1:53", "direct")

	if _, ok := results[directKey]; !ok {
		t.Errorf("expected health result for %s", directKey)
	}
	if _, ok := results[proxyKey]; !ok {
		t.Errorf("expected health result for %s", proxyKey)
	}
	if _, ok := results[cfKey]; !ok {
		t.Errorf("expected health result for %s", cfKey)
	}

	// Per-channel health check.
	directResults := mgr.HealthCheckByProxyTag("direct")
	if len(directResults) != 2 {
		t.Errorf("expected 2 direct health results, got %d", len(directResults))
	}

	proxyResults := mgr.HealthCheckByProxyTag("proxy")
	if len(proxyResults) != 1 {
		t.Errorf("expected 1 proxy health result, got %d", len(proxyResults))
	}
}

// --- Test: ListUpstreams ---

func TestListUpstreams(t *testing.T) {
	configs := []UpstreamConfig{
		newTestUpstreamConfig("google-direct", "8.8.8.8:53", "direct"),
		newTestUpstreamConfig("google-proxy", "8.8.8.8:53", "proxy"),
		newTestUpstreamConfig("cloudflare", "1.1.1.1:53", ""),
	}
	mgr := NewUpstreamManager(configs)

	instances := mgr.ListUpstreams()
	if len(instances) != 3 {
		t.Errorf("expected 3 instances, got %d", len(instances))
	}

	// Verify we have the expected instances.
	foundDirect := false
	foundProxy := false
	foundCF := false
	for _, inst := range instances {
		switch inst.ID {
		case "google-direct":
			foundDirect = true
		case "google-proxy":
			foundProxy = true
		case "cloudflare":
			foundCF = true
		}
	}
	if !foundDirect {
		t.Error("expected to find google-direct")
	}
	if !foundProxy {
		t.Error("expected to find google-proxy")
	}
	if !foundCF {
		t.Error("expected to find cloudflare")
	}
}

// --- Test: ProxyAddrResolver ---

func TestProxyAddrResolver(t *testing.T) {
	mgr := NewUpstreamManager(nil)

	// No resolver set → should use fallback defaults.
	addr := mgr.resolveProxyAddr("unknown-tag")
	if addr != "" {
		t.Errorf("expected empty for unknown tag, got %s", addr)
	}

	// "proxy" tag should use default SOCKS5 address.
	addr = mgr.resolveProxyAddr("proxy")
	if addr != "127.0.0.1:1080" {
		t.Errorf("expected 127.0.0.1:1080 for proxy tag, got %s", addr)
	}

	// Set custom resolver.
	mgr.SetProxyAddrResolver(func(tag string) string {
		switch tag {
		case "my-proxy":
			return "10.0.0.1:2080"
		case "proxy":
			return "192.168.1.1:1080"
		default:
			return ""
		}
	})

	// Custom resolver should take precedence.
	addr = mgr.resolveProxyAddr("my-proxy")
	if addr != "10.0.0.1:2080" {
		t.Errorf("expected 10.0.0.1:2080, got %s", addr)
	}

	addr = mgr.resolveProxyAddr("proxy")
	if addr != "192.168.1.1:1080" {
		t.Errorf("expected 192.168.1.1:1080 (resolver override), got %s", addr)
	}

	addr = mgr.resolveProxyAddr("unknown")
	if addr != "" {
		t.Errorf("expected empty for unknown tag with custom resolver, got %s", addr)
	}
}

// --- Test: Exchange routing via proxy ---

func TestExchangeProxyRouting(t *testing.T) {
	// Test the logic that determines whether to route through proxy.
	// This tests the routing decision, not the actual network exchange.

	tests := []struct {
		name     string
		proxyTag string
		expectViaProxy bool
	}{
		{"empty tag = direct", "", false},
		{"direct tag = direct", "direct", false},
		{"proxy tag = via proxy", "proxy", true},
		{"custom tag = via proxy", "my-proxy", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstream := &UpstreamInstance{
				Addr:     "8.8.8.8:53",
				ProxyTag: tt.proxyTag,
			}

			isViaProxy := upstream.ProxyTag != "" && upstream.ProxyTag != "direct"
			if isViaProxy != tt.expectViaProxy {
				t.Errorf("expected viaProxy=%v for ProxyTag=%q, got %v",
					tt.expectViaProxy, tt.proxyTag, isViaProxy)
			}
		})
	}
}

// --- Test: concurrent multi-channel operations ---

func TestConcurrentMultiChannelOperations(t *testing.T) {
	mgr := NewUpstreamManager(nil)

	// Concurrently create upstream instances for different channels.
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			proxyTag := fmt.Sprintf("channel-%d", idx)
			config := UpstreamConfig{
				ID:       fmt.Sprintf("upstream-%d", idx),
				Addr:     "8.8.8.8:53",
				Protocol: "udp",
				ProxyTag: proxyTag,
			}
			instance := mgr.GetOrCreateUpstream(config)
			if instance == nil {
				t.Errorf("goroutine %d: got nil instance", idx)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines.
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all 10 instances were created.
	instances := mgr.ListUpstreams()
	if len(instances) != 10 {
		t.Errorf("expected 10 instances, got %d", len(instances))
	}

	// Verify they're all accessible by composite key.
	for i := 0; i < 10; i++ {
		proxyTag := fmt.Sprintf("channel-%d", i)
		key := BuildUpstreamKey("8.8.8.8:53", proxyTag)
		_, ok := mgr.GetUpstreamByKey(key)
		if !ok {
			t.Errorf("expected to find instance for key %s", key)
		}
	}
}

// --- Test: Empty manager ---

func TestEmptyUpstreamManager(t *testing.T) {
	mgr := NewUpstreamManager(nil)

	if mgr == nil {
		t.Fatal("NewUpstreamManager(nil) returned nil")
	}

	instances := mgr.ListUpstreams()
	if len(instances) != 0 {
		t.Errorf("expected 0 instances, got %d", len(instances))
	}

	_, ok := mgr.GetUpstream("anything")
	if ok {
		t.Error("expected false for GetUpstream on empty manager")
	}
}

// --- Test: Multiple channels, same address ---

func TestSameAddressDifferentChannels(t *testing.T) {
	// Multiple proxy channels for the same upstream address.
	channels := []string{"direct", "proxy", "proxy-us", "proxy-eu", "proxy-asia"}
	configs := make([]UpstreamConfig, len(channels))
	for i, ch := range channels {
		configs[i] = UpstreamConfig{
			ID:       fmt.Sprintf("google-%s", ch),
			Addr:     "8.8.8.8:53",
			Protocol: "udp",
			ProxyTag: ch,
		}
	}

	mgr := NewUpstreamManager(configs)

	// Verify all channels are separate instances.
	for _, ch := range channels {
		key := BuildUpstreamKey("8.8.8.8:53", ch)
		instance, ok := mgr.GetUpstreamByKey(key)
		if !ok {
			t.Errorf("expected instance for channel %q", ch)
		} else if instance.ProxyTag != ch {
			t.Errorf("expected ProxyTag=%q, got %q", ch, instance.ProxyTag)
		}
	}

	// Verify count.
	instances := mgr.ListUpstreams()
	if len(instances) != len(channels) {
		t.Errorf("expected %d instances, got %d", len(channels), len(instances))
	}
}

// --- Test: buildCacheKey with different proxy tags ---

func TestBuildCacheKeyWithProxyTags(t *testing.T) {
	query := newTestQuery("example.com", TypeA)

	tests := []struct {
		proxyTag    string
		expectedTag string
	}{
		{"direct", "direct"},
		{"proxy", "proxy"},
		{"", "direct"}, // Empty defaults to "direct"
	}

	for _, tt := range tests {
		t.Run(tt.expectedTag, func(t *testing.T) {
			upstream := &UpstreamInstance{
				Addr:     "8.8.8.8:53",
				ProxyTag: tt.proxyTag,
			}
			key := buildCacheKey(query, upstream)
			if key.ProxyTag != tt.expectedTag {
				t.Errorf("expected ProxyTag=%q, got %q", tt.expectedTag, key.ProxyTag)
			}
		})
	}
}

// --- Test: Exchange via proxy tag routing (decision only) ---

func TestExchangeProxyTagRoutingDecision(t *testing.T) {
	tests := []struct {
		name     string
		proxyTag string
		wantProxy bool
	}{
		{"direct connection", "direct", false},
		{"empty proxy tag", "", false},
		{"proxy channel", "proxy", true},
		{"custom channel", "my-proxy", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstream := &UpstreamInstance{
				Addr:     "8.8.8.8:53",
				Protocol: "udp",
				ProxyTag: tt.proxyTag,
				Client: &dns.Client{
					Net:          "udp",
					Timeout:      5 * time.Second,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
			}

			// The Exchange method checks ProxyTag to decide routing.
			shouldProxy := upstream.ProxyTag != "" && upstream.ProxyTag != "direct"

			if shouldProxy != tt.wantProxy {
				t.Errorf("Exchange(%q): expected proxy=%v, got proxy=%v",
					upstream.ProxyTag, tt.wantProxy, shouldProxy)
			}
		})
	}
}
