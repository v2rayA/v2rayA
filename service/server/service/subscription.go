package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/serverObj/clash"
	"github.com/v2rayA/v2rayA/kernel/touch"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/pkg/util/log"
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
		rawQuery := url.Values{}
		if server.Plugin != "" {
			// SIP008's "plugin" is the plugin name and "plugin_opts" its options;
			// combine them into the ss:// "plugin=name;opts" parameter.
			plugin := server.Plugin
			if server.PluginOpts != "" {
				plugin += ";" + server.PluginOpts
			}
			rawQuery.Set("plugin", plugin)
		}
		u := url.URL{
			Scheme:   "ss",
			User:     url.UserPassword(server.Method, server.Password),
			Host:     net.JoinHostPort(server.Server, strconv.Itoa(server.ServerPort)),
			RawQuery: rawQuery.Encode(),
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
	// Split raw
	rows := strings.Split(strings.TrimSpace(raw), "\n")
	// Parse
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

// Known reports whether the provider sent any usage field at all.
func (sui *SubscriptionUserInfo) Known() bool {
	return sui.Download != -1 || sui.Upload != -1 || sui.Total != -1 || !sui.Expire.IsZero()
}

// String renders the usage the way the node list shows it: what is used of
// what is available, and the day it expires. It used to list every raw
// field ("download: 0 GB; upload: 0 GB; total: 107 GB; expire: 2026-10-04
// 18:59 UTC") in integer GB next to the provider's own text in GiB.
func (sui *SubscriptionUserInfo) String() string {
	const gib = 1024 * 1024 * 1024
	var outputs []string
	if sui.Download != -1 || sui.Upload != -1 {
		used := float64(max(sui.Download, 0)+max(sui.Upload, 0)) / gib
		if sui.Total > 0 {
			outputs = append(outputs, fmt.Sprintf("Used %.2f GiB / %.2f GiB", used, float64(sui.Total)/gib))
		} else {
			outputs = append(outputs, fmt.Sprintf("Used %.2f GiB", used))
		}
	} else if sui.Total > 0 {
		outputs = append(outputs, fmt.Sprintf("Total %.2f GiB", float64(sui.Total)/gib))
	}
	if !sui.Expire.IsZero() {
		outputs = append(outputs, "Expires "+sui.Expire.Local().Format("2006-01-02"))
	}
	return strings.Join(outputs, " · ")
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
	defer func() {
		if err != nil {
			var coded *common.CodedError
			if !errors.As(err, &coded) {
				err = common.Coded("SUBSCRIPTION_FETCH_FAILED", err, map[string]interface{}{
					"host":   subscriptionHost(source),
					"detail": err.Error(),
				})
			}
		}
	}()

	c := *client
	if c.Timeout < 30*time.Second {
		c.Timeout = 30 * time.Second
	}

	res, err := httpClient.HttpGetUsingSpecificClient(&c, source)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return nil, "", fmt.Errorf("subscription server answered %s", res.Status)
	}
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
	if len(infos) == 0 {
		// an update that replaced the list with nothing would drop every
		// node; treat an unparseable or empty body as a failed fetch
		return nil, "", common.Coded("SUBSCRIPTION_EMPTY", fmt.Errorf("no server found in the subscription response"), nil)
	}
	// The Subscription-Userinfo header is the standard form of the usage; a
	// STATUS= line in the body is the provider's own prose for the same
	// numbers. Showing both printed every figure twice in two formats.
	sui := parseSubscriptionUserInfo(res.Header.Get("Subscription-Userinfo"))
	if sui.Known() {
		status = sui.String()
	}
	return infos, status, nil
}

