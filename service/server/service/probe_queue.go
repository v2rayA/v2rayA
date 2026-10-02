package service

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
)

// Hold the slot until cmd.Wait returns, including cancellation and failed startup.
var probeCoreSlot = make(chan struct{}, 1)

func acquireProbeCore(ctx context.Context) error {
	select {
	case probeCoreSlot <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func releaseProbeCore() { <-probeCoreSlot }

func pingGroupCandidates(ctx context.Context, nodes []serverObj.ServerObj) []subscriptionProbeResult {
	results := make([]subscriptionProbeResult, len(nodes))
	dialer := httpClient.DirectDialer(subscriptionProbeTimeout, v2ray.IsTransparentOn(configure.GetSettingNotNil()))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range min(8, len(nodes)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				if nodes[i] == nil {
					results[i].err = fmt.Errorf("missing candidate")
					continue
				}
				start := time.Now()
				conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(nodes[i].GetHostname(), strconv.Itoa(nodes[i].GetPort())))
				results[i].latency, results[i].err = time.Since(start), err
				if err != nil {
					results[i].latency = subscriptionProbeTimeout
				}
				if conn != nil {
					conn.Close()
				}
			}
		}()
	}
	for i := range nodes {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}

// UDP transports may have no TCP listener. Their actual proxy check is authoritative.
func udpOnlyCandidate(node serverObj.ServerObj) bool {
	switch node.GetProtocol() {
	case "hysteria2", "hy2", "tuic", "juicity", "wireguard":
		return true
	}
	return false
}
