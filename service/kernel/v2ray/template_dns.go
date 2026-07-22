package v2ray

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
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
func (t *Template) setDNS() error {
	return t.generateDnsModuleConfig()
}

// generateDnsModuleConfig 生成新 DNS 模块的 JSON 配置，嵌入 xray JSON 配置文件。
// v2raya-core 启动时解析此配置并启动独立 DNS 监听器，v2rayA 不参与 DNS 查询处理。
//
// 生成的配置结构对应 core/dns/config.go 中的 DnsModuleConfig。
func (t *Template) generateDnsModuleConfig() error {
	setting := t.Setting
	if setting == nil {
		setting = configure.GetSettingNotNil()
	}

	listenAddr := setting.DnsListenAddr
	if listenAddr == "" {
		listenAddr = "0.0.0.0:52353"
	}

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
			"extra_listen_addrs": []string{"127.2.0.17:53"},
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
		"upstreams":     make([]map[string]interface{}, 0),
		"rules":         make([]map[string]interface{}, 0),
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
		}
	}

	cfg["upstreams"] = upstreams
	cfg["rules"] = rulesList

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
