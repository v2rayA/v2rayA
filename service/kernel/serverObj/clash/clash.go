// Package clash maps the proxies of a Clash/mihomo YAML config onto the
// share links v2rayA parses, so a Clash subscription imports like any other.
package clash

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"gopkg.in/yaml.v3"
)

// clashProxy is the subset of a Clash/mihomo proxy entry that maps onto the
// share links v2rayA already parses. Unknown keys are ignored.
type clashProxy struct {
	Name           string   `yaml:"name"`
	Type           string   `yaml:"type"`
	Server         string   `yaml:"server"`
	Port           int      `yaml:"port"`
	UUID           string   `yaml:"uuid"`
	Password       string   `yaml:"password"`
	Username       string   `yaml:"username"`
	Cipher         string   `yaml:"cipher"`
	AlterID        int      `yaml:"alterId"`
	TLS            bool     `yaml:"tls"`
	SNI            string   `yaml:"sni"`
	ServerName     string   `yaml:"servername"`
	SkipCertVerify bool     `yaml:"skip-cert-verify"`
	ALPN           []string `yaml:"alpn"`
	Fingerprint    string   `yaml:"client-fingerprint"`
	Network        string   `yaml:"network"`
	Flow           string   `yaml:"flow"`
	Plugin         string   `yaml:"plugin"`
	PluginOpts     struct {
		Mode string `yaml:"mode"`
		Host string `yaml:"host"`
		Path string `yaml:"path"`
		TLS  bool   `yaml:"tls"`
	} `yaml:"plugin-opts"`
	WSOpts struct {
		Path    string            `yaml:"path"`
		Headers map[string]string `yaml:"headers"`
	} `yaml:"ws-opts"`
	GRPCOpts struct {
		ServiceName string `yaml:"grpc-service-name"`
	} `yaml:"grpc-opts"`
	H2Opts struct {
		Host []string `yaml:"host"`
		Path string   `yaml:"path"`
	} `yaml:"h2-opts"`
	RealityOpts struct {
		PublicKey string `yaml:"public-key"`
		ShortID   string `yaml:"short-id"`
		SpiderX   string `yaml:"spider-x"`
	} `yaml:"reality-opts"`
	// ssr
	Protocol      string `yaml:"protocol"`
	ProtocolParam string `yaml:"protocol-param"`
	Obfs          string `yaml:"obfs"`
	ObfsParam     string `yaml:"obfs-param"`
	// hysteria2
	ObfsPassword string `yaml:"obfs-password"`
	// tuic
	CongestionController string `yaml:"congestion-controller"`
	UDPRelayMode         string `yaml:"udp-relay-mode"`
	DisableSNI           bool   `yaml:"disable-sni"`
}

type clashConfig struct {
	Proxies []clashProxy `yaml:"proxies"`
}

// Resolve parses a Clash/mihomo YAML config and returns its proxies. ok is
// false when raw is not such a config; err reports a config that is one but
// yields no usable node.
func Resolve(raw string) (infos []serverObj.ServerObj, ok bool, err error) {
	if !strings.Contains(raw, "proxies:") {
		return nil, false, nil
	}
	var cfg clashConfig
	if e := yaml.Unmarshal([]byte(raw), &cfg); e != nil || len(cfg.Proxies) == 0 {
		return nil, false, nil
	}
	unsupported := map[string]int{}
	for _, p := range cfg.Proxies {
		link, e := p.link()
		if e != nil {
			log.Warn("clash: %v: %v", p.Name, e)
			continue
		}
		if link == "" {
			unsupported[p.Type]++
			continue
		}
		obj, e := serverObj.NewFromLink(strings.SplitN(link, "://", 2)[0], link)
		if e != nil {
			log.Warn("clash: %v: %v", p.Name, e)
			continue
		}
		infos = append(infos, obj)
	}
	for t, n := range unsupported {
		log.Warn("clash: skipped %d %q proxies: type not supported", n, t)
	}
	if len(infos) == 0 {
		return nil, true, fmt.Errorf("no supported proxy in the Clash config (%d entries)", len(cfg.Proxies))
	}
	return infos, true, nil
}

func (p *clashProxy) hostPort() string {
	return net.JoinHostPort(p.Server, strconv.Itoa(p.Port))
}

func (p *clashProxy) sni() string {
	if p.SNI != "" {
		return p.SNI
	}
	return p.ServerName
}

