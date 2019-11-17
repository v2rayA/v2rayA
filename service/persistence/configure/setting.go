package configure

import (
	"V2RayA/model/vmessInfo"
)

type Setting struct {
	PacMode                    PacMode                `json:"pacMode"`
	CustomPac                  CustomPac              `json:"customPac"`
	ProxyModeWhenSubscribe     ProxyModeWhenSubscribe `json:"proxyModeWhenSubscribe"`
	PacAutoUpdateMode          AutoUpdateMode         `json:"pacAutoUpdateMode"`
	SubscriptionAutoUpdateMode AutoUpdateMode         `json:"subscriptionAutoUpdateMode"`
	TcpFastOpen                DefaultYesNo           `json:"tcpFastOpen"`
	MuxOn                      DefaultYesNo           `json:"muxOn"`
	Mux                        int                    `json:"mux"`
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

type TouchServerRaw struct {
	VmessInfo vmessInfo.VmessInfo `json:"vmessInfo"`
}

type SubscriptionRaw struct {
	Address string           `json:"address"`
	Status  string           `json:"status"` //update time, error info, etc.
	Servers []TouchServerRaw `json:"servers"`
}
