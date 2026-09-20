package serverObj

import (
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
)

// ShadowsocksR is no longer supported: the core has no ShadowsocksR outbound
// and the plugin that once served one is gone. The type stays only so a
// node stored by an earlier release still decodes, keeps its place in the
// list (the group members point at nodes by ordinal) and can be deleted;
// importing an ssr:// link is refused in FromLink.
func init() {
	EmptyRegister("shadowsocksr", func() (ServerObj, error) {
		return new(ShadowsocksR), nil
	})
	EmptyRegister("ssr", func() (ServerObj, error) {
		return new(ShadowsocksR), nil
	})
}

const ssrUnsupported = "ShadowsocksR is not supported; use Shadowsocks, VMess, VLESS or Trojan instead"

type ShadowsocksR struct {
	Name       string `json:"name"`
	Server     string `json:"server"`
	Port       int    `json:"port"`
	Password   string `json:"password"`
	Cipher     string `json:"cipher"`
	Proto      string `json:"proto"`
	ProtoParam string `json:"protoParam"`
	Obfs       string `json:"obfs"`
	ObfsParam  string `json:"obfsParam"`
	Protocol   string `json:"protocol"`
}

func (s *ShadowsocksR) Configuration(info PriorInfo) (c Configuration, err error) {
	return Configuration{}, fmt.Errorf("unsupported: %s", ssrUnsupported)
}

// ExportToURL keeps the stored node shareable to a client that still speaks SSR.
func (s *ShadowsocksR) ExportToURL() string {
	return fmt.Sprintf("ssr://%v", base64.RawURLEncoding.EncodeToString([]byte(
		fmt.Sprintf(
			"%v:%v:%v:%v:%v/?remarks=%v&protoparam=%v&obfsparam=%v",
			net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
			s.Proto,
			s.Cipher,
			s.Obfs,
			base64.RawURLEncoding.EncodeToString([]byte(s.Password)),
			base64.RawURLEncoding.EncodeToString([]byte(s.Name)),
			base64.RawURLEncoding.EncodeToString([]byte(s.ProtoParam)),
			base64.RawURLEncoding.EncodeToString([]byte(s.ObfsParam)),
		),
	)))
}

func (s *ShadowsocksR) NeedPluginPort() bool { return false }

func (s *ShadowsocksR) ProtoToShow() string { return "SSR (unsupported)" }

func (s *ShadowsocksR) GetProtocol() string { return s.Protocol }
func (s *ShadowsocksR) GetHostname() string { return s.Server }
func (s *ShadowsocksR) GetPort() int        { return s.Port }
func (s *ShadowsocksR) GetName() string     { return s.Name }
func (s *ShadowsocksR) SetName(name string) { s.Name = name }