func ResolveByLines(raw string) (infos []serverObj.ServerObj, status string, err error) {
	var sip SIP008
	if infos, sip, err = resolveSIP008(raw); err == nil {
		status = getDataUsageStatus(sip.BytesUsed, sip.BytesRemaining)
	} else if clashInfos, ok, clashErr := clash.Resolve(raw); ok {
		// a Clash config is never a link list; parsing it line by line
		// would only pick up stray URLs from its rules
		infos, err = clashInfos, clashErr
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
	subscription := configure.GetSubscription(index)
	if subscription == nil {
		return common.Coded("SUBSCRIPTION_NOT_FOUND", fmt.Errorf("subscription #%d no longer exists; reload the page", index+1), map[string]interface{}{"id": index + 1})
	}
	addr := subscription.Address
	c := httpClient.GetHttpClientAutomatically()
	resolv.CheckResolvConf()
	subscriptionInfos, status, err := ResolveSubscriptionWithClient(addr, c)
	if err != nil {
		log.Warn("Subscription fetch failed: %v", err)
		return fmt.Errorf("could not fetch subscription from %s: %w", subscriptionHost(addr), err)
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
				if sRaw.ServerObj == nil {
					log.Warn("UpdateSubscription: skipping connected server with nil ServerObj (Sub=%d, ID=%d)", cs.Sub, cs.ID)
					continue
				}
				link := sRaw.ServerObj.ExportToURL()
				link2Raw[link] = sRaw
				connectedVmessInfo2CssIndex[link] = append(connectedVmessInfo2CssIndex[link], i)
			}
		}
	}
	// Replace list with new one, and find one with the same server value as the current connection, set it as Connected; if none, disconnect
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
	// Fallback: subscription providers often rename nodes (traffic/expiry counters in
	// names) while keeping the same endpoint. Remap remaining connected servers to new
	// servers with the same protocol, hostname and port.
	looseKey := func(obj serverObj.ServerObj) string {
		return obj.GetProtocol() + "://" + net.JoinHostPort(obj.GetHostname(), strconv.Itoa(obj.GetPort()))
	}
	loose2Index := make(map[string]int)
	for i, info := range subscriptionInfos {
		k := looseKey(info)
		if _, ok := loose2Index[k]; !ok {
			loose2Index[k] = i
		}
	}
	var connectedServerChanged bool
	for link, cssIndexes := range connectedVmessInfo2CssIndex {
		if i, ok := loose2Index[looseKey(link2Raw[link].ServerObj)]; ok {
			for _, cssIndex := range cssIndexes {
				cssAfter[cssIndex].ID = i + 1
			}
			connectedServerChanged = true
			log.Info("UpdateSubscription: remapped connected server %v to %v by endpoint match",
				link2Raw[link].ServerObj.GetName(), subscriptionInfos[i].GetName())
			delete(connectedVmessInfo2CssIndex, link)
		}
	}
	// Last fallback: some providers keep node names stable but rotate IPs, which
	// defeats the endpoint match. Remap by name, but only when the name maps to
	// exactly one server on each side to avoid connecting to a different node.
	name2Index := make(map[string]int)
	nameCount := make(map[string]int)
	for i, info := range subscriptionInfos {
		nameCount[info.GetName()]++
		name2Index[info.GetName()] = i
	}
	oldNameCount := make(map[string]int)
	for link := range connectedVmessInfo2CssIndex {
		oldNameCount[link2Raw[link].ServerObj.GetName()]++
	}
	for link, cssIndexes := range connectedVmessInfo2CssIndex {
		name := link2Raw[link].ServerObj.GetName()
		if nameCount[name] != 1 || oldNameCount[name] != 1 {
			continue
		}
		i := name2Index[name]
		for _, cssIndex := range cssIndexes {
			cssAfter[cssIndex].ID = i + 1
		}
		connectedServerChanged = true
		log.Info("UpdateSubscription: remapped connected server %v (%v -> %v) by unique name match",
			name, link2Raw[link].ServerObj.GetHostname(), subscriptionInfos[i].GetHostname())
		delete(connectedVmessInfo2CssIndex, link)
	}
	for link, cssIndexes := range connectedVmessInfo2CssIndex {
		for _, cssIndex := range cssIndexes {
			if disconnectIfNecessary {
				err = Disconnect(*css.Get()[cssIndex], false)
				if err != nil {
					return fmt.Errorf("could not disconnect the server that left the subscription: %w", err)
				}
			} else {
				// Append previously connected node
				infoServerRaws = append(infoServerRaws, *link2Raw[link])
				cssAfter[cssIndex].ID = len(infoServerRaws)
			}
		}
	}
	if err := configure.OverwriteConnects(configure.NewWhiches(cssAfter)); err != nil {
		return err
	}
	subscription = configure.GetSubscription(index)
	if subscription == nil {
		return common.Coded("SUBSCRIPTION_NOT_FOUND", fmt.Errorf("subscription #%d no longer exists; reload the page", index+1), map[string]interface{}{"id": index + 1})
	}
	subscription.Servers = infoServerRaws
	subscription.Status = string(touch.NewUpdateStatus())
	subscription.Info = status
	if err := configure.SetSubscription(index, subscription); err != nil {
		return err
	}
	// A remapped connection may point at a server whose config differs from the old
	// one; the running core keeps using the old config until it is regenerated.
	if connectedServerChanged && v2ray.ProcessManager.Running() {
		if err := v2ray.UpdateV2RayConfig(); err != nil {
			log.Warn("UpdateSubscription: failed to reload core after remapping connected servers: %v", err)
		}
	}
	return nil
}