func setIf(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

// transport fills the type/host/path/serviceName parameters shared by the
// vmess, vless and trojan links.
func (p *clashProxy) transport(q url.Values) {
	network := p.Network
	if network == "" {
		network = "tcp"
	}
	q.Set("type", network)
	switch network {
	case "ws":
		setIf(q, "path", p.WSOpts.Path)
		setIf(q, "host", p.WSOpts.Headers["Host"])
	case "grpc":
		setIf(q, "serviceName", p.GRPCOpts.ServiceName)
	case "h2":
		setIf(q, "path", p.H2Opts.Path)
		if len(p.H2Opts.Host) > 0 {
			q.Set("host", p.H2Opts.Host[0])
		}
	}
}

func buildLink(scheme string, user *url.Userinfo, host string, q url.Values, name string) string {
	u := url.URL{Scheme: scheme, User: user, Host: host, Fragment: name}
	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// link renders the proxy as the share link v2rayA parses for its type; it
// returns "" for a type without a mapping.
func (p *clashProxy) link() (string, error) {
	if p.Server == "" || p.Port == 0 {
		return "", fmt.Errorf("missing server or port")
	}
	q := url.Values{}
	switch p.Type {
	case "anytls":
		setIf(q, "sni", p.sni())
		if p.SkipCertVerify {
			q.Set("allow_insecure", "1")
		}
		return buildLink("anytls", url.User(p.Password), p.hostPort(), q, p.Name), nil
	case "trojan":
		setIf(q, "sni", p.sni())
		setIf(q, "alpn", strings.Join(p.ALPN, ","))
		if p.Network != "" && p.Network != "tcp" {
			p.transport(q)
		}
		return buildLink("trojan", url.User(p.Password), p.hostPort(), q, p.Name), nil
	case "vless":
		p.transport(q)
		setIf(q, "flow", p.Flow)
		setIf(q, "fp", p.Fingerprint)
		setIf(q, "alpn", strings.Join(p.ALPN, ","))
		if p.RealityOpts.PublicKey != "" {
			q.Set("security", "reality")
			q.Set("pbk", p.RealityOpts.PublicKey)
			setIf(q, "sid", p.RealityOpts.ShortID)
			setIf(q, "spx", p.RealityOpts.SpiderX)
		} else if p.TLS {
			q.Set("security", "tls")
		} else {
			q.Set("security", "none")
		}
		setIf(q, "sni", p.sni())
		return buildLink("vless", url.User(p.UUID), p.hostPort(), q, p.Name), nil
	case "vmess":
		// v2rayN style vmess://BASE64(json), which ParseVmessURL accepts
		network := p.Network
		if network == "" {
			network = "tcp"
		}
		info := map[string]interface{}{
			"v":           "2",
			"ps":          p.Name,
			"add":         p.Server,
			"port":        strconv.Itoa(p.Port),
			"id":          p.UUID,
			"aid":         strconv.Itoa(p.AlterID),
			"scy":         p.Cipher,
			"net":         network,
			"type":        "none",
			"host":        "",
			"path":        "",
			"tls":         "",
			"sni":         p.sni(),
			"alpn":        strings.Join(p.ALPN, ","),
			"fingerprint": p.Fingerprint,
		}
		if p.TLS {
			info["tls"] = "tls"
		}
		switch network {
		case "ws":
			info["path"] = p.WSOpts.Path
			info["host"] = p.WSOpts.Headers["Host"]
		case "grpc":
			info["path"] = p.GRPCOpts.ServiceName
		case "h2":
			info["path"] = p.H2Opts.Path
			if len(p.H2Opts.Host) > 0 {
				info["host"] = p.H2Opts.Host[0]
			}
		}
		b, err := jsoniter.Marshal(info)
		if err != nil {
			return "", err
		}
		return "vmess://" + base64.StdEncoding.EncodeToString(b), nil
	case "ss":
		if p.Plugin != "" {
			opts := []string{p.Plugin}
			switch p.Plugin {
			case "obfs":
				opts[0] = "simple-obfs"
				if p.PluginOpts.Mode != "" {
					opts = append(opts, "obfs="+p.PluginOpts.Mode)
				}
				if p.PluginOpts.Host != "" {
					opts = append(opts, "obfs-host="+p.PluginOpts.Host)
				}
			case "v2ray-plugin":
				if p.PluginOpts.TLS {
					opts = append(opts, "tls")
				}
				if p.PluginOpts.Mode != "" {
					opts = append(opts, "mode="+p.PluginOpts.Mode)
				}
				if p.PluginOpts.Host != "" {
					opts = append(opts, "host="+p.PluginOpts.Host)
				}
				if p.PluginOpts.Path != "" {
					opts = append(opts, "path="+p.PluginOpts.Path)
				}
			default:
				return "", fmt.Errorf("ss plugin %q not supported", p.Plugin)
			}
			q.Set("plugin", strings.Join(opts, ";"))
		}
		userinfo := base64.RawURLEncoding.EncodeToString([]byte(p.Cipher + ":" + p.Password))
		return buildLink("ss", url.User(userinfo), p.hostPort(), q, p.Name), nil
	case "ssr":
		b64 := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
		body := fmt.Sprintf("%s:%d:%s:%s:%s:%s/?obfsparam=%s&protoparam=%s&remarks=%s",
			p.Server, p.Port, p.Protocol, p.Cipher, p.Obfs, b64(p.Password),
			b64(p.ObfsParam), b64(p.ProtocolParam), b64(p.Name))
		return "ssr://" + b64(body), nil
	case "hysteria2":
		setIf(q, "sni", p.sni())
		setIf(q, "obfs", p.Obfs)
		setIf(q, "obfs-password", p.ObfsPassword)
		return buildLink("hysteria2", url.User(p.Password), p.hostPort(), q, p.Name), nil
	case "tuic":
		setIf(q, "sni", p.sni())
		setIf(q, "congestion_control", p.CongestionController)
		setIf(q, "udp_relay_mode", p.UDPRelayMode)
		setIf(q, "alpn", strings.Join(p.ALPN, ","))
		if p.DisableSNI {
			q.Set("disable_sni", "true")
		}
		if p.SkipCertVerify {
			q.Set("allow_insecure", "true")
		}
		return buildLink("tuic", url.UserPassword(p.UUID, p.Password), p.hostPort(), q, p.Name), nil
	case "socks5", "http":
		var user *url.Userinfo
		if p.Username != "" {
			user = url.UserPassword(p.Username, p.Password)
		}
		scheme := "socks5"
		if p.Type == "http" {
			scheme = "http-proxy"
			if p.TLS {
				scheme = "https-proxy"
			}
		}
		return buildLink(scheme, user, p.hostPort(), q, p.Name), nil
	}
	return "", nil
}
