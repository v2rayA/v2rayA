package serverObj

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func init() {
	FromLinkRegister("ping-tunnel", NewPingTunnel)
	EmptyRegister("ping-tunnel", func() (ServerObj, error) {
		return new(PingTunnel), nil
	})
	FromLinkRegister("pingtunnel", NewPingTunnel)
	EmptyRegister("pingtunnel", func() (ServerObj, error) {
		return new(PingTunnel), nil
	})
}

type PingTunnel struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	Password string `json:"password"`
	Protocol string `json:"protocol"`
}

func NewPingTunnel(link string) (ServerObj, error) {
	if strings.HasPrefix(link, "pingtunnel://") {
		return ParsePingTunnelURL1(link)
	} else if strings.HasPrefix(link, "ping-tunnel://") {
		return ParsePingTunnelURL2(link)
	}
	return nil, InvalidParameterErr
}

func ParsePingTunnelURL1(u string) (data *PingTunnel, err error) {
	u = strings.TrimPrefix(u, "pingtunnel://")
	u, err = common.Base64StdDecode(u)
	if err != nil {
		log.Warn("ParsePingTunnelURL1: %v", u)
		err = fmt.Errorf("ParsePingTunnelURL1: %w", err)
		return
	}
	arr := strings.Split(u, "#")
	var ps string
	if len(arr) == 2 {
		ps, _ = url.QueryUnescape(arr[1])
	}
	u = arr[0]
	re := regexp.MustCompile(`(.+):(.+)`)
	subMatch := re.FindStringSubmatch(u)
	if len(subMatch) < 3 {
		return nil, fmt.Errorf("wrong format of pingtunnel")
	}
	passwd, err := common.Base64URLDecode(subMatch[2])
	if err != nil {
		log.Warn("ParsePingTunnelURL1: %v", subMatch[2])
		err = fmt.Errorf("ParsePingTunnelURL1: %w", err)
		return
	}
	data = &PingTunnel{
		Name:     ps,
		Server:   subMatch[1],
		Password: passwd,
		Protocol: "pingtunnel",
	}
	return data, nil
}

func ParsePingTunnelURL2(u string) (data *PingTunnel, err error) {
	U, err := url.Parse(u)
	if err != nil {
		return
	}
	data = &PingTunnel{
		Name:     U.Fragment,
		Server:   U.Host,
		Password: U.User.String(),
		Protocol: "pingtunnel",
	}
	return data, nil
}

func (p *PingTunnel) Configuration(info PriorInfo) (c Configuration, err error) {
	U := url.URL{
		Scheme: "ping-tunnel",
		Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(info.PluginPort)),
		RawQuery: url.Values{
			"password": []string{p.Password},
			"server":   []string{p.Server},
		}.Encode(),
	}
	return Configuration{
		CoreOutbound: info.PluginObj(),
		PluginChain:  U.String(),
		UDPSupport:   false,
	}, nil
}

func (p *PingTunnel) ExportToURL() string {
	U := url.URL{
		Scheme:   "ping-tunnel",
		User:     url.User(p.Password),
		Host:     p.Server,
		Fragment: p.Name,
	}
	return U.String()
}

func (p *PingTunnel) NeedPluginPort() bool {
	return true
}

func (p *PingTunnel) ProtoToShow() string {
	return p.Protocol
}

func (p *PingTunnel) GetProtocol() string {
	return p.Protocol
}

func (p *PingTunnel) GetHostname() string {
	return p.Server
}

func (p *PingTunnel) GetPort() int {
	// PingTunnel Need No Port
	return 0
}

func (p *PingTunnel) GetName() string {
	return p.Name
}

func (p *PingTunnel) SetName(name string) {
	p.Name = name
}
