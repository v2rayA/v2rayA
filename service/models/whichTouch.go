package models

import "sort"

type TouchType string

const (
	SubscriptionType = "subscription"
	ServerType       = "server"
)

/*前后端通信用*/
type WhichTouch struct {
	TYPE TouchType `json:"_type"` //Server还是Subscription
	ID   int       `json:"id"`    //ID是多少, 从1开始. 如果是Subscription的一个server, 代表这个server在该Subscription中的ID
	Sub  int       `json:"sub"`   //如果是Subscription的一个server, Sub代表Subscription的下标, 从0开始
}
type WhichTouches struct {
	Touches []WhichTouch `json:"touches"`
	sort.Interface
}

func (t WhichTouches) Len() int {
	return len(t.Touches)
}

func (t WhichTouches) Less(i, j int) bool {
	//server排在subscription前面
	quantifyType := map[TouchType]int{
		ServerType:       0,
		SubscriptionType: 1,
	}
	if t.Touches[i].TYPE == t.Touches[j].TYPE {
		return t.Touches[i].ID > t.Touches[j].ID
	}
	return quantifyType[t.Touches[i].TYPE] < quantifyType[t.Touches[j].TYPE]
}

func (t WhichTouches) Swap(i, j int) {
	t.Touches[i], t.Touches[j] = t.Touches[j], t.Touches[i]
}
func (t *WhichTouches) GetTouches() []WhichTouch {
	return t.Touches
}
