package service

import (
	"V2RayA/model/nodeData"
	"V2RayA/model/touch"
	"V2RayA/persistence/configure"
	"V2RayA/tools"
	"V2RayA/tools/httpClient"
	"bytes"
	"errors"
	"log"
	"net/http"
	"strings"
)

func ResolveSubscription(source string) (infos []*nodeData.NodeData, err error) {
	return ResolveSubscriptionWithClient(source, http.DefaultClient)
}

func ResolveSubscriptionWithClient(source string, client *http.Client) (infos []*nodeData.NodeData, err error) {
	// get请求source
	res, err := httpClient.HttpGetUsingSpecificClient(client, source)
	if err != nil {
		return
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(res.Body)
	defer res.Body.Close()
	// base64解码, raw是多行vmess
	raw, _ := tools.Base64StdDecode(buf.String())
	// 切分raw
	rows := strings.Split(strings.TrimSpace(raw), "\n")
	// 解析
	infos = make([]*nodeData.NodeData, 0)
	for _, row := range rows {
		var data *nodeData.NodeData
		data, err = ResolveURL(row)
		if err != nil {
			if !strings.Contains(err.Error(), "空地址") {
				log.Println(row, err)
			}
			err = nil
			continue
		}
		infos = append(infos, data)
	}
	return
}

func UpdateSubscription(index int, disconnectIfNecessary bool) (err error) {
	subscriptions := configure.GetSubscriptions()
	addr := subscriptions[index].Address
	c, err := httpClient.GetHttpClientAutomatically()
	if err != nil {
		reason := "尝试使用代理失败，建议修改设置为直连模式再试"
		return errors.New(reason)
	}
	infos, err := ResolveSubscriptionWithClient(addr, c)
	if err != nil {
		reason := "解析订阅地址失败: " + err.Error()
		log.Println(infos, err)
		return errors.New(reason)
	}
	tsrs := make([]configure.ServerRaw, len(infos))
	var connectedServer *configure.ServerRaw
	cs := configure.GetConnectedServer()
	toFindConnectedServer := false
	var found bool
	if cs != nil {
		connectedServer, _ = cs.LocateServer()
		if connectedServer != nil && cs.TYPE == configure.SubscriptionServerType && cs.Sub == index {
			toFindConnectedServer = true
			found = false
		}
	}
	//将列表更换为新的，并且找到一个跟现在连接的server值相等的，设为Connected，如果没有，则断开连接
	for i, info := range infos {
		tsr := configure.ServerRaw{
			VmessInfo: info.VmessInfo,
		}
		if toFindConnectedServer && connectedServer.VmessInfo == tsr.VmessInfo {
			err = configure.SetConnect(&configure.Which{
				TYPE:    configure.SubscriptionServerType,
				ID:      i + 1,
				Sub:     index,
				Latency: "",
			})
			if err != nil {
				return
			}
			toFindConnectedServer = false
			found = true
		}
		tsrs[i] = tsr
	}
	if toFindConnectedServer && !found {
		if disconnectIfNecessary {
			err = Disconnect()
			if err != nil {
				reason := "现连接的服务器已被更新且不包含在新的订阅中，在试图与其断开的过程中遇到失败"
				return errors.New(reason)
			}
		} else if connectedServer != nil {
			//将之前连接的节点append进去
			tsrs = append(tsrs, *connectedServer)
			cs.ID = len(tsrs) - 1
			err = configure.SetConnect(cs)
			if err != nil {
				return
			}
		}
	}
	subscriptions[index].Servers = tsrs
	subscriptions[index].Status = string(touch.NewUpdateStatus())
	return configure.SetSubscription(index, &subscriptions[index])
}

func ModifySubscriptionRemark(subscription touch.Subscription) (err error) {
	raw := configure.GetSubscription(subscription.ID - 1)
	if raw == nil {
		return errors.New("无法找到对应的subscription")
	}
	raw.Remarks = subscription.Remarks
	return configure.SetSubscription(subscription.ID-1, raw)
}
