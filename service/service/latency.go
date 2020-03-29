package service

import (
	"V2RayA/common/httpClient"
	"V2RayA/common/netTools/netstat"
	"V2RayA/core/shadowsocksr"
	"V2RayA/core/v2ray"
	"V2RayA/core/vmessInfo"
	"V2RayA/global"
	"V2RayA/persistence/configure"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func Ping(which []configure.Which, timeout time.Duration) ([]configure.Which, error) {
	var whiches configure.Whiches
	whiches.Set(which)
	//对要Ping的which去重
	which = whiches.GetNonDuplicated()
	//暂时关闭透明代理
	v2ray.CheckAndStopTransparentProxy()
	defer v2ray.CheckAndSetupTransparentProxy(true)
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

func TestHttpLatency(which []configure.Which, timeout time.Duration, maxParallel int) ([]configure.Which, error) {
	var whiches configure.Whiches
	whiches.Set(which)
	for i := len(which) - 1; i >= 0; i-- {
		if which[i].TYPE == configure.SubscriptionType { //去掉subscriptionType
			which = append(which[:i], which[i+1:]...)
		}
	}
	//对要Ping的which去重
	which = whiches.GetNonDuplicated()
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
	portMap := make(map[int]string)
	ssrPortMap := make(map[int]int)
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
		case "shadowsocksr", "shadowsocks", "ss", "ssr":
			//再找一个空端口
			port++
			for {
				if !isOccupiedTCPPort(nsmap, port) {
					break
				}
				port++
			}
			ssrLocalPortIfNeed = port
			ssrPortMap[i] = port
		default:
		}
		err := tmpl.AddMappingOutbound(v, v2rayInboundPort, false, ssrLocalPortIfNeed, "")
		if err != nil {
			which[i].Latency = err.Error()
			continue
		}
		portMap[i] = v2rayInboundPort
	}
	//启ssr
	//不清SSRs，防止断开当前连接
	if len(ssrPortMap) > 0 {
		for i, localPort := range ssrPortMap {
			v := vms[i]
			ssr := new(shadowsocksr.SSR)
			err := ssr.Serve(localPort, v.Net, v.ID, v.Add, v.Port, v.TLS, v.Path, v.Type, v.Host)
			if err != nil {
				return nil, err
			}
			global.SSRs.Append(*ssr)
		}
	}
	err = v2ray.WriteV2rayConfig(tmpl.ToConfigBytes())
	if err != nil {
		return nil, err
	}
	err = v2ray.RestartV2rayService()
	if err != nil {
		return nil, err
	}
	//线程并发限制
	//time.Sleep(200 * time.Millisecond)
	wg = new(sync.WaitGroup)
	cc := make(chan struct{}, maxParallel)
	for i := range which {
		if which[i].Latency != "" {
			continue
		}
		wg.Add(1)
		go func(i int) {
			cc <- struct{}{}
			defer func() { <-cc; wg.Done() }()
			httpLatency(&which[i], portMap[i], timeout)
		}(i)
	}
	wg.Wait()
	global.SSRs.ClearAll()
	if v2rayRunning && configure.GetConnectedServer() != nil {
		err = v2ray.UpdateV2RayConfig(nil)
		if err != nil {
			return which, errors.New("fail in restart v2ray-core, please connect a server")
		}
	} else {
		_ = v2ray.StopV2rayService() //没关掉那就不好意思了
	}
	return which, nil
}
func httpLatency(which *configure.Which, port string, timeout time.Duration) {
	c, err := httpClient.GetHttpClientWithProxy("socks5://127.0.0.1:" + port)
	if err != nil {
		which.Latency = err.Error()
		return
	}
	defer c.CloseIdleConnections()
	c.Timeout = timeout
	t := time.Now()
	req, _ := http.NewRequest("GET", "https://www.youtube.com", nil)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "close")
	resp, err := c.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if err != nil {
			es := strings.ToLower(err.Error())
			s, _ := which.LocateServer()
			log.Println(err, s.VmessInfo.Add+":"+s.VmessInfo.Port)
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
