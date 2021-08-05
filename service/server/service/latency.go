package service

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/common/netTools/netstat"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/plugin"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const HttpTestURL = "http://www.msftconnecttest.com/connecttest.txt"

func Ping(which []*configure.Which, timeout time.Duration) (_ []*configure.Which, err error) {
	var whiches = configure.NewWhiches(which)
	//对要Ping的which去重
	which = whiches.GetNonDuplicated()
	//暂时关闭透明代理
	v2ray.CheckAndStopTransparentProxy()
	defer func() {
		if e := v2ray.CheckAndSetupTransparentProxy(true); e != nil {
			err = newError(e).Base(err)
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

func TestHttpLatency(which []*configure.Which, timeout time.Duration, maxParallel int, showLog bool) ([]*configure.Which, error) {
	specialMode.StopDNSSupervisor()
	var whiches = configure.NewWhiches(which)
	for i := len(which) - 1; i >= 0; i-- {
		if which[i].TYPE == configure.SubscriptionType { //去掉subscriptionType
			which = append(which[:i], which[i+1:]...)
		}
	}
	which = whiches.Get()
	v2rayRunning := v2ray.IsV2RayRunning()
	wg := new(sync.WaitGroup)
	vms := make([]vmessInfo.VmessInfo, len(which))
	//init vmessInfos
	for i := range which {
		which[i].Latency = ""
		sr, err := which[i].LocateServerRaw()
		if err != nil {
			which[i].Latency = err.Error()
			continue
		}
		vms[i] = sr.VmessInfo
	}
	//modify the template based on current configuration
	var tmpl v2ray.Template
	if v2rayRunning {
		var err error
		tmpl, err = v2ray.NewTemplateFromConfig()
		if err != nil {
			return nil, err
		}
	} else {
		tmpl = v2ray.Template{}
	}
	inboundPortMap := make([]string, len(vms))
	pluginPortMap := make(map[int]int)
	port := 0
	nsmap, err := netstat.ToPortMap([]string{"tcp", "tcp6"})
	if err != nil {
		return nil, err
	}
	for i, v := range vms {
		if which[i].Latency != "" {
			continue
		}
		//find a port for the inbound
		if port == 0 {
			port = 14321 //starting port
		} else {
			port = port + 1
		}
		for {
			if !ports.IsOccupiedTCPPort(nsmap, port) {
				break
			}
			port++
		}
		v2rayInboundPort := strconv.Itoa(port)
		pluginPort := 0
		if plugin.HasProperPlugin(v) {
			// find a port for the plugin
			port++
			for {
				if !ports.IsOccupiedTCPPort(nsmap, port) {
					break
				}
				port++
			}
			pluginPort = port
			pluginPortMap[i] = port
		}
		err := tmpl.AddMappingOutbound(v, v2rayInboundPort, false, pluginPort, "socks")
		if err != nil {
			if strings.Contains(err.Error(), "unsupported") {
				which[i].Latency = "UNSUPPORTED PROTOCOL"
				continue
			}
			return nil, err
		}
		inboundPortMap[i] = v2rayInboundPort
	}
	//start plugins
	//do not clean plugins to prevent current connections disconnecting
	if len(pluginPortMap) > 0 {
		for i, localPort := range pluginPortMap {
			v := vms[i]
			var plu plugin.Plugin
			plu, err = plugin.NewPluginAndServe(localPort, v)
			if err != nil {
				return nil, err
			}
			plugin.GlobalPlugins.Add("outbound"+strconv.Itoa(localPort), plu)
		}
	}
	err = v2ray.WriteV2rayConfig(tmpl.ToConfigBytes())
	if err != nil {
		return nil, err
	}

	if err = tmpl.CheckInboundPortsOccupied(); err != nil {
		return nil, newError(err)
	}
	err = v2ray.RestartV2rayService(false)
	if err != nil {
		return nil, err
	}
	//limit the concurrency
	wg = new(sync.WaitGroup)
	cc := make(chan interface{}, maxParallel)
	for i := range which {
		if which[i].Latency != "" {
			if showLog {
				fmt.Printf("Error[%v]%v: %v\n", i+1, which[i].Latency, which[i].Link)
			}
			continue
		}
		wg.Add(1)
		go func(i int) {
			cc <- nil
			defer func() { <-cc; wg.Done() }()
			httpLatency(which[i], inboundPortMap[i], timeout)
			if showLog {
				fmt.Printf("Test done[%v]%v: %v\n", i+1, which[i].Latency, which[i].Link)
			}
		}(i)
	}
	wg.Wait()
	if v2rayRunning && configure.GetConnectedServers() != nil {
		err = v2ray.UpdateV2RayConfig()
		if err != nil {
			return which, newError("failed to restart v2ray-core, please connect a server")
		}
	} else {
		// no connected servers or v2ray was not running
		_ = v2ray.StopV2rayService(false)
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
		s, _ := which.LocateServerRaw()
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
			log.Println(err, s.VmessInfo.Add+":"+s.VmessInfo.Port)
		} else {
			which.Latency = "BAD RESPONSE"
			log.Println(resp.Status, s.VmessInfo.Add+":"+s.VmessInfo.Port)
		}
		return
	}
	_ = resp.Body.Close()
	which.Latency = fmt.Sprintf("%.0fms", time.Since(t).Seconds()*1000)
}
