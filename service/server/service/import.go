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
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/touch"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
)

func PluginManagerValidateLink(url string) bool {
	if pm := conf.GetEnvironmentConfig().PluginManager; pm != "" {
		_, err := serverObj.NewFromLink(serverObj.PluginManagerScheme, url)
		return err == nil
	} else {
		return false
	}
}

// Import guesses whether url is a server link or a subscription address
// from its scheme. Callers that know which one they hold should use
// ImportServer or ImportSubscription instead; this wrapper only exists for
// clients that predate the "kind" field of POST /api/import.
func Import(url string, which *configure.Which) (err error) {
	log.Trace("Import: received url=%v, which=%+v", url, which)
	url = strings.TrimSpace(url)
	if lines := strings.Split(url, "\n"); len(lines) >= 2 || strings.HasPrefix(url, "{") {
		return ImportServer(url, which)
	}
	// "http" and "https" are deliberately absent: a bare http(s) URL has
	// always meant a subscription here, and single HTTP proxy nodes are
	// written as http-proxy:// or https-proxy://.
	supportedPrefix := []string{"vmess", "vless", "ss", "ssr", "trojan", "trojan-go", "http-proxy",
		"https-proxy", "socks5", "http2", "juicity", "tuic", "hysteria", "hysteria2", "anytls",
		"shadowsocks", "shadowsocksr", "hy1", "hy2", "mcore", "mcp", "plugin", "wireguard"}
	for i := range supportedPrefix {
		supportedPrefix[i] += "://"
	}
	urlLower := strings.ToLower(url)
	if PluginManagerValidateLink(url) || common.HasAnyPrefix(urlLower, supportedPrefix) {
		return ImportServer(url, which)
	}
	return ImportSubscription(url)
}

// ImportServer imports one server link, or several separated by newlines,
// or a JSON config. A non-nil which with ID > 0 replaces that server instead
// of appending.
func ImportServer(url string, which *configure.Which) (err error) {
	resolv.CheckResolvConf()
	url = strings.TrimSpace(url)
	if lines := strings.Split(url, "\n"); len(lines) >= 2 || strings.HasPrefix(url, "{") {
		infos, _, err := ResolveByLines(url)
		if err != nil {
			return fmt.Errorf("failed to resolve addresses: %w", err)
		}
		for i, info := range infos {
			if err = configure.AppendServers([]*configure.ServerRaw{{ServerObj: info}}); err != nil {
				return fmt.Errorf("failed to import server %d/%d (%v): %w", i+1, len(infos), info.GetName(), err)
			}
		}
		return nil
	}
	{
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

			switch which.TYPE {
			case configure.ServerType:
				if ind < 0 || ind >= configure.GetLenServers() {
					log.Warn("Import: invalid server index: %v", ind)
					return fmt.Errorf("bad request: invalid index")
				}
				if err = configure.SetServer(ind, &configure.ServerRaw{ServerObj: obj}); err != nil {
					log.Warn("Import: SetServer failed: %v", err)
					return
				}
			case configure.SubscriptionServerType:
				// a subscription node lives inside its subscription; writing it
				// with SetServer would overwrite the standalone server at the
				// same index instead
				subs := configure.GetSubscriptions()
				if which.Sub < 0 || which.Sub >= len(subs) || ind < 0 || ind >= len(subs[which.Sub].Servers) {
					log.Warn("Import: invalid subscription server index: sub=%v ind=%v", which.Sub, ind)
					return fmt.Errorf("bad request: invalid index")
				}
				sub := subs[which.Sub]
				sub.Servers[ind] = configure.ServerRaw{ServerObj: obj}
				if err = configure.SetSubscription(which.Sub, &sub); err != nil {
					log.Warn("Import: SetSubscription failed: %v", err)
					return
				}
			}
			log.Info("Import: modified %v %v", which.TYPE, which.ID)
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
	}
	return
}

// ImportSubscription fetches url (or the base64 payload of a sub:// link),
// parses the server list it returns and stores it as a new subscription.
func ImportSubscription(url string) (err error) {
	resolv.CheckResolvConf()
	url = strings.TrimSpace(url)
	{
		source := url
		if u, err := url2.Parse(source); err == nil {
			switch strings.ToLower(u.Scheme) {
			case "sub":
				// strip the "sub://" (case-insensitive) scheme, then base64 decode
				// the payload
				payload := source[len(u.Scheme)+3:]
				var e error
				if source, e = common.Base64StdDecode(payload); e != nil {
					source, _ = common.Base64URLDecode(payload)
				}
			case "":
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

		// deduplicate using protocol://host:port as key since ServerRaw contains
		// non-comparable fields. Host:port alone is not enough: the same endpoint
		// can legitimately serve multiple protocols (e.g. vmess + trojan).
		seen := make(map[string]struct{})
		uniqueServers := make([]configure.ServerRaw, 0, len(servers))
		for _, s := range servers {
			key := fmt.Sprintf("%s://%s:%d", s.ServerObj.GetProtocol(), s.ServerObj.GetHostname(), s.ServerObj.GetPort())
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			uniqueServers = append(uniqueServers, s)
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
