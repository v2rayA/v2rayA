// Package dns provides a standalone DNS module for v2rayA.
// It directly listens on UDP/TCP ports to handle DNS queries,
// replacing the previous approach of routing DNS traffic through
// xray's dokodemo-door and dns-out.
package dns

import (
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// QueryType represents a DNS query type.
type QueryType uint16

// Supported DNS query type constants.
const (
	TypeA     QueryType = QueryType(dns.TypeA)
	TypeAAAA  QueryType = QueryType(dns.TypeAAAA)
	TypeCNAME QueryType = QueryType(dns.TypeCNAME)
	TypeTXT   QueryType = QueryType(dns.TypeTXT)
	TypeMX    QueryType = QueryType(dns.TypeMX)
	TypeSRV   QueryType = QueryType(dns.TypeSRV)
	TypeNS    QueryType = QueryType(dns.TypeNS)
	TypePTR   QueryType = QueryType(dns.TypePTR)
	TypeSOA   QueryType = QueryType(dns.TypeSOA)
)

// DnsQuery represents a normalized DNS query request.
type DnsQuery struct {
	// Name is the queried domain name (e.g. "www.example.com").
	Name string
	// QType is the DNS query type (A, AAAA, etc.).
	QType QueryType
	// QClass is the DNS query class (usually IN=1).
	QClass uint16
	// ClientIP is the IP address of the DNS client.
	ClientIP net.IP
	// IsBootstrap indicates whether this is a bootstrap query.
	IsBootstrap bool
}

// DnsResponse represents a normalized DNS query response.
type DnsResponse struct {
	// Query is the original query corresponding to this response.
	Query DnsQuery
	// RawMsg is the raw DNS response message.
	RawMsg *dns.Msg
	// Rcode is the DNS response code (e.g. dns.RcodeSuccess).
	Rcode int
	// Answer contains the answer resource records.
	Answer []dns.RR
	// Authority contains the authority resource records.
	Authority []dns.RR
	// Additional contains the additional resource records.
	Additional []dns.RR
	// TTL is the minimum TTL among all answer records.
	TTL uint32
	// Upstream is the upstream DNS server address that provided the response.
	Upstream string
	// ProxyTag is the proxy channel tag used for the query.
	ProxyTag string
	// RTT is the round-trip time of the query.
	RTT time.Duration
	// Cached indicates whether the response was served from cache.
	Cached bool
}

// DnsListenerConfig holds configuration for the DNS listener.
type DnsListenerConfig struct {
	// ListenAddr is the primary listening address, default "0.0.0.0:52353".
	ListenAddr string `json:"listen_addr"`
	// ExtraListenAddrs is a list of additional addresses to listen on.
	// For example, ["127.2.0.17:53"] to handle hijacked system DNS queries
	// directly, avoiding the need for iptables REDIRECT rules.
	ExtraListenAddrs []string `json:"extra_listen_addrs,omitempty"`
	// Timeout is the query timeout in seconds, default 5.
	Timeout int `json:"timeout"`
}

// DnsModuleConfig holds the complete DNS module configuration.
type DnsModuleConfig struct {
	Listener  DnsListenerConfig `json:"listener"`
	Cache     CacheConfig       `json:"cache"`
	Upstreams []UpstreamConfig  `json:"upstreams"`
	Rules     []RuleConfig      `json:"rules"`
	// ProxyMap maps proxy tags (e.g. "proxy") to SOCKS5 proxy addresses (e.g. "127.0.0.1:20170").
	// The DNS module connects to these SOCKS5 proxies when sending queries through tagged upstreams.
	// If a tag is not found in this map, the module falls back to common defaults ("proxy" → "127.0.0.1:1080").
	ProxyMap map[string]string `json:"proxy_map,omitempty"`
	// Bootstrap lists domain names of DNS upstream servers that need to be resolved
	// via system DNS before the module can use them. These are resolved at startup.
	Bootstrap []string `json:"bootstrap,omitempty"`
	// BootstrapDns lists saved system DNS servers (read from /etc/resolv.conf before hijack).
	// Used by resolveBootstrap to avoid chicken-and-egg: the hijacked resolv.conf
	// points to 127.2.0.17:53 (our own DNS module), but it hasn't started yet.
	BootstrapDns []string `json:"bootstrap_dns,omitempty"`
}

// CacheConfig holds DNS cache configuration.
type CacheConfig struct {
	Enabled           bool `json:"enabled"`
	Size              int  `json:"size"`
	MinTTL            int  `json:"min_ttl"`
	MaxTTL            int  `json:"max_ttl"`
	Prefetch          bool `json:"prefetch"`
	PrefetchThreshold int  `json:"prefetch_threshold"`
	NegativeCache     bool `json:"neg_cache"`
}

// UpstreamConfig holds DNS upstream server configuration.
type UpstreamConfig struct {
	ID        string `json:"id"`
	Addr      string `json:"addr"`
	Protocol  string `json:"protocol,omitempty"`
	ProxyTag  string `json:"proxy_tag,omitempty"`
	Bootstrap bool   `json:"bootstrap,omitempty"`
}

// RuleConfig holds DNS routing rule configuration.
type RuleConfig struct {
	ID           string      `json:"id,omitempty"`
	Domain       []string    `json:"domain,omitempty"`
	DomainSuffix []string    `json:"domain_suffix,omitempty"`
	DomainRegex  []string    `json:"domain_regex,omitempty"`
	IP           []string    `json:"ip,omitempty"`
	QueryType    []QueryType `json:"query_type,omitempty"`
	ClientIP     []string    `json:"client_ip,omitempty"`
	Upstream     string      `json:"upstream"`
	Action       string      `json:"action,omitempty"`
	Policy       string      `json:"policy,omitempty"`
}

// DnsRule represents a fully parsed DNS routing rule with metadata.
type DnsRule struct {
	ID           string
	Domain       []string
	DomainSuffix []string
	DomainRegex  []string
	IP           []string
	QueryType    []QueryType
	ClientIP     []string
	Upstream     string
	Action       string
	Policy       string
}

// DnsStats accumulates DNS query statistics.
type DnsStats struct {
	TotalQueries    int64
	RoutedQueries   map[string]int64 // Per-upstream query count.
	RejectedQueries int64
	BypassedQueries int64
	CacheHits       int64
	CacheMisses     int64
	AvgRTT          time.Duration
	mu              sync.Mutex
	StartedAt       time.Time // 模块启动时间
}
