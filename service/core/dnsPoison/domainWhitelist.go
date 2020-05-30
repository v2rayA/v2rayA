package dnsPoison

import (
	"sync"
)

// Three consecutive succeed proposal will add the domain to the whitelist,
// but any against will remove it from the whitelist or stop the progress.
type domainWhitelist struct {
	whitelist      map[string]interface{}
	progress       map[string]uint8
	progressMutex  sync.Mutex
	whitelistMutex sync.RWMutex
}

func newDomainWhitelist() *domainWhitelist {
	return &domainWhitelist{
		whitelist:      make(map[string]interface{}),
		progress:       make(map[string]uint8),
		progressMutex:  sync.Mutex{},
		whitelistMutex: sync.RWMutex{},
	}
}

func (w *domainWhitelist) Propose(domain string) bool {
	w.whitelistMutex.RLock()
	if _, exists := w.whitelist[domain]; exists {
		// domain exists in whitelist and no need to propose to
		w.whitelistMutex.RUnlock()
		return false
	}
	w.whitelistMutex.RUnlock()

	w.progressMutex.Lock()
	defer w.progressMutex.Unlock()
	w.progress[domain]++
	if w.progress[domain] >= 3 {
		w.progress[domain] = 0
		w.whitelistMutex.Lock()
		w.whitelist[domain] = nil
		w.whitelistMutex.Unlock()
		return true
	}
	return false
}

func (w *domainWhitelist) Against(domain string) (suc bool) {
	w.whitelistMutex.Lock()
	if _, exists := w.whitelist[domain]; exists {
		// domain exists in whitelist
		delete(w.whitelist, domain)
		suc = true
	}
	w.whitelistMutex.Unlock()

	w.progressMutex.Lock()
	w.progress[domain] = 0
	w.progressMutex.Unlock()
	return
}

func (w *domainWhitelist) Exists(domain string) bool {
	w.whitelistMutex.RLock()
	_, exists := w.whitelist[domain]
	w.whitelistMutex.RUnlock()
	return exists
}
