package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/errors"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/infra/nodeData"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//func ResolveSubscription(source string) (infos []*nodeData.NodeData, err error) {
//	return ResolveSubscriptionWithClient(source, http.DefaultClient)
//}
type SIP008 struct {
	Version        int    `json:"version"`
	Username       string `json:"username"`
	UserUUID       string `json:"user_uuid"`
	BytesUsed      uint64 `json:"bytes_used"`
	BytesRemaining uint64 `json:"bytes_remaining"`
	Servers        []struct {
		Server     string `json:"server"`
		ServerPort int    `json:"server_port"`
		Password   string `json:"password"`
		Method     string `json:"method"`
		Plugin     string `json:"plugin"`
		PluginOpts string `json:"plugin_opts"`
		Remarks    string `json:"remarks"`
		ID         string `json:"id"`
	} `json:"servers"`
}

func resolveSIP008(raw string) (infos []*nodeData.NodeData, sip SIP008, err error) {
	err = json.Unmarshal([]byte(raw), &sip)
	if err != nil {
		return
	}
	for _, server := range sip.Servers {
		arr := strings.Split(server.PluginOpts, ";")
		var obfs, path, host string
		for i := 0; i < len(arr); i++ {
			a := strings.Split(arr[i], "=")
			switch a[0] {
			case "obfs":
				obfs = a[1]
			case "obfs-path":
				path = a[1]
			case "obfs-host":
				host = a[1]
			}
		}
		infos = append(infos, &nodeData.NodeData{VmessInfo: vmessInfo.VmessInfo{
			Ps:       server.Remarks,
			Add:      server.Server,
			Port:     strconv.Itoa(server.ServerPort),
			ID:       server.Password,
			Net:      server.Method,
			Type:     obfs,
			Path:     path,
			Host:     host,
			Protocol: "ss",
		}})
	}
	return
}

func resolveByLines(raw string) (infos []*nodeData.NodeData, status string, err error) {
	// 切分raw
	rows := strings.Split(strings.TrimSpace(raw), "\n")
	// 解析
	infos = make([]*nodeData.NodeData, 0)
	for _, row := range rows {
		if strings.HasPrefix(row, "STATUS=") {
			status = strings.TrimPrefix(row, "STATUS=")
			continue
		}
		var data *nodeData.NodeData
		data, err = ResolveURL(row)
		if err != nil {
			if errors.Cause(err) != ErrorEmptyAddress {
				log.Println(row, err)
			}
			err = nil
			continue
		}
		infos = append(infos, data)
	}
	return
}

func ResolveSubscriptionWithClient(source string, client *http.Client) (infos []*nodeData.NodeData, status string, err error) {
	// get请求source
	c := *client
	c.Timeout = 30 * time.Second
	res, err := httpClient.HttpGetUsingSpecificClient(&c, source)
	if err != nil {
		return
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(res.Body)
	defer res.Body.Close()
	// base64解码
	raw, err := common.Base64StdDecode(buf.String())
	if err != nil {
		raw, _ = common.Base64URLDecode(buf.String())
	}
	return ResolveLines(raw)
}
func ResolveLines(raw string) (infos []*nodeData.NodeData, status string, err error) {
	var sip SIP008
	if infos, sip, err = resolveSIP008(raw); err == nil {
		if sip.BytesUsed != 0 {
			status = fmt.Sprintf("Used: %.2fGB", float64(sip.BytesUsed)/1024/1024/1024)
			if sip.BytesRemaining != 0 {
				status += fmt.Sprintf(" | Remaining: %.2fGB", float64(sip.BytesRemaining)/1024/1024/1024)
			}
		}
	} else {
		infos, status, err = resolveByLines(raw)
	}
	return
}

func UpdateSubscription(index int, disconnectIfNecessary bool) (err error) {
	subscriptions := configure.GetSubscriptions()
	addr := subscriptions[index].Address
	c, err := httpClient.GetHttpClientAutomatically()
	if err != nil {
		reason := "failed to get proxy"
		return newError(reason)
	}
	checkResolvConf()
	infos, status, err := ResolveSubscriptionWithClient(addr, c)
	if err != nil {
		reason := "failed to resolve subscription address: " + err.Error()
		log.Println(infos, err)
		return newError(reason)
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
				reason := "failed to disconnect previous server"
				return newError(reason)
			}
		} else if connectedServer != nil {
			//将之前连接的节点append进去
			tsrs = append(tsrs, *connectedServer)
			cs.ID = len(tsrs)
			err = configure.SetConnect(cs)
			if err != nil {
				return
			}
		}
	}
	subscriptions[index].Servers = tsrs
	subscriptions[index].Status = string(touch.NewUpdateStatus())
	subscriptions[index].Info = status
	return configure.SetSubscription(index, &subscriptions[index])
}

func ModifySubscriptionRemark(subscription touch.Subscription) (err error) {
	raw := configure.GetSubscription(subscription.ID - 1)
	if raw == nil {
		return newError("failed to find the corresponding subscription")
	}
	raw.Remarks = subscription.Remarks
	return configure.SetSubscription(subscription.ID-1, raw)
}
