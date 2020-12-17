package configure

import "github.com/v2rayA/v2rayA/core/ipforward"

type Setting struct {
	PacMode                    PacMode         `json:"pacMode"`
	ProxyModeWhenSubscribe     ProxyMode       `json:"proxyModeWhenSubscribe"`
	PacAutoUpdateMode          AutoUpdateMode  `json:"pacAutoUpdateMode"`
	SubscriptionAutoUpdateMode AutoUpdateMode  `json:"subscriptionAutoUpdateMode"`
	TcpFastOpen                DefaultYesNo    `json:"tcpFastOpen"`
	MuxOn                      DefaultYesNo    `json:"muxOn"`
	Mux                        int             `json:"mux"`
	Transparent                TransparentMode `json:"transparent"`
	IpForward                  bool            `json:"ipforward"`
	EnhancedMode               bool            `json:"enhancedMode"`
	AntiPollution              Antipollution   `json:"antipollution"`
	DnsForceMode               bool            `json:"dnsForceMode"`
}

func NewSetting() (setting *Setting) {
	return &Setting{
		PacMode:                    WhitelistMode,
		ProxyModeWhenSubscribe:     ProxyModeDirect,
		PacAutoUpdateMode:          NotAutoUpdate,
		SubscriptionAutoUpdateMode: NotAutoUpdate,
		TcpFastOpen:                Default,
		MuxOn:                      No,
		Mux:                        8,
		Transparent:                TransparentClose,
		IpForward:                  ipforward.IsIpForwardOn(),
		EnhancedMode:               false,
		AntiPollution:              AntipollutionClosed,
		DnsForceMode:               false,
	}

}

type CustomPac struct {
	DefaultProxyMode RoutingDefaultProxyMode `json:"defaultProxyMode"` //默认路由规则, proxy还是direct
	RoutingRules     []RoutingRule           `json:"routingRules"`
}

//v2rayTmpl.RoutingRule的前端友好版本
type RoutingRule struct {
	Filename  string       `json:"filename"`  //SiteDAT文件名
	Tags      []string     `json:"tags"`      //SiteDAT文件的标签
	MatchType PacMatchType `json:"matchType"` //是domain匹配还是ip匹配
	RuleType  PacRuleType  `json:"ruleType"`  //在名单上的项进行直连、代理还是拦截
}
