package serverObj

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/ntp"
	"github.com/v2rayA/v2rayA/core/coreObj"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func init() {
	FromLinkRegister("vmess", NewV2Ray)
	FromLinkRegister("vless", NewV2Ray)
	EmptyRegister("vmess", func() (ServerObj, error) {
		return new(V2Ray), nil
	})
	EmptyRegister("vless", func() (ServerObj, error) {
		return new(V2Ray), nil
	})
}

type V2Ray struct {
	Ps            string `json:"ps"`
	Add           string `json:"add"`
	Port          string `json:"port"`
	ID            string `json:"id"`
	Aid           string `json:"aid"`
	Security      string `json:"scy"`
	Net           string `json:"net"`
	Type          string `json:"type"`
	Host          string `json:"host"`
	SNI           string `json:"sni,omitempty"`
	Path          string `json:"path"`
	TLS           string `json:"tls"`
	Fingerprint   string `json:"fingerprint,omitempty"`
	PublicKey     string `json:"pbk,omitempty"`
	ShortId       string `json:"sid,omitempty"`
	SpiderX       string `json:"spx,omitempty"`
	Flow          string `json:"flow,omitempty"`
	Alpn          string `json:"alpn,omitempty"`
	AllowInsecure bool   `json:"allowInsecure"`
	Key           string `json:"key,omitempty"`
	QuicSecurity  string `json:"quicSecurity"`
	XHTTPMode              string `json:"xhttpMode,omitempty"`
	XHTTPHeaders           string `json:"xhttpHeaders,omitempty"`           // JSON-encoded map[string]string
	NoGRPCHeader           bool   `json:"noGRPCHeader,omitempty"`
	NoSSEHeader            bool   `json:"noSSEHeader,omitempty"`
	UplinkHTTPMethod       string `json:"uplinkHTTPMethod,omitempty"`
	ScMaxEachPostBytesFrom int    `json:"scMaxEachPostBytesFrom,omitempty"`
	ScMaxEachPostBytesTo   int    `json:"scMaxEachPostBytesTo,omitempty"`
	ScMinPostsIntervalFrom int    `json:"scMinPostsIntervalFrom,omitempty"`
	ScMinPostsIntervalTo   int    `json:"scMinPostsIntervalTo,omitempty"`
	ScMaxBufferedPosts     int    `json:"scMaxBufferedPosts,omitempty"`
	ScStreamUpServerFrom   int    `json:"scStreamUpServerFrom,omitempty"`
	ScStreamUpServerTo     int    `json:"scStreamUpServerTo,omitempty"`
	XPaddingBytesFrom      int    `json:"xPaddingBytesFrom,omitempty"`
	XPaddingBytesTo        int    `json:"xPaddingBytesTo,omitempty"`
	XmuxMaxConcurFrom      int    `json:"xmuxMaxConcurFrom,omitempty"`
	XmuxMaxConcurTo        int    `json:"xmuxMaxConcurTo,omitempty"`
	XmuxMaxConnFrom        int    `json:"xmuxMaxConnFrom,omitempty"`
	XmuxMaxConnTo          int    `json:"xmuxMaxConnTo,omitempty"`
	XmuxCMaxReuseFrom      int    `json:"xmuxCMaxReuseFrom,omitempty"`
	XmuxCMaxReuseTo        int    `json:"xmuxCMaxReuseTo,omitempty"`
	XmuxHMaxReqFrom        int    `json:"xmuxHMaxReqFrom,omitempty"`
	XmuxHMaxReqTo          int    `json:"xmuxHMaxReqTo,omitempty"`
	XmuxHMaxReusableFrom   int    `json:"xmuxHMaxReusableFrom,omitempty"`
	XmuxHMaxReusableTo     int    `json:"xmuxHMaxReusableTo,omitempty"`
	XmuxHKeepAlive         int64  `json:"xmuxHKeepAlive,omitempty"`
	MaxEarlyData           string `json:"maxEarlyData,omitempty"`        // WebSocket Early Data 最大字节数
	EarlyDataHeaderName string `json:"earlyDataHeaderName,omitempty"` // WebSocket Early Data 头部名称
	MultiMode           string `json:"multiMode,omitempty"`           // gRPC MultiMode
	IdleTimeout         string `json:"idleTimeout,omitempty"`         // gRPC IdleTimeout (秒)
	HealthCheckTimeout  string `json:"healthCheckTimeout,omitempty"`  // gRPC HealthCheckTimeout (秒)
	PermitWithoutStream string `json:"permitWithoutStream,omitempty"` // gRPC PermitWithoutStream
	InitialWindowsSize  string `json:"initialWindowsSize,omitempty"`  // gRPC InitialWindowsSize
	V                   string `json:"v"`
	Protocol      string `json:"protocol"`
}

