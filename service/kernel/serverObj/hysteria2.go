package serverObj

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/v2rayA/v2rayA/kernel/coreObj"
)

func init() {
	FromLinkRegister("hysteria2", NewHysteria2)
	FromLinkRegister("hy2", NewHysteria2)
	EmptyRegister("hysteria2", func() (ServerObj, error) {
		return new(Hysteria2), nil
	})
	EmptyRegister("hy2", func() (ServerObj, error) {
		return new(Hysteria2), nil
	})
}

type Hysteria2 struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Link     string `json:"link"`
}

func NewHysteria2(link string) (ServerObj, error) {
	return ParseHysteria2URL(link)
}

func ParseHysteria2URL(link string) (data *Hysteria2, err error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	portStr := u.Port()
	if portStr == "" {
		portStr = "443"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}
	return &Hysteria2{
		Name:     u.Fragment,
		Server:   u.Hostname(),
		Port:     port,
		Protocol: "hysteria2",
		Link:     link,
	}, nil
}

// hysteria2ClientSettings is the outbound "settings" JSON of xray's native
// hysteria protocol (version 2, i.e. hysteria2).
type hysteria2ClientSettings struct {
	Version int32  `json:"version"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// parseLinkParams extracts the hysteria2 URL parameters, which are the source
// of the transport config (password, sni, certificate pinning, salamander obfs).
func (s *Hysteria2) parseLinkParams() (password, sni, pinnedPeerCertSha256, verifyPeerCertByName, obfs, obfsPassword string) {
	if s.Link == "" {
		return
	}
	u, err := url.Parse(s.Link)
	if err != nil {
		return
	}
	if u.User != nil {
		password = u.User.Username()
		if pw, ok := u.User.Password(); ok {
			password += ":" + pw
		}
	}
	q := u.Query()
	return password,
		q.Get("sni"),
		q.Get("pinned_peer_cert_sha256"),
		q.Get("verify_peer_cert_by_name"),
		q.Get("obfs"),
		q.Get("obfs-password")
}

func (s *Hysteria2) Configuration(info PriorInfo) (c Configuration, err error) {
	password, sni, pinnedPeerCertSha256, verifyPeerCertByName, obfs, obfsPassword := s.parseLinkParams()

	settingsJSON, err := json.Marshal(hysteria2ClientSettings{
		Version: 2,
		Address: s.Server,
		Port:    s.Port,
	})
	if err != nil {
		return c, fmt.Errorf("hysteria2: marshal settings: %w", err)
	}

	tlsSettings := &coreObj.TLSSettings{
		PinnedPeerCertSha256: pinnedPeerCertSha256,
		VerifyPeerCertByName: verifyPeerCertByName,
	}
	if sni == "" {
		tlsSettings.ServerName = s.Server
	} else {
		tlsSettings.ServerName = sni
	}

	streamSettings := &coreObj.StreamSettings{
		Network:     "hysteria",
		Security:    "tls",
		TLSSettings: tlsSettings,
		HysteriaSettings: &coreObj.HysteriaSettings{
			Version: 2,
			Auth:    password,
		},
	}
	if obfs == "salamander" {
		maskSettings, err := json.Marshal(map[string]string{"password": obfsPassword})
		if err != nil {
			return c, fmt.Errorf("hysteria2: marshal obfs settings: %w", err)
		}
		streamSettings.FinalMask = &coreObj.FinalMask{
			Udp: []coreObj.UdpMask{{
				Type:     "salamander",
				Settings: maskSettings,
			}},
		}
	}

	return Configuration{
		CoreOutbound: coreObj.OutboundObject{
			Tag:            info.Tag,
			Protocol:       "hysteria",
			Settings:       coreObj.Settings{Inlined: settingsJSON},
			StreamSettings: streamSettings,
		},
		UDPSupport: true,
	}, nil
}

func (s *Hysteria2) ExportToURL() string {
	return s.Link
}

func (s *Hysteria2) NeedPluginPort() bool {
	return false
}

func (s *Hysteria2) ProtoToShow() string {
	return fmt.Sprintf("hysteria2")
}

func (s *Hysteria2) GetProtocol() string {
	return s.Protocol
}

func (s *Hysteria2) GetHostname() string {
	return s.Server
}

func (s *Hysteria2) GetPort() int {
	return s.Port
}

func (s *Hysteria2) GetName() string {
	return s.Name
}

func (s *Hysteria2) SetName(name string) {
	s.Name = name
}
