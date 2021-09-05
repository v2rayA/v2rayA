package service

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"strings"
	"time"
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
			//修改
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
			//新建
			//后端NodeData转前端TouchServerRaw压入TouchRaw.Servers
			err = configure.AppendServers([]*configure.ServerRawV2{{ServerObj: obj}})
		}
	} else {
		//不是ss://也不是vmess://，有可能是订阅地址
		if strings.HasPrefix(url, "sub://") {
			var e error
			url, e = common.Base64StdDecode(url[6:])
			if e != nil {
				url, _ = common.Base64URLDecode(url[6:])
			}
		}
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + url
		}
		c, err := httpClient.GetHttpClientAutomatically()
		if err != nil {
			return err
		}
		c.Timeout = 90 * time.Second
		infos, status, err := ResolveSubscriptionWithClient(url, c)
		if err != nil {
			return fmt.Errorf("failed to resolve subscription address: %w", err)
		}
		//后端NodeData转前端TouchServerRaw压入TouchRaw.Subscriptions.Servers
		servers := make([]configure.ServerRawV2, len(infos))
		for i, v := range infos {
			servers[i] = configure.ServerRawV2{ServerObj: v}
		}
		//去重
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
			Address: url,
			Status:  string(touch.NewUpdateStatus()),
			Servers: uniqueServers,
			Info:    status,
		}})
	}
	return
}
