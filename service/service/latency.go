package service

import (
	"V2RayA/model/v2ray"
	"V2RayA/model/vmessInfo"
	"V2RayA/persistence/configure"
	"V2RayA/tools"
	"errors"
	"fmt"
	"net"
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
	_ = CheckAndStopTransparentProxy()
	defer CheckAndSetupTransparentProxy(true)
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
	//将要测试的节点全部解析成ip
	v2rayRunning := v2ray.IsV2RayRunning()
	wg := new(sync.WaitGroup)
	vms := make([]vmessInfo.VmessInfo, len(which))
	for i := range which {
		which[i].Latency = ""
		sr, err := which[i].LocateServer()
		if err != nil {
			which[i].Latency = err.Error()
			continue
		}
		vms[i] = sr.VmessInfo
		if net.ParseIP(vms[i].Add) == nil {
			var hosts []string
			hosts, err = net.LookupHost(vms[i].Add)
			if err != nil || len(hosts) <= 0 {
				if err != nil {
					which[i].Latency = err.Error()
				} else {
					which[i].Latency = "dns解析失败: " + vms[i].Add
				}
				continue
			}
			vms[i].Add = hosts[0]
		}
	}
	//写v2ray配置
	var tmpl v2ray.Template
	if v2rayRunning {
		var err error
		tmpl, err = v2ray.NewTemplateFromConfig()
		if err != nil {
			return nil, err
		}
		//暂时关闭透明代理
		//_ = CheckAndStopTransparentProxy()
		//defer CheckAndSetupTransparentProxy(true)
	} else {
		tmpl = v2ray.NewTemplate()
	}
	portMap := make(map[int]string)
	port := 0
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
			if occupied, which := tools.IsPortOccupied(strconv.Itoa(port), "tcp"); occupied && !strings.Contains(which, "v2ray") {
				port++
			} else {
				break
			}
		}
		sPort := strconv.Itoa(port)
		portMap[i] = sPort
		err := tmpl.AddMappingOutbound(v, sPort, false)
		if err != nil {
			which[i].Latency = err.Error()
			continue
		}
	}
	err := v2ray.WriteV2rayConfig(tmpl.ToConfigBytes())
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
	if v2rayRunning && configure.GetConnectedServer() != nil {
		err = v2ray.UpdateV2rayWithConnectedServer()
		if err != nil {
			return which, errors.New("V2Ray重启失败，请手动连接一个节点")
		}
	} else {
		_ = v2ray.StopV2rayService() //没关掉那就不好意思了
	}
	return which, nil
}
func httpLatency(which *configure.Which, port string, timeout time.Duration) {
	c, err := tools.GetHttpClientWithProxy("socks5://127.0.0.1:" + port)
	if err != nil {
		which.Latency = err.Error()
		return
	}
	c.Timeout = timeout
	t := time.Now()
	req, _ := http.NewRequest("HEAD", "https://www.google.com", nil)
	resp, err := c.Do(req)
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
		return
	}
	_ = resp.Body.Close()
	which.Latency = fmt.Sprintf("%.0fms", time.Since(t).Seconds()*1000)
}
