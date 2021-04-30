package v2ray

import (
	"fmt"
	"github.com/v2rayA/v2rayA/core/dnsPoison/entity"
	"github.com/v2rayA/v2rayA/db/configure"
	"os"
	"strings"
	"time"
)

const (
	resolverFile  = "/etc/resolv.conf"
	checkInterval = 3 * time.Second
)

type DnsHijacker struct {
	ticker  *time.Ticker
	fakedns bool
}

func NewDnsHijacker() *DnsHijacker {
	hij := DnsHijacker{
		ticker:  time.NewTicker(checkInterval),
		fakedns: entity.ShouldDnsPoisonOpen() == 2,
	}
	hij.HijackDNS()
	go func() {
		for range hij.ticker.C {
			hij.HijackDNS()
		}
	}()
	return &hij
}
func (h *DnsHijacker) Close() error {
	h.ticker.Stop()
	return nil
}

var hijacker *DnsHijacker

func (h *DnsHijacker) HijackDNS() error {
	alternatives := []string{
		"223.5.5.5",
		"119.29.29.29",
		"223.6.6.6",
		"180.76.76.76",
		"114.114.114.114",
		"208.67.222.222",
	}
	m := make(map[string]struct{})
	userSetDns := configure.GetDnsListNotNil()
	for _, dns := range userSetDns {
		m[dns] = struct{}{}
	}
	cnt := 0
	maxcnt := 2
	if h.fakedns {
		// fakedns
		alternatives = append([]string{"127.0.0.1"}, alternatives...)
	}
	var builder strings.Builder
	builder.WriteString("# v2rayA DNS hijack\n")
	for _, dns := range alternatives {
		// should not be duplicated with user preset dns
		if _, ok := m[dns]; !ok {
			builder.WriteString("nameserver " + dns + "\n")
			cnt++
			if cnt >= maxcnt {
				break
			}
		}
	}
	err := os.WriteFile(resolverFile, []byte(builder.String()), os.FileMode(0644))
	if err != nil {
		err = fmt.Errorf("failed to hijackDNS: [write] %v", err)
	}
	return err
}

func resetDnsHijacker() {
	if hijacker != nil {
		hijacker.Close()
	}
	hijacker = NewDnsHijacker()
}

func removeDnsHijacker() {
	if hijacker != nil {
		hijacker.Close()
		hijacker = nil
	}
}
