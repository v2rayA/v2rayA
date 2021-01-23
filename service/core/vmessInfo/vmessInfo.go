package vmessInfo

import (
	"encoding/base64"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"net"
	"net/url"
	"strings"
)

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
	TLS           string `json:"tls"`
	Flow          string `json:"flow,omitempty"`
	V             string `json:"v"`
	AllowInsecure bool   `json:"allowInsecure"`
	Protocol      string `json:"protocol"`
}

func (v *VmessInfo) ExportToURL() string {
	switch v.Protocol {
	case "vless":
		//FIXME: 临时方案
		fallthrough
	case "", "vmess":
		if v.V == "" {
			v.V = "2"
		}
		b, _ := jsoniter.Marshal(v)
		return "vmess://" + base64.StdEncoding.EncodeToString(b)
	case "ss":
		/* ss://BASE64(method:password)@server:port#name */
		u := &url.URL{
			Scheme:   "ss",
			User:     url.User(base64.URLEncoding.EncodeToString([]byte(v.Net + ":" + v.ID))),
			Host:     net.JoinHostPort(v.Add, v.Port),
			Path:     "/",
			RawQuery: "",
			Fragment: v.Ps,
		}
		if v.Type != "" {
			a := []string{
				`obfs-local`,
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
	case "pingtunnel":
		// pingtunnel://server:URLBASE64(passwd)#URLBASE64(remarks)
		return fmt.Sprintf("pingtunnel://%v", base64.URLEncoding.EncodeToString([]byte(
			fmt.Sprintf("%v:%v#%v",
				v.Add,
				base64.URLEncoding.EncodeToString([]byte(v.ID)),
				common.UrlEncoded(v.Ps),
			),
		)))
	case "trojan":
		// trojan://passwd@server:port#URLESCAPE(remarks)
		u := &url.URL{
			Scheme:   "trojan",
			User:     url.User(v.ID),
			Host:     net.JoinHostPort(v.Add, v.Port),
			Path:     "/",
			RawQuery: "",
			Fragment: v.Ps,
		}
		q := u.Query()
		if v.Host != "" {
			q.Set("sni", v.Host)
		}
		if v.AllowInsecure {
			q.Set("allowInsecure", "1")
		}
		u.RawQuery = q.Encode()
		return u.String()
	}
	return ""
}
