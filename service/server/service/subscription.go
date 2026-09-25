package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/v2rayA/v2rayA/conf"
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
	return resolveSubscriptionWithContext(context.Background(), source, client)
}

func resolveSubscriptionWithContext(ctx context.Context, source string, client *http.Client) (infos []serverObj.ServerObj, status string, err error) {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("v2rayA/%s WebRequestHelper", conf.Version))
	res, err := c.Do(req)
	if err != nil && ctx.Err() == nil {
		direct := *httpClient.DirectSubscriptionClient()
		direct.Timeout = c.Timeout
		res, err = direct.Do(req)
	}
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
	c := subscriptionHTTPClient()
	resolv.CheckResolvConf()
	subscriptionInfos, status, err := ResolveSubscriptionWithClient(addr, c)
	if err != nil {
		log.Warn("Subscription fetch failed: %v", err)
		return fmt.Errorf("could not fetch subscription from %s: %w", subscriptionHost(addr), err)
	}
	return storeSubscriptionUpdate(index, subscription, subscriptionInfos, status, disconnectIfNecessary)
}

func storeSubscriptionUpdate(index int, old *configure.SubscriptionRaw, nodes []serverObj.ServerObj, info string, disconnect bool) error {
	next := *old
	next.Servers = make([]configure.ServerRaw, len(nodes))
	for i, node := range nodes {
		next.Servers[i].ServerObj = node
	}
	next.Status, next.Info = string(touch.NewUpdateStatus()), info
	previous := configure.GetConnectedServers()
	affected := false
	updated := configure.NewNodeRefs(nil)
	for _, ref := range previous.Get() {
		copy := *ref
		if ref.TYPE == configure.SubscriptionServerType && ref.Sub == index {
			if ref.ID <= 0 || ref.ID > len(old.Servers) {
				return fmt.Errorf("invalid connected server reference")
			}
			raw := old.Servers[ref.ID-1]
			copy.ID = remapSubscriptionNode(raw.ServerObj, nodes)
			if copy.ID == 0 {
				if disconnect || configure.GetOutboundSetting(ref.Outbound).AutoAdd {
					affected = true
					continue
				}
				next.Servers = append(next.Servers, raw)
				copy.ID = len(next.Servers)
			}
			if copy.ID != ref.ID || next.Servers[copy.ID-1].ServerObj.ExportToURL() != raw.ServerObj.ExportToURL() {
				affected = true
			}
		}
		updated.Add(copy)
	}
	if !affected {
		return configure.SetSubscriptionAndConnects(index, &next, updated)
	}
	return ApplyGroupConfig(func() func() error {
		return func() error { return configure.SetSubscriptionAndConnects(index, old, previous) }
	}, func() error { return configure.SetSubscriptionAndConnects(index, &next, updated) })
}

// Prefer full identity; accept a renamed/rotated endpoint only when its match
// is unambiguous. In particular, two accounts on one endpoint are not interchangeable.
func remapSubscriptionNode(old serverObj.ServerObj, nodes []serverObj.ServerObj) int {
	if old == nil {
		return 0
	}
	for i, node := range nodes {
		if node.ExportToURL() == old.ExportToURL() {
			return i + 1
		}
	}
	for _, match := range []func(serverObj.ServerObj) bool{
		func(n serverObj.ServerObj) bool {
			return n.GetProtocol() == old.GetProtocol() && n.GetHostname() == old.GetHostname() && n.GetPort() == old.GetPort()
		},
		func(n serverObj.ServerObj) bool {
			return old.GetName() != "" && n.GetName() == old.GetName() && n.GetProtocol() == old.GetProtocol()
		},
	} {
		found := 0
		for i, node := range nodes {
			if match(node) {
				if found != 0 {
					found = -1
					break
				}
				found = i + 1
			}
		}
		if found > 0 {
			return found
		}
	}
	return 0
}

func ModifySubscriptionRemark(subscription touch.Subscription) error {
	raw := configure.GetSubscription(subscription.ID - 1)
	if raw == nil {
		return common.Coded("SUBSCRIPTION_NOT_FOUND", fmt.Errorf("subscription does not exist"), nil)
	}
	if subscription.AutoUpdate && (subscription.UpdateIntervalMinutes == nil || subscription.FailureIntervalMinutes == nil) {
		return fmt.Errorf("both update intervals are required")
	}
	regular, failure := raw.UpdateIntervalMinutes, raw.FailureIntervalMinutes
	if subscription.UpdateIntervalMinutes != nil {
		regular = *subscription.UpdateIntervalMinutes
	}
	if subscription.FailureIntervalMinutes != nil {
		failure = *subscription.FailureIntervalMinutes
	}
	if failure == 0 && !subscription.AutoUpdate {
		failure = 1
	}
	if regular < 0 || regular > 525600 || failure < 1 || failure > 525600 {
		return fmt.Errorf("regular interval must be 0–525600 minutes; failure interval must be 1–525600 minutes")
	}
	raw.Remarks, raw.Address = subscription.Remarks, subscription.Address
	raw.AutoUpdate, raw.UpdateIntervalMinutes, raw.FailureIntervalMinutes = subscription.AutoUpdate, regular, failure
	return configure.SetSubscription(subscription.ID-1, raw)
}

func subscriptionHTTPClient() *http.Client {
	if configure.GetSettingNotNil().ProxyModeWhenSubscribe == configure.ProxyModeDirect {
		return httpClient.DirectSubscriptionClient()
	}
	return httpClient.GetHttpClientAutomatically()
}
