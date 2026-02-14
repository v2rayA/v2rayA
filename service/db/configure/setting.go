package configure

import (
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/ipforward"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

type Setting struct {
	LogLevel                           string          `json:"logLevel"`
	RulePortMode                       RulePortMode    `json:"pacMode"`
	ProxyModeWhenSubscribe             ProxyMode       `json:"proxyModeWhenSubscribe"`
	GFWListAutoUpdateMode              AutoUpdateMode  `json:"pacAutoUpdateMode"`
	GFWListAutoUpdateIntervalHour      int             `json:"pacAutoUpdateIntervalHour"`
	SubscriptionAutoUpdateMode         AutoUpdateMode  `json:"subscriptionAutoUpdateMode"`
	SubscriptionAutoUpdateIntervalHour int             `json:"subscriptionAutoUpdateIntervalHour"`
	TcpFastOpen                        DefaultYesNo    `json:"tcpFastOpen"`
	MuxOn                              DefaultYesNo    `json:"muxOn"`
	Mux                                int             `json:"mux"`
	InboundSniffing                    InboundSniffing `json:"inboundSniffing"`
	Transparent                        TransparentMode `json:"transparent"`
	IpForward                          bool            `json:"ipforward"`
	RouteOnly                          bool            `json:"routeOnly"`
	PortSharing                        bool            `json:"portSharing"`
	SpecialMode                        SpecialMode     `json:"specialMode"`
	TransparentType                    TransparentType `json:"transparentType"`
	AntiPollution                      Antipollution   `json:"antipollution"`
	TunFakeIP                          bool            `json:"tunFakeIP"`
	TunIPv6                            bool            `json:"tunIPv6"`
	TunStrictRoute                     bool            `json:"tunStrictRoute"`
}

func NewSetting() (setting *Setting) {
	return &Setting{
		LogLevel:                           "info",
		RulePortMode:                       WhitelistMode,
		ProxyModeWhenSubscribe:             ProxyModeDirect,
		GFWListAutoUpdateMode:              NotAutoUpdate,
		GFWListAutoUpdateIntervalHour:      0,
		SubscriptionAutoUpdateMode:         NotAutoUpdate,
		SubscriptionAutoUpdateIntervalHour: 0,
		TcpFastOpen:                        Default,
		MuxOn:                              No,
		Mux:                                8,
		InboundSniffing:                    "http,tls,quic",
		Transparent:                        TransparentClose,
		IpForward:                          ipforward.IsIpForwardOn(),
		PortSharing:                        false,
		SpecialMode:                        SpecialModeNone,
		TransparentType:                    TransparentRedirect,
		AntiPollution:                      AntipollutionClosed,
		TunFakeIP:                          true,
		TunIPv6:                            false,
		TunStrictRoute:                     false,
	}
}

func (s *Setting) FillEmpty() {
	if err := common.FillEmpty(s, GetSettingNotNil()); err != nil {
		log.Warn("FillEmpty: %v:", err)
	}
}

type CustomPac struct {
	DefaultProxyMode RoutingDefaultProxyMode `json:"defaultProxyMode"` //默认路由规则, proxy还是direct
	RoutingRules     []RoutingRule           `json:"routingRules"`
}

// v2rayTmpl.RoutingRule的前端友好版本
type RoutingRule struct {
	Filename  string       `json:"filename"`  //SiteDAT文件名
	Tags      []string     `json:"tags"`      //SiteDAT文件的标签
	MatchType PacMatchType `json:"matchType"` //是domain匹配还是ip匹配
	RuleType  PacRuleType  `json:"ruleType"`  //在名单上的项进行直连、代理还是拦截
}
