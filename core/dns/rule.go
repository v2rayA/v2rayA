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

// OrMatcher matches if ANY of its child matchers match (OR logic).
// It is used to group matchers from the same rule field together,
// so that multiple domain patterns from a geosite expansion are OR-ed.
type OrMatcher struct {
	Matchers []RuleMatcher
}

// Match returns true if any child matcher matches.
func (m *OrMatcher) Match(query *DnsQuery) bool {
	for _, matcher := range m.Matchers {
		if matcher.Match(query) {
			return true
		}
	}
	return false
}

// Type returns the matcher type.
func (m *OrMatcher) Type() string {
	return "or_group"
}

// Value returns a description of the group.
func (m *OrMatcher) Value() string {
	return fmt.Sprintf("%d matchers", len(m.Matchers))
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
//
// Matching semantics: within each field group, matchers are OR-ed (any one match
// is sufficient); across different field groups, matchers are AND-ed (all groups
// must match). For example, a rule with domain=["geosite:cn"] and query_type=[A]
// matches if the domain is in geosite:cn AND the query type is A.
//
// If a resolver is provided, geosite tags (e.g. "geosite:cn") are expanded into
// individual domain matchers, all grouped into a single OR group.
func BuildMatchers(rule *DnsRule, resolver GeositeResolver) ([]RuleMatcher, error) {
	// Collect matchers by field group. Each group uses OR logic internally.
	var domainMatchers []RuleMatcher
	var ipMatchers []RuleMatcher
	var qtypeMatchers []RuleMatcher
	var clientIPMatchers []RuleMatcher

	// ── Domain group (OR): exact + suffix + regex + geosite expansion ──

	for _, d := range rule.Domain {
		if strings.HasPrefix(d, "geosite:") {
			if resolver != nil {
				doms, suffs := resolver(strings.TrimPrefix(d, "geosite:"))
				for _, dom := range doms {
					domainMatchers = append(domainMatchers, &DomainRule{Domain: dom, IsSuffix: false, IsRegex: false})
				}
				for _, suf := range suffs {
					domainMatchers = append(domainMatchers, &DomainRule{Domain: suf, IsSuffix: true, IsRegex: false})
				}
			}
			// Without resolver, geosite tags are silently skipped (no matcher created).
			continue
		}
		domainMatchers = append(domainMatchers, &DomainRule{Domain: d, IsSuffix: false, IsRegex: false})
	}

	for _, d := range rule.DomainSuffix {
		if strings.HasPrefix(d, "geosite:") {
			if resolver != nil {
				doms, suffs := resolver(strings.TrimPrefix(d, "geosite:"))
				for _, dom := range doms {
					domainMatchers = append(domainMatchers, &DomainRule{Domain: dom, IsSuffix: false, IsRegex: false})
				}
				for _, suf := range suffs {
					domainMatchers = append(domainMatchers, &DomainRule{Domain: suf, IsSuffix: true, IsRegex: false})
				}
			}
			continue
		}
		domainMatchers = append(domainMatchers, &DomainRule{Domain: d, IsSuffix: true, IsRegex: false})
	}

	for _, d := range rule.DomainRegex {
		re, err := regexp.Compile(d)
		if err != nil {
			return nil, fmt.Errorf("invalid domain regex %q: %w", d, err)
		}
		domainMatchers = append(domainMatchers, &DomainRule{Domain: d, IsRegex: true, Re: re})
	}

	// ── IP group (OR) ──
	for _, ipStr := range rule.IP {
		cidr, err := parseCIDR(ipStr)
		if err != nil {
			return nil, fmt.Errorf("invalid IP CIDR %q: %w", ipStr, err)
		}
		ipMatchers = append(ipMatchers, &IPRule{CIDR: cidr})
	}

	// ── QueryType group (OR) ──
	for _, qt := range rule.QueryType {
		qtypeMatchers = append(qtypeMatchers, &QueryTypeRule{QType: qt})
	}

	// ── ClientIP group (OR) ──
	for _, cipStr := range rule.ClientIP {
		cidr, err := parseCIDR(cipStr)
		if err != nil {
			return nil, fmt.Errorf("invalid client IP CIDR %q: %w", cipStr, err)
		}
		clientIPMatchers = append(clientIPMatchers, &ClientIPRule{CIDR: cidr})
	}

	// ── Build final list: each group is OR internally, AND across groups ──
	var matchers []RuleMatcher

	if len(domainMatchers) > 0 {
		if len(domainMatchers) == 1 {
			matchers = append(matchers, domainMatchers[0])
		} else {
			matchers = append(matchers, &OrMatcher{Matchers: domainMatchers})
		}
	}
	if len(ipMatchers) > 0 {
		if len(ipMatchers) == 1 {
			matchers = append(matchers, ipMatchers[0])
		} else {
			matchers = append(matchers, &OrMatcher{Matchers: ipMatchers})
		}
	}
	if len(qtypeMatchers) > 0 {
		if len(qtypeMatchers) == 1 {
			matchers = append(matchers, qtypeMatchers[0])
		} else {
			matchers = append(matchers, &OrMatcher{Matchers: qtypeMatchers})
		}
	}
	if len(clientIPMatchers) > 0 {
		if len(clientIPMatchers) == 1 {
			matchers = append(matchers, clientIPMatchers[0])
		} else {
			matchers = append(matchers, &OrMatcher{Matchers: clientIPMatchers})
		}
	}

	return matchers, nil
}

// MatchAll checks if all matcher groups match the query (AND across groups).
// Within each group (e.g. OrMatcher), the semantics is OR (any one matches).
// Empty matchers list means no constraints — matches everything.
func MatchAll(matchers []RuleMatcher, query *DnsQuery) bool {
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
