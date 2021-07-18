package configure

type (
	AutoUpdateMode          string
	ProxyMode               string
	PacMode                 string
	PacRuleType             string
	PacMatchType            string
	RoutingDefaultProxyMode string
	TouchType               string
	DefaultYesNo            string
	TransparentMode         string
	TransparentType         string
	Antipollution           string
	SpecialMode             string
)

const (
	TransparentClose     = TransparentMode("close")
	TransparentProxy     = TransparentMode("proxy")
	TransparentWhitelist = TransparentMode("whitelist")
	TransparentGfwlist   = TransparentMode("gfwlist")
	TransparentPac       = TransparentMode("pac")

	TransparentTproxy   = TransparentType("tproxy")
	TransparentRedirect = TransparentType("redirect")

	Default = DefaultYesNo("default")
	Yes     = DefaultYesNo("yes")
	No      = DefaultYesNo("no")

	NotAutoUpdate         = AutoUpdateMode("none")
	AutoUpdate            = AutoUpdateMode("auto_update")
	AutoUpdateAtIntervals = AutoUpdateMode("auto_update_at_intervals")

	ProxyModeDirect = ProxyMode("direct")
	ProxyModePac    = ProxyMode("pac")
	ProxyModeProxy  = ProxyMode("proxy")

	WhitelistMode = PacMode("whitelist")
	GfwlistMode   = PacMode("gfwlist")
	CustomMode    = PacMode("custom")
	RoutingAMode  = PacMode("routingA")

	DirectRule = PacRuleType("direct")
	ProxyRule  = PacRuleType("proxy")
	BlockRule  = PacRuleType("block")

	DomainMatchRule = PacMatchType("domain")
	IpMatchRule     = PacMatchType("ip")

	DefaultDirectMode = RoutingDefaultProxyMode("direct")
	DefaultProxyMode  = RoutingDefaultProxyMode("proxy")
	DefaultBlockMode  = RoutingDefaultProxyMode("block")

	SubscriptionType       = TouchType("subscription")
	ServerType             = TouchType("server")
	SubscriptionServerType = TouchType("subscriptionServer")

	AntipollutionDnsForward = Antipollution("dnsforward")
	AntipollutionDoH        = Antipollution("doh")
	AntipollutionAntiHijack = Antipollution("none")     // 历史原因，none代表“仅防止dns劫持”，不代表关闭
	AntipollutionClosed     = Antipollution("closed")
	AntipollutionAdvanced   = Antipollution("advanced") // 自定义

	SpecialModeNone       = SpecialMode("none")
	SpecialModeSupervisor = SpecialMode("supervisor")
	SpecialModeFakeDns    = SpecialMode("fakedns")
)

const (
	RoutingATemplate = `default: proxy

# write your own rules below
domain(geosite:google-scholar)->proxy
domain(geosite:category-scholar-!cn,geosite:category-scholar-cn,domain:qq.com)->direct

ip(geoip:private, geoip:cn)->direct
domain(geosite:cn)->direct`
)
