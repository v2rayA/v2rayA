package models

import "time"

/*Touch是前后端通信形式，其结构设计和前端统一*/
type SubscriptionStatus string
type Touch struct {
	Servers       []TouchServer  `json:"servers"`
	Subscriptions []Subscription `json:"subscriptions"`
	ConnectedServer *WhichTouch       `json:"connectedServer"` //冗余一个信息，方便查找
}
type TouchServer struct {
	ID        int    `json:"id"`
	TYPE      string `json:"_type"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Net       string `json:"net"`
	Connected bool   `json:"connected"`
}
type Subscription struct {
	ID      int                `json:"id"`
	TYPE    string             `json:"_type"`
	Host    string             `json:"host"`
	Status  SubscriptionStatus `json:"status"`
	Servers []TouchServer      `json:"servers"`
}

func NewUpdateStatus() SubscriptionStatus {
	return SubscriptionStatus("上次更新：" + time.Now().Format("2006-1-2 15:04:05"))
}
