package touch

import (
	"time"
)

/*
Touch是树型结构的前后端通信形式，其结构设计和前端统一。
*/
type SubscriptionStatus string
type Touch struct {
	Servers         []TouchServer  `json:"servers"`
	Subscriptions   []Subscription `json:"subscriptions"`
	ConnectedServer *WhichTouch    `json:"connectedServer"` //冗余一个信息，方便查找
}
type TouchServer struct {
	ID          int       `json:"id"`
	TYPE        TouchType `json:"_type"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Net         string    `json:"net"`
	Connected   bool      `json:"connected"`
	PingLatency string    `json:"pingLatency"`
}
type Subscription struct {
	ID      int                `json:"id"`
	TYPE    TouchType          `json:"_type"`
	Host    string             `json:"host"`
	Status  SubscriptionStatus `json:"status"`
	Servers []TouchServer      `json:"servers"`
}

func NewUpdateStatus() SubscriptionStatus {
	return SubscriptionStatus("上次更新：" + time.Now().Local().Format("2006-1-2 15:04:05"))
}
func NewUpdateFailStatus(reason string) SubscriptionStatus {
	return SubscriptionStatus(time.Now().Local().Format("2006-1-2 15:04:05") + "尝试更新失败："+reason)
}
