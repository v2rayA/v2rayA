// protocol spec:
// https://github.com/esrrhs/pingtunnel

package pingtunnel

import (
	"fmt"
	"github.com/mzz2017/go-engine/src/loggo"
	"github.com/mzz2017/go-engine/src/pingtunnel"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"net"
	"net/url"
	"strconv"
	"time"
)

type PingTunnel struct {
	client        *pingtunnel.Client
	key           int
	remote        string
	listenAddress string
	localPort     int
}

func init() {
	plugin.RegisterServer("ping-tunnel", NewPingTunnelDialer)
}

func NewPingTunnel(s string, p plugin.Proxy) (*PingTunnel, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	key, err := strconv.Atoi(u.Query().Get("password"))
	if err != nil {
		return nil, fmt.Errorf("password must be a string of numbers")
	}
	return &PingTunnel{
		key:           key,
		listenAddress: u.Host,
		remote:        u.Query().Get("server"),
		localPort:     0,
	}, nil
}

func NewPingTunnelDialer(s string, p plugin.Proxy) (plugin.Server, error) {
	return NewPingTunnel(s, p)
}

func (tunnel *PingTunnel) LocalPort() int {
	return tunnel.localPort
}

func (tunnel *PingTunnel) Close() (err error) {
	if tunnel.client != nil {
		tunnel.client.Stop()
	}
	start := time.Now()
	for {
		conn, e := net.Dial("tcp", tunnel.listenAddress)
		if e == nil {
			conn.Close()
		}
		if time.Since(start) > 3*time.Second {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// ListenAndServe serves pingtunnel requests.
func (tunnel *PingTunnel) ListenAndServe() error {
	//go s.ListenAndServeUDP()
	return tunnel.ListenAndServeTCP()
}

func (tunnel *PingTunnel) ListenAddr() string {
	return tunnel.listenAddress
}

// ListenAndServeTCP listen and serve on tcp port.
func (tunnel *PingTunnel) ListenAndServeTCP() error {
	listen := tunnel.listenAddress
	target := ""
	server := tunnel.remote
	timeout := 60
	key := tunnel.key
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
	return tunnel.client.Run()
}
