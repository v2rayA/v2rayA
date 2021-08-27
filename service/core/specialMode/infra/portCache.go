package infra

import (
	"sync"
	"time"
)

type localPort string
type portCache struct {
	pool  map[localPort]*time.Timer
	mutex sync.RWMutex
	sync.RWMutex
}

func newPortCache() *portCache {
	return &portCache{
		pool:  make(map[localPort]*time.Timer),
		mutex: sync.RWMutex{},
	}
}

func (p *portCache) Set(port localPort, timeout time.Duration) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if _, ok := p.pool[port]; ok {
		p.pool[port].Reset(timeout)
		return
	}
	p.pool[port] = time.AfterFunc(timeout, func() {
		p.mutex.Lock()
		defer p.mutex.Unlock()
		delete(p.pool, port)
	})
}

func (p *portCache) Exists(port localPort) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	_, ok := p.pool[port]
	return ok
}
