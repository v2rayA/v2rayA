package service

import (
	"github.com/gorilla/websocket"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"sync"
	"time"
)

const (
	writeWait = v2ray.ApiFeedInterval
)

type MessageHandler struct {
	conn  *websocket.Conn
	boxes map[string]*v2ray.Box
}
type Message struct {
	ProduceTime int64       `json:"produce_time"`
	Type        string      `json:"type"`
	Body        interface{} `json:"body"`
}

func NewMessageHandler(conn *websocket.Conn) *MessageHandler {
	h := &MessageHandler{
		conn:  conn,
		boxes: make(map[string]*v2ray.Box),
	}
	for _, product := range v2ray.ApiProducts {
		box := v2ray.ApiFeed.SubscribeMessage(product)
		h.boxes[product] = box
	}
	return h
}

// Read should be invoked only once
func (h *MessageHandler) Read() {
	defer func() {
		for product := range h.boxes {
			h.boxes[product].Cancel()
		}
		h.conn.Close()
	}()
	for {
		if _, _, err := h.conn.ReadMessage(); err != nil {
			//log.Println(err)
			break
		}
	}
}

// Write should be invoked only once
func (h *MessageHandler) Write() {
	defer func() {
		h.conn.Close()
	}()
	var wg sync.WaitGroup
	for _, box := range h.boxes {
		wg.Add(1)
		go func(box *v2ray.Box) {
			defer wg.Done()
			for msg := range box.Messages {
				_ = h.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := h.conn.WriteJSON(Message{
					ProduceTime: msg.ProduceTime,
					Type:        msg.Product,
					Body:        msg.Body,
				}); err != nil {
					return
				}
			}
		}(box)
	}
	wg.Wait()
}
