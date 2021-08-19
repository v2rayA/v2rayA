package simpleobfs

import (
	"fmt"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/plugin"
	"github.com/v2rayA/v2rayA/plugin/infra/simpleobfs"
	"log"
	"net"
	"net/url"
)

type SimpleObfs struct {
	s *plugin.Server
}
type Params struct {
	Address  string
	Port     string
	Path     string
	ObfsType string
	Host     string
}

func init() {
	plugin.RegisterPlugin("simpleobfs", NewSimpleObfsPlugin)
}

func NewSimpleObfsPlugin(localPort int, v vmessInfo.VmessInfo) (plugin plugin.Plugin, err error) {
	plugin = new(SimpleObfs)
	err = plugin.Serve(localPort, v)
	return
}

func (so *SimpleObfs) LocalPort() int {
	return so.s.LocalPort
}

func ParseVmess(v vmessInfo.VmessInfo) (s string, err error) {
	params := Params{
		Address:  v.Add,
		Port:     v.Port,
		Path:     v.Path,
		ObfsType: v.Type,
		Host:     v.Host,
	}
	u, err := url.Parse(fmt.Sprintf(
		"simpleobfs://%v",
		net.JoinHostPort(params.Address, params.Port),
	))
	if err != nil {
		log.Println(err)
		return
	}
	q := u.Query()
	q.Set("type", params.ObfsType)
	q.Set("host", params.Host)
	q.Set("path", params.Path)
	u.RawQuery = q.Encode()
	s = u.String()
	return
}

func (so *SimpleObfs) Serve(localPort int, v vmessInfo.VmessInfo) (err error) {
	so.s = plugin.NewServer(localPort)
	s, err := ParseVmess(v)
	if err != nil {
		return
	}
	p, _ := simpleobfs.NewProxy(s)
	return so.s.Serve(p, "tcp->"+net.JoinHostPort(v.Add, v.Port))
}

func (so *SimpleObfs) Close() error {
	return so.s.Close()
}

func (so *SimpleObfs) SupportUDP() bool {
	return false
}
