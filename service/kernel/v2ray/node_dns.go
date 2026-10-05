package v2ray

import (
	"encoding/json"
	"fmt"
	"net"
	"runtime"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

type NodeDNSOption struct {
	Value    string `json:"value"`
	URL      string `json:"url"`
	Category string `json:"category"`
}

type NodeDNSOptions struct {
	Options  []NodeDNSOption `json:"options"`
	Warnings []string        `json:"warnings,omitempty"`
}

func CollectNodeDNSOptions(setting *configure.Setting, rules []configure.DnsRule) NodeDNSOptions {
	system, err := readOriginalResolv(resolvPath, resolvBackupPath)
	listeners := append([]string{dnsModuleListenAddr(setting)}, dnsModuleExtraListenAddrs(setting)...)
	listeners = append(listeners, tunDNSListeners(setting)...)
	// Also exclude the running generation when listeners are being changed.
	if p := ProcessManager.Process(); p != nil && p.template != nil && p.template.Setting != nil {
		listeners = append(listeners, dnsModuleListenAddr(p.template.Setting))
		listeners = append(listeners, tunDNSListeners(p.template.Setting)...)
		// Avoid probing ports already owned by the running module.
		listeners = append(listeners, runningDNSListeners(p.template)...)
	}
	result := collectNodeDNSOptions(rules, system, listeners, resolv.FallbackDNSURLs())
	if err != nil {
		result.Warnings = append(result.Warnings, "NODE_DNS_SYSTEM_READ_FAILED")
	}
	return result
}

func runningDNSListeners(t *Template) []string {
	var cfg struct {
		Listener struct {
			Address string   `json:"listen_addr"`
			Extra   []string `json:"extra_listen_addrs"`
		} `json:"listener"`
	}
	_ = json.Unmarshal(t.DnsModuleConfig, &cfg)
	return append([]string{cfg.Listener.Address}, cfg.Listener.Extra...)
}

func collectNodeDNSOptions(rules []configure.DnsRule, system, listeners, fallback []string) NodeDNSOptions {
	groups := map[string][]string{"direct": {}, "localhost": system, "proxy": {}, "fallback": fallback}
	for _, rule := range configure.MigrateDnsRules(rules) {
		if rule.Action != "route" || rule.Outbound == "block" || rule.Outbound == "reject" || rule.Outbound == "blackhole" {
			continue
		}
		if rule.Upstream == "localhost" {
			continue
		}
		category := "proxy"
		if rule.Outbound == "direct" {
			category = "direct"
		}
		groups[category] = append(groups[category], rule.Upstream)
	}
	result := NodeDNSOptions{Options: make([]NodeDNSOption, 0)}
	seen := make(map[string]bool)
	auto := ""
	for _, category := range []string{"direct", "localhost", "proxy", "fallback"} {
		for _, address := range groups[category] {
			endpoint, err := resolv.ParseIPDNS(address)
			if err != nil || isOwnDNS(endpoint, listeners) || seen[endpoint.URL] {
				continue
			}
			// Intercepted resolver/TUN addresses cannot stand in for system DNS.
			if category == "localhost" && endpoint.IP.Equal(net.ParseIP("127.2.0.17")) {
				continue
			}
			seen[endpoint.URL] = true
			result.Options = append(result.Options, NodeDNSOption{Value: endpoint.URL, URL: endpoint.URL, Category: category})
			if auto == "" && category != "proxy" {
				auto = endpoint.URL
			}
		}
	}
	if auto != "" {
		result.Options = append([]NodeDNSOption{{Value: "auto", URL: auto, Category: "auto"}}, result.Options...)
	}
	return result
}

func isOwnDNS(endpoint *resolv.IPDNSEndpoint, listeners []string) bool {
	if endpoint == nil {
		return true
	}
	for _, listener := range listeners {
		host, port, err := net.SplitHostPort(listener)
		if err != nil || port != endpoint.Port {
			continue
		}
		ip := net.ParseIP(host)
		if ip != nil && ip.Equal(endpoint.IP) {
			return true
		}
		if host == "" || (ip != nil && ip.IsUnspecified()) {
			if endpoint.IP.IsLoopback() {
				return true
			}
			addrs, _ := net.InterfaceAddrs()
			for _, addr := range addrs {
				local, _, _ := net.ParseCIDR(addr.String())
				if local.Equal(endpoint.IP) {
					return true
				}
			}
		}
	}
	return false
}

func (options NodeDNSOptions) Select(value string) (*resolv.IPDNSEndpoint, error) {
	if value == "" {
		value = "auto"
	}
	if value != "auto" {
		endpoint, err := resolv.ParseIPDNS(value)
		if err != nil {
			return nil, common.Coded("NODE_DNS_INVALID", err, nil)
		}
		value = endpoint.URL
	}
	for _, option := range options.Options {
		if option.Value == value {
			return resolv.ParseIPDNS(option.URL)
		}
	}
	return nil, common.Coded("NODE_DNS_INVALID", fmt.Errorf("node DNS source is no longer available; refresh or choose another endpoint or auto"), nil)
}

func SelectNodeDNS(setting *configure.Setting, rules []configure.DnsRule) (*resolv.IPDNSEndpoint, error) {
	return CollectNodeDNSOptions(setting, rules).Select(setting.NodeDns)
}

// PlannedDNSHijack follows the policy that the next generation will apply.
// Latency tasks instead use DNSHijackActive, which records successful setup.
func PlannedDNSHijack(setting *configure.Setting) bool {
	if !IsTransparentOn(setting) {
		return false
	}
	if setting.TransparentType == configure.TransparentTun && runtime.GOOS != "linux" {
		if !setting.TunAutoRoute {
			return false
		}
		if runtime.GOOS == "darwin" {
			endpoint, _ := resolv.ParseIPDNS("127.0.0.1:53")
			listeners := append([]string{dnsModuleListenAddr(setting)}, dnsModuleExtraListenAddrs(setting)...)
			return isOwnDNS(endpoint, listeners)
		}
		return true
	}
	install, _ := dnsRedirectPolicy(runtime.GOOS, setting)
	return install && shouldLocalDnsListen(setting)
}

func DNSHijackActive() bool { return ProcessManager.Running() && dnsHijackActive.Load() }

// AddNodeDNSDomains extends the temporary latency configuration to every
// tested node, including nodes absent from the normal connected set.
func (t *Template) AddNodeDNSDomains(servers []serverObj.ServerObj) error {
	if t.NodeDNS == nil {
		return fmt.Errorf("node DNS endpoint is missing")
	}
	var cfg map[string]interface{}
	if len(t.DnsModuleConfig) == 0 {
		infos := make([]serverInfo, 0, len(servers))
		for _, server := range servers {
			if server != nil {
				infos = append(infos, serverInfo{Info: server})
			}
		}
		if err := t.generateDnsModuleConfig(infos); err != nil {
			return err
		}
	}
	if err := json.Unmarshal(t.DnsModuleConfig, &cfg); err != nil {
		return err
	}
	var domains []string
	for _, server := range servers {
		if server != nil && net.ParseIP(server.GetHostname()) == nil {
			domains = append(domains, server.GetHostname())
		}
	}
	rules, _ := cfg["rules"].([]interface{})
	upstreams, _ := cfg["upstreams"].([]interface{})
	found := false
	for _, item := range rules {
		rule := item.(map[string]interface{})
		if rule["id"] == "rule-node" {
			previous, _ := rule["domain_suffix"].([]interface{})
			for _, domain := range previous {
				domains = append(domains, domain.(string))
			}
			rule["domain_suffix"] = common.Deduplicate(domains)
			found = true
		}
	}
	endpoint := t.NodeDNS
	if !found && len(domains) > 0 {
		rules = append([]interface{}{map[string]interface{}{"id": "rule-node", "upstream": "upstream-node", "action": "route", "policy": "single", "domain_suffix": common.Deduplicate(domains)}}, rules...)
		upstreams = append([]interface{}{map[string]interface{}{"id": "upstream-node"}}, upstreams...)
	}
	for _, item := range upstreams {
		upstream := item.(map[string]interface{})
		if upstream["id"] == "upstream-node" {
			upstream["addr"] = endpoint.UpstreamAddress()
			upstream["protocol"] = endpoint.Protocol()
			upstream["server_name"] = endpoint.IP.String()
			upstream["proxy_tag"] = "direct"
		}
	}
	cfg["rules"], cfg["upstreams"] = rules, upstreams
	raw, err := json.Marshal(cfg)
	if err == nil {
		t.DnsModuleConfig = raw
	}
	return err
}

func moduleOwnsDNSListener(address string) bool {
	p := ProcessManager.Process()
	if p == nil || p.template == nil {
		return false
	}
	for _, listener := range runningDNSListeners(p.template) {
		if listener == address {
			return true
		}
	}
	return false
}

func tunDNSListeners(setting *configure.Setting) []string {
	if setting.TransparentType != configure.TransparentTun || !IsTransparentOn(setting) {
		return nil
	}
	return []string{net.JoinHostPort(tunGateway4, "53"), net.JoinHostPort(tunGateway6, "53")}
}
