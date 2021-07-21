package service

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/common/netTools/netstat"
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

func Ping(which []*configure.Which, timeout time.Duration) (_ []*configure.Which, err error) {
	var whiches configure.Whiches
	whiches.Set(which)
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

func isOccupiedTCPPort(nsmap map[string]map[int][]*netstat.Socket, port int) bool {
	v := nsmap["tcp"][port]
	v6 := nsmap["tcp6"][port]
	v = append(v, v6...)
	for _, v := range v {
		if v.State != netstat.Close {
			return true
		}
	}
	return false
}

func TestHttpLatency(which []*configure.Which, timeout time.Duration, maxParallel int, showLog bool) ([]*configure.Which, error) {
	specialMode.StopDNSSupervisor()
	var whiches configure.Whiches
	whiches.Set(which)
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
		sr, err := which[i].LocateServer()
		if err != nil {
			which[i].Latency = err.Error()
			continue
		}
		vms[i] = sr.VmessInfo
	}
	//写v2ray配置
	var tmpl v2ray.Template
	if v2rayRunning {
		var err error
		tmpl, err = v2ray.NewTemplateFromConfig()
		if err != nil {
			return nil, err
		}
	} else {
		tmpl = v2ray.NewTemplate()
	}
	portMap := make([]string, len(vms))
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
		//找到一个未被占用的高端口
		if port == 0 {
			port = 14321 //起始端口
		} else {
			port = port + 1
		}
		for {
			if !isOccupiedTCPPort(nsmap, port) {
				break
			}
			port++
		}
		v2rayInboundPort := strconv.Itoa(port)
		ssrLocalPortIfNeed := 0
		switch strings.ToLower(v.Protocol) {
		case "vmess", "vless", "trojan", "":
			//pass
		case "shadowsocks", "ss":
			var donotneedport bool
			if v.Type == "" {
				switch v.Net {
				case "aes-128-cfb", "aes-192-cfb", "aes-256-cfb", "aes-128-ctr", "aes-192-ctr", "aes-256-ctr", "aes-128-ofb", "aes-192-ofb", "aes-256-ofb", "des-cfb", "bf-cfb", "cast5-cfb", "rc4-md5", "chacha20", "chacha20-ietf", "salsa20", "camellia-128-cfb", "camellia-192-cfb", "camellia-256-cfb", "idea-cfb", "rc2-cfb", "seed-cfb":
					//ssr插件接simpleobfs插件
				default:
					donotneedport = true
				}
			}
			if donotneedport {
				break
			}
			//有可能是simpleobfs
			fallthrough
		default:
			if !plugin.IsProtocolValid(v) {
				which[i].Latency = "UNSUPPORTED PROTOCOL"
				continue
			}
			//再找一个空端口
			port++
			for {
				if !isOccupiedTCPPort(nsmap, port) {
					break
				}
				port++
			}
			ssrLocalPortIfNeed = port
			pluginPortMap[i] = port
		}
		err := tmpl.AddMappingOutbound(v, v2rayInboundPort, false, ssrLocalPortIfNeed, "")
		if err != nil {
			if strings.Contains(err.Error(), "unsupported") {
				which[i].Latency = "UNSUPPORTED PROTOCOL"
				continue
			}
			return nil, err
		}
		portMap[i] = v2rayInboundPort
	}
	//启plugin
	//不清plugins，防止断开当前连接
	if len(pluginPortMap) > 0 {
		for i, localPort := range pluginPortMap {
			v := vms[i]
			var plu plugin.Plugin
			plu, err = plugin.NewPlugin(localPort, v)
			if err != nil {
				return nil, err
			}
			plugin.GlobalPlugins.Append(plu)
		}
	}
	err = v2ray.WriteV2rayConfig(tmpl.ToConfigBytes())
	if err != nil {
		return nil, err
	}

	if occupied, port, pname := tmpl.CheckInboundPortsOccupied(); occupied {
		return nil, newError("Port ", port, " is occupied by ", pname)
	}
	err = v2ray.RestartV2rayService()
	if err != nil {
		return nil, err
	}
	//time.Sleep(200 * time.Millisecond)
	//线程并发限制
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
			httpLatency(which[i], portMap[i], timeout)
			if showLog {
				fmt.Printf("Test done[%v]%v: %v\n", i+1, which[i].Latency, which[i].Link)
			}
		}(i)
	}
	wg.Wait()
	if v2rayRunning && configure.GetConnectedServer() != nil {
		err = v2ray.UpdateV2RayConfig(nil)
		if err != nil {
			return which, newError("failed to restart v2ray-core, please connect a server")
		}
	} else {
		_ = v2ray.StopV2rayService() //没关掉那就不好意思了
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
	req, _ := http.NewRequest("GET", "http://www.msftconnecttest.com/connecttest.txt", nil)
	//req, _ := http.NewRequest("GET", "http://www.gstatic.com/generate_204", nil)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", "curl/7.70.0")
	resp, err := c.Do(req)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 400 {
		s, _ := which.LocateServer()
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
