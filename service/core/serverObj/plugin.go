package serverObj

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/coreObj"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"os/exec"
	"strconv"
	"strings"
)

var ErrUnsupportedPlugin = fmt.Errorf("unsupported plugin")

const PluginManagerScheme = "plugin-manager"

func init() {
	FromLinkRegister(PluginManagerScheme, NewPlugin)
	EmptyRegister(PluginManagerScheme, func() (ServerObj, error) {
		return new(Plugin), nil
	})
}

type Plugin struct {
	Name           string `json:"pm_name"`
	Host           string `json:"pm_add"`
	Port           string `json:"pm_port"`
	ProtocolToShow string `json:"pm_protocol"`
	Link           string `json:"pm_link"`

	Protocol string `json:"protocol"`
}

func NewPlugin(link string) (ServerObj, error) {
	if pm := conf.GetEnvironmentConfig().PluginManager; pm != "" {
		b, err := exec.Command(pm,
			"--stage=parse",
			fmt.Sprintf("--link=%v", link),
			fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config),
		).CombinedOutput()
		if err != nil {
			log.Info("[PluginManager] failed to parse %v: %v", link, string(b))
			return nil, err
		}
		nExpected := 4
		fields := strings.SplitN(strings.TrimSpace(string(b)), ";", nExpected)
		if len(fields) < nExpected {
			return nil, fmt.Errorf("[PluginManager] insufficient return value: %v. Expected number: %v", fields, nExpected)
		}
		return &Plugin{
			Name:           fields[0],
			Host:           fields[1],
			Port:           fields[2],
			ProtocolToShow: fields[3],
			Link:           link,
			Protocol:       PluginManagerScheme,
		}, nil
	}
	return nil, ErrUnsupportedPlugin
}

func (v *Plugin) Configuration(info PriorInfo) (c Configuration, err error) {
	if pm := conf.GetEnvironmentConfig().PluginManager; pm != "" {
		b, err := exec.Command(pm,
			"--stage=configuration",
			fmt.Sprintf("--link=%v", v.Link),
			fmt.Sprintf("--port=%v", info.PluginPort),
			fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config),
		).CombinedOutput()
		sOut := strings.TrimSpace(string(b))
		if err != nil {
			return Configuration{}, fmt.Errorf("[PluginManager] Error when Configuration: %w: %v", err, sOut)
		}
		var core coreObj.OutboundObject
		if err := jsoniter.UnmarshalFromString(sOut, &core); err != nil {
			return Configuration{}, fmt.Errorf("[PluginManager] Error when Configuration: failed to parse json output")
		}
		core.Tag = info.Tag
		return Configuration{
			CoreOutbound:            core,
			PluginChain:             "",
			UDPSupport:              true,
			PluginManagerServerLink: v.Link,
		}, nil
	} else {
		return Configuration{}, ErrUnsupportedPlugin
	}
}

func (v *Plugin) ExportToURL() string {
	if v.Link == "" {
		log.Warn("You may need to re-import the node for PluginManager to take effect")
	}
	return v.Link
}

func (v *Plugin) NeedPluginPort() bool {
	return true
}

func (v *Plugin) ProtoToShow() string {
	return v.ProtocolToShow + "(plugin)"
}

func (v *Plugin) GetProtocol() string {
	return v.ProtocolToShow
}

func (v *Plugin) GetHostname() string {
	return v.Host
}

func (v *Plugin) GetPort() int {
	p, _ := strconv.Atoi(v.Port)
	return p
}

func (v *Plugin) GetName() string {
	return v.Name
}

func (v *Plugin) SetName(name string) {
	v.Name = name
}
