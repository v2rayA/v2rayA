package service

import (
	"V2RayA/model/nodeData"
	"V2RayA/model/touch"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"V2RayA/tools/httpClient"
	"errors"
	"fmt"
	"strings"
)

func Import(url string, which *configure.Which) (err error) {
	if strings.HasPrefix(url, "vmess://") || strings.HasPrefix(url, "ss://") || strings.HasPrefix(url, "ssr://") {
		var n *nodeData.NodeData
		n, err = ResolveURL(url)
		if err != nil {
			return
		}
		fmt.Println(n)
		if which != nil {
			//修改
			ind := which.ID - 1
			if which.TYPE != configure.ServerType || ind < 0 || ind >= configure.GetLenServers() {
				return errors.New("节点参数有误")
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
				err = v2ray.UpdateV2rayWithConnectedServer()
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
			return errors.New("尝试使用代理失败，建议修改设置为直连模式再试" + err.Error())
		}
		infos, err := ResolveSubscriptionWithClient(url, c)
		if err != nil {
			return errors.New("解析订阅地址失败" + err.Error())
		}
		//后端NodeData转前端TouchServerRaw压入TouchRaw.Subscriptions.Servers
		servers := make([]configure.ServerRaw, len(infos))
		for i, v := range infos {
			servers[i] = *v.ToServerRaw()
		}
		//去重
		unique := make(map[configure.ServerRaw]struct{})
		for _, s := range servers {
			unique[s] = struct{}{}
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
