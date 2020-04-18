package plugins

import (
	"v2rayA/core/vmessInfo"
	"strings"
	"sync"
)

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
	r.Plugins = append(r.Plugins, plugin)
}

type PluginCreator func(localPort int, v vmessInfo.VmessInfo) (plugin Plugin, err error)

var pluginMap = make(map[string]PluginCreator)

func RegisterPlugin(protocol string, pluginCreator PluginCreator) {
	pluginMap[protocol] = pluginCreator
}

func NewPlugin(localPort int, v vmessInfo.VmessInfo) (plugin Plugin, err error) {
	v.Protocol = strings.ToLower(v.Protocol)
	creator, ok := pluginMap[v.Protocol]
	if !ok {
		return nil, newError("unregistered protocol ", v.Protocol)
	}
	return creator(localPort, v)
}

func IsProtocolValid(v vmessInfo.VmessInfo) bool {
	_, ok := pluginMap[v.Protocol]
	return ok
}
