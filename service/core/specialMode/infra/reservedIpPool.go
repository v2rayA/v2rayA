package infra

import (
	lru2 "github.com/v2rayA/v2rayA/infra/dataStructure/lru"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"sync"
)

type ReservedIpPool struct {
	domainLRU              *lru2.LRU
	lastInsertedReservedIP reservedIP
	m                      map[string]reservedIP
	sync.Mutex
}

func NewReservedIpPool() *ReservedIpPool {
	return &ReservedIpPool{
		// (256 - 240) * 256 * 256 * (256 - 2)
		domainLRU:              lru2.New(lru2.FixedLength, 0b10000000000000000000000000000-0b100000000000000000000*2),
		lastInsertedReservedIP: 0,
		m:                      make(map[string]reservedIP),
	}
}

func (u *ReservedIpPool) Lookup(domain string) [4]byte {
	u.Lock()
	defer u.Unlock()
	out, removed := u.domainLRU.GetOrInsert(domain, func() (val interface{}) {
		return domain
	})
	if len(removed) > 0 {
		// full and replace
		d := out.(string)
		ip, ok := u.m[d]
		if !ok {
			log.Fatal("cannot find the ip of deleting domain: %v", d)
		}
		u.m[domain] = ip
		delete(u.m, d)
		return ip.IP()
	} else {
		// exists or (not full and insert)
		ip, ok := u.m[domain]
		if ok {
			return ip.IP()
		}
		u.lastInsertedReservedIP = u.lastInsertedReservedIP.Next()
		u.m[domain] = u.lastInsertedReservedIP
		return u.lastInsertedReservedIP.IP()
	}
}
