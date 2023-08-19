package vmessInfo

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// Deprecated: use serverObj instead.
type VmessInfo struct {
	Ps            string `json:"ps"`
	Add           string `json:"add"`
	Port          string `json:"port"`
	ID            string `json:"id"`
	Aid           string `json:"aid"`
	Net           string `json:"net"`
	Type          string `json:"type"`
	Host          string `json:"host"`
	Path          string `json:"path"`
	SNI           string `json:"sni"`
	TLS           string `json:"tls"`
	Flow          string `json:"flow,omitempty"`
	Alpn          string `json:"alpn,omitempty"` // VLESS only
	V             string `json:"v"`
	AllowInsecure bool   `json:"allowInsecure"`
	Protocol      string `json:"protocol"`
}

func setValue(values *url.Values, key string, value string) {
	if value == "" {
		return
	}
	values.Set(key, value)
}

func (v *VmessInfo) ExportToURL() string {
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
		case "mkcp", "kcp":
			setValue(&query, "headerType", v.Type)
			setValue(&query, "seed", v.Path)
		case "tcp":
			setValue(&query, "headerType", v.Type)
			setValue(&query, "host", v.Host)
			setValue(&query, "path", v.Path)
		case "grpc":
			setValue(&query, "serviceName", v.Path)
		}
		if v.TLS != "none" {
			setValue(&query, "sni", v.SNI)
			setValue(&query, "alpn", v.Alpn)
		}
		if v.TLS == "xtls" {
			setValue(&query, "flow", v.Flow)
		}

		U := url.URL{
			Scheme:   "vless",
			User:     url.User(v.ID),
			Host:     net.JoinHostPort(v.Add, v.Port),
			RawQuery: query.Encode(),
			Fragment: v.Ps,
		}
		return U.String()
	case "", "vmess":
		v.V = "2"
		b, _ := jsoniter.Marshal(v)
		return "vmess://" + base64.StdEncoding.EncodeToString(b)
	case "ss":
		/* ss://BASE64(method:password)@server:port#name */
		u := &url.URL{
			Scheme:   "ss",
			User:     url.User(base64.URLEncoding.EncodeToString([]byte(v.Net + ":" + v.ID))),
			Host:     net.JoinHostPort(v.Add, v.Port),
			Fragment: v.Ps,
		}
		if v.Type != "" {
			a := []string{
				`simple-obfs`,
				`obfs=` + v.Type,
				`obfs-host=` + v.Host,
			}
			if v.Type == "http" {
				a = append(a, `obfs-path=`+v.Path)
			}
			plugin := strings.Join(a, ";")
			q := u.Query()
			q.Set("plugin", plugin)
			u.RawQuery = q.Encode()
		}
		return u.String()
	case "ssr":
		/* ssr://server:port:proto:method:obfs:URLBASE64(password)/?remarks=URLBASE64(remarks)&protoparam=URLBASE64(protoparam)&obfsparam=URLBASE64(obfsparam)) */
		return fmt.Sprintf("ssr://%v", base64.URLEncoding.EncodeToString([]byte(
			fmt.Sprintf(
				"%v:%v:%v:%v:%v/?remarks=%v&protoparam=%v&obfsparam=%v",
				net.JoinHostPort(v.Add, v.Port),
				v.Type,
				v.Net,
				v.TLS,
				base64.URLEncoding.EncodeToString([]byte(v.ID)),
				base64.URLEncoding.EncodeToString([]byte(v.Ps)),
				base64.URLEncoding.EncodeToString([]byte(v.Host)),
				base64.URLEncoding.EncodeToString([]byte(v.Path)),
			),
		)))
	case "trojan", "trojan-go":
		// trojan://passwd@server:port#URLESCAPE(remarks)
		u := &url.URL{
			Scheme:   "trojan",
			User:     url.User(v.ID),
			Host:     net.JoinHostPort(v.Add, v.Port),
			RawQuery: "",
			Fragment: v.Ps,
		}
		q := u.Query()
		if v.AllowInsecure {
			q.Set("allowInsecure", "1")
		}
		if v.Protocol == "trojan-go" {
			u.Scheme = "trojan-go"
			if v.Host != "" {
				fields := strings.SplitN(v.Host, ",", 2)
				q.Set("sni", fields[0])
				q.Set("host", fields[1])
			}
			q.Set("encryption", v.Type)
			q.Set("type", v.Net)
			q.Set("path", v.Path)
		} else {
			if v.Host != "" {
				q.Set("sni", v.Host)
			}
		}
		u.RawQuery = q.Encode()
		return u.String()
	case "http", "https":
		var user *url.Userinfo
		if v.ID != "" && v.Aid != "" {
			user = url.UserPassword(v.ID, v.Aid)
		}
		u := &url.URL{
			Scheme:   v.Protocol,
			User:     user,
			Host:     net.JoinHostPort(v.Add, v.Port),
			Fragment: v.Ps,
		}
		return u.String()
	}
	return ""
}
