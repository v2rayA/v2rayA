package service

import (
	"V2RayA/model/nodeData"
	"V2RayA/model/touch"
	"V2RayA/persistence/configure"
	"V2RayA/tools"
	"errors"
	"strings"
)

func Import(url string) (err error) {
	if strings.HasPrefix(url, "vmess://") || strings.HasPrefix(url, "ss://") {
		var n *nodeData.NodeData
		n, err = ResolveURL(url)
		if err != nil {
			return
		}
		//后端NodeData转前端TouchServerRaw压入TouchRaw.Servers
		err = configure.AppendServer(n.ToTouchServerRaw())
	} else {
		//不是ss://也不是vmess://，有可能是订阅地址
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + url
		}
		c, err := tools.GetHttpClientAutomatically()
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
			servers[i] = *v.ToTouchServerRaw()
		}
		err = configure.AppendSubscription(&configure.SubscriptionRaw{
			Address: url,
			Status:  string(touch.NewUpdateStatus()),
			Servers: servers,
		})
	}
	return
}