func ModifySubscriptionRemark(subscription touch.Subscription) (err error) {
	raw := configure.GetSubscription(subscription.ID - 1)
	if raw == nil {
		return common.Coded("SUBSCRIPTION_NOT_FOUND", fmt.Errorf("subscription #%d does not exist; reload the page", subscription.ID), map[string]interface{}{"id": subscription.ID})
	}
	raw.Remarks = subscription.Remarks
	raw.Address = subscription.Address
	raw.AutoSelect = subscription.AutoSelect
	return configure.SetSubscription(subscription.ID-1, raw)
}

func SelectServersFromSubscription(index int, shouldDisconnect bool) (err error) {
	var subscriptionServer configure.Which
	subscriptionServer.TYPE = "subscriptionServer"
	subscriptionServer.Sub = index // Subscription IDs start with 0
	subscriptionServer.Outbound = "proxy"
	if shouldDisconnect {
		connections := configure.GetConnectedServersByOutbound(subscriptionServer.Outbound)
		if connections == nil {
			return nil
		}
		remaining := make([]configure.Which, 0, connections.Len())
		var found bool
		for _, connected := range connections.Get() {
			if connected.TYPE == configure.SubscriptionServerType && connected.Sub == index {
				found = true
				continue
			}
			remaining = append(remaining, *connected)
		}
		if !found {
			return nil
		}
		return ReplaceOutboundConnections(subscriptionServer.Outbound, remaining)
	}

	for i := 1; i < configure.GetLenSubscriptionServers(index)+1; i++ {
		subscriptionServer.ID = i // Server IDs start with 1
		sub := configure.GetSubscription(index)
		if sub == nil {
			return common.Coded("SUBSCRIPTION_NOT_FOUND", fmt.Errorf("subscription #%d no longer exists", index+1), map[string]interface{}{"id": index + 1})
		}
		serverObj := sub.Servers[i-1].ServerObj // ServerObj IDs start with 0
		if serverObj == nil {
			log.Warn("[AutoSelect] Skipping server %d in subscription %d: nil ServerObj", i, index)
			continue
		}
		serverName := serverObj.GetName()

		// Workaround for partial SS support in v2fly and xray
		isSupported, _ := IsSupported(subscriptionServer)
		if !isSupported {
			log.Info("[AutoSelect] Skipping unsupported server %v", serverName)
			continue
		}

		err := Connect(&subscriptionServer)
		if err == nil {
			log.Info("[AutoSelect] Automatically selected server: %v", serverName)
		} else {
			log.Error("[AutoSelect] Failed to connect to server: %v", serverName)
			return err
		}
	}
	return nil
}

func AutoSelectServersFromSubscriptions(shouldDisconnect bool) (err error) {
	for i := 0; i < configure.GetLenSubscriptions(); i++ {
		subscription := configure.GetSubscription(i)
		if subscription == nil {
			log.Warn("[AutoSelect] Failed to read subscription at index %d, skipping", i)
			continue
		}
		if subscription.AutoSelect {
			if shouldDisconnect {
				log.Info("[AutoSelect] Automatically disconnecting servers from subscription: %v", subscription.Address)
				err := SelectServersFromSubscription(i, true)
				if err != nil {
					log.Error("[AutoSelect] Failed to disconnect servers from subscription: %v", subscription.Address)
					return err
				}
			} else {
				log.Info("[AutoSelect] Automatically selecting servers from subscription: %v", subscription.Address)
				err := SelectServersFromSubscription(i, false)
				if err != nil {
					log.Error("[AutoSelect] Failed to select servers from subscription: %v", subscription.Address)
					return err
				}
			}
		}
	}
	return nil
}
