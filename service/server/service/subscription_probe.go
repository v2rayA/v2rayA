package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/coreObj"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db/configure"
)

const subscriptionProbeTimeout = 5 * time.Second

func probeSubscription(servers []serverObj.ServerObj, probeURL string) []subscriptionProbeResult {
	return probeSubscriptionWithContext(context.Background(), servers, probeURL)
}

func probeSubscriptionWithContext(ctx context.Context, servers []serverObj.ServerObj, probeURL string) []subscriptionProbeResult {
	results := make([]subscriptionProbeResult, len(servers))
	// Limit extra core processes on routers; still wait for the entire subscription.
	jobs := make(chan int)
	var wg sync.WaitGroup
	for worker := 0; worker < 2; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i].latency, results[i].err = probeSubscriptionServerWithContext(ctx, servers[i], probeURL, subscriptionProbeTimeout)
			}
		}()
	}
	for i := range servers {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}

func probeSubscriptionServer(server serverObj.ServerObj, probeURL string, timeout time.Duration) (time.Duration, error) {
	return probeSubscriptionServerWithContext(context.Background(), server, probeURL, timeout)
}

func probeSubscriptionServerWithContext(parent context.Context, server serverObj.ServerObj, probeURL string, timeout time.Duration) (time.Duration, error) {
	if err := parent.Err(); err != nil {
		return 0, err
	}
	u, err := url.Parse(probeURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return 0, fmt.Errorf("invalid HTTP probe URL")
	}
	if server == nil || server.NeedPluginPort() {
		return 0, fmt.Errorf("server requires an external plugin and cannot be probed in isolation")
	}
	bin, err := where.GetV2rayBinPath()
	if err != nil {
		return 0, err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	tmpl := v2ray.NewEmptyTemplate(configure.GetSettingNotNil())
	defer tmpl.Close()
	if err := tmpl.InsertMappingOutbound(server, strconv.Itoa(port), false, 0, "http"); err != nil {
		return 0, err
	}
	if len(tmpl.Plugins) > 0 || len(tmpl.PluginManagerInfoList) > 0 {
		return 0, fmt.Errorf("server requires an external plugin and cannot be probed in isolation")
	}
	tmpl.Inbounds[0].Listen = "127.0.0.1"
	tmpl.Routing.DomainStrategy = "AsIs"
	tmpl.Log = &coreObj.Log{Access: "none", Error: "none", Loglevel: "none"}
	file, err := os.CreateTemp("", "v2raya-subscription-*.json")
	if err != nil {
		return 0, err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(tmpl.ToConfigBytes()); err != nil {
		file.Close()
		return 0, err
	}
	if err = file.Close(); err != nil {
		return 0, err
	}
	startupTimeout := time.Duration(conf.GetEnvironmentConfig().CoreStartupTimeout) * time.Second
	if startupTimeout <= 0 {
		startupTimeout = 15 * time.Second
	}
	ctx, cancel := context.WithTimeout(parent, startupTimeout+timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "run", "--config="+file.Name())
	cmd.Env = append(os.Environ(), "XRAY_LOCATION_ASSET="+asset.GetV2rayLocationAssetOverride(), "V2RAY_CONF_GEOLOADER=memconservative")
	listener.Close()
	if err = cmd.Start(); err != nil {
		return 0, err
	}
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	defer func() {
		cancel()
		<-done
	}()
	readyCtx, readyCancel := context.WithTimeout(ctx, startupTimeout)
	defer readyCancel()
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return 0, fmt.Errorf("probe core exited before becoming ready")
		case <-readyCtx.Done():
			return 0, fmt.Errorf("probe core startup timeout")
		case <-ticker.C:
			conn, e := net.DialTimeout("tcp", address, 100*time.Millisecond)
			if e == nil {
				conn.Close()
				return probeHTTP(ctx, address, probeURL, timeout)
			}
		}
	}
}

func probeHTTP(ctx context.Context, proxyAddress, probeURL string, timeout time.Duration) (time.Duration, error) {
	transport := &http.Transport{Proxy: http.ProxyURL(&url.URL{Scheme: "http", Host: proxyAddress})}
	defer transport.CloseIdleConnections()
	client := &http.Client{
		Transport:     transport,
		Timeout:       timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Cache-Control", "no-cache")
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("probe returned HTTP %d", resp.StatusCode)
	}
	return time.Since(start), nil
}
