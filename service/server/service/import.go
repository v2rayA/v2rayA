package service

import (
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/infra/nodeData"
	"strings"
	"time"
)

func Import(url string, which *configure.Which) (err error) {
	//log.Println(url)
	checkResolvConf()
	url = strings.TrimSpace(url)
	if lines := strings.Split(url, "\n"); len(lines) >= 2 {
		infos, _, err := ResolveLines(url)
		if err != nil {
			return newError("failed to resolve addresses").Base(err)
		}
		for _, info := range infos {
			err = configure.AppendServers([]*configure.ServerRaw{info.ToServerRaw()})
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
		strings.HasPrefix(url, "trojan-go://") {
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
			err = configure.AppendServers([]*configure.ServerRaw{n.ToServerRaw()})
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
			return newError("failed to resolve subscription address").Base(err)
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
		err = configure.AppendSubscriptions([]*configure.SubscriptionRaw{{
			Address: url,
			Status:  string(touch.NewUpdateStatus()),
			Servers: uniqueServers,
			Info:    status,
		}})
	}
	return
}
