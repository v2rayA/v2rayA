package dnsPoison

import (
	"log"
	"sync"
	"v2rayA/dataStructure/lru"
)

type ReservedIpPool struct {
	domainLRU              *lru.LRU
	lastInsertedReservedIP reservedIP
	m                      map[string]reservedIP
	sync.Mutex
}

func NewReservedIpPool() *ReservedIpPool {
	return &ReservedIpPool{
		domainLRU:              lru.New(0b10000000000000000000000000000 - 0b100000000000000000000*2),
		lastInsertedReservedIP: 0,
		m:                      make(map[string]reservedIP),
	}
}

func (u *ReservedIpPool) Lookup(domain string) [4]byte {
	u.Lock()
	defer u.Unlock()
	out := u.domainLRU.ShiftOrInsert(domain)
	if out != nil {
		// full and replace
		d := out.(string)
		ip, ok := u.m[d]
		if !ok {
			log.Fatal("cannot find the ip of deleting domain:", d)
		}
		u.m[domain] = ip
		delete(u.m, d)
		return ip.IP()
	} else {
		// not full and (exists or insert)
		ip, ok := u.m[domain]
		if ok {
			return ip.IP()
		}
		u.lastInsertedReservedIP = u.lastInsertedReservedIP.Next()
		u.m[domain] = u.lastInsertedReservedIP
		return u.lastInsertedReservedIP.IP()
	}
}
