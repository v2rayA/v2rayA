package v2ray

import (
	"fmt"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"os"
	"time"
)

const (
	resolverFile  = "/etc/resolv.conf"
	checkInterval = 3 * time.Second
)

type ResolvHijacker struct {
	ticker   *time.Ticker
	localDNS bool
}

func NewResolvHijacker() *ResolvHijacker {
	hij := ResolvHijacker{
		ticker:   time.NewTicker(checkInterval),
		localDNS: specialMode.ShouldLocalDnsListen(),
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

var hijacker *ResolvHijacker

func (h *ResolvHijacker) HijackResolv() error {
	err := os.WriteFile(resolverFile, []byte(`# v2rayA DNS hijack
223.5.5.5
119.29.29.29
`), os.FileMode(0644))
	if err != nil {
		err = fmt.Errorf("failed to hijackDNS: [write] %v", err)
	}
	return err
}

func resetResolvHijacker() {
	if hijacker != nil {
		hijacker.Close()
	}
	hijacker = NewResolvHijacker()
}

func removeResolvHijacker() {
	if hijacker != nil {
		hijacker.Close()
		hijacker = nil
	}
}
