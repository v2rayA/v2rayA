package configure

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/db"
)

// DnsRule 代表一条 DNS 服务器规则。
// Server 是 DNS 服务器地址，例如 "8.8.8.8"、"https://dns.google/dns-query"、"localhost"。
// Domains 是换行分隔的域名模式列表；空字符串表示该服务器作为"兜底"DNS。
// Outbound 是路由出口标签，例如 "direct" 或 "proxy"。
type DnsRule struct {
	Server   string `json:"server"`
	Domains  string `json:"domains"`
	Outbound string `json:"outbound"`
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
