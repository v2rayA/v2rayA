package service

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/core/coreObj"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const HttpTestURL = "https://gstatic.com/generate_204"

func Ping(which []*configure.Which, timeout time.Duration) (_ []*configure.Which, err error) {
	var whiches = configure.NewWhiches(which)
	//对要Ping的which去重
	which = whiches.GetNonDuplicated()
	//暂时关闭透明代理
	v2ray.ProcessManager.CheckAndStopTransparentProxy(nil)
	defer func() {
		if e := v2ray.ProcessManager.CheckAndSetupTransparentProxy(true, nil); e != nil {
			err = fmt.Errorf("Ping: %v: %v", e, err)
		}
	}()
	//多线程异步ping
	wg := new(sync.WaitGroup)
	for i, v := range which {
		if v.TYPE == configure.SubscriptionType { //subscription不能ping
			continue
		}
		wg.Add(1)
		go func(i int) {
			_ = which[i].Ping(timeout)
			wg.Done()
		}(i)
	}
	wg.Wait()
	for i := len(which) - 1; i >= 0; i-- {
		if which[i].TYPE == configure.SubscriptionType { //不返回subscriptionType
			which = append(which[:i], which[i+1:]...)
		}
	}
	return which, nil
}

func addHosts(tmpl *v2ray.Template, vms []serverObj.ServerObj) {
	if tmpl.DNS == nil {
		tmpl.DNS = new(coreObj.DNS)
	}
	if tmpl.DNS.Hosts == nil {
		tmpl.DNS.Hosts = make(coreObj.Hosts)
	}
	const concurrency = 5
	var mu sync.Mutex
	var limit = make(chan struct{}, concurrency)
	var wg = sync.WaitGroup{}
	for _, v := range vms {
		if net.ParseIP(v.GetHostname()) == nil {
			wg.Add(1)
			go func(addr string) {
				limit <- struct{}{}
				defer func() {
					wg.Done()
					<-limit
				}()
				ips, err := resolv.LookupHost(addr)
				if err != nil {
					return
				}
				if len(ips) > 0 {
					ips = v2ray.FilterIPs(ips)
					mu.Lock()
					tmpl.DNS.Hosts[addr] = ips
					mu.Unlock()
				}
			}(v.GetHostname())
		}
	}
	wg.Wait()
}

