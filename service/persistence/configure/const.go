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
	Antipollution           string
)

const (
	TransparentClose     = TransparentMode("close")
	TransparentProxy     = TransparentMode("proxy")
	TransparentWhitelist = TransparentMode("whitelist")
	TransparentGfwlist   = TransparentMode("gfwlist")

	Default = DefaultYesNo("default")
	Yes     = DefaultYesNo("yes")
	No      = DefaultYesNo("no")

	NotAutoUpdate = AutoUpdateMode("none")
	AutoUpdate    = AutoUpdateMode("auto_update")

	ProxyModeDirect = ProxyMode("direct")
	ProxyModePac    = ProxyMode("pac")
	ProxyModeProxy  = ProxyMode("proxy")

	WhitelistMode = PacMode("whitelist")
	GfwlistMode   = PacMode("gfwlist")
	CustomMode    = PacMode("custom")

	DirectRule = PacRuleType("direct")
	ProxyRule  = PacRuleType("proxy")
	BlockRule  = PacRuleType("block")

	DomainMatchRule = PacMatchType("domain")
	IpMatchRule     = PacMatchType("ip")

	DefaultDirectMode = RoutingDefaultProxyMode("direct")
	DefaultProxyMode  = RoutingDefaultProxyMode("proxy")

	SubscriptionType       = TouchType("subscription")
	ServerType             = TouchType("server")
	SubscriptionServerType = TouchType("subscriptionServer")

	DnsForward        = Antipollution("dnsforward")
	DoH               = Antipollution("doh")
	AntipollutionNone = Antipollution("none")
)
