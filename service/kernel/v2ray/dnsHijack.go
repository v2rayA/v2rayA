package v2ray

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const checkInterval = 3 * time.Second

// Overridden by the tests; production always uses these paths.
var (
	resolvPath       = "/etc/resolv.conf"
	resolvBackupPath = "/etc/resolv.conf.v2raya_backup"
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
	ticker    *time.Ticker
	done      chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
	localDNS  bool
}

func NewResolvHijacker() *ResolvHijacker {
	if runtime.GOOS != "linux" {
		return nil
	}
	hij := ResolvHijacker{
		ticker:   time.NewTicker(checkInterval),
		done:     make(chan struct{}),
		localDNS: ShouldLocalDnsListen(),
	}
	hij.HijackResolv()
	hij.wg.Add(1)
	go func() {
		defer hij.wg.Done()
		for {
			select {
			case <-hij.ticker.C:
				hij.HijackResolv()
			case <-hij.done:
				return
			}
		}
	}()
	return &hij
}
func (h *ResolvHijacker) Close() error {
	h.closeOnce.Do(func() {
		h.ticker.Stop()
		close(h.done)
	})
	// Wait for a tick that is already writing: restoring while it runs put the
	// hijacked file straight back.
	h.wg.Wait()
	return nil
}

const HijackFlag = "# v2rayA DNS hijack"

// fallbackResolver is the nameserver behind the module's: the first direct
// port-53 upstream of the DNS rules, a public one when there is none.
func fallbackResolver() string {
	for _, s := range directDnsServers() {
		if host, port, err := net.SplitHostPort(s); err == nil && port == "53" {
			return host
		}
	}
	return "119.29.29.29"
}

const (
	symlinkMarker = "# v2rayA saved symlink: "
	missingMarker = "# v2rayA: no resolv.conf"
	emptyMarker   = "# v2rayA: empty resolv.conf"
)

var hijacker *ResolvHijacker
var hijackerMu sync.Mutex

// HijackResolv 将 /etc/resolv.conf 的 nameserver 设置为 127.2.0.17。
// 当新 DNS 模块启用时，127.2.0.17:53 的流量被 iptables 规则重定向到 :52353（新 DNS 模块端口）。
// 当使用旧 xray DNS 时，127.2.0.17:53 的流量被 dns-in (dokodemo-door) 捕获。
func (h *ResolvHijacker) HijackResolv() error {
	if runtime.GOOS != "linux" {
		return nil
	}
	if err := backupResolv(); err != nil {
		log.Warn("DNS hijack: %v", err)
	}
	// /etc/resolv.conf is a symlink on any systemd-resolved, resolvconf or
	// openresolv system. Writing through it overwrites the resolver's own
	// file; replace the link with a regular file instead.
	if fi, err := os.Lstat(resolvPath); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		if err := os.Remove(resolvPath); err != nil {
			log.Warn("DNS hijack: cannot replace the %v link: %v", resolvPath, err)
		}
	}
	err := os.WriteFile(resolvPath,
		[]byte(HijackFlag+"\nnameserver 127.2.0.17\nnameserver "+fallbackResolver()+"\n"),
		os.FileMode(0644),
	)
	if err != nil {
		err = fmt.Errorf("failed to hijackDNS: [write] %v", err)
	}
	return err
}

