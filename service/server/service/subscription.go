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
	"net/http"
	"strconv"
	"strings"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/infra/nodeData"
	"github.com/v2rayA/v2rayA/pkg/util/log"
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
func resolveOOCv1(raw string) (infos []*nodeData.NodeData, oocv1 OOCv1, err error) {
	err = json.Unmarshal([]byte(raw), &oocv1)
	if err != nil {
		return
	}

	// Detect protocols
	// TODO: Emit a warning if unsupported protocols exist.
	hasShadowsocks := false
	for _, protocol := range oocv1.Protocols {
		if protocol == "shadowsocks" {
			hasShadowsocks = true
		}
	}

	if hasShadowsocks {
		for _, server := range oocv1.Shadowsocks {
			// Drop servers with plugins
			// TODO: Emit a warning.
			if server.PluginName != "" {
				continue
			}

			infos = append(infos, &nodeData.NodeData{
				VmessInfo: vmessInfo.VmessInfo{
					Ps:       server.Name,
					Add:      server.Address,
					Port:     strconv.Itoa(server.Port),
					ID:       server.Password,
					Net:      server.Method,
					Protocol: "ss",
				},
			})
		}
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
			if !errors.Is(err, ErrorEmptyAddress) {
				log.Warn("resolveByLines: %v: %v", err, row)
			}
			err = nil
			continue
		}
		infos = append(infos, data)
	}
	return
}

func ResolveSubscriptionWithClient(source string, client *http.Client) (infos []*nodeData.NodeData, status string, err error) {
	// Check if source is OOCv1 API token.
	var url string
	var token OOCv1ApiToken
	if err = json.Unmarshal([]byte(source), &token); err == nil {
		url = token.BaseUrl + "/" + token.Secret + "/ooc/v1/" + token.UserId
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
		url = source
	}

	res, err := httpClient.HttpGetUsingSpecificClient(client, url)
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
	subscriptions := configure.GetSubscriptions()
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
	infoServerRaws := make([]configure.ServerRaw, len(subscriptionInfos))
	css := configure.GetConnectedServers()
	cssAfter := css.Get()
	connectedVmessInfo2CssIndex := make(map[vmessInfo.VmessInfo][]int)
	for i, cs := range css.Get() {
		if cs.TYPE == configure.SubscriptionServerType && cs.Sub == index {
			if sRaw, err := cs.LocateServerRaw(); err != nil {
				return err
			} else {
				connectedVmessInfo2CssIndex[sRaw.VmessInfo] = append(connectedVmessInfo2CssIndex[sRaw.VmessInfo], i)
			}
		}
	}
	//将列表更换为新的，并且找到一个跟现在连接的server值相等的，设为Connected，如果没有，则断开连接
	for i, info := range subscriptionInfos {
		infoServerRaw := configure.ServerRaw{
			VmessInfo: info.VmessInfo,
		}
		if cssIndexes, ok := connectedVmessInfo2CssIndex[infoServerRaw.VmessInfo]; ok {
			for _, cssIndex := range cssIndexes {
				cssAfter[cssIndex].ID = i + 1
			}
			delete(connectedVmessInfo2CssIndex, infoServerRaw.VmessInfo)
		}
		infoServerRaws[i] = infoServerRaw
	}
	for vi, cssIndexes := range connectedVmessInfo2CssIndex {
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
				infoServerRaws = append(infoServerRaws, configure.ServerRaw{
					VmessInfo: vi,
					Latency:   "",
				})
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
