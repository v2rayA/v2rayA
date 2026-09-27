package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/kernel/v2ray/where"
)

const (
	subscriptionProbeTimeout   = 5 * time.Second
	subscriptionSpeedProbeURL  = "https://speed.cloudflare.com/__down?bytes=262144"
	subscriptionSpeedProbeSize = int64(256 * 1024)
	subscriptionMinSpeed       = int64(100 * 1024)
)

func probeSubscription(servers []serverObj.ServerObj, probeURL string) []subscriptionProbeResult {
	return probeSubscriptionWithContext(context.Background(), servers, probeURL)
}

func probeSubscriptionWithContext(ctx context.Context, servers []serverObj.ServerObj, probeURL string) []subscriptionProbeResult {
	servers = cloneProbeNodes(servers)
	results := make([]subscriptionProbeResult, len(servers))
	// Limit extra core processes on routers; still wait for the entire subscription.
	jobs := make(chan int)
	var wg sync.WaitGroup
	for worker := 0; worker < 2; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = probeSubscriptionServerMeasurementWithContext(ctx, servers[i], probeURL, subscriptionProbeTimeout)
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

// Configuration normalises some protocols in place. Probes therefore receive
// objects reconstructed from their links instead of pointers stored in the
// catalog. Plugin-managed nodes are deliberately left nil: configuring one can
// run an external process that cannot be cancelled by this worker.
func cloneProbeNodes(servers []serverObj.ServerObj) []serverObj.ServerObj {
	clones := make([]serverObj.ServerObj, len(servers))
	for i, server := range servers {
		if server == nil {
			continue
		}
		if _, plugin := server.(*serverObj.Plugin); plugin {
			continue
		}
		clone, err := ResolveURL(server.ExportToURL())
		if err == nil {
			clones[i] = clone
		}
	}
	return clones
}

func probeSubscriptionServer(server serverObj.ServerObj, probeURL string, timeout time.Duration) (time.Duration, error) {
	return probeSubscriptionServerWithContext(context.Background(), server, probeURL, timeout)
}

func probeSubscriptionServerWithContext(parent context.Context, server serverObj.ServerObj, probeURL string, timeout time.Duration) (time.Duration, error) {
	result := probeSubscriptionServerMeasurementWithContext(parent, server, probeURL, timeout)
	return result.latency, result.err
}

func probeSubscriptionServerMeasurementWithContext(parent context.Context, server serverObj.ServerObj, probeURL string, timeout time.Duration) subscriptionProbeResult {
	if err := parent.Err(); err != nil {
		return subscriptionProbeResult{err: err}
	}
	u, err := url.Parse(probeURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return subscriptionProbeResult{err: fmt.Errorf("invalid HTTP probe URL")}
	}
	if server == nil {
		return subscriptionProbeResult{err: fmt.Errorf("server requires an external plugin and cannot be probed in isolation")}
	}
	bin, err := where.GetV2rayBinPath()
	if err != nil {
		return subscriptionProbeResult{err: err}
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return subscriptionProbeResult{err: err}
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	tmpl := v2ray.NewEmptyTemplate(configure.GetSettingNotNil())
	defer tmpl.Close()
	if err := tmpl.InsertMappingOutbound(server, strconv.Itoa(port), false, 0, "http"); err != nil {
		return subscriptionProbeResult{err: err}
	}
	tmpl.Inbounds[0].Listen = "127.0.0.1"
	// Candidate connections must not be intercepted by the main group's
	// transparent proxy, especially while that group is empty and blocking.
	tmpl.SetOutboundSockopt()
	tmpl.Routing.DomainStrategy = "AsIs"
	tmpl.Log = &coreObj.Log{Access: "none", Error: "none", Loglevel: "none"}
	file, err := os.CreateTemp("", "v2raya-subscription-*.json")
	if err != nil {
		return subscriptionProbeResult{err: err}
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(tmpl.ToConfigBytes()); err != nil {
		file.Close()
		return subscriptionProbeResult{err: err}
	}
	if err = file.Close(); err != nil {
		return subscriptionProbeResult{err: err}
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
		return subscriptionProbeResult{err: err}
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
			return subscriptionProbeResult{err: fmt.Errorf("probe core exited before becoming ready")}
		case <-readyCtx.Done():
			return subscriptionProbeResult{err: fmt.Errorf("probe core startup timeout")}
		case <-ticker.C:
			conn, e := net.DialTimeout("tcp", address, 100*time.Millisecond)
			if e == nil {
				conn.Close()
				return probeHTTP(ctx, address, probeURL, timeout)
			}
		}
	}
}

func probeHTTP(ctx context.Context, proxyAddress, probeURL string, timeout time.Duration) subscriptionProbeResult {
	transport := &http.Transport{Proxy: http.ProxyURL(&url.URL{Scheme: "http", Host: proxyAddress})}
	defer transport.CloseIdleConnections()
	client := &http.Client{
		Transport:     transport,
		Timeout:       timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return probeHTTPWithClient(ctx, client, probeURL, subscriptionSpeedProbeURL)
}

type concurrentProbeResult struct {
	latency    time.Duration
	throughput int64
	latencyRun bool
	speedRun   bool
	err        error
}

// probeHTTPWithClient checks the configured reachability URL and a bounded
// throughput sample concurrently. The configured URL determines reachability;
// an unavailable speed sample remains unknown so it cannot disconnect every
// otherwise reachable candidate.
func probeHTTPWithClient(ctx context.Context, client *http.Client, probeURL, speedURL string) subscriptionProbeResult {
	probeCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make(chan concurrentProbeResult, 2)
	go func() {
		latency, err := probeLatency(probeCtx, client, probeURL)
		results <- concurrentProbeResult{latency: latency, latencyRun: true, err: err}
	}()
	go func() {
		throughput, err := probeThroughput(probeCtx, client, speedURL)
		results <- concurrentProbeResult{throughput: throughput, speedRun: err == nil}
	}()

	measurement := subscriptionProbeResult{}
	for range 2 {
		result := <-results
		if result.latencyRun && result.err != nil {
			cancel()
			return subscriptionProbeResult{err: result.err}
		}
		if result.latencyRun {
			measurement.latency = result.latency
		} else if result.speedRun {
			measurement.throughput = result.throughput
			measurement.speedMeasured = true
		}
	}
	if err := ctx.Err(); err != nil {
		measurement.err = err
	}
	return measurement
}

func probeLatency(ctx context.Context, client *http.Client, probeURL string) (time.Duration, error) {
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

func probeThroughput(ctx context.Context, client *http.Client, speedURL string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, speedURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Cache-Control", "no-cache")
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("speed probe returned HTTP %d", resp.StatusCode)
	}
	start := time.Now()
	bytesRead, err := io.CopyN(io.Discard, resp.Body, subscriptionSpeedProbeSize)
	elapsed := time.Since(start)
	throughput := probeSpeed(bytesRead, elapsed)
	if err != nil && bytesRead == 0 {
		return 0, fmt.Errorf("speed probe downloaded no data: %w", err)
	}
	return throughput, nil
}

func probeSpeed(bytesRead int64, elapsed time.Duration) int64 {
	if elapsed <= 0 {
		return subscriptionMinSpeed
	}
	return int64(float64(bytesRead) / elapsed.Seconds())
}
