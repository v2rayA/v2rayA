package shadowsocksr

import (
	"sync"
)

type SSRs struct {
	SSRs  []SSR
	mutex sync.Mutex
}

func (r *SSRs) ClearAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, ssr := range r.SSRs {
		if ssr.IsRunning() {
			_ = ssr.Close()
		}
	}
	r.SSRs = make([]SSR, 0)
}

func (r *SSRs) Append(ssr SSR) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.SSRs = append(r.SSRs, ssr)
}
