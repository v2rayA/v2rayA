// protocol spec:
// https://github.com/esrrhs/pingtunnel

package pingtunnel

import (
	"fmt"
	"github.com/mzz2017/go-engine/src/loggo"
	"github.com/mzz2017/go-engine/src/pingtunnel"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"net"
	"strconv"
	"time"
)

type PingTunnel struct {
	client    *pingtunnel.Client
	localPort int
}

func init() {
	plugin.RegisterPlugin("pingtunnel", NewPingTunnelPlugin)
}

func NewPingTunnelPlugin(localPort int, v vmessInfo.VmessInfo) (plugin plugin.Plugin, err error) {
	plugin = new(PingTunnel)
	err = plugin.Serve(localPort, v)
	return
}

func (tunnel *PingTunnel) LocalPort() int {
	return tunnel.localPort
}

func (tunnel *PingTunnel) Serve(localPort int, v vmessInfo.VmessInfo) (err error) {
	tunnel.localPort = localPort
	listen := ":" + strconv.Itoa(localPort)
	target := ""
	server := v.Add
	timeout := 60
	key, err := strconv.Atoi(v.ID)
	if err != nil {
		return fmt.Errorf("password must be a string of numbers")
	}
	tcpmode := 1
	tcpmode_buffersize := 1 * 1024 * 1024 //1MB
	tcpmode_maxwin := 10000
	tcpmode_resend_timems := 400
	tcpmode_compress := 0
	tcpmode_stat := 0
	open_sock5 := 1
	maxconn := 0
	loggo.Ini(loggo.Config{
		Level:     loggo.NameToLevel("ERROR"),
		Prefix:    "pingtunnel",
		MaxDay:    3,
		NoLogFile: true,
		NoPrint:   log.ParseLevel(conf.GetEnvironmentConfig().LogLevel) < log.ParseLevel("debug"),
	})
	c, err := pingtunnel.NewClient(listen, server, target, timeout, key,
		tcpmode, tcpmode_buffersize, tcpmode_maxwin, tcpmode_resend_timems, tcpmode_compress,
		tcpmode_stat, open_sock5, maxconn, nil)
	if err != nil {
		return fmt.Errorf("[PingTunnel] Serve: %w", err)
	}
	tunnel.client = c
	return c.Run()
}

func (tunnel *PingTunnel) Close() (err error) {
	if tunnel.client != nil {
		tunnel.client.Stop()
	}
	start := time.Now()
	port := strconv.Itoa(tunnel.localPort)
	for {
		var o bool
		o, _, err := ports.IsPortOccupied([]string{port + ":tcp"})
		if err != nil {
			return err
		}
		if !o {
			break
		}
		conn, e := net.Dial("tcp", ":"+port)
		if e == nil {
			conn.Close()
		}
		if time.Since(start) > 3*time.Second {
			log.Warn("PingTunnel.Close: timeout: %v", tunnel.localPort)
			return fmt.Errorf("[PingTunnel] Close: timeout")
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func (tunnel *PingTunnel) SupportUDP() bool {
	return false
}
