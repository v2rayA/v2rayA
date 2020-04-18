package service

import (
	"V2RayA/common/httpClient"
	"V2RayA/core/nodeData"
	"V2RayA/core/touch"
	"V2RayA/core/v2ray"
	"V2RayA/persistence/configure"
	"strings"
	"time"
)

func Import(url string, which *configure.Which) (err error) {
	url = strings.TrimSpace(url)
	if strings.HasPrefix(url, "vmess://") ||
		strings.HasPrefix(url, "ss://") ||
		strings.HasPrefix(url, "ssr://") ||
		strings.HasPrefix(url, "pingtunnel://") {
		var n *nodeData.NodeData
		n, err = ResolveURL(url)
		if err != nil {
			return
		}
		if which != nil {
			//修改
			ind := which.ID - 1
			if which.TYPE != configure.ServerType || ind < 0 || ind >= configure.GetLenServers() {
				return newError("bad request")
			}
			var sr *configure.ServerRaw
			sr, err = which.LocateServer()
			if err != nil {
				return
			}
			sr.VmessInfo = n.VmessInfo
			err = configure.SetServer(ind, n.ToServerRaw())
			cs := configure.GetConnectedServer()
			if cs != nil && which.TYPE == cs.TYPE && which.ID == cs.ID {
				err = v2ray.UpdateV2RayConfig(nil)
			}
		} else {
			//新建
			//后端NodeData转前端TouchServerRaw压入TouchRaw.Servers
			err = configure.AppendServer(n.ToServerRaw())
		}
	} else {
		//不是ss://也不是vmess://，有可能是订阅地址
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + url
		}
		c, err := httpClient.GetHttpClientAutomatically()
		if err != nil {
			return err
		}
		c.Timeout = 90 * time.Second
		infos, err := ResolveSubscriptionWithClient(url, c)
		if err != nil {
			return newError("fail in resolving subscription address").Base(err)
		}
		//后端NodeData转前端TouchServerRaw压入TouchRaw.Subscriptions.Servers
		servers := make([]configure.ServerRaw, len(infos))
		for i, v := range infos {
			servers[i] = *v.ToServerRaw()
		}
		//去重
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
		err = configure.AppendSubscription(&configure.SubscriptionRaw{
			Address: url,
			Status:  string(touch.NewUpdateStatus()),
			Servers: uniqueServers,
		})
	}
	return
}
