package configure

import (
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/kernel/ipforward"
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
	TransparentType                    TransparentType `json:"transparentType"`
	TproxyExcludedInterfaces           string          `json:"tproxyExcludedInterfaces"`
	TunBypassInterfaces                string          `json:"tunBypassInterfaces"`
	TunAutoRoute                       bool            `json:"tunAutoRoute"`
	TunRouteShellType                  string          `json:"tunRouteShellType"`
	TunRouteShellPath                  string          `json:"tunRouteShellPath"`
	TunSetupScript                     string          `json:"tunSetupScript"`
	TunTeardownScript                  string          `json:"tunTeardownScript"`
	TunProcessBackend                  string          `json:"tunProcessBackend"`
	TunExcludeProcesses                string          `json:"tunExcludeProcesses"`
	SsBackend                          string          `json:"ssBackend"`
	TrojanBackend                      string          `json:"trojanBackend"`
	// 新 DNS 模块监听配置
	DnsListenAddr string `json:"dnsListenAddr"` // 监听地址，默认 "0.0.0.0:52353"

	// DNS 缓存配置
	DnsCacheEnabled  bool `json:"dnsCacheEnabled"`  // 启用缓存，默认 true
	DnsCacheSize     int  `json:"dnsCacheSize"`     // 缓存大小，默认 4096
	DnsCacheMinTTL   int  `json:"dnsCacheMinTTL"`   // 最小 TTL，默认 60
	DnsCacheMaxTTL   int  `json:"dnsCacheMaxTTL"`   // 最大 TTL，默认 86400
	DnsPrefetch      bool `json:"dnsPrefetch"`      // 启用预取，默认 true
	DnsNegativeCache bool `json:"dnsNegativeCache"` // 启用负缓存，默认 true
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
		TransparentType:                    TransparentRedirect,
		TproxyExcludedInterfaces:           "docker*,veth*,wg*,ppp*,br-*",
		TunAutoRoute:                       true,
		// 新 DNS 模块默认值
		DnsListenAddr:    "0.0.0.0:52353",
		DnsCacheEnabled:  true,
		DnsCacheSize:     4096,
		DnsCacheMinTTL:   60,
		DnsCacheMaxTTL:   86400,
		DnsPrefetch:      true,
		DnsNegativeCache: true,
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
