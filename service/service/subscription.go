package service

import (
	"V2RayA/model/nodeData"
	"V2RayA/model/touch"
	"V2RayA/persistence/configure"
	"V2RayA/tools"
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
	res, err := client.Get(source)
	if err != nil {
		return
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(res.Body)
	defer res.Body.Close()
	// base64解码, raw是多行vmess
	raw, err := tools.Base64StdDecode(buf.String())
	if err != nil {
		return
	}
	// 切分raw
	rows := strings.Split(strings.TrimSpace(raw), "\n")
	// 解析
	infos = make([]*nodeData.NodeData, 0)
	for _, row := range rows {
		var data *nodeData.NodeData
		data, err = ResolveURL(row)
		if err != nil {
			log.Println(row, err)
			err = nil
			continue
		}
		infos = append(infos, data)
	}
	return
}

func UpdateSubscription(index int) (err error) {
	subscriptions := configure.GetSubscriptions()
	addr := subscriptions[index].Address
	c, err := tools.GetHttpClientAutomatically()
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
	tsrs := make([]configure.TouchServerRaw, len(infos))
	var connectedServer *configure.TouchServerRaw
	if cs := configure.GetConnectedServer(); cs != nil {
		connectedServer, _ = cs.LocateServer()
	}
	//将列表更换为新的，并且找到一个跟现在连接的server值相等的，设为Connected，如果没有，则断开连接
	finishFindConnected := false
	for i, info := range infos {
		tsr := configure.TouchServerRaw{
			VmessInfo: info.VmessInfo,
		}
		if !finishFindConnected && connectedServer != nil && connectedServer.VmessInfo == tsr.VmessInfo {
			err = configure.SetConnect(&configure.Which{
				TYPE:        configure.SubscriptionServerType,
				ID:          i + 1,
				Sub:         index,
				PingLatency: nil,
			})
			if err != nil {
				return
			}
			finishFindConnected = true
		}
		tsrs[i] = tsr
	}
	if !finishFindConnected {
		err = Disconnect()
		if err != nil {
			reason := "现连接的服务器已被更新且不包含在新的订阅中，在试图与其断开的过程中遇到失败"
			return errors.New(reason)
		}
	}
	subscriptions[index].Servers = tsrs
	subscriptions[index].Status = string(touch.NewUpdateStatus())
	return configure.SetSubscription(index, &subscriptions[index])
}
