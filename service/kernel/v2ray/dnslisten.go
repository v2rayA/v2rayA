package v2ray

// 本文件保存原 specialMode 包中与 redirect 透明代理相关的 DNS 监听辅助函数。
// supervisor / fakedns 相关代码已随 specialMode 包一同移除。

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/common/netTools/netstat"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
)

var (
	DnsPortCheckFailedErr = fmt.Errorf("failed to check dns port occupation")
	DnsPortOccupied       = fmt.Errorf("dns port 53 is occupied")
)

// ShouldLocalDnsListen 在透明代理启用时返回 true，
// 表示需要劫持 /etc/resolv.conf 将系统 DNS 指向 127.2.0.17:53，
// 使 DNS 流量可以被 iptables/nftables 规则捕获后重定向到 52353。
//
// 新 DNS 模块架构下，无论 Redirect 还是 TProxy 模式，都需要劫持 DNS：
//
//	/etc/resolv.conf → 127.2.0.17:53
//	  → iptables REDIRECT/TPROXY --dport 53 → 52353
//	  → v2raya-core DNS 模块处理
//
// 不依赖 xray 的 dns-in/dns-out，DNS 模块直接查询上游并返回结果。
func ShouldLocalDnsListen() bool {
	setting := configure.GetSettingNotNil()
	if setting.Transparent == configure.TransparentClose {
		return false
	}
	if conf.GetEnvironmentConfig().Lite {
		return false
	}
	return true
}

var couldListenCache struct {
	couldListenLocalhost bool
	err                  error
	lastUpdate           time.Time
	mu                   sync.Mutex
}

// CouldLocalDnsListen 检查 53 端口是否可用，结果缓存 3 秒。
func CouldLocalDnsListen() (couldListenLocalhost bool, err error) {
	couldListenCache.mu.Lock()
	defer couldListenCache.mu.Unlock()
	if time.Since(couldListenCache.lastUpdate) < 3*time.Second {
		return couldListenCache.couldListenLocalhost, couldListenCache.err
	}
	defer func() {
		couldListenCache.lastUpdate = time.Now()
		couldListenCache.couldListenLocalhost = couldListenLocalhost
		couldListenCache.err = err
	}()
	occupied, sockets, err := ports.IsPortOccupied([]string{"53:udp"})
	if err != nil {
		return false, fmt.Errorf("%w: %v", DnsPortCheckFailedErr, err)
	}
	if err = netstat.FillProcesses(sockets); err != nil {
		return false, fmt.Errorf("%w: %v. please try again later", DnsPortCheckFailedErr, err)
	}
	var occupiedErr error
	if occupied {
		for _, socket := range sockets {
			p := socket.Proc
			if p == nil {
				continue
			}
			if p.PPID == strconv.Itoa(os.Getpid()) {
				continue
			}
			occupiedErr = fmt.Errorf("%w by %v(%v)", DnsPortOccupied, p.Name, p.PID)
			if socket.LocalAddress.IP.Equal(net.ParseIP("127.2.0.17")) {
				return false, occupiedErr
			}
			if socket.LocalAddress.IP.IsUnspecified() {
				return false, occupiedErr
			}
		}
	}
	return true, occupiedErr
}
