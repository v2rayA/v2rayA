package service

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/kernel/v2ray/where"
)

func ValidateNodeProbe(mode, target string) error {
	if mode != "" && mode != "proxy" {
		return common.Coded("BAD_REQUEST", fmt.Errorf("unknown node check mode %q", mode), nil)
	}
	if target == "" {
		return nil
	}
	u, err := url.Parse(target)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
		return common.Coded("INVALID_TEST_URL", fmt.Errorf("test URL must be an HTTP(S) URL without credentials"), map[string]interface{}{"testUrl": target})
	}
	return nil
}

// Keep probe configuration and lifetime independent of the running service core.
func startProbeCore(ctx context.Context, tmpl *v2ray.Template, ports []string) (func(), error) {
	bin, err := where.GetV2rayBinPath()
	if err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp("", "v2raya-probe-")
	if err != nil {
		return nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, tmpl.ToConfigBytes(), 0600); err != nil {
		cleanup()
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, bin, "run", "--config="+path)
	if err := cmd.Start(); err != nil {
		cancel()
		cleanup()
		return nil, err
	}
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()
	stop := func() { cancel(); <-done; cleanup() }
	readyCtx, readyCancel := context.WithTimeout(ctx, 5*time.Second)
	defer readyCancel()
	for _, port := range ports {
		if port == "" {
			continue
		}
		for {
			conn, err := (&net.Dialer{Timeout: 100 * time.Millisecond}).DialContext(readyCtx, "tcp", "127.0.0.1:"+port)
			if err == nil {
				_ = conn.Close()
				break
			}
			select {
			case <-done:
				stop()
				return nil, fmt.Errorf("probe core exited before opening its listeners")
			case <-readyCtx.Done():
				stop()
				return nil, fmt.Errorf("probe core startup: %w", readyCtx.Err())
			case <-time.After(20 * time.Millisecond):
			}
		}
	}
	return stop, nil
}
