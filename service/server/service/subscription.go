package service

import (
	"bytes"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//func ResolveSubscription(source string) (infos []*nodeData.NodeData, err error) {
//	return ResolveSubscriptionWithClient(source, http.DefaultClient)
//}

// OOCv1ApiToken is used to exchange API access info for OOCv1.
type OOCv1ApiToken struct {
	Version    int    `json:"version"`
	BaseUrl    string `json:"baseUrl"`
	Secret     string `json:"secret"`
	UserId     string `json:"userId"`
	CertSha256 string `json:"certSha256"`
}

// OOCv1 contains fields for all supported protocols
// for easy serialization and deserialization.
type OOCv1 struct {
	Username       string   `json:"username"`
	BytesUsed      uint64   `json:"bytesUsed"`
	BytesRemaining uint64   `json:"bytesRemaining"`
	ExpiryDate     int64    `json:"expiryDate"`
	Protocols      []string `json:"protocols"`

	Shadowsocks []OOCv1Shadowsocks `json:"shadowsocks"`
}

// OOCv1Shadowsocks represents a Shadowsocks server
// in the `shadowsocks` array of an OOCv1 JSON document.
type OOCv1Shadowsocks struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Address         string `json:"address"`
	Port            int    `json:"port"`
	Method          string `json:"method"`
	Password        string `json:"password"`
	PluginName      string `json:"pluginName"`
	PluginVersion   string `json:"pluginVersion"`
	PluginOptions   string `json:"pluginOptions"`
	PluginArguments string `json:"pluginArguments"`
}

// resolveOOCv1 unmarshals a raw OOCv1 JSON document into an OOCv1 value.
func resolveOOCv1(raw string) (infos []serverObj.ServerObj, oocv1 OOCv1, err error) {
	err = json.Unmarshal([]byte(raw), &oocv1)
	if err != nil {
		return
	}

	// Detect protocols
	// Emit a warning if unsupported protocols exist.
	for _, protocol := range oocv1.Protocols {
		if protocol != "shadowsocks" {
			log.Warn("OOCv1: unsupported protocol: %v", protocol)
		}
	}

	for _, server := range oocv1.Shadowsocks {
		var sip003 *serverObj.Sip003
		switch server.PluginName {
		case "simple-obfs", "obfs-local":
			sip003 = &serverObj.Sip003{
				Name: server.PluginName,
				Opts: serverObj.ParseSip003Opts(server.PluginOptions),
			}
		case "":
			// no plugin
		default:
			log.Warn("failed to parse (OOCv1): %v: unsupported plugin: %v", server.Name, server.PluginName)
			continue
		}
		u := url.URL{
			Scheme:   "ss",
			User:     url.UserPassword(server.Method, server.Password),
			Host:     net.JoinHostPort(server.Address, strconv.Itoa(server.Port)),
			Fragment: server.Name,
		}
		if sip003 != nil {
			u.RawQuery = url.Values{"plugin": []string{sip003.String()}}.Encode()
		}
		obj, err := serverObj.NewFromLink(u.Scheme, u.String())
		if err != nil {
			log.Warn("failed to parse (OOCv1): %v: %v", server.Name, err)
			continue
		}
		infos = append(infos, obj)
	}

	return
}

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
	err = json.Unmarshal([]byte(raw), &sip)
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

func ResolveSubscriptionWithClient(source string, client *http.Client) (infos []serverObj.ServerObj, status string, err error) {
	c := *client
	if c.Timeout < 30*time.Second {
		c.Timeout = 30 * time.Second
	}

	// Check if source is OOCv1 API token.
	var u string
	var token OOCv1ApiToken
	if err = json.Unmarshal([]byte(source), &token); err == nil {
		u = token.BaseUrl + "/" + token.Secret + "/ooc/v1/" + token.UserId
		if token.CertSha256 != "" {
			client = &http.Client{
				Transport: &http.Transport{TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
						h := crypto.SHA256.New()
						for _, line := range rawCerts {
							h.Write(line)
						}
						fingerprint := hex.EncodeToString(h.Sum(nil))
						if fingerprint == token.CertSha256 {
							return fmt.Errorf("server certificate fingerprint mismatch (actual: %s, expected: %s)", fingerprint, token.CertSha256)
						}
						return nil
					},
				}},
			}
		}
	} else {
		u = source
	}

	res, err := httpClient.HttpGetUsingSpecificClient(client, u)
	client.Timeout = 30 * time.Second
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(res.Body)
	defer res.Body.Close()

	// base64 decode
	raw, err := common.Base64StdDecode(buf.String())
	if err != nil {
		raw, _ = common.Base64URLDecode(buf.String())
	}
	return ResolveLines(raw)
}

func ResolveLines(raw string) (infos []serverObj.ServerObj, status string, err error) {
	var oocv1 OOCv1
	var sip SIP008
	if infos, oocv1, err = resolveOOCv1(raw); err == nil && len(oocv1.Protocols) > 0 {
		status = "Username: " + oocv1.Username + " | " + getDataUsageStatus(oocv1.BytesUsed, oocv1.BytesRemaining)
	} else if infos, sip, err = resolveSIP008(raw); err == nil {
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
	subscriptions := configure.GetSubscriptionsV2()
	addr := subscriptions[index].Address
	c, err := httpClient.GetHttpClientAutomatically()
	if err != nil {
		reason := "failed to get proxy"
		return fmt.Errorf("UpdateSubscription: %v", reason)
	}
	resolv.CheckResolvConf()
	subscriptionInfos, status, err := ResolveSubscriptionWithClient(addr, c)
	if err != nil {
		reason := "failed to resolve subscription address: " + err.Error()
		log.Warn("UpdateSubscription: %v: %v", err, subscriptionInfos)
		return fmt.Errorf("UpdateSubscription: %v", reason)
	}
	infoServerRaws := make([]configure.ServerRawV2, len(subscriptionInfos))
	css := configure.GetConnectedServers()
	cssAfter := css.Get()
	// serverObj.ServerObj is a pointer(interface), and shouldn't be as a key
	link2Raw := make(map[string]*configure.ServerRawV2)
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
		infoServerRaw := configure.ServerRawV2{
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
	raw := configure.GetSubscriptionV2(subscription.ID - 1)
	if raw == nil {
		return fmt.Errorf("failed to find the corresponding subscription")
	}
	raw.Remarks = subscription.Remarks
	return configure.SetSubscription(subscription.ID-1, raw)
}
