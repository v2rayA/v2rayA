// protocol spec:
// https://github.com/esrrhs/pingtunnel

package pingtunnel

import (
	"github.com/mzz2017/go-engine/src/loggo"
	"github.com/mzz2017/go-engine/src/pingtunnel"
	"github.com/mzz2017/v2rayA/common/netTools/ports"
	"github.com/mzz2017/v2rayA/core/vmessInfo"
	"github.com/mzz2017/v2rayA/global"
	"github.com/mzz2017/v2rayA/plugin"
	"log"
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

func (tunnel *PingTunnel) Serve(localPort int, v vmessInfo.VmessInfo) (err error) {
	tunnel.localPort = localPort
	listen := ":" + strconv.Itoa(localPort)
	target := ""
	server := v.Add
	timeout := 60
	key, err := strconv.Atoi(v.ID)
	if err != nil {
		return newError("password must be a string of numbers")
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
		NoPrint:   global.Version != "debug",
	})
	c, err := pingtunnel.NewClient(listen, server, target, timeout, key,
		tcpmode, tcpmode_buffersize, tcpmode_maxwin, tcpmode_resend_timems, tcpmode_compress,
		tcpmode_stat, open_sock5, maxconn, nil)
	if err != nil {
		return newError().Base(err)
	}
	tunnel.client = c
	return c.Run()
}

func (tunnel *PingTunnel) Close() (err error) {
	if tunnel.client != nil {
		tunnel.client.Stop()
	}
	start := time.Now()
	for {
		var o bool
		o, _, err := ports.IsPortOccupied([]string{strconv.Itoa(tunnel.localPort) + ":tcp"})
		if err != nil {
			return err
		}
		if !o {
			break
		}
		if time.Since(start) > 3*time.Second {
			log.Println("PingTunnel.Close: timeout", tunnel.localPort)
			return newError("PingTunnel.Close: timeout")
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func (tunnel *PingTunnel) SupportUDP() bool {
	return false
}
