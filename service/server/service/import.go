package service

import (
	"fmt"
	url2 "net/url"
	"strings"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/pkg/util/log"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
)

func PluginManagerValidateLink(url string) bool {
	if pm := conf.GetEnvironmentConfig().PluginManager; pm != "" {
		_, err := serverObj.NewFromLink(serverObj.PluginManagerScheme, url)
		return err == nil
	} else {
		return false
	}
}

func Import(url string, which *configure.Which) (err error) {
	log.Trace("Import: %v, which: %+v", url, which)
	resolv.CheckResolvConf()
	url = strings.TrimSpace(url)
	if lines := strings.Split(url, "\n"); len(lines) >= 2 || strings.HasPrefix(url, "{") {
		infos, _, err := ResolveByLines(url)
		if err != nil {
			return fmt.Errorf("failed to resolve addresses: %w", err)
		}
		for _, info := range infos {
			err = configure.AppendServers([]*configure.ServerRaw{{ServerObj: info}})
		}
		return err
	}
	supportedPrefix := []string{"vmess", "vless", "ss", "ssr", "trojan", "trojan-go", "http-proxy",
		"https-proxy", "socks5", "http2", "juicity", "tuic", "hysteria", "hysteria2", "anytls",
		"shadowsocks", "shadowsocksr", "hy1", "hy2", "mcore", "mcp", "pingtunnel", "plugin"}
	for i := range supportedPrefix {
		supportedPrefix[i] += "://"
	}
	urlLower := strings.ToLower(url)
	if PluginManagerValidateLink(url) || common.HasAnyPrefix(urlLower, supportedPrefix) {
		var obj serverObj.ServerObj
		obj, err = ResolveURL(url)
		if err != nil {
			log.Warn("ResolveURL failed: %v", err)
			return
		}
		if which != nil && which.ID > 0 {
			// the request is to modify a server
			ind := which.ID - 1
			log.Info("Import: modifying server. ind=%v, which.ID=%v, which.TYPE=%v", ind, which.ID, which.TYPE)
			if which.TYPE != configure.ServerType {
				// Also support modifying subscription servers if the frontend allows it
				// but for now, we primarily care about ServerType
				if which.TYPE != configure.SubscriptionServerType {
					log.Warn("Import: unsupported touch type for modification: %v", which.TYPE)
					return fmt.Errorf("bad request: unsupported touch type")
				}
			}

			if which.TYPE == configure.ServerType && (ind < 0 || ind >= configure.GetLenServers()) {
				log.Warn("Import: invalid server index: %v", ind)
				return fmt.Errorf("bad request: invalid index")
			}

			if err = configure.SetServer(ind, &configure.ServerRaw{ServerObj: obj}); err != nil {
				log.Warn("Import: SetServer failed: %v", err)
				return
			}
			log.Info("Import: SetServer success for index %v", ind)
			css := configure.GetConnectedServers()
			if css.Len() > 0 {
				for _, cs := range css.Get() {
					if which.TYPE == cs.TYPE && which.ID == cs.ID {
						log.Info("Import: updating connected v2ray config")
						if err = v2ray.UpdateV2RayConfig(); err != nil {
							log.Warn("Import: UpdateV2RayConfig failed: %v", err)
							return
						}
					}
				}
			}
		} else {
			// append a server
			log.Info("Import: appending a new server")
			err = configure.AppendServers([]*configure.ServerRaw{{ServerObj: obj}})
		}
	} else {
		// subscription
		source := url
		if u, err := url2.Parse(source); err == nil {
			if u.Scheme == "sub" {
				var e error
				source, e = common.Base64StdDecode(source[6:])
				if e != nil {
					source, _ = common.Base64URLDecode(source[6:])
				}
			} else if u.Scheme == "" {
				u.Scheme = "http"
				source = u.String()
			}
		}
		c := httpClient.GetHttpClientAutomatically()
		c.Timeout = 90 * time.Second
		infos, status, err := ResolveSubscriptionWithClient(source, c)
		if err != nil {
			return fmt.Errorf("failed to resolve subscription address: %w", err)
		}

		// info to serverRawV2
		servers := make([]configure.ServerRaw, len(infos))
		for i, v := range infos {
			servers[i] = configure.ServerRaw{ServerObj: v}
		}

		// deduplicate
		unique := make(map[configure.ServerRaw]interface{})
		for _, s := range servers {
			unique[s] = nil
		}
		uniqueServers := make([]configure.ServerRaw, 0)
		for _, s := range servers {
			if _, ok := unique[s]; ok {
				uniqueServers = append(uniqueServers, s)
				delete(unique, s)
			}
		}
		err = configure.AppendSubscriptions([]*configure.SubscriptionRaw{{
			Address: source,
			Status:  string(touch.NewUpdateStatus()),
			Servers: uniqueServers,
			Info:    status,
		}})
	}
	return
}
