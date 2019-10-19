package models

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type Template struct {
	Log      Log        `json:"log"`
	Inbounds []Inbounds `json:"inbounds"`
	Outbound Outbound   `json:"outbound"`
}
type Log struct {
	Access   string `json:"access"`
	Error    string `json:"error"`
	Loglevel string `json:"loglevel"`
}
type Sniffing struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
}
type Inbounds struct {
	Port     int      `json:"port"`
	Listen   string   `json:"listen"`
	Protocol string   `json:"protocol"`
	Sniffing Sniffing `json:"sniffing"`
	Settings struct {
		Auth    string      `json:"auth"`
		UDP     bool        `json:"udp"`
		IP      interface{} `json:"ip"`
		Clients interface{} `json:"clients"`
	} `json:"settings"`
	StreamSettings interface{} `json:"streamSettings"`
}
type Users struct {
	ID       string `json:"id"`
	AlterID  int    `json:"alterId"`
	Security string `json:"security"`
}
type Vnext struct {
	Address string  `json:"address"`
	Port    int     `json:"port"`
	Users   []Users `json:"users"`
}
type Settings struct {
	Vnext   []Vnext     `json:"vnext"`
	Servers interface{} `json:"servers"`
}
type TLSSettings struct {
	AllowInsecure bool        `json:"allowInsecure"`
	ServerName    interface{} `json:"serverName"`
}
type Headers struct {
	Host string `json:"Host"`
}
type WsSettings struct {
	ConnectionReuse bool    `json:"connectionReuse"`
	Path            string  `json:"path"`
	Headers         Headers `json:"headers"`
}
type StreamSettings struct {
	Network      string      `json:"network"`
	Security     string      `json:"security"`
	TLSSettings  interface{} `json:"tlsSettings"`
	TCPSettings  interface{} `json:"tcpSettings"`
	KcpSettings  interface{} `json:"kcpSettings"`
	WsSettings   interface{} `json:"wsSettings"`
	HTTPSettings interface{} `json:"httpSettings"`
}
type Mux struct {
	Enabled     bool `json:"enabled"`
	Concurrency int  `json:"concurrency"`
}
type Outbound struct {
	Tag            string         `json:"tag"`
	Protocol       string         `json:"protocol"`
	Settings       Settings       `json:"settings"`
	StreamSettings StreamSettings `json:"streamSettings"`
	Mux            Mux            `json:"mux"`
}
type TCPSettings struct {
	ConnectionReuse bool `json:"connectionReuse"`
	Header          struct {
		Type    string `json:"type"`
		Request struct {
			Version string   `json:"version"`
			Method  string   `json:"method"`
			Path    []string `json:"path"`
			Headers struct {
				Host           []string `json:"Host"`
				UserAgent      []string `json:"User-Agent"`
				AcceptEncoding []string `json:"Accept-Encoding"`
				Connection     []string `json:"Connection"`
				Pragma         string   `json:"Pragma"`
			} `json:"headers"`
		} `json:"request"`
		Response struct {
			Version string `json:"version"`
			Status  string `json:"status"`
			Reason  string `json:"reason"`
			Headers struct {
				ContentType      []string `json:"Content-Type"`
				TransferEncoding []string `json:"Transfer-Encoding"`
				Connection       []string `json:"Connection"`
				Pragma           string   `json:"Pragma"`
			} `json:"headers"`
		} `json:"response"`
	} `json:"header"`
}
type KcpSettings struct {
	Mtu              int  `json:"mtu"`
	Tti              int  `json:"tti"`
	UplinkCapacity   int  `json:"uplinkCapacity"`
	DownlinkCapacity int  `json:"downlinkCapacity"`
	Congestion       bool `json:"congestion"`
	ReadBufferSize   int  `json:"readBufferSize"`
	WriteBufferSize  int  `json:"writeBufferSize"`
	Header           struct {
		Type     string      `json:"type"`
		Request  interface{} `json:"request"`
		Response interface{} `json:"response"`
	} `json:"header"`
}
type HttpSettings struct {
	Path string   `json:"path"`
	Host []string `json:"host"`
}

func NewTemplate() (tmpl *Template) {
	tmpl = new(Template)
	return
}

/*
根据传入的 vmess://xxxxx 填充模板
*/
func (t *Template) ImportFromURL(v VmessInfo) error {
	var tmpljson struct {
		Template     Template     `json:"template"`
		TCPSettings  TCPSettings  `json:"tcpSettings"`
		WsSettings   WsSettings   `json:"wsSettings"`
		TLSSettings  TLSSettings  `json:"tlsSettings"`
		KcpSettings  KcpSettings  `json:"kcpSettings"`
		HttpSettings HttpSettings `json:"httpSettings"`
	}
	// 读入模板json，该json是精心准备过的，直接unmarshal到tmpljson上
	raw, err := ioutil.ReadFile("models/template.json")
	if err != nil {
		log.Fatal(err)
	}
	_ = json.Unmarshal(raw, &tmpljson)
	// 其中Template是基础配置，替换掉*t即可
	*t = tmpljson.Template
	// 进行适配性修改
	t.Outbound.Settings.Vnext[0].Address = v.Add
	port, _ := strconv.Atoi(v.Port)
	t.Outbound.Settings.Vnext[0].Port = port
	t.Outbound.Settings.Vnext[0].Users[0].ID = v.ID
	aid, _ := strconv.Atoi(v.Aid)
	t.Outbound.Settings.Vnext[0].Users[0].AlterID = aid
	if strings.ToLower(v.TLS) == "tls" {
		t.Outbound.StreamSettings.Security = "tls"
		t.Outbound.StreamSettings.TLSSettings = tmpljson.TLSSettings
	}
	t.Outbound.StreamSettings.Network = v.Net
	// 根据传输协议(network)修改streamSettings
	switch strings.ToLower(v.Net) {
	case "ws":
		tmpljson.WsSettings.Headers.Host = v.Host
		tmpljson.WsSettings.Path = v.Path
		t.Outbound.StreamSettings.WsSettings = tmpljson.WsSettings
	case "kcp":
		tmpljson.KcpSettings.Header.Type = v.Type
		t.Outbound.StreamSettings.KcpSettings = tmpljson.KcpSettings
	case "tcp":
		if strings.ToLower(v.Type) != "none" {
			tmpljson.TCPSettings.Header.Request.Headers.Host = strings.Split(v.Host, ",")
			t.Outbound.StreamSettings.TCPSettings = tmpljson.TCPSettings
		}
	case "h2":
		tmpljson.HttpSettings.Host = strings.Split(v.Host, ",")
		tmpljson.HttpSettings.Path = v.Path
		t.Outbound.StreamSettings.HTTPSettings = tmpljson.HttpSettings
	}
	return nil
}
