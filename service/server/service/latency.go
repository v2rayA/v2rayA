package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const HttpTestURL = "https://gstatic.com/generate_204"

func Ping(which []*configure.Which, timeout time.Duration) (_ []*configure.Which, err error) {
	var whiches = configure.NewWhiches(which)
	// Deduplicate whiches to Ping
	which = whiches.GetNonDuplicated()
	// Do not remove interception for a dashboard probe: unrelated traffic
	// could otherwise escape directly during this window. Only the probe's
	// TCP and DNS sockets bypass interception: marked on Linux, bound to
	// the physical egress interface for macOS and Windows TUN.
	dialer := httpClient.DirectDialer(timeout, v2ray.IsTransparentOn(configure.GetSettingNotNil()))
	lookup, _, err := nodeLookup(dialer)
	if err != nil {
		return nil, err
	}
	// Multi-threaded asynchronous ping
	loc := configure.NewLocator()
	wg := new(sync.WaitGroup)
	for i, v := range which {
		if v.TYPE == configure.SubscriptionType { // subscriptions cannot be pinged
			continue
		}
		wg.Add(1)
		go func(i int) {
			_ = which[i].PingWithLookup(loc, timeout, dialer, lookup)
			wg.Done()
		}(i)
	}
	wg.Wait()
	for i := len(which) - 1; i >= 0; i-- {
		if which[i].TYPE == configure.SubscriptionType { // do not return subscriptionType
			which = append(which[:i], which[i+1:]...)
		}
	}
	return which, nil
}

func addHosts(tmpl *v2ray.Template, vms []serverObj.ServerObj, lookup func(string) ([]string, error)) map[string]error {
	failures := make(map[string]error)
	// entries whose server could not be located are left nil by the caller
	located := vms[:0:0]
	for _, v := range vms {
		if v != nil {
			located = append(located, v)
		}
	}
	vms = located
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
	seen := make(map[string]bool)
	for _, v := range vms {
		if net.ParseIP(v.GetHostname()) == nil && !seen[v.GetHostname()] {
			seen[v.GetHostname()] = true
			wg.Add(1)
			go func(addr string) {
				limit <- struct{}{}
				defer func() {
					wg.Done()
					<-limit
				}()
				mu.Lock()
				delete(tmpl.DNS.Hosts, addr)
				mu.Unlock()
				ips, err := lookup(addr)
				if err != nil {
					mu.Lock()
					failures[addr] = err
					mu.Unlock()
					return
				}
				if len(ips) > 0 {
					ips = v2ray.FilterIPs(ips)
					if len(ips) == 0 {
						mu.Lock()
						failures[addr] = fmt.Errorf("node DNS %s: no usable IP addresses", addr)
						mu.Unlock()
						return
					}
					mu.Lock()
					tmpl.DNS.Hosts[addr] = ips
					mu.Unlock()
				}
			}(v.GetHostname())
		}
	}
	wg.Wait()
	return failures
}

