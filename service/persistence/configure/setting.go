package configure

import (
	"V2RayA/model/vmessInfo"
	"time"
)

type Setting struct {
	PacMode                    PacMode                `json:"pacMode"`
	CustomPac                  CustomPac              `json:"customPac"`
	ProxyModeWhenSubscribe     ProxyModeWhenSubscribe `json:"proxyModeWhenSubscribe"`
	PacAutoUpdateMode          AutoUpdateMode         `json:"pacAutoUpdateMode"`
	PacAutoUpdateTime          int                    `json:"pacAutoUpdateTime"` //时间戳
	SubscriptionAutoUpdateMode AutoUpdateMode         `json:"subscriptionAutoUpdateMode"`
	SubscriptionAutoUpdateTime int                    `json:"subscriptionAutoUpdateTime"` //时间戳
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
		PacAutoUpdateMode:          DoNotUpdatePac,
		PacAutoUpdateTime:          int(21 * time.Hour / time.Millisecond), //凌晨5点
		SubscriptionAutoUpdateMode: DoNotUpdatePac,
		SubscriptionAutoUpdateTime: int(21 * time.Hour / time.Millisecond), //凌晨5点
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
