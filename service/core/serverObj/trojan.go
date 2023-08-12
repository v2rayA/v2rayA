package serverObj

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/coreObj"
)

func init() {
	FromLinkRegister("trojan", NewTrojan)
	FromLinkRegister("trojan-go", NewTrojan)
	EmptyRegister("trojan", func() (ServerObj, error) {
		return new(Trojan), nil
	})
	EmptyRegister("trojan-go", func() (ServerObj, error) {
		return new(Trojan), nil
	})
}

type Trojan struct {
	Name          string `json:"name"`
	Server        string `json:"server"`
	Port          int    `json:"port"`
	Password      string `json:"password"`
	Sni           string `json:"sni"`
	Type          string `json:"type"`
	Encryption    string `json:"encryption"`
	Host          string `json:"host"`
	Path          string `json:"path"`
	AllowInsecure bool   `json:"allowInsecure"`
	Protocol      string `json:"protocol"`
}

func NewTrojan(link string) (ServerObj, error) {
	return ParseTrojanURL(link)
}

func ParseTrojanURL(u string) (data *Trojan, err error) {
	//trojan://password@server:port#escape(remarks)
	t, err := url.Parse(u)
	if err != nil {
		err = fmt.Errorf("invalid trojan format")
		return
	}
	allowInsecure := t.Query().Get("allowInsecure")
	sni := t.Query().Get("peer")
	if sni == "" {
		sni = t.Query().Get("sni")
	}
	if sni == "" {
		sni = t.Hostname()
	}
	port, err := strconv.Atoi(t.Port())
	if err != nil {
		return nil, ErrInvalidParameter
	}
	data = &Trojan{
		Name:          t.Fragment,
		Server:        t.Hostname(),
		Port:          port,
		Password:      t.User.Username(),
		Sni:           sni,
		AllowInsecure: allowInsecure == "1" || allowInsecure == "true",
		Protocol:      "trojan",
	}
	if t.Scheme == "trojan-go" {
		data.Protocol = "trojan-go"
		data.Encryption = t.Query().Get("encryption")
		data.Host = t.Query().Get("host")
		data.Path = t.Query().Get("path")
		data.Type = t.Query().Get("type")
		data.AllowInsecure = false
	}
	return data, nil
}

func (t *Trojan) Configuration(info PriorInfo) (c Configuration, err error) {
	if t.Protocol == "trojan-go" {
		// "trojanc -> tcp:// -> ss,ws,tls"
		tcpListener := url.URL{
			Scheme: "tcp",
			Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(info.PluginPort)),
		}
		tls := url.URL{
			Scheme: "tls",
			Host:   net.JoinHostPort(t.Server, strconv.Itoa(t.Port)),
			RawQuery: url.Values{
				"sni":           []string{t.Sni},
				"allowInsecure": []string{common.BoolToString(t.AllowInsecure)},
			}.Encode(),
		}
		chain := []string{tcpListener.String(), tls.String()}
		if t.Type == "ws" {
			ws := url.URL{
				Scheme: "ws",
				Host:   net.JoinHostPort(t.Server, strconv.Itoa(t.Port)),
				RawQuery: url.Values{
					"host": []string{t.Host},
					"path": []string{t.Path},
				}.Encode(),
			}
			chain = append(chain, ws.String())
		}
		if strings.HasPrefix(t.Encryption, "ss;") {
			fields := strings.SplitN(t.Encryption, ";", 3)
			ss := url.URL{
				Scheme: "ss",
				Host:   net.JoinHostPort(t.Server, strconv.Itoa(t.Port)),
				User:   url.UserPassword(fields[1], fields[2]),
			}
			chain = append(chain, ss.String())
		}
		return Configuration{
			CoreOutbound: coreObj.OutboundObject{
				Tag:      info.Tag,
				Protocol: "trojan",
				Settings: coreObj.Settings{
					Servers: []coreObj.Server{{
						Address:  "127.0.0.1",
						Port:     info.PluginPort,
						Password: t.Password,
					}},
				},
			},
			PluginChain: strings.Join(chain, ","),
			UDPSupport:  true,
		}, nil
	}
	core := coreObj.OutboundObject{
		Tag:      info.Tag,
		Protocol: "trojan",
		Settings: coreObj.Settings{
			Servers: []coreObj.Server{{
				Address:  t.Server,
				Port:     t.Port,
				Password: t.Password,
			}},
		},
		StreamSettings: &coreObj.StreamSettings{
			Network:  "tcp",
			Security: "tls",
			TLSSettings: &coreObj.TLSSettings{
				ServerName:    t.Sni,
				AllowInsecure: t.AllowInsecure,
			},
		},
	}
	return Configuration{
		CoreOutbound: core,
		PluginChain:  "",
		UDPSupport:   true,
	}, nil
}

func (t *Trojan) ExportToURL() string {
	u := &url.URL{
		Scheme:   "trojan",
		User:     url.User(t.Password),
		Host:     net.JoinHostPort(t.Server, strconv.Itoa(t.Port)),
		Fragment: t.Name,
	}
	q := u.Query()
	if t.AllowInsecure {
		q.Set("allowInsecure", "1")
	}
	setValue(&q, "sni", t.Sni)

	if t.Protocol == "trojan-go" {
		u.Scheme = "trojan-go"
		setValue(&q, "host", t.Host)
		setValue(&q, "encryption", t.Encryption)
		setValue(&q, "type", t.Type)
		setValue(&q, "path", t.Path)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (t *Trojan) NeedPluginPort() bool {
	return t.Protocol == "trojan-go"
}

func (t *Trojan) ProtoToShow() string {
	if t.Protocol == "trojan" {
		return t.Protocol
	}
	if t.Encryption == "" {
		return fmt.Sprintf("%v(%v)", t.Protocol, t.Type)
	}
	return fmt.Sprintf("%v(%v+%v)", t.Protocol, t.Type, strings.Split(t.Encryption, ";")[0])
}

func (t *Trojan) GetProtocol() string {
	return t.Protocol
}

func (t *Trojan) GetHostname() string {
	return t.Server
}

func (t *Trojan) GetPort() int {
	return t.Port
}

func (t *Trojan) GetName() string {
	return t.Name
}

func (t *Trojan) SetName(name string) {
	t.Name = name
}