// backupResolv records what /etc/resolv.conf was before the first hijack: its
// target when it is a symlink (systemd-resolved, resolvconf and openresolv all
// use one), otherwise its content. Without this, stopping the proxy left the
// machine on the hard-coded public resolvers and the symlink gone for good.
func backupResolv() error {
	if _, err := os.Lstat(resolvBackupPath); err == nil {
		// Already taken; never overwrite it with a hijacked file.
		return nil
	}
	fi, err := os.Lstat(resolvPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Record that there was no file at all.
			return writeResolvBackup([]byte(missingMarker + "\n"))
		}
		return fmt.Errorf("cannot inspect %v: %w", resolvPath, err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(resolvPath)
		if err != nil {
			return fmt.Errorf("cannot read the %v link: %w", resolvPath, err)
		}
		return writeResolvBackup([]byte(symlinkMarker + target + "\n"))
	}
	b, err := os.ReadFile(resolvPath)
	if err != nil {
		return fmt.Errorf("cannot read %v: %w", resolvPath, err)
	}
	if strings.HasPrefix(string(b), HijackFlag) {
		// A hijacked file from a previous run that was never restored: keep
		// looking for the real backup instead of saving our own work.
		return nil
	}
	if len(b) == 0 {
		// an empty original is legitimate; an empty backup is what a
		// crash mid-write leaves, so the two must not look alike
		return writeResolvBackup([]byte(emptyMarker + "\n"))
	}
	return writeResolvBackup(b)
}

func writeResolvBackup(content []byte) (err error) {
	temp, err := os.CreateTemp(filepath.Dir(resolvBackupPath), "."+filepath.Base(resolvBackupPath)+".*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}()
	if err = temp.Chmod(0644); err != nil {
		return err
	}
	if _, err = temp.Write(content); err != nil {
		return err
	}
	if err = temp.Sync(); err != nil {
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, resolvBackupPath)
}

// restoreResolv puts back what backupResolv saved. It reports whether the
// original configuration is back.
func restoreResolv() bool {
	b, err := os.ReadFile(resolvBackupPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warn("DNS hijack: cannot read backup %v: %v", resolvBackupPath, err)
		}
		return false
	}
	if len(b) == 0 {
		log.Warn("DNS hijack: backup %v is empty", resolvBackupPath)
		return false
	}
	content := string(b)
	switch {
	case strings.HasPrefix(content, symlinkMarker):
		target := strings.TrimSpace(strings.TrimPrefix(content, symlinkMarker))
		if target == "" {
			return false
		}
		if err := os.Remove(resolvPath); err != nil && !os.IsNotExist(err) {
			log.Warn("DNS hijack: cannot replace %v: %v", resolvPath, err)
			return false
		}
		if err := os.Symlink(target, resolvPath); err != nil {
			log.Warn("DNS hijack: cannot restore the %v link to %v: %v", resolvPath, target, err)
			return false
		}
	case strings.HasPrefix(strings.TrimSpace(content), emptyMarker):
		if err := os.WriteFile(resolvPath, nil, 0644); err != nil {
			log.Warn("DNS hijack: cannot restore an empty %v: %v", resolvPath, err)
			return false
		}
		_ = os.Remove(resolvBackupPath)
		return true
	case strings.HasPrefix(strings.TrimSpace(content), missingMarker):
		if err := os.Remove(resolvPath); err != nil && !os.IsNotExist(err) {
			log.Warn("DNS hijack: cannot remove %v: %v", resolvPath, err)
			return false
		}
	default:
		if err := os.WriteFile(resolvPath, b, 0644); err != nil {
			log.Warn("DNS hijack: cannot restore %v: %v", resolvPath, err)
			return false
		}
	}
	_ = os.Remove(resolvBackupPath)
	return true
}

func resetResolvHijacker() {
	if runtime.GOOS != "linux" {
		return
	}
	hijackerMu.Lock()
	defer hijackerMu.Unlock()
	if hijacker != nil {
		hijacker.Close()
	}
	hijacker = NewResolvHijacker()
}

func removeResolvHijacker() {
	if runtime.GOOS != "linux" {
		return
	}
	hijackerMu.Lock()
	defer hijackerMu.Unlock()
	if hijacker != nil {
		hijacker.Close()
		if hijacker.localDNS && !restoreResolv() {
			// No usable backup: leave a working resolver behind rather than a
			// file pointing at a listener that is gone.
			log.Warn("DNS hijack: no backup of %v to restore, writing public resolvers instead", resolvPath)
			os.WriteFile(resolvPath,
				[]byte(HijackFlag+"\nnameserver "+fallbackResolver()+"\n"),
				os.FileMode(0644),
			)
		}
		hijacker = nil
	}
}
