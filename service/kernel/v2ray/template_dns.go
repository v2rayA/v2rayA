package v2ray

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/iptables"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

type Addr struct {
	host string
	port string
	udp  bool
}

// setDNS 生成新 DNS 模块的配置，嵌入 xray JSON 供 v2raya-core 读取。
// v2raya-core 启动时解析此配置并启动独立 DNS 监听器，v2rayA 不参与 DNS 查询处理。
func (t *Template) setDNS(serverInfos []serverInfo) error {
	if !dnsServiceEnabled(t.Setting) {
		t.DnsModuleConfig = nil
		return nil
	}
	return t.generateDnsModuleConfig(serverInfos)
}

// dnsModuleListenAddr is the DNS module's primary listener. The stored
// setting defaults to 0.0.0.0:52353 (NewSetting and MigrateSetting write
// it), so that value is treated as "not chosen": it stays on loopback
// unless the box serves other hosts — port sharing, or IP forwarding for
// a LAN behind it, whose clients' queries the PREROUTING REDIRECT delivers
// to this host's LAN address. Any other value is the user's and is used
// as is.
func dnsModuleListenAddr(setting *configure.Setting) string {
	const stock = "0.0.0.0:52353"
	if setting.DnsListenAddr != "" && setting.DnsListenAddr != stock {
		return setting.DnsListenAddr
	}
	if setting.PortSharing || setting.IpForward {
		return stock
	}
	return "127.0.0.1:52353"
}

// dnsModuleExtraListenAddrs adds the 127.2.0.17:53 listener the resolv.conf
// hijack points at, but only when port 53 is free for it: a resolver bound
// to the wildcard address would make the bind fail and the core exit.
func dnsModuleExtraListenAddrs(setting *configure.Setting) []string {
	addrs := []string{}
	// Every extra address here exists to catch a query that a REDIRECT rule
	// or a hijacked resolver sends to it. Service mode leaves the firewall
	// and the resolvers alone, so the module keeps only its configured
	// address there.
	if !dnsInterceptionEnabled(setting) {
		return addrs
	}
	// The ip6tables REDIRECT sends an application's IPv6 DNS query to ::1,
	// which a loopback IPv4 primary listener does not cover. A wildcard
	// primary listener is dual-stack already and a second bind would fail.
	host, _, _ := net.SplitHostPort(dnsModuleListenAddr(setting))
	if ip := net.ParseIP(host); ip != nil && !ip.IsUnspecified() && ip.To4() != nil && iptables.IsIPv6Supported() {
		addrs = append(addrs, net.JoinHostPort("::1", dnsModulePort(setting)))
	}
	// macOS's system resolver is pointed at loopback port 53 in tun mode
	// (see tun_core_darwin.go); the Linux resolv.conf hijack points at
	// 127.2.0.17. Nothing sends to either on Windows.
	switch runtime.GOOS {
	case "darwin":
		if setting.TransparentType == configure.TransparentTun && IsTransparentOn(setting) {
			// The system resolver is pointed at loopback in tun mode. If
			// another resolver already owns the port, leave it to that one
			// rather than start a core that cannot bind.
			if err := probeListen("127.0.0.1:53"); err != nil && !moduleOwnsDNSListener("127.0.0.1:53") {
				log.Warn("DNS module will not listen on 127.0.0.1:53, the system resolver keeps using whatever answers there: %v", err)
				return addrs
			}
			return append(addrs, "127.0.0.1:53")
		}
		return addrs
	case "linux":
	default:
		return addrs
	}
	// The tproxy rules hand LAN clients' queries to 127.2.0.17 on the
	// module's port and the OUTPUT REDIRECT lands local queries on
	// 127.0.0.1; a primary listener on a specific address covers neither.
	if ip := net.ParseIP(host); ip != nil && !ip.IsUnspecified() {
		if !ip.IsLoopback() {
			addrs = append(addrs, net.JoinHostPort("127.0.0.1", dnsModulePort(setting)))
		}
		addrs = append(addrs, net.JoinHostPort("127.2.0.17", dnsModulePort(setting)))
	}
	if could, err := CouldLocalDnsListen(); !could {
		log.Warn("DNS module will not listen on 127.2.0.17:53: %v", err)
		return addrs
	}
	return append(addrs, "127.2.0.17:53")
}

