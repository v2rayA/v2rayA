package infra

import (
	"sync"
)

// NOTE: The following description is obsolete

// Three consecutive succeed proposal will add the domain to the whitelist,
// but any against will remove it from the whitelist or stop the progress.
type domainBlacklist struct {
	list          map[string]interface{}
	progress      map[string]int8
	progressMutex sync.Mutex
	listMutex     sync.RWMutex
}

func newDomainBlacklist() *domainBlacklist {
	return &domainBlacklist{
		list:          make(map[string]interface{}),
		progress:      make(map[string]int8),
		progressMutex: sync.Mutex{},
		listMutex:     sync.RWMutex{},
	}
}

func (w *domainBlacklist) Propose(domain string) bool {
	w.listMutex.RLock()
	if _, exists := w.list[domain]; exists {
		// domain exists in list and no need to propose to
		w.listMutex.RUnlock()
		return false
	}
	w.listMutex.RUnlock()

	w.progressMutex.Lock()
	defer w.progressMutex.Unlock()
	w.progress[domain]++
	if w.progress[domain] >= 1 {
		w.progress[domain] = 0
		w.listMutex.Lock()
		w.list[domain] = nil
		w.listMutex.Unlock()
		return true
	}
	return false
}

func (w *domainBlacklist) Against(domain string) (suc bool) {
	w.listMutex.Lock()
	if _, exists := w.list[domain]; exists {
		// domain exists in list
		delete(w.list, domain)
		suc = true
	}
	w.listMutex.Unlock()

	w.progressMutex.Lock()
	// NOTE: in this version we ban this domain
	w.progress[domain] = 0
	w.progressMutex.Unlock()
	return
}

func (w *domainBlacklist) Exists(domain string) bool {
	w.listMutex.RLock()
	_, exists := w.list[domain]
	w.listMutex.RUnlock()
	return exists
}
