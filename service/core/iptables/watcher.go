package iptables

import (
	"time"
)

// LocalIPWatcher invokes functions when interface IPs change
type LocalIPWatcher struct {
	ticker      *time.Ticker
	cidrPool    map[string]struct{}
	AddedFunc   func(cidr string)
	RemovedFunc func(cidr string)
	UpdateFunc  func(cidrs []string)
}

func NewLocalIPWatcher(interval time.Duration, AddedFunc func(cidr string), RemovedFunc func(cidr string)) *LocalIPWatcher {
	w := LocalIPWatcher{
		ticker:      time.NewTicker(interval),
		cidrPool:    make(map[string]struct{}),
		AddedFunc:   AddedFunc,
		RemovedFunc: RemovedFunc,
	}
	w.SyncIP()
	go func() {
		for range w.ticker.C {
			w.SyncIP()
		}
	}()
	return &w
}

func (w *LocalIPWatcher) Close() error {
	w.AddedFunc = func(cidr string) {}
	w.RemovedFunc = w.AddedFunc
	w.ticker.Stop()
	return nil
}

func (w *LocalIPWatcher) SyncIP() {
	cidrs, err := GetLocalCIDR()
	if err != nil {
		return
	}
	m := make(map[string]struct{})

	for _, cidr := range cidrs {
		m[cidr] = struct{}{}
		if _, ok := w.cidrPool[cidr]; !ok {
			w.AddedFunc(cidr)
		}
	}
	for cidr := range w.cidrPool {
		if _, ok := m[cidr]; !ok {
			w.RemovedFunc(cidr)
		}
	}
	w.cidrPool = m
}
