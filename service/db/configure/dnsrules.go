package configure

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/db"
)

// DnsRule 代表一条 DNS 服务器规则。
// Server 是 DNS 服务器地址，例如 "8.8.8.8"、"https://dns.google/dns-query"、"localhost"。
// Domains 是换行分隔的域名模式列表；空字符串表示该服务器作为"兜底"DNS。
// Outbound 是路由出口标签，例如 "direct" 或 "proxy"。
//
// 新字段（向后兼容扩展）：
//   - Upstream: 上游 DNS 地址（与 Server 语义一致，新配置使用）
//   - Domain: 域名匹配模式（与 Domains 语义一致，新配置使用）
//   - IP: IP 匹配（CIDR 格式）
//   - QueryType: DNS 查询类型 A/AAAA/CNAME/TXT/MX/SRV/NS/PTR/SOA/*
//   - ClientIP: 客户端 IP 匹配
//   - Action: route/reject/bypass
//   - Policy: single/parallel/fallback
//   - RuleID: 规则唯一标识
type DnsRule struct {
	// 原始字段（向后兼容）
	Server   string `json:"server"`
	Domains  string `json:"domains"`
	Outbound string `json:"outbound"`

	// 新扩展字段
	ID        int    `json:"id" gorm:"primaryKey"`
	Domain    string `json:"domain"`
	IP        string `json:"ip"`
	QueryType string `json:"queryType"`
	ClientIP  string `json:"clientIP"`
	Upstream  string `json:"upstream"`
	Action    string `json:"action"`
	Policy    string `json:"policy"`
	RuleID    string `json:"ruleId"`
}

// DefaultDnsRules 是内置的默认 DNS 规则。
var DefaultDnsRules = []DnsRule{
	{Server: "localhost", Domains: "geosite:private", Outbound: "direct"},
	{Server: "223.5.5.5", Domains: "geosite:cn", Outbound: "direct"},
	{Server: "8.8.8.8", Domains: "", Outbound: "proxy"},
}

// GetDnsRulesNotNil 返回用户配置的 DNS 规则，若未配置则返回 DefaultDnsRules。
func GetDnsRulesNotNil() []DnsRule {
	b, err := db.GetRaw("system", "dnsRules")
	if err == nil && len(b) > 0 {
		var rules []DnsRule
		if e := jsoniter.Unmarshal(b, &rules); e == nil && len(rules) > 0 {
			return rules
		}
	}
	return DefaultDnsRules
}

// SetDnsRules 将 DNS 规则持久化到数据库。
func SetDnsRules(rules []DnsRule) error {
	return db.Set("system", "dnsRules", rules)
}
