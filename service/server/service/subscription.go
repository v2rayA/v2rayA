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
	"github.com/v2rayA/v2rayA/kernel/ipforward"
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

const maxSubscriptionDocumentSize int64 = 32 << 20

func readSubscriptionBody(resp *http.Response) ([]byte, error) {
	if resp.ContentLength > maxSubscriptionDocumentSize {
		return nil, fmt.Errorf("subscription document exceeds the 32 MiB limit")
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxSubscriptionDocumentSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > maxSubscriptionDocumentSize {
		return nil, fmt.Errorf("subscription document exceeds the 32 MiB limit")
	}
	return b, nil
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
	b, err := readSubscriptionBody(res)
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
	loc := configure.NewLocator()
	for i, cs := range css.Get() {
		if cs.TYPE == configure.SubscriptionServerType && cs.Sub == index {
			if sRaw, err := loc.Locate(cs); err != nil {
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
	subscription = configure.GetSubscription(index)
	if subscription == nil {
		return common.Coded("SUBSCRIPTION_NOT_FOUND", fmt.Errorf("subscription #%d no longer exists; reload the page", index+1), map[string]interface{}{"id": index + 1})
	}
	subscription.Servers = infoServerRaws
	subscription.Status = string(touch.NewUpdateStatus())
	subscription.Info = status
	if err := configure.SetSubscriptionAndConnects(index, subscription, configure.NewNodeRefs(cssAfter)); err != nil {
		return err
	}
	// A remapped connection may point at a server whose config differs from the old
	// one; the running core keeps using the old config until it is regenerated.
	if connectedServerChanged && v2ray.ProcessManager.Running() {
		if err := v2ray.UpdateV2RayConfig(); err != nil {
			return subscriptionCoreApplyError(err)
		}
	}
	return nil
}

func subscriptionCoreApplyError(err error) error {
	return fmt.Errorf("subscription stored, but the core could not apply it: %w", err)
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
	var subscriptionServer configure.NodeRef
	subscriptionServer.TYPE = "subscriptionServer"
	subscriptionServer.Sub = index // Subscription IDs start with 0
	subscriptionServer.Outbound = "proxy"
	if shouldDisconnect {
		connections := configure.GetConnectedServersByOutbound(subscriptionServer.Outbound)
		if connections == nil {
			return nil
		}
		remaining := make([]configure.NodeRef, 0, connections.Len())
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

	// One replacement for the whole subscription: connecting the members one
	// by one restarted the core once per node.
	sub := configure.GetSubscription(index)
	if sub == nil {
		return common.Coded("SUBSCRIPTION_NOT_FOUND", fmt.Errorf("subscription #%d no longer exists", index+1), map[string]interface{}{"id": index + 1})
	}
	backup := configure.GetConnectedServersByOutbound(subscriptionServer.Outbound)
	var existing []*configure.NodeRef
	if backup != nil {
		existing = backup.Get()
	}
	members := autoSelectMembers(index, sub, existing)
	if len(members) == len(existing) {
		return nil
	}
	// What Connect did once per node: the asset check and the ip forward
	// reconciliation, then the store, then the core.
	if err := checkSupport(nil); err != nil {
		return err
	}
	if setting := GetSetting(); setting.IpForward != ipforward.IsIpForwardOn() {
		if e := ipforward.WriteIpForward(setting.IpForward); e != nil {
			log.Warn("[AutoSelect] %v", e)
		}
	}
	if err := ReplaceOutboundConnections(subscriptionServer.Outbound, members); err != nil {
		return err
	}
	// Connect started the core for a selection made while it was stopped,
	// and dropped the selection when that failed.
	if !v2ray.ProcessManager.Running() {
		if err := v2ray.UpdateV2RayConfig(); err != nil {
			if backup != nil && backup.Len() > 0 {
				_ = configure.OverwriteConnects(backup)
			} else {
				_ = configure.ClearConnects(subscriptionServer.Outbound)
			}
			return err
		}
	}
	return nil
}

// autoSelectMembers appends every supported node of the subscription to the
// members already in the proxy group.
func autoSelectMembers(index int, sub *configure.SubscriptionRaw, existing []*configure.NodeRef) []configure.NodeRef {
	members := make([]configure.NodeRef, 0, len(existing)+len(sub.Servers))
	for _, connected := range existing {
		members = append(members, *connected)
	}
	for i, server := range sub.Servers {
		if server.ServerObj == nil {
			log.Warn("[AutoSelect] Skipping server %d in subscription %d: nil ServerObj", i+1, index)
			continue
		}
		serverName := server.ServerObj.GetName()
		// Workaround for partial SS support in v2fly and xray
		if supported, _ := isSupportedObj(server.ServerObj); !supported {
			log.Info("[AutoSelect] Skipping unsupported server %v", serverName)
			continue
		}
		members = append(members, configure.NodeRef{TYPE: configure.SubscriptionServerType, ID: i + 1, Sub: index, Outbound: "proxy"})
		log.Info("[AutoSelect] Automatically selected server: %v", serverName)
	}
	return members
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
