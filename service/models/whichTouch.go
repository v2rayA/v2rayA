package models

import "sort"

type TouchType string

const (
	SubscriptionType       = "subscription"
	ServerType             = "server"
	SubscriptionServerType = "subscriptionServer"
)

type WhichTouch struct {
	TYPE        TouchType `json:"_type"` //Server还是Subscription
	ID          int       `json:"id"`    //代表某个subscription或某个server的ID是多少, 从1开始. 如果是SubscriptionServer, 代表这个server在该Subscription中的ID
	Sub         int       `json:"sub"`   //仅当TYPE为SubscriptionServer时有效, 代表Subscription的下标, 从0开始.
	PingLatency *string   `json:"pingLatency,omitempty"`
}

/*
WhichTouches是线性结构的前后端通信形式，其结构设计和前端统一。
*/
type WhichTouches struct {
	Touches        []WhichTouch `json:"touches"`
	sort.Interface `json:"-"`
}

func (t WhichTouches) Len() int {
	return len(t.Touches)
}

func (t WhichTouches) Less(i, j int) bool {
	//server排在subscription前面
	quantifyType := map[TouchType]int{
		ServerType:             0,
		SubscriptionType:       1,
		SubscriptionServerType: 2,
	}
	if t.Touches[i].TYPE == t.Touches[j].TYPE {
		return t.Touches[i].ID > t.Touches[j].ID
	}
	return quantifyType[t.Touches[i].TYPE] < quantifyType[t.Touches[j].TYPE]
}

func (t WhichTouches) Swap(i, j int) {
	t.Touches[i], t.Touches[j] = t.Touches[j], t.Touches[i]
}

/*
对touches排序，先按类型排，再按下标排。

排序规则：

server < subscription

大下标 < 小下标
*/
func (t WhichTouches) Sort() {
	sort.Sort(t)
}

func (t *WhichTouches) GetTouches() []WhichTouch {
	return t.Touches
}

func (t *WhichTouches) SetTouches(wt []WhichTouch) {
	t.Touches = wt
}

/*去重，并做下标范围检测，只保留符合下标范围的项*/
func (t *WhichTouches) GetNonDuplicatedTouches(tr *TouchRaw) (wts []WhichTouch) {
	ts := make(map[WhichTouch]struct{})
	//下标范围检测，并利用map的key值无重复特性去重
	for i := range t.Touches {
		ind := t.Touches[i].ID - 1
		v := t.Touches[i]
		switch v.TYPE {
		case SubscriptionType:
			if ind >= 0 && ind < len(tr.Subscriptions) {
				ts[v] = struct{}{}
			}
		case ServerType:
			if ind >= 0 && ind < len(tr.Servers) {
				ts[v] = struct{}{}
			}
		case SubscriptionServerType:
			if v.Sub >= 0 && v.Sub < len(tr.Subscriptions) && ind >= 0 && ind < len(tr.Subscriptions[v.Sub].Servers) {
				ts[v] = struct{}{}
			}
		}
	}
	//还原回slice
	wts = make([]WhichTouch, 0)
	for k := range ts {
		wts = append(wts, k)
	}
	return
}
