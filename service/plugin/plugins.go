package plugin

import (
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"strings"
	"sync"
)

var GlobalPlugins Plugins

type Plugin interface {
	Serve(localPort int, v vmessInfo.VmessInfo) (err error)
	Close() (err error)
	SupportUDP() bool //TODO: support udp
}

type Plugins struct {
	Plugins []Plugin
	mutex   sync.Mutex
}

func (r *Plugins) CloseAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, ssr := range r.Plugins {
		_ = ssr.Close()
	}
	r.Plugins = make([]Plugin, 0)
}

func (r *Plugins) Append(plugin Plugin) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if plugin == nil {
		return
	}
	r.Plugins = append(r.Plugins, plugin)
}

type PluginCreator func(localPort int, v vmessInfo.VmessInfo) (plugin Plugin, err error)

var pluginMap = make(map[string]PluginCreator)

func RegisterPlugin(protocol string, pluginCreator PluginCreator) {
	pluginMap[protocol] = pluginCreator
}

func NewPlugin(localPort int, v vmessInfo.VmessInfo) (plugin Plugin, err error) {
	v.Protocol = strings.ToLower(v.Protocol)
	switch v.Protocol {
	case "shadowsocks", "ss":
		if v.Type == "" {
			return nil, nil
		} else if v.Type == "http" || v.Type == "tls" {
			v.Protocol = "simpleobfs"
		}
	}
	creator, ok := pluginMap[v.Protocol]
	if !ok {
		return nil, newError("unregistered protocol ", v.Protocol)
	}
	return creator(localPort, v)
}

func IsProtocolValid(v vmessInfo.VmessInfo) bool {
	switch v.Protocol {
	case "shadowsocks", "ss":
		if v.Type == "" {
			return false
		} else if v.Type == "http" || v.Type == "tls" {
			v.Protocol = "simpleobfs"
		}
	}
	_, ok := pluginMap[v.Protocol]
	return ok
}
