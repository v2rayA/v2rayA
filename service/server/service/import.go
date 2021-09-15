package service

import (
	"fmt"
	url2 "net/url"
	"strings"
	"time"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
)

func Import(url string, which *configure.Which) (err error) {
	//log.Trace(url)
	resolv.CheckResolvConf()
	url = strings.TrimSpace(url)
	if lines := strings.Split(url, "\n"); len(lines) >= 2 {
		infos, _, err := ResolveLines(url)
		if err != nil {
			return fmt.Errorf("failed to resolve addresses: %w", err)
		}
		for _, info := range infos {
			err = configure.AppendServers([]*configure.ServerRawV2{{ServerObj: info}})
		}
		return err
	}
	if strings.HasPrefix(url, "vmess://") ||
		strings.HasPrefix(url, "vless://") ||
		strings.HasPrefix(url, "ss://") ||
		strings.HasPrefix(url, "ssr://") ||
		strings.HasPrefix(url, "pingtunnel://") ||
		strings.HasPrefix(url, "ping-tunnel://") ||
		strings.HasPrefix(url, "trojan://") ||
		strings.HasPrefix(url, "trojan-go://") ||
		strings.HasPrefix(url, "http-proxy://") ||
		strings.HasPrefix(url, "https-proxy://") ||
		strings.HasPrefix(url, "http2://") {
		var obj serverObj.ServerObj
		obj, err = ResolveURL(url)
		if err != nil {
			return
		}
		if which != nil {
			// the request is to modify a server
			ind := which.ID - 1
			if which.TYPE != configure.ServerType || ind < 0 || ind >= configure.GetLenServers() {
				return fmt.Errorf("bad request")
			}
			var sr *configure.ServerRawV2
			sr, err = which.LocateServerRaw()
			if err != nil {
				return
			}
			sr.ServerObj = obj
			if err = configure.SetServer(ind, &configure.ServerRawV2{ServerObj: obj}); err != nil {
				return
			}
			css := configure.GetConnectedServers()
			if css.Len() > 0 {
				for _, cs := range css.Get() {
					if which.TYPE == cs.TYPE && which.ID == cs.ID {
						if err = v2ray.UpdateV2RayConfig(); err != nil {
							return
						}
					}
				}
			}
		} else {
			// append a server
			err = configure.AppendServers([]*configure.ServerRawV2{{ServerObj: obj}})
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
		} else {
			// maybe it is a OOCv1 token
		}
		c, err := httpClient.GetHttpClientAutomatically()
		if err != nil {
			return err
		}
		c.Timeout = 90 * time.Second
		infos, status, err := ResolveSubscriptionWithClient(source, c)
		if err != nil {
			return fmt.Errorf("failed to resolve subscription address: %w", err)
		}

		// info to serverRawV2
		servers := make([]configure.ServerRawV2, len(infos))
		for i, v := range infos {
			servers[i] = configure.ServerRawV2{ServerObj: v}
		}

		// deduplicate
		unique := make(map[configure.ServerRawV2]interface{})
		for _, s := range servers {
			unique[s] = nil
		}
		uniqueServers := make([]configure.ServerRawV2, 0)
		for _, s := range servers {
			if _, ok := unique[s]; ok {
				uniqueServers = append(uniqueServers, s)
				delete(unique, s)
			}
		}
		err = configure.AppendSubscriptions([]*configure.SubscriptionRawV2{{
			Address: source,
			Status:  string(touch.NewUpdateStatus()),
			Servers: uniqueServers,
			Info:    status,
		}})
	}
	return
}
