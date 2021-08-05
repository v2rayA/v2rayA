package v2ray

import (
	"sync"
	"time"
)

type Feed struct {
	boxes   map[string][]*chan Message
	mu      sync.Mutex
	boxSize int
}
type Message struct {
	ProduceTime time.Time
	Body        interface{}
}

func NewSubscriptions(boxSize int) *Feed {
	if boxSize <= 0 {
		return nil
	}
	return &Feed{
		boxes:   make(map[string][]*chan Message),
		boxSize: boxSize,
	}
}

func (s *Feed) BoxSize() int {
	return s.boxSize
}

func (s *Feed) RegisterProduct(product string) {
	s.mu.Lock()
	s.boxes[product] = nil
	s.mu.Unlock()
}

func (s *Feed) SubscribeMessage(product string) (box *chan Message, cancel func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.boxes[product]; !ok {
		return nil, nil
	}
	b := make(chan Message, s.boxSize)
	s.boxes[product] = append(s.boxes[product], &b)
	index := len(s.boxes[product]) - 1
	cancel = func() {
		if index >= len(s.boxes[product]) || s.boxes[product][index] != &b {
			index = -1
			for i := range s.boxes[product] {
				if s.boxes[product][i] == &b {
					index = i
					break
				}
			}
		}
		if index == -1 {
			// the cancel function is invoked more than once
			return
		}
		s.boxes[product] = append(s.boxes[product][:index], s.boxes[product][index+1:]...)
	}
	return &b, cancel
}

func (s *Feed) ProductMessage(product string, message interface{}) (numConsumer int) {
	msg := Message{
		ProduceTime: time.Now(),
		Body:        message,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	consumers, ok := s.boxes[product]
	if !ok {
		return 0
	}
	var cnt int
	for _, consumer := range consumers {
		select {
		case *consumer <- msg:
			cnt++
		default:
		}
	}
	return cnt
}
