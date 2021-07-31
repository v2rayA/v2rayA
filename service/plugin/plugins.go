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
	LocalPort() int
	SupportUDP() bool //TODO: support udp
}

type Plugins struct {
	Plugins map[string]Plugin
	mutex   sync.Mutex
}

func (r *Plugins) CloseAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, ssr := range r.Plugins {
		_ = ssr.Close()
	}
	r.Plugins = make(map[string]Plugin)
}

func (r *Plugins) Close(outbound string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if plu, ok := r.Plugins[outbound]; ok {
		plu.Close()
		delete(r.Plugins, outbound)
	}
}

func (r *Plugins) Add(outbound string, plugin Plugin) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if plugin == nil {
		return
	}
	r.Plugins[outbound] = plugin
}

type Creator func(localPort int, v vmessInfo.VmessInfo) (plugin Plugin, err error)

var pluginMap = make(map[string]Creator)

func RegisterPlugin(protocol string, pluginCreator Creator) {
	pluginMap[protocol] = pluginCreator
}

func preprocess(v *vmessInfo.VmessInfo) (needPlugin bool) {
	switch v.Protocol {
	case "", "vmess", "vless", "trojan":
		return false
	case "shadowsocks", "ss":
		if v.Type == "" {
			switch v.Net {
			case "aes-128-cfb", "aes-192-cfb", "aes-256-cfb", "aes-128-ctr", "aes-192-ctr", "aes-256-ctr", "aes-128-ofb", "aes-192-ofb", "aes-256-ofb", "des-cfb", "bf-cfb", "cast5-cfb", "rc4-md5", "chacha20", "chacha20-ietf", "salsa20", "camellia-128-cfb", "camellia-192-cfb", "camellia-256-cfb", "idea-cfb", "rc2-cfb", "seed-cfb":
				//使用ssr插件
				RazorSS(v)
				v.Protocol = "ssr"
			default:
				//不需要插件
				return false
			}
		} else if v.Type == "http" || v.Type == "tls" {
			switch v.Net {
			case "aes-128-cfb", "aes-192-cfb", "aes-256-cfb", "aes-128-ctr", "aes-192-ctr", "aes-256-ctr", "aes-128-ofb", "aes-192-ofb", "aes-256-ofb", "des-cfb", "bf-cfb", "cast5-cfb", "rc4-md5", "chacha20", "chacha20-ietf", "salsa20", "camellia-128-cfb", "camellia-192-cfb", "camellia-256-cfb", "idea-cfb", "rc2-cfb", "seed-cfb":
				//ssr插件接simpleobfs插件
				v.Protocol = "ssrplugin-simpleobfs"
			default:
				//simpleobfs插件
				v.Protocol = "simpleobfs"
			}
		}
	}
	return true
}

// New a plugin and serve. If no plugin is needed, returns nil, nil.
func NewPluginAndServe(localPort int, v vmessInfo.VmessInfo) (plugin Plugin, err error) {
	v.Protocol = strings.ToLower(v.Protocol)
	if ok := preprocess(&v); !ok {
		return nil, nil
	}
	creator, ok := pluginMap[v.Protocol]
	if !ok {
		return nil, newError("unregistered protocol ", v.Protocol)
	}
	return creator(localPort, v)
}
func IsProtocolValid(v vmessInfo.VmessInfo) bool {
	preprocess(&v)
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

func RazorSS(ss *vmessInfo.VmessInfo) {
	ss.TLS = "plain"
	ss.Type = "origin"
	ss.Path = ""
	ss.Host = ""
}
