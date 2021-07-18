package lru

// LRU is Non Thread Safe
type LRU struct {
	maxsize uint64
	size    uint64
	head    *node
	tail    *node
	m       map[interface{}]*node
}
type node struct {
	value interface{}
	pre   *node
	next  *node
}

// with head node
func New(maxsize uint64) *LRU {
	head := new(node)
	return &LRU{
		maxsize: maxsize,
		size:    0,
		head:    head,
		tail:    head,
		m:       make(map[interface{}]*node),
	}
}

func (lru *LRU) pushLeft(value interface{}) (out interface{}) {
	fn := lru.head.next
	p := &node{
		value: value,
		pre:   lru.head,
		next:  fn,
	}
	if p.next == nil {
		lru.tail = p
	}
	if fn != nil {
		fn.pre = p
	}
	lru.head.next = p
	lru.m[value] = p
	if lru.size+1 > lru.maxsize {
		out = lru.tail.value
		lru.tail = lru.tail.pre
		lru.tail.next = nil
	} else {
		lru.size++
	}
	return out
}

func (lru *LRU) ShiftOrInsert(value interface{}) (out interface{}) {
	p, ok := lru.m[value]
	if !ok {
		out = lru.pushLeft(value)
		return
	}
	pre := p.pre
	next := p.next
	if lru.tail == p {
		lru.tail = pre
	}
	pre.next = next
	if next != nil {
		next.pre = pre
	}
	fn := lru.head.next
	if fn != nil {
		fn.pre = p
	}
	p.next = fn
	p.pre = lru.head
	lru.head.next = p
	return nil
}

func (lru *LRU) Full() bool {
	return lru.size == lru.maxsize
}

func (lru *LRU) Size() uint64 {
	return lru.size
}

func (lru *LRU) MaxSize() uint64 {
	return lru.maxsize
}

/*  transfer the tail to the head */
func (lru *LRU) Boost() (value interface{}) {
	value = lru.tail.value
	lru.ShiftOrInsert(lru.tail)
	return
}
