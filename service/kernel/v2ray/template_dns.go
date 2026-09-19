package v2ray

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/iptables"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

type Addr struct {
	host string
	port string
	udp  bool
}

func parseDnsAddr(addr string) Addr {
	// 223.5.5.5
	if net.ParseIP(addr) != nil {
		return Addr{
			host: addr,
			port: "53",
			udp:  true,
		}
	}
	// dns.google:53
	if host, port, err := net.SplitHostPort(addr); err == nil {
		if _, err = strconv.Atoi(port); err == nil {
			return Addr{
				host: host,
				port: port,
				udp:  true,
			}
		}
	}
	// tcp://8.8.8.8:53, https://dns.google/dns-query, quic://dns.nextdns.io
	if strings.Contains(addr, "://") {
		if u, err := url.Parse(addr); err == nil {
			udp := false
			if u.Scheme == "quic" {
				udp = true
			}
			return Addr{
				host: u.Hostname(),
				port: u.Port(),
				udp:  udp,
			}
		}
	}
	// dns.google, dns.pub, etc.
	return Addr{
		host: addr,
		port: "53",
		udp:  true,
	}
}

type DnsRouting struct {
	DirectDomains []Addr
	ProxyDomains  []Addr
	DirectIPs     []Addr
	ProxyIPs      []Addr
}

// setDNS 生成新 DNS 模块的配置，嵌入 xray JSON 供 v2raya-core 读取。
// v2raya-core 启动时解析此配置并启动独立 DNS 监听器，v2rayA 不参与 DNS 查询处理。
func (t *Template) setDNS(serverInfos []serverInfo) error {
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
			if err := probeListen("127.0.0.1:53"); err != nil {
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
	// 此时 /etc/resolv.conf 尚未被劫持，读取的是真实的系统 DNS。
	bootstrapDns := getSystemDnsServers()

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
		"egress_interface": tunEgressInterfaceIfTun(setting),
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

		// 如果是域名地址，加入 bootstrap 列表，由 v2raya-core 用系统 DNS 解析
		if !strings.Contains(upstreamAddr, "://") {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				host = addr
			}
			if net.ParseIP(host) == nil {
				bootstrapList = append(bootstrapList, host)
			}
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
	if len(nodeDomains) > 0 {
		nodeUpstreamAddr := "223.5.5.5:53"
		for _, s := range bootstrapDns {
			if host, _, err := net.SplitHostPort(s); err == nil {
				if ip := net.ParseIP(host); ip != nil && !ip.IsLoopback() {
					nodeUpstreamAddr = s
					break
				}
			}
		}
		// 不能 append 到末尾：模块以最后一个非 bootstrap 上游作为默认上游
		upstreams = append([]map[string]interface{}{{
			"id":        "upstream-node",
			"addr":      nodeUpstreamAddr,
			"protocol":  "udp",
			"proxy_tag": "direct",
			"bootstrap": false,
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
	}

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

	return nil
}

// getSystemDnsServers 读取当前系统的 DNS 服务器列表（从 /etc/resolv.conf）。
// 在劫持发生前调用，保存原始 DNS 供 v2raya-core bootstrap 使用。
func getSystemDnsServers() []string {
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil
	}
	var servers []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nameserver") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ip := parts[1]
				if net.ParseIP(ip) != nil {
					servers = append(servers, net.JoinHostPort(ip, "53"))
				}
			}
		}
	}
	return servers
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
