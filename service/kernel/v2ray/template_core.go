package v2ray

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/kernel/v2ray/where"
)

type Template struct {
	Log       *coreObj.Log             `json:"log,omitempty"`
	Inbounds  []coreObj.Inbound        `json:"inbounds"`
	Outbounds []coreObj.OutboundObject `json:"outbounds"`
	Routing   struct {
		DomainStrategy string                `json:"domainStrategy"`
		DomainMatcher  string                `json:"domainMatcher,omitempty"`
		Rules          []coreObj.RoutingRule `json:"rules"`
		Balancers      []coreObj.Balancer    `json:"balancers,omitempty"`
	} `json:"routing"`
	DNS              *coreObj.DNS              `json:"dns,omitempty"`
	MultiObservatory *coreObj.MultiObservatory `json:"multiObservatory,omitempty"`
	Observatory      *coreObj.ObservatoryItem  `json:"observatory,omitempty"`
	API              *coreObj.APIObject        `json:"api,omitempty"`
	// DnsModuleConfig 是新 DNS 模块的配置，由 v2raya-core 启动时解析并启动 DNS 监听器。
	DnsModuleConfig json.RawMessage `json:"dns_module,omitempty"`

	Variant       where.Variant          `json:"-"`
	CoreVersion   string                 `json:"-"`
	OutboundTags  []string               `json:"-"`
	ApiCloses     []func()               `json:"-"`
	ApiPort       int                    `json:"-"`
	Setting       *configure.Setting     `json:"-"`
	serverInfoMap map[string]*serverInfo `json:"-"` // outbound tag -> server info
}

func (t *Template) Close() error {
	for _, f := range t.ApiCloses {
		f()
	}
	return nil
}

func NewTemplate(serverInfos []serverInfo, setting *configure.Setting) (t *Template, err error) {
	serverData := NewServerData(serverInfos)
	if setting != nil {
		setting.FillEmpty()
	} else {
		setting = configure.GetSettingNotNil()
	}

	var tmplJson Template
	// read template json
	raw := []byte(TemplateJson)
	err = jsoniter.Unmarshal(raw, &tmplJson)
	if err != nil {
		return nil, fmt.Errorf("error occurs while reading template json, please check whether templateJson variable is correct json format")
	}
	tmplJson.Variant, tmplJson.CoreVersion, _ = where.GetV2rayServiceVersion()
	t = &tmplJson
	t.Setting = setting
	// log
	logLevel := setting.LogLevel
	if logLevel == "" {
		logLevel = conf.GetEnvironmentConfig().LogLevel
	}
	logLevel = strings.ToLower(logLevel)
	t.Log = new(coreObj.Log)
	switch logLevel {
	case "trace", "debug":
		t.Log.Loglevel = "debug"
		t.Log.Access = ""
		t.Log.Error = ""
	case "info":
		t.Log.Loglevel = "info"
		t.Log.Access = ""
		t.Log.Error = "none"
	case "warn", "warning":
		t.Log.Loglevel = "warning"
		t.Log.Access = "none"
		t.Log.Error = ""
	case "error":
		t.Log.Loglevel = "error"
		t.Log.Access = "none"
		t.Log.Error = ""
	default:
		t.Log = nil
	}
	// resolve Outbounds
	_, outboundTags, err := t.resolveOutbounds(serverData)
	if err != nil {
		return nil, err
	}
	t.OutboundTags = outboundTags

	//set inbound ports according to the setting
	if err = t.setInbound(setting); err != nil {
		return nil, err
	}
	// When TinyTun is active it handles DNS routing natively; v2ray only forwards traffic.
	// Skip the DNS module for that mode.
	isTinyTunMode := setting.TransparentType == configure.TransparentTun && IsTransparentOn(setting)
	if !isTinyTunMode {
		//生成新 DNS 模块配置
		if err = t.setDNS(); err != nil {
			return nil, err
		}
	}
	// 路由域名匹配器
	t.Routing.DomainMatcher = "mph"
	//rule port routing
	if err = t.setRulePortRouting(); err != nil {
		return nil, err
	}
	//transparent routing
	if IsTransparentOn(setting) {
		if err = t.setTransparentRouting(); err != nil {
			return nil, err
		}
	}
	//set vmess inbound routing
	t.setVmessInboundRouting()
	// set api
	if t.API == nil {
		if _, err = t.SetAPI(serverData); err != nil {
			return nil, err
		}
	}
	// set routing whitelist
	var whitelist []Addr
	for _, info := range serverInfos {
		port := ""
		if info.Info.GetPort() != 0 {
			port = strconv.Itoa(info.Info.GetPort())
		}
		whitelist = append(whitelist, Addr{
			host: info.Info.GetHostname(),
			port: port,
		})
	}
	t.setWhitelistRouting(whitelist)

	t.updatePrivateRouting()

	// add spare tire outbound routing. Fix: https://github.com/v2rayA/v2rayA/issues/447
	t.Routing.Rules = append(t.Routing.Rules, coreObj.RoutingRule{Type: "field", Port: "0-65535", OutboundTag: "proxy"})

	// Set group routing. This should be put in the end of routing setters.
	t.setGroupRouting()

	t.optimizeGeoipMemoryOccupation()

	//set outboundSockopt
	t.SetOutboundSockopt()

	//set inbound listening address and routing
	t.setDualStack()

	//check if there are any duplicated tags
	if err = t.checkDuplicatedTags(); err != nil {
		return nil, err
	}
	//check if there are any duplicated inbound ports
	if err = t.checkDuplicatedInboundSockets(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Template) ToConfigBytes() []byte {
	b, _ := jsoniter.Marshal(t)
	return b
}

func WriteV2rayConfig(content []byte) (err error) {
	err = os.WriteFile(asset.GetV2rayConfigPath(), content, os.FileMode(0600))
	if err != nil {
		return fmt.Errorf("WriteV2rayConfig: %w", err)
	}
	return
}

func NewEmptyTemplate(setting *configure.Setting) (t *Template) {
	t = new(Template)
	t.Variant, t.CoreVersion, _ = where.GetV2rayServiceVersion()
	if setting != nil {
		setting.FillEmpty()
	} else {
		setting = configure.GetSettingNotNil()
	}
	t.Setting = setting
	return t
}
