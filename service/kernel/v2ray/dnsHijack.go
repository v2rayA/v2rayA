package v2ray

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

const (
	resolverFile  = "/etc/resolv.conf"
	checkInterval = 3 * time.Second
)

// ResolvHijacker 劫持系统 DNS 配置，将命名服务器指向 127.2.0.17:53。
// 其流量路径为：
//
//	/etc/resolv.conf → 127.2.0.17:53 → iptables/nftables DNS 规则 → 重定向到 52353 (新 DNS 模块)
//
// 旧路径（xray DNS 模式）：
//
//	/etc/resolv.conf → 127.2.0.17:53 → dns-in (dokodemo-door, port 53) → xray DNS 路由 → dns-out
//
// 新 DNS 模块启用时，劫持后的 53 端口流量被 iptables REDIRECT/TPROXY 规则捕获，
// 导向 DNS 模块的监听端口 52353，由 UpstreamManager 进行解析。
type ResolvHijacker struct {
	ticker   *time.Ticker
	localDNS bool
}

func NewResolvHijacker() *ResolvHijacker {
	if runtime.GOOS != "linux" {
		return nil
	}
	hij := ResolvHijacker{
		ticker:   time.NewTicker(checkInterval),
		localDNS: ShouldLocalDnsListen(),
	}
	hij.HijackResolv()
	go func() {
		for range hij.ticker.C {
			hij.HijackResolv()
		}
	}()
	return &hij
}
func (h *ResolvHijacker) Close() error {
	h.ticker.Stop()
	return nil
}

const HijackFlag = "# v2rayA DNS hijack"

var hijacker *ResolvHijacker

// HijackResolv 将 /etc/resolv.conf 的 nameserver 设置为 127.2.0.17。
// 当新 DNS 模块启用时，127.2.0.17:53 的流量被 iptables 规则重定向到 :52353（新 DNS 模块端口）。
// 当使用旧 xray DNS 时，127.2.0.17:53 的流量被 dns-in (dokodemo-door) 捕获。
func (h *ResolvHijacker) HijackResolv() error {
	if runtime.GOOS != "linux" {
		return nil
	}
	err := os.WriteFile(resolverFile,
		[]byte(HijackFlag+"\nnameserver 127.2.0.17\nnameserver 119.29.29.29\n"),
		os.FileMode(0644),
	)
	if err != nil {
		err = fmt.Errorf("failed to hijackDNS: [write] %v", err)
	}
	return err
}

func resetResolvHijacker() {
	if runtime.GOOS != "linux" {
		return
	}
	if hijacker != nil {
		hijacker.Close()
	}
	hijacker = NewResolvHijacker()
}

func removeResolvHijacker() {
	if runtime.GOOS != "linux" {
		return
	}
	if hijacker != nil {
		hijacker.Close()
		if hijacker.localDNS {
			os.WriteFile(resolverFile,
				[]byte(HijackFlag+"\nnameserver 223.6.6.6\nnameserver 119.29.29.29\n"),
				os.FileMode(0644),
			)
		}
		hijacker = nil
	}
}
