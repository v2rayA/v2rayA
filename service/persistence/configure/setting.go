package configure

type Setting struct {
	PacMode                    PacMode         `json:"pacMode"`
	CustomPac                  CustomPac       `json:"customPac"`
	ProxyModeWhenSubscribe     ProxyMode       `json:"proxyModeWhenSubscribe"`
	PacAutoUpdateMode          AutoUpdateMode  `json:"pacAutoUpdateMode"`
	SubscriptionAutoUpdateMode AutoUpdateMode  `json:"subscriptionAutoUpdateMode"`
	TcpFastOpen                DefaultYesNo    `json:"tcpFastOpen"`
	MuxOn                      DefaultYesNo    `json:"muxOn"`
	Mux                        int             `json:"mux"`
	Transparent                TransparentMode `json:"transparent"` //当透明代理开启时将覆盖端口单独的配置
	IpForward                  bool            `json:"ipforward"`
	DnsForward                 DefaultYesNo    `json:"dnsforward"`
}

func NewSetting() (setting *Setting) {
	return &Setting{
		PacMode: WhitelistMode,
		CustomPac: CustomPac{
			URL:              "",
			DefaultProxyMode: DefaultDirectMode,
			RoutingRules:     []RoutingRule{},
		},
		ProxyModeWhenSubscribe:     ProxyModeDirect,
		PacAutoUpdateMode:          NotAutoUpdate,
		SubscriptionAutoUpdateMode: NotAutoUpdate,
		TcpFastOpen:                Default,
		MuxOn:                      No,
		Mux:                        8,
		Transparent:                TransparentClose,
	}

}

type CustomPac struct {
	URL              string                  `json:"url"`              //SiteDAT文件的URL
	DefaultProxyMode RoutingDefaultProxyMode `json:"defaultProxyMode"` //默认路由规则, proxy还是direct
	RoutingRules     []RoutingRule           `json:"routingRules"`
}

//v2rayTmpl.RoutingRule的前端友好版本
type RoutingRule struct {
	Tags      []string     `json:"tags"`      //SiteDAT文件的标签
	MatchType PacMatchType `json:"matchType"` //是domain匹配还是ip匹配
	RuleType  PacRuleType  `json:"ruleType"`  //在名单上的项进行直连、代理还是拦截
}
