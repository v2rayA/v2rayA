package dns

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// GeositeResolver resolves a geosite tag (e.g. "cn", "geolocation-!cn")
// into domain and suffix patterns. Returns (domains, suffixes) where
// domains are for exact/full matching and suffixes for suffix matching.
// Nil means the tag is not recognized/loaded.
type GeositeResolver func(tag string) (domains, suffixes []string)

// RuleMatcher is the interface for DNS query rule matchers.
type RuleMatcher interface {
	// Match checks whether the given query matches this rule.
	Match(query *DnsQuery) bool
	// Type returns the matcher type description (e.g., "domain", "ip", "querytype", "clientip").
	Type() string
	// Value returns the matcher value description.
	Value() string
}

// DomainRule matches DNS queries by domain name.
// It supports exact match, suffix match (wildcard), and regex match.
type DomainRule struct {
	Domain   string
	IsSuffix bool // true: match suffix like .example.com; false: exact match
	IsRegex  bool // true: regex match
	Re       *regexp.Regexp
}

// Match checks if the query domain matches this domain rule.
func (r *DomainRule) Match(query *DnsQuery) bool {
	name := strings.TrimSuffix(query.Name, ".")

	if r.IsRegex {
		return r.Re.MatchString(name)
	}

	if r.IsSuffix {
		// Suffix match: domain ".example.com" matches "sub.example.com" or "example.com"
		// If the rule domain starts with ".", treat it as a suffix match for subdomains.
		domain := strings.TrimSuffix(r.Domain, ".")
		if strings.HasPrefix(domain, ".") {
			// ".example.com" matches anything ending with ".example.com"
			return strings.HasSuffix(name, domain)
		}
		// "example.com" matches "example.com" itself and "sub.example.com"
		if name == domain {
			return true
		}
		return strings.HasSuffix(name, "."+domain)
	}

	// Exact match
	domain := strings.TrimSuffix(r.Domain, ".")
	return name == domain
}

// Type returns the matcher type.
func (r *DomainRule) Type() string {
	if r.IsRegex {
		return "domain_regex"
	}
	if r.IsSuffix {
		return "domain_suffix"
	}
	return "domain"
}

// Value returns the matcher value.
func (r *DomainRule) Value() string {
	return r.Domain
}

// IPRule matches DNS queries by response IP CIDR range.
// Note: This matcher is used for response IP-based filtering.
// For query-time matching, it checks against the query name's resolved IPs
// (pre-match mode) or is used in post-processing.
type IPRule struct {
	CIDR *net.IPNet
}

// Match checks if the client IP or any embedded IP matches this rule.
// For request-time use, this typically matches the query's ClientIP against
// the CIDR range. Response-time matching would check answer records.
func (r *IPRule) Match(query *DnsQuery) bool {
	if r.CIDR == nil {
		return false
	}
	return r.CIDR.Contains(query.ClientIP)
}

// Type returns the matcher type.
func (r *IPRule) Type() string {
	return "ip"
}

// Value returns the matcher value.
func (r *IPRule) Value() string {
	if r.CIDR == nil {
		return ""
	}
	return r.CIDR.String()
}

// QueryTypeRule matches DNS queries by query type (A, AAAA, etc.).
type QueryTypeRule struct {
	QType QueryType
}

// Match checks if the query type matches this rule.
func (r *QueryTypeRule) Match(query *DnsQuery) bool {
	return r.QType == query.QType
}

// Type returns the matcher type.
func (r *QueryTypeRule) Type() string {
	return "query_type"
}

// Value returns the matcher value.
func (r *QueryTypeRule) Value() string {
	return fmt.Sprintf("%d", uint16(r.QType))
}

// ClientIPRule matches DNS queries by client IP CIDR range.
type ClientIPRule struct {
	CIDR *net.IPNet
}

// Match checks if the query's client IP matches this rule.
func (r *ClientIPRule) Match(query *DnsQuery) bool {
	if r.CIDR == nil || query.ClientIP == nil {
		return false
	}
	return r.CIDR.Contains(query.ClientIP)
}

// Type returns the matcher type.
func (r *ClientIPRule) Type() string {
	return "client_ip"
}

// Value returns the matcher value.
func (r *ClientIPRule) Value() string {
	if r.CIDR == nil {
		return ""
	}
	return r.CIDR.String()
}

