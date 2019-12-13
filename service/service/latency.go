package service

import (
	"V2RayA/model/v2ray"
	"V2RayA/model/vmessInfo"
	"V2RayA/persistence/configure"
	"V2RayA/tools"
	"errors"
	"fmt"
	"net"
	"strconv"
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
	//暂时关闭透明代理
	_ = CheckAndStopTransparentProxy()
	defer CheckAndSetupTransparentProxy(true)
	//全部解析成ip
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
	tmpl := v2ray.NewTemplate()
	for i, v := range vms {
		if which[i].Latency != "" {
			continue
		}
		err := tmpl.AddMappingOutbound(v, strconv.Itoa(54322+i))
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
			httpLatency(&which[i], strconv.Itoa(54322+i), timeout)
		}(i)
	}
	wg.Wait()
	if configure.GetConnectedServer() != nil {
		err = v2ray.UpdateV2rayWithConnectedServer()
		if err != nil {
			return which, errors.New("V2Ray重启失败，请手动连接一个节点")
		}
	}
	return which, nil
}
func httpLatency(which *configure.Which, port string, timeout time.Duration) {
	c, err := tools.GetHttpClientWithProxy("socks5://localhost:" + port)
	if err != nil {
		which.Latency = err.Error()
		return
	}
	c.Timeout = timeout
	t := time.Now()
	resp, err := tools.HttpGetUsingSpecificClient(c, "https://google.com")
	if err != nil {
		which.Latency = err.Error()
		return
	}
	_ = resp.Body.Close()
	which.Latency = fmt.Sprintf("%.0fms", time.Since(t).Seconds()*1000)
}
