package lru

import (
	"github.com/v2rayA/v2rayA/infra/dataStructure/linklist"
	"sync"
	"time"
)

type LimitStrategy int

const (
	FixedLength LimitStrategy = iota
	FixedTimeout
)

type LRU struct {
	list         *linklist.Linklist
	index        map[interface{}]*linklist.Node
	reverseIndex map[*linklist.Node]interface{}
	mutex        sync.Mutex
	strategy     LimitStrategy
	limit        int64
}

type EncapsulatedValue struct {
	Value       interface{}
	LastUseTime time.Time
}

func New(strategy LimitStrategy, limit int64) *LRU {
	return &LRU{
		index:        make(map[interface{}]*linklist.Node),
		reverseIndex: make(map[*linklist.Node]interface{}),
		list:         linklist.NewLinklist(),
		strategy:     strategy,
		limit:        limit,
	}
}

func (l *LRU) GetOrInsert(key interface{}, valFunc func() (val interface{})) (val interface{}, removed []*EncapsulatedValue) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	val = l.get(key)
	if val == nil {
		val := valFunc()
		return val, l.insert(key, val)
	}
	return val, nil
}

func (l *LRU) Get(key interface{}) (value interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.get(key)
}

func (l *LRU) get(key interface{}) (value interface{}) {
	v, ok := l.index[key]
	if !ok {
		return nil
	}
	l.list.Promote(v)
	ev := v.Val.(*EncapsulatedValue)
	ev.LastUseTime = time.Now()
	return ev.Value
}

func (l *LRU) Insert(key interface{}, val interface{}) (removed []*EncapsulatedValue) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.insert(key, val)
}

func (l *LRU) insert(key interface{}, val interface{}) (removed []*EncapsulatedValue) {
	ev := &EncapsulatedValue{
		Value:       val,
		LastUseTime: time.Now(),
	}
	node := l.list.PushFront(ev)
	l.index[key] = node
	l.reverseIndex[node] = key
	switch l.strategy {
	case FixedLength:
		if int64(len(l.index)) > l.limit {
			back := l.list.Back()
			removed = []*EncapsulatedValue{back.Val.(*EncapsulatedValue)}
			key := l.reverseIndex[back]
			l.list.Remove(back)
			delete(l.index, key)
			delete(l.reverseIndex, back)
		}
	case FixedTimeout:
		now := time.Now()
		// pop timeout exceeded nodes until the last node does not exceed
		for {
			back := l.list.Back()
			if back == nil {
				break
			}
			ev := back.Val.(*EncapsulatedValue)
			if int64(now.Sub(ev.LastUseTime)) < l.limit {
				break
			}
			removed = append(removed, ev)
			key := l.reverseIndex[back]
			l.list.Remove(back)
			delete(l.index, key)
			delete(l.reverseIndex, back)
		}
	}
	return
}