func TestHttpLatency(which []*configure.Which, timeout time.Duration, maxParallel int, showLog bool) ([]*configure.Which, error) {
	var whiches = configure.NewWhiches(which)
	for i := len(which) - 1; i >= 0; i-- {
		if which[i].TYPE == configure.SubscriptionType { //去掉subscriptionType
			which = append(which[:i], which[i+1:]...)
		}
	}
	which = whiches.Get()
	v2rayRunning := v2ray.ProcessManager.Running()
	wg := new(sync.WaitGroup)
	vms := make([]serverObj.ServerObj, len(which))
	//init vmessInfos
	for i := range which {
		which[i].Latency = ""
		sr, err := which[i].LocateServerRaw()
		if err != nil {
			which[i].Latency = err.Error()
			continue
		}
		vms[i] = sr.ServerObj
	}
	//modify the template based on current configuration
	var (
		tmpl *v2ray.Template
		err  error
	)
	if v2rayRunning {
		tmpl, err = v2ray.NewTemplateFromConnectedServers(nil)
		if err != nil {
			if !errors.Is(err, v2ray.NoConnectedServerErr) {
				log.Warn("NewTemplateFromConnectedServers: %v", err)
			}
		}
	}
	if tmpl == nil {
		tmpl = v2ray.NewEmptyTemplate(&configure.Setting{
			RulePortMode:  configure.WhitelistMode,
			TcpFastOpen:   configure.Default,
			MuxOn:         configure.No,
			Transparent:   configure.TransparentClose,
			SpecialMode:   configure.SpecialModeNone,
			AntiPollution: configure.AntipollutionClosed,
		})
		tmpl.SetAPI(nil)
	}
	inboundPortMap := make([]string, len(vms))
	pluginPortMap := make(map[int]int)
	var toClose []io.Closer
	defer func() {
		for _, l := range toClose {
			_ = l.Close()
		}
	}()
	for i, v := range vms {
		if which[i].Latency != "" {
			continue
		}
		//find a port for the inbound
		t := time.Now()
		var port int
		for {
			l, err := net.Listen("tcp", "0.0.0.0:0")
			if err == nil {
				port = l.Addr().(*net.TCPAddr).Port
				toClose = append(toClose, l)
				l2, err2 := net.ListenPacket("udp", "0.0.0.0:"+strconv.Itoa(port))
				if err2 == nil {
					toClose = append(toClose, l2)
					break
				}
			}
			if time.Since(t) > 3*time.Second {
				return nil, fmt.Errorf("timeout: failed to find availble ports")
			}
		}
		v2rayInboundPort := strconv.Itoa(port)
		pluginPort := 0
		if v.NeedPluginPort() {
			// find a port for the plugin
			for {
				l, err := net.Listen("tcp", "127.0.0.1:0")
				if err == nil {
					toClose = append(toClose, l)
					port = l.Addr().(*net.TCPAddr).Port
					l2, err2 := net.ListenPacket("udp", "127.0.0.1:"+strconv.Itoa(port))
					if err2 == nil {
						toClose = append(toClose, l2)
						break
					}
				}
				if time.Since(t) > 3*time.Second {
					return nil, fmt.Errorf("timeout: failed to find availble ports")
				}
			}
			pluginPort = port
			pluginPortMap[i] = port
		}
		err := tmpl.InsertMappingOutbound(v, v2rayInboundPort, false, pluginPort, "socks")
		if err != nil {
			if strings.Contains(err.Error(), "unsupported") {
				which[i].Latency = "UNSUPPORTED PROTOCOL"
				continue
			}
			return nil, err
		}
		inboundPortMap[i] = v2rayInboundPort
	}
	for _, l := range toClose {
		_ = l.Close()
	}
	toClose = nil
	time.Sleep(30 * time.Millisecond)
	tmpl.Routing.DomainStrategy = "AsIs"
	addHosts(tmpl, vms)
	tmpl.SetOutboundSockopt()
	if err := v2ray.ProcessManager.Start(tmpl); err != nil {
		return nil, err
	}
	//limit the concurrency
	wg = new(sync.WaitGroup)
	cc := make(chan interface{}, maxParallel)
	for i := range which {
		if which[i].Latency != "" {
			if showLog {
				log.Warn("Error[%v]%v: %v", i+1, which[i].Latency, which[i].Link)
			}
			continue
		}
		wg.Add(1)
		go func(i int) {
			cc <- nil
			defer func() { <-cc; wg.Done() }()
			httpLatency(which[i], inboundPortMap[i], timeout)
			if showLog {
				log.Info("Test done[%v]%v: %v", i+1, which[i].Latency, which[i].Link)
			}
		}(i)
	}
	wg.Wait()
	if v2rayRunning && configure.GetConnectedServers() != nil {
		err := v2ray.UpdateV2RayConfig()
		if err != nil {
			return which, fmt.Errorf("cannot restart v2ray-core: %w", err)
		}
	} else {
		v2ray.ProcessManager.Stop(true)
	}
	if err := configure.NewWhiches(which).SaveLatencies(); err != nil {
		return nil, fmt.Errorf("failed to save the latency test result: %v", err)
	}
	return which, nil
}
func httpLatency(which *configure.Which, port string, timeout time.Duration) {
	c, err := httpClient.GetHttpClientWithProxy("socks5://127.0.0.1:" + port)
	if err != nil {
		which.Latency = "SYSTEM ERROR"
		return
	}
	defer c.CloseIdleConnections()
	c.Timeout = timeout
	t := time.Now()
	// NOT follow redirects
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	req, _ := http.NewRequest("GET", HttpTestURL, nil)
	//req, _ := http.NewRequest("GET", "http://www.gstatic.com/generate_204", nil)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", "curl/7.70.0")
	resp, err := c.Do(req)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 400 {
		if err != nil {
			es := strings.ToLower(err.Error())
			switch {
			case strings.Contains(es, "eof"):
				which.Latency = "NOT STABLE"
			case strings.Contains(es, "does not look like a tls handshake"):
				which.Latency = "INVALID"
			case strings.Contains(es, "timeout"):
				which.Latency = "TIMEOUT"
			default:
				which.Latency = err.Error()
			}
		} else {
			which.Latency = "BAD RESPONSE"
		}
		return
	}
	_ = resp.Body.Close()
	which.Latency = fmt.Sprintf("%.0fms", time.Since(t).Seconds()*1000)
}
