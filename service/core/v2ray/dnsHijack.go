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
