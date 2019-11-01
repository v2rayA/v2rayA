package models

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

const templateJson = `{
  "template": {
    "log": {
      "access": "/dev/null",
      "error": "/dev/null",
      "loglevel": "/dev/null"
    },
    "inbounds": [
      {
        "port": 10800,
        "listen": "0.0.0.0",
        "protocol": "socks",
        "sniffing": {
          "enabled": true,
          "destOverride": ["http", "tls"]
        },
        "settings": {
          "auth": "noauth",
          "udp": true,
          "ip": null,
          "clients": null
        },
        "streamSettings": null,
        "tag": "socks"
      },
      {
        "port": 10801,
        "listen": "0.0.0.0",
        "protocol": "http",
        "sniffing": {
          "enabled": true,
          "destOverride": ["http", "tls"]
        },
        "tag": "http"
      },
      {
        "port": 10802,
        "listen": "0.0.0.0",
        "protocol": "http",
        "sniffing": {
          "enabled": true,
          "destOverride": ["http", "tls"]
        },
        "tag": "pac"
      }
    ],
    "outbounds": [
      {
        "tag": "proxy",
        "protocol": "vmess",
        "settings": {
          "vnext": null,
          "servers": null
        },
        "streamSettings": null,
        "mux": null
      },
      {
        "protocol": "freedom",
        "settings": {},
        "tag": "direct"
      }
    ],
    "routing": {
      "domainStrategy": "IPOnDemand",
      "rules": [
        {
          "type": "field",
          "inboundTag": [
            "pac"
          ],
          "outboundTag": "direct",
          "domain": ["geosite:cn"]
        },
        {
          "type": "field",
          "inboundTag": [
            "pac"
          ],
          "outboundTag": "direct",
          "ip": [
            "geoip:cn",
            "geoip:private"
          ]
        }
      ]
    }
  },
  "tcpSettings": {
    "connectionReuse": true,
    "header": {
      "type": "http",
      "request": {
        "version": "1.1",
        "method": "GET",
        "path": ["/"],
        "headers": {
          "Host": ["host"],
          "User-Agent": [
            "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.75 Safari/537.36",
            "Mozilla/5.0 (iPhone; CPU iPhone OS 10_0_2 like Mac OS X) AppleWebKit/601.1 (KHTML, like Gecko) CriOS/53.0.2785.109 Mobile/14A456 Safari/601.1.46"
          ],
          "Accept-Encoding": ["gzip, deflate"],
          "Connection": ["keep-alive"],
          "Pragma": "no-cache"
        }
      },
      "response": {
        "version": "1.1",
        "status": "200",
        "reason": "OK",
        "headers": {
          "Content-Type": ["application/octet-stream", "video/mpeg"],
          "Transfer-Encoding": ["chunked"],
          "Connection": ["keep-alive"],
          "Pragma": "no-cache"
        }
      }
    }
  },
  "wsSettings": {
    "connectionReuse": true,
    "path": "",
    "headers": {
      "Host": "host"
    }
  },
  "tlsSettings": {
    "allowInsecure": true,
    "serverName": null
  },
  "kcpSettings": {
    "mtu": 1350,
    "tti": 50,
    "uplinkCapacity": 12,
    "downlinkCapacity": 100,
    "congestion": false,
    "readBufferSize": 2,
    "writeBufferSize": 2,
    "header": {
      "type": "none",
      "request": null,
      "response": null
    }
  },
  "httpSettings": {
    "path": "path",
    "host": ["host"]
  },
  "streamSettings": {
    "network": "ws",
    "security": "",
    "tlsSettings": null,
    "tcpSettings": null,
    "kcpSettings": null,
    "wsSettings": null,
    "httpSettings": null
  },
  "mux": {
    "enabled": false,
    "concurrency": 8
  }
}

`

/*对应template.json*/
type TmplJson struct {
	Template       Template       `json:"template"`
	TCPSettings    TCPSettings    `json:"tcpSettings"`
	WsSettings     WsSettings     `json:"wsSettings"`
	TLSSettings    TLSSettings    `json:"tlsSettings"`
	KcpSettings    KcpSettings    `json:"kcpSettings"`
	HttpSettings   HttpSettings   `json:"httpSettings"`
	StreamSettings StreamSettings `json:"streamSettings"`
	Mux            Mux            `json:"mux"`
}

