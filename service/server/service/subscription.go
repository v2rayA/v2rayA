package service

import (
	"bytes"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

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

func resolveSIP008(raw string) (infos []serverObj.ServerObj, sip SIP008, err error) {
	err = jsoniter.Unmarshal([]byte(raw), &sip)
	if err != nil {
		return
	}
	for _, server := range sip.Servers {
		u := url.URL{
			Scheme:   "ss",
			User:     url.UserPassword(server.Method, server.Password),
			Host:     net.JoinHostPort(server.Server, strconv.Itoa(server.ServerPort)),
			RawQuery: url.Values{"plugin": []string{server.PluginOpts}}.Encode(),
			Fragment: server.Remarks,
		}
		obj, err := serverObj.NewFromLink("shadowsocks", u.String())
		if err != nil {
			return nil, SIP008{}, err
		}
		infos = append(infos, obj)
	}
	return
}

func resolveByLines(raw string) (infos []serverObj.ServerObj, status string, err error) {
	// 切分raw
	rows := strings.Split(strings.TrimSpace(raw), "\n")
	// 解析
	infos = make([]serverObj.ServerObj, 0)
	for _, row := range rows {
		if strings.HasPrefix(row, "STATUS=") {
			status = strings.TrimPrefix(row, "STATUS=")
			continue
		}
		var data serverObj.ServerObj
		data, err = ResolveURL(row)
		if err != nil {
			if !errors.Is(err, EmptyAddressErr) {
				log.Warn("resolveByLines: %v: %v", err, row)
			}
			err = nil
			continue
		}
		infos = append(infos, data)
	}
	return
}

type SubscriptionUserInfo struct {
	Upload   int64
	Download int64
	Total    int64
	Expire   time.Time
}

func (sui *SubscriptionUserInfo) String() string {
	var outputs []string
	if sui.Download != -1 {
		outputs = append(outputs, fmt.Sprintf("download: %v GB", sui.Download/1e9))
	}
	if sui.Upload != -1 {
		outputs = append(outputs, fmt.Sprintf("upload: %v GB", sui.Upload/1e9))
	}
	if sui.Total != -1 {
		outputs = append(outputs, fmt.Sprintf("total: %v GB", sui.Total/1e9))
	}
	if !sui.Expire.IsZero() {
		outputs = append(outputs, fmt.Sprintf("expire: %v UTC", sui.Expire.Format("2006-01-02 15:04")))
	}
	return strings.Join(outputs, "; ")
}

func parseSubscriptionUserInfo(str string) SubscriptionUserInfo {
	fields := strings.Split(str, ";")
	sui := SubscriptionUserInfo{
		Upload:   -1,
		Download: -1,
		Total:    -1,
		Expire:   time.Time{},
	}
	for _, field := range fields {
		field = strings.TrimSpace(field)
		kv := strings.SplitN(field, "=", 2)
		if len(kv) < 2 {
			continue
		}
		v, e := strconv.ParseInt(kv[1], 10, 64)
		if e != nil {
			continue
		}
		switch kv[0] {
		case "upload":
			sui.Upload = v
		case "download":
			sui.Download = v
		case "total":
			sui.Total = v
		case "expire":
			sui.Expire = time.Unix(v, 0).UTC()
		}
	}
	return sui
}
func trapBOM(fileBytes []byte) []byte {
	trimmedBytes := bytes.Trim(fileBytes, "\xef\xbb\xbf")
	return trimmedBytes
}
func ResolveSubscriptionWithClient(source string, client *http.Client) (infos []serverObj.ServerObj, status string, err error) {
	c := *client
	if c.Timeout < 30*time.Second {
		c.Timeout = 30 * time.Second
	}

	res, err := httpClient.HttpGetUsingSpecificClient(client, source)
	if err != nil {
		return
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, "", err
	}
	// base64 decode. trapBOM due to https://github.com/v2rayA/v2rayA/issues/612
	raw, err := common.Base64StdDecode(string(trapBOM(b)))
	if err != nil {
		raw, _ = common.Base64URLDecode(string(b))
	}
	infos, status, err = ResolveByLines(raw)
	if err != nil {
		return nil, "", err
	}
	subscriptionUserInfo := res.Header.Get("Subscription-Userinfo")
	sui := parseSubscriptionUserInfo(subscriptionUserInfo)
	if len(status) > 0 {
		status = sui.String() + "|" + status
	} else {
		status = sui.String()
	}
	return infos, status, nil
}

func ResolveByLines(raw string) (infos []serverObj.ServerObj, status string, err error) {
	var sip SIP008
	if infos, sip, err = resolveSIP008(raw); err == nil {
		status = getDataUsageStatus(sip.BytesUsed, sip.BytesRemaining)
	} else {
		infos, status, err = resolveByLines(raw)
	}
	return
}

func getDataUsageStatus(bytesUsed, bytesRemaining uint64) (status string) {
	if bytesUsed != 0 {
		status = fmt.Sprintf("Used: %.2f GiB", float64(bytesUsed)/1024/1024/1024)
		if bytesRemaining != 0 {
			status += fmt.Sprintf(" | Remaining: %.2f GiB", float64(bytesRemaining)/1024/1024/1024)
		}
	}
	return
}

func UpdateSubscription(index int, disconnectIfNecessary bool) (err error) {
	subscriptions := configure.GetSubscriptions()
	addr := subscriptions[index].Address
	c := httpClient.GetHttpClientAutomatically()
	resolv.CheckResolvConf()
	subscriptionInfos, status, err := ResolveSubscriptionWithClient(addr, c)
	if err != nil {
		reason := "failed to resolve subscription address: " + err.Error()
		log.Warn("UpdateSubscription: %v: %v", err, subscriptionInfos)
		return fmt.Errorf("UpdateSubscription: %v", reason)
	}
	infoServerRaws := make([]configure.ServerRaw, len(subscriptionInfos))
	css := configure.GetConnectedServers()
	cssAfter := css.Get()
	// serverObj.ServerObj is a pointer(interface), and shouldn't be as a key
	link2Raw := make(map[string]*configure.ServerRaw)
	connectedVmessInfo2CssIndex := make(map[string][]int)
	for i, cs := range css.Get() {
		if cs.TYPE == configure.SubscriptionServerType && cs.Sub == index {
			if sRaw, err := cs.LocateServerRaw(); err != nil {
				return err
			} else {
				link := sRaw.ServerObj.ExportToURL()
				link2Raw[link] = sRaw
				connectedVmessInfo2CssIndex[link] = append(connectedVmessInfo2CssIndex[link], i)
			}
		}
	}
	//将列表更换为新的，并且找到一个跟现在连接的server值相等的，设为Connected，如果没有，则断开连接
	for i, info := range subscriptionInfos {
		infoServerRaw := configure.ServerRaw{
			ServerObj: info,
		}
		link := infoServerRaw.ServerObj.ExportToURL()
		if cssIndexes, ok := connectedVmessInfo2CssIndex[link]; ok {
			for _, cssIndex := range cssIndexes {
				cssAfter[cssIndex].ID = i + 1
			}
			delete(connectedVmessInfo2CssIndex, link)
		}
		infoServerRaws[i] = infoServerRaw
	}
	for link, cssIndexes := range connectedVmessInfo2CssIndex {
		for _, cssIndex := range cssIndexes {
			if disconnectIfNecessary {
				err = Disconnect(*css.Get()[cssIndex], false)
				if err != nil {
					reason := "failed to disconnect previous server"
					return fmt.Errorf("UpdateSubscription: %v", reason)
				}
			} else {
				// 将之前连接的节点append进去
				// TODO: 变更ServerRaw时可能需要考虑
				infoServerRaws = append(infoServerRaws, *link2Raw[link])
				cssAfter[cssIndex].ID = len(infoServerRaws)
			}
		}
	}
	if err := configure.OverwriteConnects(configure.NewWhiches(cssAfter)); err != nil {
		return err
	}
	subscriptions[index].Servers = infoServerRaws
	subscriptions[index].Status = string(touch.NewUpdateStatus())
	subscriptions[index].Info = status
	return configure.SetSubscription(index, &subscriptions[index])
}

func ModifySubscriptionRemark(subscription touch.Subscription) (err error) {
	raw := configure.GetSubscription(subscription.ID - 1)
	if raw == nil {
		return fmt.Errorf("failed to find the corresponding subscription")
	}
	raw.Remarks = subscription.Remarks
	return configure.SetSubscription(subscription.ID-1, raw)
}
