package serverObj

import (
	"fmt"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/coreObj"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"net/url"
)

var InvalidParameterErr = fmt.Errorf("invalid parameters")

type ServerObj interface {
	Configuration(info PriorInfo) (c Configuration, err error)
	ExportToURL() string
	NeedPluginPort() bool
	ProtoToShow() string
	GetProtocol() string
	GetHostname() string
	GetPort() int
	GetName() string
	SetName(name string)
}

type Configuration struct {
	CoreOutbound            coreObj.OutboundObject
	ExtraOutbounds          []coreObj.OutboundObject
	PluginChain             string // The first is a server plugin, and the others are client plugins. Split by ",".
	UDPSupport              bool
	PluginManagerServerLink string
}

type PriorInfo struct {
	Variant     where.Variant
	CoreVersion string
	Tag         string
	PluginPort  int
}

func (info *PriorInfo) PluginObj() coreObj.OutboundObject {
	return coreObj.OutboundObject{
		Tag:      info.Tag,
		Protocol: "socks",
		Settings: coreObj.Settings{
			Servers: []coreObj.Server{
				{
					Address: "127.0.0.1",
					Port:    info.PluginPort,
				},
			}},
	}
}

type FromLinkCreator func(link string) (ServerObj, error)
type EmptyCreator func() (ServerObj, error)

var fromLinkCreators = make(map[string]FromLinkCreator)
var emptyCreators = make(map[string]EmptyCreator)

func FromLinkRegister(name string, creator FromLinkCreator) {
	fromLinkCreators[name] = creator
}
func EmptyRegister(name string, creator EmptyCreator) {
	emptyCreators[name] = creator
}
func New(name string) (ServerObj, error) {
	if creator, ok := emptyCreators[name]; ok {
		return creator()
	} else if pm := conf.GetEnvironmentConfig().PluginManager; pm != "" {
		// we do not support to override build-in protocols
		creator := emptyCreators[PluginManagerScheme]
		return creator()
	} else {
		return nil, fmt.Errorf("unsupported link type: %v", name)
	}
}
func NewFromLink(name string, link string) (ServerObj, error) {
	if creator, ok := fromLinkCreators[name]; ok {
		return creator(link)
	} else if pm := conf.GetEnvironmentConfig().PluginManager; pm != "" {
		// we do not support to override build-in protocols
		creator := fromLinkCreators[PluginManagerScheme]
		return creator(link)
	} else {
		return nil, fmt.Errorf("unsupported link type: %v", name)
	}
}

func setValue(values *url.Values, key string, value string) {
	if value == "" {
		return
	}
	values.Set(key, value)
}
