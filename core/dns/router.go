package dns

import (
	"fmt"
	"log"
	"sync"
)

// RouteResult represents the result of routing a DNS query.
type RouteResult struct {
	// UpstreamID is the ID of the matched upstream server.
	UpstreamID string
	// UpstreamAddr is the address of the matched upstream server.
	UpstreamAddr string
	// Action is the routing action: route / reject / bypass.
	Action string
	// RuleID is the ID of the matched rule.
	RuleID string
	// Policy is the query policy: single / parallel / fallback.
	Policy string
	// ProxyTag is the proxy channel tag associated with the upstream.
	// Empty or "direct" means direct connection; other values indicate
	// a specific proxy outbound tag.
	ProxyTag string
}

// Router is the DNS query routing engine.
// It maintains an ordered list of rules and matches incoming queries
// against them to determine the upstream server.
type Router struct {
	rules            []*DnsRule        // Ordered list of routing rules.
	matchers         [][]RuleMatcher   // Pre-built matchers for each rule.
	upstreamMap      map[string]string // Upstream ID → upstream address mapping.
	upstreamProxyTag map[string]string // Upstream ID → proxy tag mapping.
	defaultUpstream  string            // Default upstream ID used when no rule matches.
	geositeResolver  GeositeResolver   // Resolves geosite tags (set externally).
	mu               sync.RWMutex
}

// NewRouter creates a new Router with the given rules and default upstream.
func NewRouter(rules []*DnsRule, upstreams []UpstreamConfig, defaultUpstream string) (*Router, error) {
	r := &Router{
		rules:            make([]*DnsRule, 0),
		matchers:         make([][]RuleMatcher, 0),
		upstreamMap:      make(map[string]string),
		upstreamProxyTag: make(map[string]string),
		defaultUpstream:  defaultUpstream,
	}

	// Build upstream address and proxy tag mapping.
	for _, u := range upstreams {
		if u.ID != "" {
			r.upstreamMap[u.ID] = u.Addr
			if u.ProxyTag != "" {
				r.upstreamProxyTag[u.ID] = u.ProxyTag
			}
		}
	}

	// Load and compile rules.
	if err := r.UpdateRules(rules); err != nil {
		return nil, fmt.Errorf("new router: %w", err)
	}

	return r, nil
}

// Route matches a DNS query against the rule list and returns the routing result.
// Rules are evaluated in order; the first matching rule determines the result.
// If no rule matches, the default upstream is used.
func (r *Router) Route(query *DnsQuery) *RouteResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Iterate rules in order.
	for i, rule := range r.rules {
		if i >= len(r.matchers) {
			continue
		}
		matchers := r.matchers[i]

		// Check if all matcher groups match (AND across groups, OR within each group).
		if MatchAll(matchers, query) {
			// Rule matched — determine upstream address.
			upstreamAddr := r.upstreamMap[rule.Upstream]

			log.Printf("[dns router] rule %q matched query %s %s → upstream=%s action=%s",
				rule.ID, query.Name, rule.Upstream, upstreamAddr, rule.Action)

			action := rule.Action
			if action == "" {
				action = "route"
			}
			policy := rule.Policy
			if policy == "" {
				policy = "single"
			}

			// Get proxy tag for the matched upstream.
			proxyTag := r.upstreamProxyTag[rule.Upstream]

			return &RouteResult{
				UpstreamID:   rule.Upstream,
				UpstreamAddr: upstreamAddr,
				Action:       action,
				RuleID:       rule.ID,
				Policy:       policy,
				ProxyTag:     proxyTag,
			}
		}
	}

	// No rule matched — use default upstream.
	log.Printf("[dns router] no rule matched for %s %d → using default upstream %s",
		query.Name, query.QType, r.defaultUpstream)

	defaultAddr := r.upstreamMap[r.defaultUpstream]
	defaultProxyTag := r.upstreamProxyTag[r.defaultUpstream]
	return &RouteResult{
		UpstreamID:   r.defaultUpstream,
		UpstreamAddr: defaultAddr,
		Action:       "route",
		RuleID:       "",
		Policy:       "single",
		ProxyTag:     defaultProxyTag,
	}
}

// UpdateRules replaces the current rule set with new rules.
// It rebuilds all matchers for the new rules, expanding geosite tags using
// the configured resolver.
func (r *Router) UpdateRules(rules []*DnsRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	newRules := make([]*DnsRule, len(rules))
	newMatchers := make([][]RuleMatcher, len(rules))

	for i, rule := range rules {
		// Validate and build matchers, passing the geosite resolver.
		matchers, err := BuildMatchers(rule, r.geositeResolver)
		if err != nil {
			return fmt.Errorf("rule %d (id=%q): %w", i, rule.ID, err)
		}

		// Copy the rule.
		r := *rule
		newRules[i] = &r
		newMatchers[i] = matchers
	}

	r.rules = newRules
	r.matchers = newMatchers

	log.Printf("[dns router] updated rules: %d rules loaded", len(rules))
	return nil
}

// ListRules returns a copy of all registered rules.
func (r *Router) ListRules() []*DnsRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rules := make([]*DnsRule, len(r.rules))
	for i, rule := range r.rules {
		cp := *rule
		rules[i] = &cp
	}
	return rules
}

// SetGeositeResolver sets the geosite tag resolver function.
func (r *Router) SetGeositeResolver(resolver GeositeResolver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.geositeResolver = resolver
}

// Reload reloads the router configuration from a DnsModuleConfig.
func (r *Router) Reload(cfg *DnsModuleConfig) error {
	if cfg == nil {
		return fmt.Errorf("router reload: config is nil")
	}

	// Convert RuleConfigs to DnsRules.
	rules := make([]*DnsRule, len(cfg.Rules))
	for i, rc := range cfg.Rules {
		rules[i] = &DnsRule{
			ID:           rc.ID,
			Domain:       rc.Domain,
			DomainSuffix: rc.DomainSuffix,
			DomainRegex:  rc.DomainRegex,
			IP:           rc.IP,
			QueryType:    rc.QueryType,
			ClientIP:     rc.ClientIP,
			Upstream:     rc.Upstream,
			Action:       rc.Action,
			Policy:       rc.Policy,
		}
	}

	// Rebuild upstream mapping.
	r.mu.Lock()
	r.upstreamMap = make(map[string]string)
	r.upstreamProxyTag = make(map[string]string)
	for _, u := range cfg.Upstreams {
		if u.ID != "" {
			r.upstreamMap[u.ID] = u.Addr
			if u.ProxyTag != "" {
				r.upstreamProxyTag[u.ID] = u.ProxyTag
			}
		}
	}
	r.mu.Unlock()

	return r.UpdateRules(rules)
}