// queryInt parses an integer query parameter; returns 0 if missing or invalid.
func queryInt(u *url.URL, key string) int {
	v := u.Query().Get(key)
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

// setIntQuery sets an integer URL query parameter only when non-zero.
func setIntQuery(q *url.Values, key string, val int) {
	if val != 0 {
		q.Set(key, strconv.Itoa(val))
	}
}

func NewV2Ray(link string) (ServerObj, error) {
	if strings.HasPrefix(link, "vmess://") {
		return ParseVmessURL(link)
	} else if strings.HasPrefix(link, "vless://") {
		return ParseVlessURL(link)
	}
	return nil, ErrInvalidParameter
}

func ParseVlessURL(vless string) (data *V2Ray, err error) {
	u, err := url.Parse(vless)
	if err != nil {
		return nil, err
	}
	data = &V2Ray{
		Ps:            u.Fragment,
		Add:           u.Hostname(),
		Port:          u.Port(),
		ID:            u.User.String(),
		Aid:           u.Query().Get("aid"),
		Net:           u.Query().Get("type"),
		Type:          u.Query().Get("headerType"),
		Host:          u.Query().Get("host"),
		SNI:           u.Query().Get("sni"),
		Path:          u.Query().Get("path"),
		TLS:           u.Query().Get("security"),
		Fingerprint:   u.Query().Get("fp"),
		PublicKey:     u.Query().Get("pbk"),
		ShortId:       u.Query().Get("sid"),
		SpiderX:       u.Query().Get("spx"),
		Flow:          u.Query().Get("flow"),
		Alpn:          u.Query().Get("alpn"),
		AllowInsecure: u.Query().Get("allowInsecure") == "true",
		Key:           u.Query().Get("key"),
		V:             vless,
		Protocol:      "vless",
	}
	if data.Net == "" {
		data.Net = "tcp"
	}
	if data.Net == "grpc" {
		data.Path = u.Query().Get("serviceName")
	}
	if data.Type == "" {
		data.Type = "none"
	}
	if data.Host == "" {
		data.Host = u.Query().Get("host")
	}
	if data.TLS == "" {
		data.TLS = "none"
	}
	if data.Net == "mkcp" || data.Net == "kcp" {
		data.Path = u.Query().Get("seed")
	}
	if data.Net == "quic" {
		data.QuicSecurity = u.Query().Get("quicSecurity")
	}
	if data.Net == "xhttp" {
		data.XHTTPMode = u.Query().Get("xhttpMode")
		if data.XHTTPMode == "" {
			data.XHTTPMode = "auto"
		}
		data.XHTTPHeaders = u.Query().Get("xhttpHeaders")
		data.NoGRPCHeader = u.Query().Get("noGRPCHeader") == "true"
		data.NoSSEHeader = u.Query().Get("noSSEHeader") == "true"
		data.UplinkHTTPMethod = u.Query().Get("uplinkHTTPMethod")
		data.ScMaxEachPostBytesFrom = queryInt(u, "scMaxEachPostBytesFrom")
		data.ScMaxEachPostBytesTo = queryInt(u, "scMaxEachPostBytesTo")
		data.ScMinPostsIntervalFrom = queryInt(u, "scMinPostsIntervalFrom")
		data.ScMinPostsIntervalTo = queryInt(u, "scMinPostsIntervalTo")
		data.ScMaxBufferedPosts = queryInt(u, "scMaxBufferedPosts")
		data.ScStreamUpServerFrom = queryInt(u, "scStreamUpServerFrom")
		data.ScStreamUpServerTo = queryInt(u, "scStreamUpServerTo")
		data.XPaddingBytesFrom = queryInt(u, "xPaddingBytesFrom")
		data.XPaddingBytesTo = queryInt(u, "xPaddingBytesTo")
		data.XmuxMaxConcurFrom = queryInt(u, "xmuxMaxConcurFrom")
		data.XmuxMaxConcurTo = queryInt(u, "xmuxMaxConcurTo")
		data.XmuxMaxConnFrom = queryInt(u, "xmuxMaxConnFrom")
		data.XmuxMaxConnTo = queryInt(u, "xmuxMaxConnTo")
		data.XmuxCMaxReuseFrom = queryInt(u, "xmuxCMaxReuseFrom")
		data.XmuxCMaxReuseTo = queryInt(u, "xmuxCMaxReuseTo")
		data.XmuxHMaxReqFrom = queryInt(u, "xmuxHMaxReqFrom")
		data.XmuxHMaxReqTo = queryInt(u, "xmuxHMaxReqTo")
		data.XmuxHMaxReusableFrom = queryInt(u, "xmuxHMaxReusableFrom")
		data.XmuxHMaxReusableTo = queryInt(u, "xmuxHMaxReusableTo")
		data.XmuxHKeepAlive = int64(queryInt(u, "xmuxHKeepAlive"))
	}
	data.MaxEarlyData = u.Query().Get("maxEarlyData")
	data.EarlyDataHeaderName = u.Query().Get("earlyDataHeaderName")
	data.MultiMode = u.Query().Get("multiMode")
	data.IdleTimeout = u.Query().Get("idleTimeout")
	data.HealthCheckTimeout = u.Query().Get("healthCheckTimeout")
	data.PermitWithoutStream = u.Query().Get("permitWithoutStream")
	data.InitialWindowsSize = u.Query().Get("initialWindowsSize")
	return data, nil
}

func ParseVmessURL(vmess string) (data *V2Ray, err error) {
	var info V2Ray
	// perform base64 decoding and unmarshal to VmessInfo
	raw, err := common.Base64StdDecode(vmess[8:])
	if err != nil {
		raw, err = common.Base64URLDecode(vmess[8:])
	}
	if err != nil {
		// not in json format, try to resolve as vmess://BASE64(Security:ID@Add:Port)?remarks=Ps&obfsParam=Host&Path=Path&obfs=Net&tls=TLS
		var u *url.URL
		u, err = url.Parse(vmess)
		if err != nil {
			return
		}
		re := regexp.MustCompile(`.*:(.+)@(.+):(\d+)`)
		s := strings.Split(vmess[8:], "?")[0]
		s, err = common.Base64StdDecode(s)
		if err != nil {
			s, _ = common.Base64URLDecode(s)
		}
		subMatch := re.FindStringSubmatch(s)
		if subMatch == nil {
			err = fmt.Errorf("unrecognized vmess address")
			return
		}
		q := u.Query()
		ps := q.Get("remarks")
		if ps == "" {
			ps = q.Get("remark")
		}
		obfs := q.Get("obfs")
		obfsParam := q.Get("obfsParam")
		path := q.Get("path")
		if obfs == "kcp" || obfs == "mkcp" {
			m := make(map[string]string)
			//cater to v2rayN definition
			_ = jsoniter.Unmarshal([]byte(obfsParam), &m)
			path = m["seed"]
			obfsParam = ""
		}
		aid := q.Get("alterId")
		if aid == "" {
			aid = q.Get("aid")
		}
		security := q.Get("scy")
		if security == "" {
			security = q.Get("security")
		}
		sni := q.Get("sni")
		info = V2Ray{
			ID:            subMatch[1],
			Add:           subMatch[2],
			Port:          subMatch[3],
			Ps:            ps,
			Host:          obfsParam,
			Path:          path,
			SNI:           sni,
			Net:           obfs,
			Aid:           aid,
			Security:      security,
			TLS:           map[string]string{"1": "tls"}[q.Get("tls")],
			AllowInsecure: false,
		}
		if info.Net == "websocket" {
			info.Net = "ws"
		}
	} else {
		// fuzzily parse allowInsecure
		if allowInsecure := gjson.Get(raw, "allowInsecure"); allowInsecure.Exists() {
			if newRaw, err := sjson.Set(raw, "allowInsecure", allowInsecure.Bool()); err == nil {
				raw = newRaw
			}
		}
		err = jsoniter.Unmarshal([]byte(raw), &info)
		if err != nil {
			return
		}
	}
	// correct the wrong vmess as much as possible
	if strings.HasPrefix(info.Host, "/") && info.Path == "" {
		info.Path = info.Host
		info.Host = ""
	}
	if info.Aid == "" {
		info.Aid = "0"
	}
	info.Protocol = "vmess"
	return &info, nil
}

func (v *V2Ray) Configuration(info PriorInfo) (c Configuration, err error) {
	core := coreObj.OutboundObject{
		Tag:      info.Tag,
		Protocol: v.Protocol,
	}
	port, _ := strconv.Atoi(v.Port)
	switch strings.ToLower(v.Protocol) {
	case "vmess", "vless":
		id := v.ID
		network := v.Net
		if l := len([]byte(id)); l < 32 || l > 36 {
			id = common.StringToUUID5(id)
		}
		core.StreamSettings = &coreObj.StreamSettings{
			Network: network,
		}
		switch strings.ToLower(v.Protocol) {
		case "vmess":
			if ok, t, err := ntp.IsDatetimeSynced(); err == nil && !ok {
				return Configuration{}, fmt.Errorf("please sync datetime first. Your datetime is %v, and the "+
					"correct datetime is %v", time.Now().Local().Format(ntp.DisplayFormat), t.Local().Format(ntp.DisplayFormat))
			}
			security := v.Security
			if security == "" {
				security = "auto"
			}
			var aid int
			if _aid, err := strconv.Atoi(v.Aid); err == nil {
				aid = _aid
			}
			core.Settings.Vnext = []coreObj.Vnext{
				{
					Address: v.Add,
					Port:    port,
					Users: []coreObj.User{
						{
							ID:       id,
							AlterID:  aid,
							Security: security,
						},
					},
				},
			}
		case "vless":
			security := v.Security
			if security == "" {
				security = "auto"
			}
			core.Settings.Vnext = []coreObj.Vnext{
				{
					Address: v.Add,
					Port:    port,
					Users: []coreObj.User{
						{
							ID:         id,
							Encryption: "none",
						},
					},
				},
			}
		// if network == "tcp" {
		// 	tcpSetting := coreObj.TCPSettings{
		// 		Header: coreObj.TCPHeader{
		// 			Type: "none",
		// 		},
		// 	}
		// 	core.StreamSettings.TCPSettings = &tcpSetting
		// }
		}
		// 根据传输协议(network)修改streamSettings
		//TODO: QUIC
		switch strings.ToLower(v.Net) {
		case "grpc":
			if v.Path == "" {
				v.Path = "GunService"
			}
			grpcSettings := coreObj.GrpcSettings{
				ServiceName: v.Path,
			}
			// 解析 gRPC MultiMode
			if v.MultiMode != "" {
				if mm, err := strconv.ParseBool(v.MultiMode); err == nil {
					grpcSettings.MultiMode = mm
				}
			}
			// 解析 gRPC IdleTimeout
			if v.IdleTimeout != "" {
				if it, err := strconv.Atoi(v.IdleTimeout); err == nil {
					grpcSettings.IdleTimeout = it
				}
			}
			// 解析 gRPC HealthCheckTimeout
			if v.HealthCheckTimeout != "" {
				if hct, err := strconv.Atoi(v.HealthCheckTimeout); err == nil {
					grpcSettings.HealthCheckTimeout = hct
				}
			}
			// 解析 gRPC PermitWithoutStream
			if v.PermitWithoutStream != "" {
				if pws, err := strconv.ParseBool(v.PermitWithoutStream); err == nil {
					grpcSettings.PermitWithoutStream = pws
				}
			}
			// 解析 gRPC InitialWindowsSize
			if v.InitialWindowsSize != "" {
				if iws, err := strconv.Atoi(v.InitialWindowsSize); err == nil {
					grpcSettings.InitialWindowsSize = iws
				}
			}
			core.StreamSettings.GrpcSettings = &grpcSettings
		case "ws":
			wsSettings := coreObj.WsSettings{
				Path: v.Path,
				Headers: coreObj.Headers{
					Host: v.Host,
				},
			}
			// 解析 WebSocket Early Data
			if v.MaxEarlyData != "" {
				if med, err := strconv.Atoi(v.MaxEarlyData); err == nil && med > 0 {
					wsSettings.MaxEarlyData = med
					if v.EarlyDataHeaderName != "" {
						wsSettings.EarlyDataHeaderName = v.EarlyDataHeaderName
					}
				}
			}
			core.StreamSettings.WsSettings = &wsSettings
		case "mkcp", "kcp":
			core.StreamSettings.KcpSettings = &coreObj.KcpSettings{
				Mtu:              1350,
				Tti:              50,
				UplinkCapacity:   12,
				DownlinkCapacity: 100,
				Congestion:       false,
				ReadBufferSize:   2,
				WriteBufferSize:  2,
				Header: coreObj.KcpHeader{
					Type: v.Type,
				},
				Seed: v.Path,
			}
		case "tcp":
			if strings.ToLower(v.Type) == "http" {
				tcpSetting := coreObj.TCPSettings{
					ConnectionReuse: true,
					Header: coreObj.TCPHeader{
						Type: "http",
						Request: coreObj.HTTPRequest{
							Version: "1.1",
							Method:  "GET",
							Path:    strings.Split(v.Path, ","),
							Headers: coreObj.HTTPReqHeader{
								Host: strings.Split(v.Host, ","),
								UserAgent: []string{
									"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.75 Safari/537.36",
									"Mozilla/5.0 (iPhone; CPU iPhone OS 10_0_2 like Mac OS X) AppleWebKit/601.1 (KHTML, like Gecko) CriOS/53.0.2785.109 Mobile/14A456 Safari/601.1.46",
								},
								AcceptEncoding: []string{"gzip, deflate"},
								Connection:     []string{"keep-alive"},
								Pragma:         "no-cache",
							},
						},
						Response: coreObj.HTTPResponse{
							Version: "1.1",
							Status:  "200",
							Reason:  "OK",
							Headers: coreObj.HTTPRespHeader{
								ContentType:      []string{"application/octet-stream", "video/mpeg"},
								TransferEncoding: []string{"chunked"},
								Connection:       []string{"keep-alive"},
								Pragma:           "no-cache",
							},
						},
					},
				}
				tcpSetting.Header.Request.Headers.Host = strings.Split(v.Host, ",")
				if v.Path != "" {
					tcpSetting.Header.Request.Path = strings.Split(v.Path, ",")
					for i := range tcpSetting.Header.Request.Path {
						if !strings.HasPrefix(tcpSetting.Header.Request.Path[i], "/") {
							tcpSetting.Header.Request.Path[i] = "/" + tcpSetting.Header.Request.Path[i]
						}
					}
				}
				core.StreamSettings.TCPSettings = &tcpSetting
			}
		case "h2", "http":
			if v.Host != "" {
				core.StreamSettings.HTTPSettings = &coreObj.HttpSettings{
					Path: v.Path,
					Host: strings.Split(v.Host, ","),
				}
			} else {
				core.StreamSettings.HTTPSettings = &coreObj.HttpSettings{
					Path: v.Path,
				}
			}
		case "quic":
			core.StreamSettings.QuicSettings = &coreObj.QuicSettings{
				Header: coreObj.KcpHeader{
					Type: v.Type,
				},
				Key:      v.Key,
				Security: v.QuicSecurity,
			}
		case "xhttp":
			xs := &coreObj.XHTTPSettings{
				Path:             v.Path,
				Host:             v.Host,
				Mode:             v.XHTTPMode,
				NoGRPCHeader:     v.NoGRPCHeader,
				NoSSEHeader:      v.NoSSEHeader,
				UplinkHTTPMethod: v.UplinkHTTPMethod,
			}
			// Parse custom headers
			if v.XHTTPHeaders != "" {
				var hdrs map[string]string
				if err := jsoniter.UnmarshalFromString(v.XHTTPHeaders, &hdrs); err == nil {
					xs.Headers = hdrs
				}
			}
			// Range configs
			if v.ScMaxEachPostBytesFrom != 0 || v.ScMaxEachPostBytesTo != 0 {
				xs.ScMaxEachPostBytes = &coreObj.XHTTPRangeConfig{From: int32(v.ScMaxEachPostBytesFrom), To: int32(v.ScMaxEachPostBytesTo)}
			}
			if v.ScMinPostsIntervalFrom != 0 || v.ScMinPostsIntervalTo != 0 {
				xs.ScMinPostsIntervalMs = &coreObj.XHTTPRangeConfig{From: int32(v.ScMinPostsIntervalFrom), To: int32(v.ScMinPostsIntervalTo)}
			}
			if v.ScMaxBufferedPosts != 0 {
				xs.ScMaxBufferedPosts = int64(v.ScMaxBufferedPosts)
			}
			if v.ScStreamUpServerFrom != 0 || v.ScStreamUpServerTo != 0 {
				xs.ScStreamUpServerSecs = &coreObj.XHTTPRangeConfig{From: int32(v.ScStreamUpServerFrom), To: int32(v.ScStreamUpServerTo)}
			}
			if v.XPaddingBytesFrom != 0 || v.XPaddingBytesTo != 0 {
				xs.XPaddingBytes = &coreObj.XHTTPRangeConfig{From: int32(v.XPaddingBytesFrom), To: int32(v.XPaddingBytesTo)}
			}
			// Xmux
			xmux := &coreObj.XHTTPXmux{HKeepAlivePeriod: v.XmuxHKeepAlive}
			if v.XmuxMaxConcurFrom != 0 || v.XmuxMaxConcurTo != 0 {
				xmux.MaxConcurrency = &coreObj.XHTTPRangeConfig{From: int32(v.XmuxMaxConcurFrom), To: int32(v.XmuxMaxConcurTo)}
			}
			if v.XmuxMaxConnFrom != 0 || v.XmuxMaxConnTo != 0 {
				xmux.MaxConnections = &coreObj.XHTTPRangeConfig{From: int32(v.XmuxMaxConnFrom), To: int32(v.XmuxMaxConnTo)}
			}
			if v.XmuxCMaxReuseFrom != 0 || v.XmuxCMaxReuseTo != 0 {
				xmux.CMaxReuseTimes = &coreObj.XHTTPRangeConfig{From: int32(v.XmuxCMaxReuseFrom), To: int32(v.XmuxCMaxReuseTo)}
			}
			if v.XmuxHMaxReqFrom != 0 || v.XmuxHMaxReqTo != 0 {
				xmux.HMaxRequestTimes = &coreObj.XHTTPRangeConfig{From: int32(v.XmuxHMaxReqFrom), To: int32(v.XmuxHMaxReqTo)}
			}
			if v.XmuxHMaxReusableFrom != 0 || v.XmuxHMaxReusableTo != 0 {
				xmux.HMaxReusableSecs = &coreObj.XHTTPRangeConfig{From: int32(v.XmuxHMaxReusableFrom), To: int32(v.XmuxHMaxReusableTo)}
			}
			// Only attach xmux if any field is set
			if xmux.MaxConcurrency != nil || xmux.MaxConnections != nil ||
				xmux.CMaxReuseTimes != nil || xmux.HMaxRequestTimes != nil ||
				xmux.HMaxReusableSecs != nil || xmux.HKeepAlivePeriod != 0 {
				xs.Xmux = xmux
			}
			core.StreamSettings.XHTTPSettings = xs
		default:
			return Configuration{}, fmt.Errorf("unexpected transport type: %v", v.Net)
		}
		if strings.ToLower(v.TLS) == "tls" {
			core.StreamSettings.Security = "tls"
			core.StreamSettings.TLSSettings = &coreObj.TLSSettings{}
			if v.AllowInsecure {
				core.StreamSettings.TLSSettings.AllowInsecure = true
			}
			// SNI
			if v.SNI != "" {
				core.StreamSettings.TLSSettings.ServerName = v.SNI
			} else if v.Host != "" {
				core.StreamSettings.TLSSettings.ServerName = v.Host
			}
			// Alpn
			if v.Alpn != "" {
				alpn := strings.Split(v.Alpn, ",")
				for i := range alpn {
					alpn[i] = strings.TrimSpace(alpn[i])
				}
				core.StreamSettings.TLSSettings.Alpn = alpn
			}
			// uTLS fingerprint
			core.StreamSettings.TLSSettings.Fingerprint = v.Fingerprint
		} else if strings.ToLower(v.TLS) == "xtls" {
			core.StreamSettings.Security = "xtls"
			core.StreamSettings.XTLSSettings = &coreObj.TLSSettings{}
			// SNI
			if v.SNI != "" {
				core.StreamSettings.XTLSSettings.ServerName = v.SNI
			} else if v.Host != "" {
				core.StreamSettings.XTLSSettings.ServerName = v.Host
			}
			if v.AllowInsecure {
				core.StreamSettings.XTLSSettings.AllowInsecure = true
			}
			if v.Alpn != "" {
				alpn := strings.Split(v.Alpn, ",")
				for i := range alpn {
					alpn[i] = strings.TrimSpace(alpn[i])
				}
				core.StreamSettings.XTLSSettings.Alpn = alpn
			}
		} else if strings.ToLower(v.TLS) == "reality" {
			core.StreamSettings.Security = "reality"
			core.StreamSettings.RealitySettings = &coreObj.RealitySettings{
				ServerName:  v.SNI,
				Fingerprint: v.Fingerprint,
				Show:        false,
				PublicKey:   v.PublicKey,
				ShortID:     v.ShortId,
				SpiderX:     v.SpiderX,
			}
		}
		// Flow
		if v.Flow != "" {
			vnext := core.Settings.Vnext.([]coreObj.Vnext)
			vnext[0].Users[0].Flow = v.Flow
			core.Settings.Vnext = vnext
		}
	}
	return Configuration{
		CoreOutbound: core,
		PluginChain:  "",
		UDPSupport:   true,
	}, nil
}

func (v *V2Ray) ExportToURL() string {
	switch v.Protocol {
	case "vless":
		// https://github.com/XTLS/Xray-core/issues/91
		var query = make(url.Values)
		setValue(&query, "type", v.Net)
		setValue(&query, "security", v.TLS)
		switch v.Net {
		case "websocket", "ws", "http", "h2":
			setValue(&query, "path", v.Path)
			setValue(&query, "host", v.Host)
			if v.MaxEarlyData != "" {
				setValue(&query, "maxEarlyData", v.MaxEarlyData)
			}
			if v.EarlyDataHeaderName != "" {
				setValue(&query, "earlyDataHeaderName", v.EarlyDataHeaderName)
			}
		case "mkcp", "kcp":
			setValue(&query, "headerType", v.Type)
			setValue(&query, "seed", v.Path)
		case "tcp":
			setValue(&query, "headerType", v.Type)
			setValue(&query, "host", v.Host)
			setValue(&query, "path", v.Path)
		case "grpc":
			setValue(&query, "serviceName", v.Path)
			if v.MultiMode != "" {
				setValue(&query, "multiMode", v.MultiMode)
			}
		case "quic":
			setValue(&query, "headerType", v.Type)
			setValue(&query, "key", v.Key)
			setValue(&query, "quicSecurity", v.QuicSecurity)
		case "xhttp":
			setValue(&query, "path", v.Path)
			setValue(&query, "host", v.Host)
			setValue(&query, "xhttpMode", v.XHTTPMode)
			if v.XHTTPHeaders != "" {
				setValue(&query, "xhttpHeaders", v.XHTTPHeaders)
			}
			if v.NoGRPCHeader {
				setValue(&query, "noGRPCHeader", "true")
			}
			if v.NoSSEHeader {
				setValue(&query, "noSSEHeader", "true")
			}
			if v.UplinkHTTPMethod != "" {
				setValue(&query, "uplinkHTTPMethod", v.UplinkHTTPMethod)
			}
			setIntQuery(&query, "scMaxEachPostBytesFrom", v.ScMaxEachPostBytesFrom)
			setIntQuery(&query, "scMaxEachPostBytesTo", v.ScMaxEachPostBytesTo)
			setIntQuery(&query, "scMinPostsIntervalFrom", v.ScMinPostsIntervalFrom)
			setIntQuery(&query, "scMinPostsIntervalTo", v.ScMinPostsIntervalTo)
			setIntQuery(&query, "scMaxBufferedPosts", v.ScMaxBufferedPosts)
			setIntQuery(&query, "scStreamUpServerFrom", v.ScStreamUpServerFrom)
			setIntQuery(&query, "scStreamUpServerTo", v.ScStreamUpServerTo)
			setIntQuery(&query, "xPaddingBytesFrom", v.XPaddingBytesFrom)
			setIntQuery(&query, "xPaddingBytesTo", v.XPaddingBytesTo)
			setIntQuery(&query, "xmuxMaxConcurFrom", v.XmuxMaxConcurFrom)
			setIntQuery(&query, "xmuxMaxConcurTo", v.XmuxMaxConcurTo)
			setIntQuery(&query, "xmuxMaxConnFrom", v.XmuxMaxConnFrom)
			setIntQuery(&query, "xmuxMaxConnTo", v.XmuxMaxConnTo)
			setIntQuery(&query, "xmuxCMaxReuseFrom", v.XmuxCMaxReuseFrom)
			setIntQuery(&query, "xmuxCMaxReuseTo", v.XmuxCMaxReuseTo)
			setIntQuery(&query, "xmuxHMaxReqFrom", v.XmuxHMaxReqFrom)
			setIntQuery(&query, "xmuxHMaxReqTo", v.XmuxHMaxReqTo)
			setIntQuery(&query, "xmuxHMaxReusableFrom", v.XmuxHMaxReusableFrom)
			setIntQuery(&query, "xmuxHMaxReusableTo", v.XmuxHMaxReusableTo)
			if v.XmuxHKeepAlive != 0 {
				setValue(&query, "xmuxHKeepAlive", strconv.FormatInt(v.XmuxHKeepAlive, 10))
			}
		}
		if v.TLS != "none" {
			setValue(&query, "flow", v.Flow)
			setValue(&query, "sni", v.SNI)
			setValue(&query, "alpn", v.Alpn)
			setValue(&query, "allowInsecure", strconv.FormatBool(v.AllowInsecure))
			setValue(&query, "fp", v.Fingerprint)
			if v.TLS == "reality" {
				setValue(&query, "pbk", v.PublicKey)
				setValue(&query, "sid", v.ShortId)
				setValue(&query, "spx", v.SpiderX)
			}
		}

		U := url.URL{
			Scheme:   "vless",
			User:     url.User(v.ID),
			Host:     net.JoinHostPort(v.Add, v.Port),
			RawQuery: query.Encode(),
			Fragment: v.Ps,
		}
		return U.String()
	case "vmess":
		v.V = "2"
		b, _ := jsoniter.Marshal(v)
		return "vmess://" + strings.TrimSuffix(base64.StdEncoding.EncodeToString(b), "=")
	}
	log.Warn("unexpected protocol: %v", v.Protocol)
	return ""
}

func (v *V2Ray) NeedPluginPort() bool {
	return false
}

func (v *V2Ray) ProtoToShow() string {
	if v.TLS != "" && v.TLS != "none" {
		return fmt.Sprintf("%v(%v+%v)", v.Protocol, v.Net, v.TLS)
	}
	return fmt.Sprintf("%v(%v)", v.Protocol, v.Net)
}

func (v *V2Ray) GetProtocol() string {
	return v.Protocol
}

func (v *V2Ray) GetHostname() string {
	return v.Add
}

func (v *V2Ray) GetPort() int {
	p, _ := strconv.Atoi(v.Port)
	return p
}

func (v *V2Ray) GetName() string {
	return v.Ps
}

func (v *V2Ray) SetName(name string) {
	v.Ps = name
}