// CheckDnsUpstream rejects an upstream the DNS module cannot query. The
// module speaks plain UDP and TCP, DNS over TLS and DNS over HTTPS;
// quic:// (DoQ) used to be accepted here and then failed on every query.
func CheckDnsUpstream(upstream string) error {
	scheme, rest, found := strings.Cut(upstream, "://")
	if !found {
		if strings.TrimSpace(upstream) == "" {
			return fmt.Errorf("DNS upstream is empty")
		}
		return nil
	}
	switch strings.ToLower(scheme) {
	case "udp", "tcp", "tls":
		if rest == "" {
			return fmt.Errorf("DNS upstream %q has no address after the scheme", upstream)
		}
		return nil
	case "https":
		if u, err := url.Parse(upstream); err != nil || u.Hostname() == "" {
			return fmt.Errorf("DNS upstream %q is not a URL with a host", upstream)
		}
		return nil
	case "quic":
		return fmt.Errorf("DNS upstream %q: DNS over QUIC is not supported; use an address (8.8.8.8), tls://host or https://host/dns-query", upstream)
	default:
		return fmt.Errorf("DNS upstream %q: unknown scheme %q; use an address, tcp://, tls:// or https://", upstream, scheme)
	}
}

// directDnsServers lists the plain udp/tcp IP upstreams of the DNS rules
// that go out directly, host:port. They come before any built-in public
// resolver wherever the service or the core needs one of its own.
func directDnsServers() []string {
	var out []string
	for _, rule := range configure.MigrateDnsRules(configure.GetDnsRulesNotNil()) {
		if rule.Outbound != "" && rule.Outbound != "direct" {
			continue
		}
		if rule.Action != "route" {
			continue
		}
		endpoint, err := resolv.ParseIPDNS(rule.Upstream)
		if err != nil || (endpoint.Scheme != "udp" && endpoint.Scheme != "tcp") || endpoint.IP.IsLoopback() {
			continue
		}
		out = append(out, endpoint.Address())
	}
	return common.Deduplicate(out)
}

func init() {
	resolv.PreferredServers = directDnsServers
}

