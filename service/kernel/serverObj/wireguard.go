package serverObj

import (
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/v2rayA/v2rayA/kernel/coreObj"
)

func init() {
	FromLinkRegister("wireguard", NewWireGuard)
	EmptyRegister("wireguard", func() (ServerObj, error) {
		return new(WireGuard), nil
	})
}

type WireGuard struct {
	Name        string `json:"name"`
	Server      string `json:"server"`
	Port        int    `json:"port"`
	PublicKey   string `json:"publicKey"`
	SecretKey   string `json:"secretKey"`
	Address     string `json:"address"`
	PrivateKey  string `json:"privateKey"`
	PreSharedKey string `json:"preSharedKey"`
	AllowedIPs  string `json:"allowedIPs"`
	KeepAlive   int    `json:"keepAlive"`
	Workers     int    `json:"workers"`
	Mtu         int    `json:"mtu"`
	Reserved    string `json:"reserved"`
	KernelMode  bool   `json:"kernelMode"`
	Protocol    string `json:"protocol"`
}

func NewWireGuard(link string) (ServerObj, error) {
	return ParseWireGuardURL(link)
}

func ParseWireGuardURL(link string) (data *WireGuard, err error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}
	q := u.Query()
	keepAlive, _ := strconv.Atoi(q.Get("keepAlive"))
	workers, _ := strconv.Atoi(q.Get("workers"))
	mtu, _ := strconv.Atoi(q.Get("mtu"))

	privateKey := ""
	if u.User != nil {
		privateKey = u.User.Username()
	}

	return &WireGuard{
		Name:         u.Fragment,
		Server:       u.Hostname(),
		Port:         port,
		PrivateKey:   privateKey,
		PublicKey:    q.Get("publicKey"),
		Address:      q.Get("address"),
		PreSharedKey: q.Get("preSharedKey"),
		AllowedIPs:   q.Get("allowedIPs"),
		KeepAlive:    keepAlive,
		Workers:      workers,
		Mtu:          mtu,
		Reserved:     q.Get("reserved"),
		KernelMode:   q.Get("kernelMode") == "true",
		Protocol:     "wireguard",
	}, nil
}

func (s *WireGuard) Configuration(info PriorInfo) (c Configuration, err error) {
	// Parse AllowedIPs
	allowedIPs := strings.Split(s.AllowedIPs, ",")
	if len(allowedIPs) == 0 || s.AllowedIPs == "" {
		allowedIPs = []string{"0.0.0.0/0", "::/0"}
	}

	// Parse Reserved
	var reserved []int
	if s.Reserved != "" {
		parts := strings.Split(s.Reserved, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				v, _ := strconv.Atoi(p)
				reserved = append(reserved, v)
			}
		}
	}

	// Build WireGuard settings
	settings := coreObj.WireGuardSettings{
		SecretKey: s.PrivateKey,
		Address:   []string{s.Address},
		Peers: []coreObj.WireGuardPeer{{
			Endpoint:     net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
			PublicKey:    s.PublicKey,
			PreSharedKey: s.PreSharedKey,
			AllowedIPs:   allowedIPs,
			KeepAlive:    s.KeepAlive,
		}},
	}

	if s.Workers > 0 {
		settings.Workers = s.Workers
	}
	if s.Mtu > 0 {
		settings.Mtu = s.Mtu
	}
	if len(reserved) > 0 {
		settings.Reserved = reserved
	}
	settings.KernelMode = s.KernelMode

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return c, err
	}

	return Configuration{
		CoreOutbound: coreObj.OutboundObject{
			Tag:      info.Tag,
			Protocol: "wireguard",
			Settings: coreObj.Settings{Inlined: settingsJSON},
		},
		UDPSupport: true,
	}, nil
}

func (s *WireGuard) ExportToURL() string {
	var query = make(url.Values)
	setValue(&query, "publicKey", s.PublicKey)
	setValue(&query, "address", s.Address)
	setValue(&query, "preSharedKey", s.PreSharedKey)
	setValue(&query, "allowedIPs", s.AllowedIPs)
	if s.KeepAlive > 0 {
		setValue(&query, "keepAlive", strconv.Itoa(s.KeepAlive))
	}
	if s.Workers > 0 {
		setValue(&query, "workers", strconv.Itoa(s.Workers))
	}
	if s.Mtu > 0 {
		setValue(&query, "mtu", strconv.Itoa(s.Mtu))
	}
	if s.Reserved != "" {
		setValue(&query, "reserved", s.Reserved)
	}
	if s.KernelMode {
		setValue(&query, "kernelMode", "true")
	}

	u := url.URL{
		Scheme:   "wireguard",
		User:     url.User(s.PrivateKey),
		Host:     net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		RawQuery: query.Encode(),
		Fragment: s.Name,
	}
	return u.String()
}

func (s *WireGuard) NeedPluginPort() bool { return false }
func (s *WireGuard) ProtoToShow() string  { return "WireGuard" }
func (s *WireGuard) GetProtocol() string  { return s.Protocol }
func (s *WireGuard) GetHostname() string  { return s.Server }
func (s *WireGuard) GetPort() int         { return s.Port }
func (s *WireGuard) GetName() string      { return s.Name }
func (s *WireGuard) SetName(name string)  { s.Name = name }