// parseCIDR parses a CIDR string and returns the IPNet.
// Supports both regular CIDR (e.g., "10.0.0.0/8") and single IP (e.g., "10.0.0.1").
func parseCIDR(s string) (*net.IPNet, error) {
	if strings.Contains(s, "/") {
		_, cidr, err := net.ParseCIDR(s)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", s, err)
		}
		return cidr, nil
	}
	// Single IP: treat as /32 for IPv4 or /128 for IPv6
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address %q", s)
	}
	if ip.To4() != nil {
		_, cidr, err := net.ParseCIDR(s + "/32")
		if err != nil {
			return nil, err
		}
		return cidr, nil
	}
	_, cidr, err := net.ParseCIDR(s + "/128")
	if err != nil {
		return nil, err
	}
	return cidr, nil
}

// BuildMatchers constructs a list of RuleMatcher from a DnsRule configuration.
// Each rule field generates one or more matchers. All matchers must match (AND logic).
// If a resolver is provided, geosite tags (e.g. "geosite:cn") are expanded.
func BuildMatchers(rule *DnsRule, resolver GeositeResolver) ([]RuleMatcher, error) {
	var matchers []RuleMatcher

	// Build domain exact matchers, expanding geosite tags if resolver is available.
	for _, d := range rule.Domain {
		if strings.HasPrefix(d, "geosite:") {
			if resolver != nil {
				doms, suffs := resolver(strings.TrimPrefix(d, "geosite:"))
				for _, dom := range doms {
					matchers = append(matchers, &DomainRule{Domain: dom, IsSuffix: false, IsRegex: false})
				}
				for _, suf := range suffs {
					matchers = append(matchers, &DomainRule{Domain: suf, IsSuffix: true, IsRegex: false})
				}
			}
			// Without resolver, geosite tags are silently skipped (no matcher created).
			continue
		}
		matchers = append(matchers, &DomainRule{Domain: d, IsSuffix: false, IsRegex: false})
	}

	// Build domain suffix matchers (also check for geosite patterns).
	for _, d := range rule.DomainSuffix {
		if strings.HasPrefix(d, "geosite:") {
			if resolver != nil {
				doms, suffs := resolver(strings.TrimPrefix(d, "geosite:"))
				for _, dom := range doms {
					matchers = append(matchers, &DomainRule{Domain: dom, IsSuffix: false, IsRegex: false})
				}
				for _, suf := range suffs {
					matchers = append(matchers, &DomainRule{Domain: suf, IsSuffix: true, IsRegex: false})
				}
			}
			continue
		}
		matchers = append(matchers, &DomainRule{Domain: d, IsSuffix: true, IsRegex: false})
	}

	// Build domain regex matchers
	for _, d := range rule.DomainRegex {
		re, err := regexp.Compile(d)
		if err != nil {
			return nil, fmt.Errorf("invalid domain regex %q: %w", d, err)
		}
		matchers = append(matchers, &DomainRule{Domain: d, IsRegex: true, Re: re})
	}

	// Build IP CIDR matchers
	for _, ipStr := range rule.IP {
		cidr, err := parseCIDR(ipStr)
		if err != nil {
			return nil, fmt.Errorf("invalid IP CIDR %q: %w", ipStr, err)
		}
		matchers = append(matchers, &IPRule{CIDR: cidr})
	}

	// Build query type matchers
	for _, qt := range rule.QueryType {
		matchers = append(matchers, &QueryTypeRule{QType: qt})
	}

	// Build client IP CIDR matchers
	for _, cipStr := range rule.ClientIP {
		cidr, err := parseCIDR(cipStr)
		if err != nil {
			return nil, fmt.Errorf("invalid client IP CIDR %q: %w", cipStr, err)
		}
		matchers = append(matchers, &ClientIPRule{CIDR: cidr})
	}

	return matchers, nil
}

// MatchAny checks if all matchers in the list match the query (AND logic).
// Returns true only if every matcher matches.
func MatchAny(matchers []RuleMatcher, query *DnsQuery) bool {
	if len(matchers) == 0 {
		// Empty matchers means no constraints — matches everything.
		return true
	}
	for _, m := range matchers {
		if !m.Match(query) {
			return false
		}
	}
	return true
}