// generateDnsModuleConfig 生成新 DNS 模块的 JSON 配置，嵌入 xray JSON 配置文件。
// v2raya-core 启动时解析此配置并启动独立 DNS 监听器，v2rayA 不参与 DNS 查询处理。
//
// 生成的配置结构对应 core/dns/config.go 中的 DnsModuleConfig。
func (t *Template) generateDnsModuleConfig(serverInfos []serverInfo) error {
	setting := t.Setting
	if setting == nil {
		setting = configure.GetSettingNotNil()
	}

	listenAddr := dnsModuleListenAddr(setting)

	// 获取 SOCKS 入站端口（用于 proxy_map）
	socksPort := 20170
	if p := configure.GetPortsNotNil(); p != nil && p.Socks5 > 0 {
		socksPort = p.Socks5
	}

	// 读取当前系统 DNS（保存原始配置，用于 v2raya-core 的 bootstrap 解析）。
	// 已劫持时从原始备份读取，避免把模块自身地址作为上游。
	// 规则里直连的明文上游排在其后，公共 DNS 只在这些都不可用时才轮到。
	bootstrapDns := common.Deduplicate(append(getSystemDnsServers(), directDnsServers()...))

	cfg := map[string]interface{}{
		"listener": map[string]interface{}{
			"listen_addr":        listenAddr,
			"extra_listen_addrs": dnsModuleExtraListenAddrs(setting),
			"timeout":            5,
		},
		"cache": map[string]interface{}{
			"enabled":   setting.DnsCacheEnabled,
			"size":      setting.DnsCacheSize,
			"min_ttl":   setting.DnsCacheMinTTL,
			"max_ttl":   setting.DnsCacheMaxTTL,
			"prefetch":  setting.DnsPrefetch,
			"neg_cache": setting.DnsNegativeCache,
		},
		"proxy_map":     make(map[string]interface{}),
		"bootstrap":     make([]string, 0),
		"bootstrap_dns": bootstrapDns,
		// The module's own upstream sockets bind here on Windows and macOS,
		// where there is no socket mark to keep them out of the TUN.
		"egress_interface": TunEgressInterfaceIfTun(setting),
		"upstreams":        make([]map[string]interface{}, 0),
		"rules":            make([]map[string]interface{}, 0),
	}

	// 应用默认值
	cache := cfg["cache"].(map[string]interface{})
	if cache["size"].(int) <= 0 {
		cache["size"] = 4096
	}
	if cache["min_ttl"].(int) <= 0 {
		cache["min_ttl"] = 60
	}
	if cache["max_ttl"].(int) <= 0 {
		cache["max_ttl"] = 86400
	}

	// 获取并迁移 DNS 规则
	rules := configure.GetDnsRulesNotNil()
	migrated := configure.MigrateDnsRules(rules)

	// 扫描所有 DNS 规则中引用的出站标签，构建 proxy_map
	// （仅当标签非 direct/block 时创建映射，指向本地 SOCKS 入站端口）
	proxyMap := cfg["proxy_map"].(map[string]interface{})
	bootstrapList := cfg["bootstrap"].([]string)

	// 用于上游去重的 key
	type upstreamKey struct {
		addr string
		tag  string
	}
	seenUpstream := make(map[upstreamKey]int) // key → index in upstreams list
	upstreams := cfg["upstreams"].([]map[string]interface{})
	rulesList := cfg["rules"].([]map[string]interface{})
	defaultUpstream := ""

	for _, rule := range migrated {
		upstreamAddr := rule.Upstream
		if upstreamAddr == "" {
			upstreamAddr = rule.Server
		}
		if upstreamAddr == "" {
			continue
		}

		// "localhost" 表示本地系统 DNS 解析器，映射为 127.0.0.1
		if upstreamAddr == "localhost" {
			upstreamAddr = "127.0.0.1"
		}

		// 解析协议和地址，确保端口默认值正确
		proto := "udp"
		addr := upstreamAddr
		needDefaultPort := false

		if strings.Contains(upstreamAddr, "://") {
			if strings.HasPrefix(upstreamAddr, "https://") {
				proto = "https"
				// DoH 地址保留完整 URL，端口由 URL 隐含
			} else if strings.HasPrefix(upstreamAddr, "tcp://") {
				proto = "tcp"
				addr = strings.TrimPrefix(upstreamAddr, "tcp://")
				needDefaultPort = true
			} else if strings.HasPrefix(upstreamAddr, "tls://") {
				proto = "tcp-tls"
				addr = strings.TrimPrefix(upstreamAddr, "tls://")
				needDefaultPort = true
			} else if strings.HasPrefix(upstreamAddr, "quic://") {
				proto = "quic"
				addr = strings.TrimPrefix(upstreamAddr, "quic://")
				needDefaultPort = true
			} else {
				addr = upstreamAddr
				needDefaultPort = true
			}
		} else {
			needDefaultPort = true
		}

		// 如果地址中没有端口号，根据协议类型补充默认端口
		if needDefaultPort {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				switch proto {
				case "https":
					// DoH URL 包含完整地址，不需追加端口
				case "tcp-tls", "tls":
					addr = net.JoinHostPort(addr, "853")
				default:
					addr = net.JoinHostPort(addr, "53")
				}
			}
		}

		if endpoint, err := resolv.ParseIPDNS(upstreamAddr); err == nil {
			addr, proto = endpoint.UpstreamAddress(), endpoint.Protocol()
		}

		// 如果是域名地址，加入 bootstrap 列表，由 v2raya-core 用系统 DNS 解析
		bootHost := addr
		if proto == "https" {
			if u, err := url.Parse(upstreamAddr); err == nil {
				bootHost = u.Hostname()
			}
		} else if host, _, err := net.SplitHostPort(addr); err == nil {
			bootHost = host
		}
		if bootHost != "" && net.ParseIP(bootHost) == nil {
			bootstrapList = append(bootstrapList, bootHost)
		}

		outboundTag := rule.Outbound
		if outboundTag == "" {
			outboundTag = "direct"
		}

		// 构建 proxy_map：为每个非直连的出站标签创建 SOCKS5 映射
		if outboundTag != "direct" && outboundTag != "block" {
			if _, exists := proxyMap[outboundTag]; !exists {
				proxyMap[outboundTag] = fmt.Sprintf("127.0.0.1:%d", socksPort)
			}
		}

		// 上游去重（相同地址+代理标签的上游只创建一个实例）
		key := upstreamKey{addr: addr, tag: outboundTag}
		upstreamID := fmt.Sprintf("upstream-%d", len(upstreams))
		if idx, ok := seenUpstream[key]; ok {
			upstreamID = fmt.Sprintf("upstream-%d", idx)
		} else {
			seenUpstream[key] = len(upstreams)
			upstreams = append(upstreams, map[string]interface{}{
				"id":        upstreamID,
				"addr":      addr,
				"protocol":  proto,
				"proxy_tag": outboundTag,
				"bootstrap": false,
			})
		}

		// 解析域名匹配规则
		var domains, suffixes []string
		domainStr := rule.Domain
		if domainStr == "" {
			domainStr = rule.Domains
		}
		if domainStr != "" {
			for _, d := range strings.Split(strings.TrimSpace(domainStr), "\n") {
				d = strings.TrimSpace(d)
				if d == "" {
					continue
				}
				if strings.HasPrefix(d, "domain:") {
					suffixes = append(suffixes, strings.TrimPrefix(d, "domain:"))
				} else if strings.HasPrefix(d, "geosite:") || strings.HasPrefix(d, "keyword:") {
					domains = append(domains, d)
				} else {
					domains = append(domains, d)
				}
			}
		}

		// 解析 IP 匹配
		var ips []string
		if rule.IP != "" {
			for _, ip := range strings.Split(rule.IP, ",") {
				ip = strings.TrimSpace(ip)
				if ip != "" {
					ips = append(ips, ip)
				}
			}
		}

		// 解析客户端 IP
		var clientIPs []string
		if rule.ClientIP != "" {
			for _, cidr := range strings.Split(rule.ClientIP, ",") {
				cidr = strings.TrimSpace(cidr)
				if cidr != "" {
					clientIPs = append(clientIPs, cidr)
				}
			}
		}

		// 构建规则配置（仅在有匹配条件时）
		if len(domains) > 0 || len(suffixes) > 0 || len(clientIPs) > 0 || len(ips) > 0 {
			ruleID := rule.RuleID
			if ruleID == "" {
				ruleID = fmt.Sprintf("rule-%d", len(rulesList))
			}

			action := rule.Action
			if action == "" {
				action = "route"
			}

			policy := rule.Policy
			if policy == "" {
				policy = "single"
			}

			rc := map[string]interface{}{
				"id":            ruleID,
				"upstream":      upstreamID,
				"action":        action,
				"policy":        policy,
				"domain":        domains,
				"domain_suffix": suffixes,
				"ip":            ips,
				"client_ip":     clientIPs,
			}

			// 解析查询类型
			if rule.QueryType != "" && rule.QueryType != "*" {
				var qtypes []string
				for _, qt := range strings.Split(rule.QueryType, ",") {
					qt = strings.TrimSpace(qt)
					if qt != "" {
						qtypes = append(qtypes, strings.ToUpper(qt))
					}
				}
				rc["query_type"] = qtypes
			}

			rulesList = append(rulesList, rc)
		} else {
			defaultUpstream = upstreamID
		}
	}
	// Without a catch-all rule the module keeps its own fallback (the last
	// upstream in the list), which is what every earlier release did.

	// 节点服务器域名必须直连解析：若命中默认（代理）上游，查询经 dispatcher
	// 路由到代理出站，而代理出站连接节点又需要解析节点域名，形成死锁，
	// 导致节点域名永远解析失败、代理通道完全不可用。
	var nodeDomains []string
	for _, info := range serverInfos {
		host := info.Info.GetHostname()
		if host != "" && net.ParseIP(host) == nil {
			nodeDomains = append(nodeDomains, host)
		}
	}
	nodeDomains = common.Deduplicate(nodeDomains)
	endpoint := t.NodeDNS
	if endpoint == nil {
		var err error
		endpoint, err = SelectNodeDNS(setting, rules)
		if err != nil {
			return err
		}
		t.NodeDNS = endpoint
	}
	// 不能 append 到末尾：模块以最后一个非 bootstrap 上游作为默认上游
	upstreams = append([]map[string]interface{}{{
		"id":          "upstream-node",
		"addr":        endpoint.UpstreamAddress(),
		"protocol":    endpoint.Protocol(),
		"server_name": endpoint.IP.String(),
		"proxy_tag":   "direct",
		"bootstrap":   false,
	}}, upstreams...)
	rulesList = append([]map[string]interface{}{{
		"id":            "rule-node",
		"upstream":      "upstream-node",
		"action":        "route",
		"policy":        "single",
		"domain":        nil,
		"domain_suffix": nodeDomains,
		"ip":            nil,
		"client_ip":     nil,
	}}, rulesList...)

	cfg["upstreams"] = upstreams
	cfg["rules"] = rulesList
	cfg["default_upstream"] = defaultUpstream

	// bootstrap 列表去重
	bootstrapList = common.Deduplicate(bootstrapList)
	cfg["bootstrap"] = bootstrapList

	// 序列化为 JSON
	raw, err := jsoniter.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal dns module config: %w", err)
	}
	t.DnsModuleConfig = json.RawMessage(raw)

	return t.setNodeDNS(nodeDomains)
}

