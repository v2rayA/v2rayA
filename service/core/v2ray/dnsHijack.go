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

const HijackFlag = "# v2rayA DNS hijack"

var hijacker *ResolvHijacker

func (h *ResolvHijacker) HijackResolv() error {
	err := os.WriteFile(resolverFile,
		[]byte(HijackFlag+"\nnameserver 127.0.0.1\nnameserver 114.114.114.114\n"),
		os.FileMode(0644),
	)
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
		if hijacker.localDNS {
			os.WriteFile(resolverFile,
				[]byte(HijackFlag+"\nnameserver 223.6.6.6\nnameserver 114.114.114.114\n"),
				os.FileMode(0644),
			)
		}
		hijacker = nil

	}
}
