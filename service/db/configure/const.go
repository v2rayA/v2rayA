package configure

type (
	AutoUpdateMode          string
	ProxyMode               string
	RulePortMode            string
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
	TransparentClose      = TransparentMode("close")
	TransparentProxy      = TransparentMode("proxy") // proxy all traffic
	TransparentWhitelist  = TransparentMode("whitelist")
	TransparentGfwlist    = TransparentMode("gfwlist")
	TransparentFollowRule = TransparentMode("pac")

	TransparentTproxy   = TransparentType("tproxy")
	TransparentRedirect = TransparentType("redirect")
	TransparentSystemProxy = TransparentType("system_proxy")

	Default = DefaultYesNo("default")
	Yes     = DefaultYesNo("yes")
	No      = DefaultYesNo("no")

	NotAutoUpdate         = AutoUpdateMode("none")
	AutoUpdate            = AutoUpdateMode("auto_update")
	AutoUpdateAtIntervals = AutoUpdateMode("auto_update_at_intervals")

	ProxyModeDirect = ProxyMode("direct")
	ProxyModePac    = ProxyMode("pac")
	ProxyModeProxy  = ProxyMode("proxy")

	WhitelistMode = RulePortMode("whitelist")
	GfwlistMode   = RulePortMode("gfwlist")
	CustomMode    = RulePortMode("custom")
	RoutingAMode  = RulePortMode("routingA")

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
	AntipollutionAntiHijack = Antipollution("none") // 历史原因，none代表“仅防止dns劫持”，不代表关闭
	AntipollutionClosed     = Antipollution("closed")
	AntipollutionAdvanced   = Antipollution("advanced") // 自定义

	SpecialModeNone       = SpecialMode("none")
	SpecialModeSupervisor = SpecialMode("supervisor")
	SpecialModeFakeDns    = SpecialMode("fakedns")
)

const (
	RoutingATemplate = `default: proxy
# write your own rules below
domain(domain:mail.qq.com)->direct

domain(geosite:google-scholar)->proxy
domain(geosite:category-scholar-!cn, geosite:category-scholar-cn)->direct
domain(geosite:geolocation-!cn, geosite:google)->proxy
domain(geosite:cn)->direct
ip(geoip:hk,geoip:mo)->proxy
ip(geoip:private, geoip:cn)->direct`
)