package linklist

// linklist with head node and tail node
type Node struct {
	prior *Node
	next  *Node
	Val   interface{}
}

func (p *Node) Next() *Node {
	return p.next
}

func (p *Node) Prior() *Node {
	return p.prior
}

type Linklist struct {
	head *Node
	tail *Node
}

func NewLinklist() *Linklist {
	head := new(Node)
	tail := new(Node)
	head.next = tail
	tail.prior = head
	return &Linklist{
		head: head,
		tail: tail,
	}
}

func (l *Linklist) Front() *Node {
	node := l.head.next
	if node == l.tail {
		return nil
	}
	return node
}

func (l *Linklist) Back() *Node {
	node := l.tail.prior
	if node == l.head {
		return nil
	}
	return node
}

func (l *Linklist) Head() *Node {
	return l.head
}

func (l *Linklist) Tail() *Node {
	return l.tail
}

func (l *Linklist) Empty() bool {
	return l.head.next == l.tail
}

func (l *Linklist) InsertAfter(prior *Node, val interface{}) *Node {
	if prior == l.tail {
		return nil
	}
	p := new(Node)
	p.Val = val
	p.prior = prior
	p.next = prior.next
	prior.next.prior = p
	prior.next = p
	return p
}

func (l *Linklist) PushFront(val interface{}) *Node {
	return l.InsertAfter(l.head, val)
}

func (l *Linklist) PushBack(val interface{}) *Node {
	return l.InsertAfter(l.tail.prior, val)
}

func (l *Linklist) Promote(p *Node) {
	if p == l.Front() {
		return
	}
	p.prior.next = p.next
	p.next.prior = p.prior
	p.prior = l.head
	p.next = l.head.next
	l.head.next.prior = p
	l.head.next = p
}

func (l *Linklist) Demote(p *Node) {
	if p == l.Back() {
		return
	}
	p.prior.next = p.next
	p.next.prior = p.prior
	p.prior = l.tail.prior
	p.next = l.tail
	l.tail.prior.next = p
	l.tail.prior = p
}

func (l *Linklist) Remove(p *Node) {
	if p == l.head || p == l.tail {
		return
	}
	p.prior.next = p.next
	p.next.prior = p.prior
}
