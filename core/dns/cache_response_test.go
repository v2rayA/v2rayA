package dns

import (
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestCacheHitDecreasesRecordTTLs(t *testing.T) {
	config := newTestCacheConfig()
	config.MinTTL = 100
	cache := NewDnsCache(config)
	key := newTestCacheKey("example.com", TypeA)
	response := newTestDNSResponse("example.com", TypeA, dns.RcodeSuccess, 150)
	cache.Set(key, response, 150)

	entry := cache.entries[key].Value.(*CacheEntry)
	entry.StoredAt = time.Now().Add(-75 * time.Second)
	entry.ExpiresAt = time.Now().Add(time.Minute)

	cached, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got := cached.Answer[0].Header().Ttl; got != 100 {
		t.Errorf("cached answer TTL = %d, want 100", got)
	}
	if got := cached.RawMsg.Answer[0].Header().Ttl; got != 100 {
		t.Errorf("cached raw message TTL = %d, want 100", got)
	}
	if got := cached.TTL; got != 100 {
		t.Errorf("cached response TTL = %d, want 100", got)
	}
	if got := response.Answer[0].Header().Ttl; got != 150 {
		t.Errorf("stored answer TTL = %d, want 150", got)
	}
}

func TestNegativeResponsesAreNotCachedWhenDisabled(t *testing.T) {
	for _, rcode := range []int{dns.RcodeNameError, dns.RcodeServerFailure, dns.RcodeRefused} {
		config := newTestCacheConfig()
		config.NegativeCache = false
		cache := NewDnsCache(config)
		key := newTestCacheKey("negative.example.com", TypeA)
		response := newTestDNSResponse("negative.example.com", TypeA, rcode, 0)

		cache.Set(key, response, 0)
		if _, ok := cache.Get(key); ok {
			t.Errorf("rcode %d was cached with negative caching disabled", rcode)
		}
	}
}