// getSystemDnsServers 读取当前系统的 DNS 服务器列表（从 /etc/resolv.conf）。
// 在劫持发生前调用，保存原始 DNS 供 v2raya-core bootstrap 使用。
func getSystemDnsServers() []string {
	servers, _ := readOriginalResolv(resolvPath, resolvBackupPath)
	return servers
}

func readOriginalResolv(path, backup string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(string(data), HijackFlag) {
		data, err = os.ReadFile(backup)
		if err != nil {
			return nil, err
		}
		content := strings.TrimSpace(string(data))
		switch {
		case strings.HasPrefix(content, symlinkMarker):
			target := strings.TrimSpace(strings.TrimPrefix(content, symlinkMarker))
			if target == "" {
				return nil, fmt.Errorf("empty resolver symlink backup")
			}
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(path), target)
			}
			data, err = os.ReadFile(target)
			if err != nil {
				return nil, err
			}
		case content == missingMarker || content == emptyMarker:
			return nil, nil
		case content == "":
			return nil, fmt.Errorf("empty resolver backup")
		}
	}
	if strings.HasPrefix(string(data), HijackFlag) {
		return nil, fmt.Errorf("resolver backup is hijacked")
	}
	var servers []string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "nameserver" {
			if ip := net.ParseIP(fields[1]); ip != nil {
				servers = append(servers, net.JoinHostPort(ip.String(), "53"))
			}
		}
	}
	return common.Deduplicate(servers), nil
}

// probeListen reports whether both the UDP and the TCP side of addr can be
// bound right now.
func probeListen(addr string) error {
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	pc.Close()
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	l.Close()
	return nil
}