func TestHttpLatency(which []*configure.Which, timeout time.Duration, maxParallel int, showLog bool, customTestUrl string) ([]*configure.Which, error) {
	if customTestUrl != "" {
		testURL, err := url.Parse(customTestUrl)
		if err != nil || (testURL.Scheme != "http" && testURL.Scheme != "https") || testURL.Hostname() == "" {
			return nil, common.Coded("INVALID_TEST_URL", fmt.Errorf("test URL %q must be an HTTP or HTTPS URL with a host", customTestUrl), map[string]interface{}{"testUrl": customTestUrl})
		}
	}
	var whiches = configure.NewWhiches(which)
	which = whiches.Get()
	for i := len(which) - 1; i >= 0; i-- {
		if which[i].TYPE == configure.SubscriptionType { // remove subscriptionType
			which = append(which[:i], which[i+1:]...)
		}
	}
	if len(which) == 0 {
		return which, nil
	}
	dialer := httpClient.DirectDialer(timeout, v2ray.IsTransparentOn(configure.GetSettingNotNil()))
	lookup, endpoint, lookupErr := nodeLookup(dialer)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if endpoint == nil {
		lookup = resolv.LookupHost
	}
	v2rayRunning := v2ray.ProcessManager.Running()
	wg := new(sync.WaitGroup)
	vms := make([]serverObj.ServerObj, len(which))
	//init vmessInfos
	loc := configure.NewLocator()
	for i := range which {
		which[i].Latency = ""
		sr, err := loc.Locate(&which[i].NodeRef)
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
		tmpl, err = v2ray.NewTemplateFromConnectedServers(nil, endpoint)
		if err != nil {
			if !errors.Is(err, v2ray.NoConnectedServerErr) {
				log.Warn("NewTemplateFromConnectedServers: %v", err)
			}
		}
	}
	if tmpl == nil {
		tmpl = v2ray.NewEmptyTemplate(&configure.Setting{
			RulePortMode: configure.WhitelistMode,
			TcpFastOpen:  configure.Default,
			MuxOn:        configure.No,
			Transparent:  configure.TransparentClose,
		})
		tmpl.SetAPI(nil)
	}
	// Until the process manager owns the template, its API producers are
	// ours to stop on an early return.
	handedOver := false
	defer func() {
		if !handedOver {
			_ = tmpl.Close()
		}
	}()
	inboundPortMap := make([]string, len(vms))
	pluginPortMap := make(map[int]int)
	listenAddr := "127.0.0.1"
	if tmpl.Setting != nil && tmpl.Setting.PortSharing {
		listenAddr = "0.0.0.0"
	}
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
			l, err := net.Listen("tcp", listenAddr+":0")
			if err == nil {
				port = l.Addr().(*net.TCPAddr).Port
				toClose = append(toClose, l)
				l2, err2 := net.ListenPacket("udp", listenAddr+":"+strconv.Itoa(port))
				if err2 == nil {
					toClose = append(toClose, l2)
					break
				}
			}
			if time.Since(t) > 3*time.Second {
				return nil, fmt.Errorf("could not find a free local port for the latency test within 3 s")
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
					return nil, fmt.Errorf("could not find a free local port for the latency test within 3 s")
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
		tmpl.Inbounds[len(tmpl.Inbounds)-1].Listen = listenAddr
		inboundPortMap[i] = v2rayInboundPort
	}
	for _, l := range toClose {
		_ = l.Close()
	}
	toClose = nil
	time.Sleep(30 * time.Millisecond)
	tmpl.Routing.DomainStrategy = "AsIs"
	if endpoint != nil {
		tmpl.NodeDNS = endpoint
		if err := tmpl.AddNodeDNSDomains(vms); err != nil {
			return nil, err
		}
	}
	failures := addHosts(tmpl, vms, lookup)
	if endpoint != nil {
		for i, v := range vms {
			if v != nil {
				if err := failures[v.GetHostname()]; err != nil {
					which[i].Latency = err.Error()
				}
			}
		}
	}
	tmpl.SetOutboundSockopt()
	v2ray.ProcessManager.SetLatencyTesting(true)
	handedOver = true
	if err := v2ray.ProcessManager.Start(tmpl); err != nil {
		v2ray.ProcessManager.SetLatencyTesting(false)
		if v2rayRunning && configure.GetConnectedServers() != nil {
			if restoreErr := v2ray.UpdateV2RayConfig(); restoreErr != nil {
				return nil, fmt.Errorf("%v; cannot restart v2ray-core: %w", err, restoreErr)
			}
		}
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
			httpLatency(which[i], inboundPortMap[i], timeout, customTestUrl)
			if showLog {
				log.Info("Test done[%v]%v: %v", i+1, which[i].Latency, which[i].Link)
			}
		}(i)
	}
	wg.Wait()
	v2ray.ProcessManager.SetLatencyTesting(false)
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
func httpLatency(which *configure.Which, port string, timeout time.Duration, customTestUrl string) {
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
	testUrl := HttpTestURL
	if len(customTestUrl) != 0 {
		testUrl = customTestUrl
	}
	req, err := http.NewRequest("GET", testUrl, nil)
	if err != nil {
		which.Latency = "SYSTEM ERROR"
		return
	}
	//req, _ := http.NewRequest("GET", "http://www.gstatic.com/generate_204", nil)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", "curl/7.70.0")
	resp, err := c.Do(req)
	setHTTPLatencyResult(which, resp, err, t)
}

func setHTTPLatencyResult(which *configure.Which, resp *http.Response, err error, started time.Time) {
	if resp != nil {
		defer resp.Body.Close()
	}
	if resp == nil && err == nil {
		which.Latency = "SYSTEM ERROR"
		return
	}
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 400 {
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				which.Latency = "TIMEOUT"
				return
			}
			var recErr tls.RecordHeaderError
			if errors.As(err, &recErr) {
				which.Latency = "INVALID"
				return
			}
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				which.Latency = "NOT STABLE"
				return
			}
			log.Debug("HTTP latency proxy request failed: %T", err)
			which.Latency = "PROXY ERROR"
		} else {
			which.Latency = "BAD RESPONSE"
		}
		return
	}
	which.Latency = fmt.Sprintf("%.0fms", time.Since(started).Seconds()*1000)
}

func isSupportedObj(obj serverObj.ServerObj) (bool, error) {
	tmpl := v2ray.NewEmptyTemplate(&configure.Setting{
		RulePortMode: configure.WhitelistMode,
		TcpFastOpen:  configure.Default,
		MuxOn:        configure.No,
		Transparent:  configure.TransparentClose,
	})
	// The template is thrown away: SetAPI would start a traffic producer
	// that nothing closes.
	err := tmpl.InsertMappingOutbound(obj, "0", false, 0, "socks")
	if err != nil {
		if strings.Contains(err.Error(), "unsupported") {
			return false, err
		}
	}
	return true, nil
}

func nodeLookup(dialer *net.Dialer) (func(string) ([]string, error), *resolv.IPDNSEndpoint, error) {
	if !v2ray.DNSHijackActive() {
		return func(host string) ([]string, error) { return resolv.LookupHostWithDialer(host, dialer) }, nil, nil
	}
	setting := configure.GetSettingNotNil()
	endpoint, err := v2ray.SelectNodeDNS(setting, configure.GetDnsRulesNotNil())
	if err != nil {
		return nil, nil, err
	}
	dnsDialer := dialer
	// A physical-interface binding cannot deliver to a resolver on this host.
	// Linux keeps the mark even for local sockets to bypass OUTPUT interception.
	if runtime.GOOS != "linux" && resolv.IsLocalIP(endpoint.IP) {
		dnsDialer = &net.Dialer{Timeout: dialer.Timeout}
	}
	return func(host string) ([]string, error) { return resolv.LookupNode(host, endpoint, dnsDialer) }, endpoint, nil
}