type Template struct {
	Log       Log        `json:"log"`
	Inbounds  []Inbound  `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
	Routing   struct {
		DomainStrategy string `json:"domainStrategy"`
		Rules          []struct {
			Type        string   `json:"type"`
			OutboundTag string   `json:"outboundTag"`
			Domain      []string `json:"domain,omitempty"`
			InboundTag  []string `json:"inboundTag"`
			IP          []string `json:"ip,omitempty"`
		} `json:"rules"`
	} `json:"routing"`
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
type Inbound struct {
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
	Tag            string      `json:"tag,omitempty"`
}
type User struct {
	ID       string `json:"id"`
	AlterID  int    `json:"alterId"`
	Security string `json:"security"`
}
type Vnext struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Users   []User `json:"users"`
}
type Server struct {
	Address  string `json:"address"`
	Method   string `json:"method"`
	Ota      bool   `json:"ota"`
	Password string `json:"password"`
	Port     int    `json:"port"`
}
type Settings struct {
	Vnext   interface{} `json:"vnext"`
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
	Network      string        `json:"network"`
	Security     string        `json:"security"`
	TLSSettings  *TLSSettings  `json:"tlsSettings"`
	TCPSettings  *TCPSettings  `json:"tcpSettings"`
	KcpSettings  *KcpSettings  `json:"kcpSettings"`
	WsSettings   *WsSettings   `json:"wsSettings"`
	HTTPSettings *HttpSettings `json:"httpSettings"`
}
type Mux struct {
	Enabled     bool `json:"enabled"`
	Concurrency int  `json:"concurrency"`
}
type Outbound struct {
	Tag            string          `json:"tag"`
	Protocol       string          `json:"protocol"`
	Settings       Settings        `json:"settings"`
	StreamSettings *StreamSettings `json:"streamSettings"`
	Mux            *Mux            `json:"mux"`
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
根据传入的 VmessInfo 填充模板
当协议是shadowsocks时，v.Net对应Method，v.ID对应Password
*/
func (t *Template) FillWithVmessInfo(v VmessInfo) error {
	var tmplJson TmplJson
	// 读入模板json
	raw := []byte(templateJson)
	err := json.Unmarshal(raw, &tmplJson)
	if err != nil {
		return errors.New("读入模板json出错，请检查templateJson变量是否是正确的json格式")
	}
	// 其中Template是基础配置，替换掉*t即可
	*t = tmplJson.Template
	// 进行适配性修改
	if v.Protocol == "" {
		v.Protocol = "vmess"
	}
	t.Outbounds[0].Protocol = v.Protocol
	port, _ := strconv.Atoi(v.Port)
	aid, _ := strconv.Atoi(v.Aid)
	switch strings.ToLower(v.Protocol) {
	case "vmess":
		t.Outbounds[0].Settings.Vnext = []Vnext{
			{
				Address: v.Add,
				Port:    port,
				Users: []User{
					{
						ID:       v.ID,
						AlterID:  aid,
						Security: "auto",
					},
				},
			},
		}
		t.Outbounds[0].StreamSettings = &tmplJson.StreamSettings
		t.Outbounds[0].StreamSettings.Network = v.Net
		// 根据传输协议(network)修改streamSettings
		switch strings.ToLower(v.Net) {
		case "ws":
			tmplJson.WsSettings.Headers.Host = v.Host
			tmplJson.WsSettings.Path = v.Path
			t.Outbounds[0].StreamSettings.WsSettings = &tmplJson.WsSettings
		case "mkcp", "kcp":
			tmplJson.KcpSettings.Header.Type = v.Type
			t.Outbounds[0].StreamSettings.KcpSettings = &tmplJson.KcpSettings
		case "tcp":
			if strings.ToLower(v.Type) != "none" { //那就是http无疑了
				tmplJson.TCPSettings.Header.Request.Headers.Host = strings.Split(v.Host, ",")
				t.Outbounds[0].StreamSettings.TCPSettings = &tmplJson.TCPSettings
			}
		case "h2", "http":
			tmplJson.HttpSettings.Host = strings.Split(v.Host, ",")
			tmplJson.HttpSettings.Path = v.Path
			t.Outbounds[0].StreamSettings.HTTPSettings = &tmplJson.HttpSettings
		}
		if strings.ToLower(v.TLS) == "tls" {
			t.Outbounds[0].StreamSettings.Security = "tls"
			t.Outbounds[0].StreamSettings.TLSSettings = &tmplJson.TLSSettings
		}
		t.Outbounds[0].Mux = &tmplJson.Mux
	case "shadowsocks":
		t.Outbounds[0].Settings.Servers = []Server{
			{
				Address:  v.Add,
				Port:     port,
				Method:   v.Net,
				Password: v.ID,
				Ota:      false, //避免chacha20无法工作
			},
		}
	default:
		return errors.New("不支持的协议: " + v.Protocol)
	}
	return nil
}

func (t *Template) ToConfigBytes() []byte {
	b, _ := json.Marshal(t)
	return b
}
