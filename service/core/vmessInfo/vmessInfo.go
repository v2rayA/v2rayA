package vmessInfo

import (
	"encoding/base64"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/mzz2017/v2rayA/common"
	"net/url"
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
		nameField := ""
		if v.Ps != "" {
			nameField = "#" + v.Ps
		}
		return fmt.Sprintf(
			"ss://%v@%v:%v%v",
			base64.URLEncoding.EncodeToString([]byte(v.Net+":"+v.ID)),
			v.Add,
			v.Port,
			nameField,
		)
	case "ssr":
		/* ssr://server:port:proto:method:obfs:URLBASE64(password)/?remarks=URLBASE64(remarks)&protoparam=URLBASE64(protoparam)&obfsparam=URLBASE64(obfsparam)) */
		return fmt.Sprintf("ssr://%v", base64.URLEncoding.EncodeToString([]byte(
			fmt.Sprintf(
				"%v:%v:%v:%v:%v:%v/?remarks=%v&protoparam=%v&obfsparam=%v",
				v.Add,
				v.Port,
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
		u, _ := url.Parse(fmt.Sprintf(
			"trojan://%v@%v:%v",
			v.ID,
			v.Add,
			v.Port,
		))
		u.Fragment = v.Ps
		q := u.Query()
		if v.Host != "" {
			q.Set("peer", v.Host)
		}
		if v.AllowInsecure {
			q.Set("allowInsecure", "1")
		}
		u.RawQuery = q.Encode()
		return u.String()
	}
	return ""
}
