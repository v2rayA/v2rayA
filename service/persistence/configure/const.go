package configure

type (
	AutoUpdateMode          string
	ProxyModeWhenSubscribe  string
	PacMode                 string
	PacRuleType             string
	PacMatchType            string
	RoutingDefaultProxyMode string
	TouchType               string
)

const (
	DoNotUpdatePac = AutoUpdateMode("none")
	UpdatePac      = AutoUpdateMode("auto_update")

	ProxyModeDirect = ProxyModeWhenSubscribe("direct")
	ProxyModePac    = ProxyModeWhenSubscribe("pac")
	ProxyModeProxy  = ProxyModeWhenSubscribe("proxy")

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
)
