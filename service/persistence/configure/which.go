package configure

import (
	"errors"
	"fmt"
	"github.com/sparrc/go-ping"
	"sort"
	"time"
)

type Whiches struct {
	Touches        []Which `json:"touches"`
	sort.Interface `json:"-"`
}

func (ws Whiches) Len() int {
	return len(ws.Touches)
}

func (ws Whiches) Less(i, j int) bool {
	//server排在subscription前面
	quantifyType := map[TouchType]int{
		ServerType:             0,
		SubscriptionType:       1,
		SubscriptionServerType: 2,
	}
	if ws.Touches[i].TYPE == ws.Touches[j].TYPE {
		return ws.Touches[i].ID > ws.Touches[j].ID
	}
	return quantifyType[ws.Touches[i].TYPE] < quantifyType[ws.Touches[j].TYPE]
}

func (ws Whiches) Swap(i, j int) {
	ws.Touches[i], ws.Touches[j] = ws.Touches[j], ws.Touches[i]
}

/*
对which排序，先按类型排，再按下标排。

排序规则：

server < subscription

大下标 < 小下标
*/
func (ws Whiches) Sort() {
	sort.Sort(ws)
}

func (ws *Whiches) Get() []Which {
	return ws.Touches
}

func (ws *Whiches) Set(wt []Which) {
	ws.Touches = wt
}

/*去重，并做下标范围检测，只保留符合下标范围的项*/
func (ws *Whiches) GetNonDuplicated() (w []Which) {
	ts := make(map[Which]struct{})
	//下标范围检测，并利用map的key值无重复特性去重
	for i := range ws.Touches {
		ind := ws.Touches[i].ID - 1
		v := ws.Touches[i]
		switch v.TYPE {
		case SubscriptionType:
			if ind >= 0 && ind < GetLenSubscriptions() {
				ts[v] = struct{}{}
			}
		case ServerType:
			if ind >= 0 && ind < GetLenServers() {
				ts[v] = struct{}{}
			}
		case SubscriptionServerType:
			if v.Sub >= 0 && v.Sub < GetLenSubscriptions() && ind >= 0 && ind < GetLenSubscriptionServers(v.Sub) {
				ts[v] = struct{}{}
			}
		}
	}
	//还原回slice
	w = make([]Which, 0)
	for k := range ts {
		w = append(w, k)
	}
	return
}

type Which struct {
	TYPE        TouchType `json:"_type"` //Server还是Subscription
	ID          int       `json:"id"`    //代表某个subscription或某个server的ID是多少, 从1开始. 如果是SubscriptionServer, 代表这个server在该Subscription中的ID
	Sub         int       `json:"sub"`   //仅当TYPE为SubscriptionServer时有效, 代表Subscription的下标, 从0开始.
	PingLatency *string   `json:"pingLatency,omitempty"`
}

func (w *Which) Ping(count int, timeout time.Duration) (err error) {
	if w.TYPE == SubscriptionType {
		return errors.New("subscription不能ping")
	}
	tsr, err := w.LocateServer()
	if err != nil {
		return
	}
	pinger, err := ping.NewPinger(tsr.VmessInfo.Add)
	if err != nil {
		return
	}
	pinger.Count = count
	pinger.Timeout = timeout
	pinger.SetPrivileged(true)
	pinger.Run()
	s := pinger.Statistics()
	w.PingLatency = new(string)
	*w.PingLatency = fmt.Sprintf("平均: %dms, 最快: %dms, 最慢: %dms. 丢包: %d/%d(%.1f%%)", int(s.AvgRtt.Seconds()*1000), int(s.MinRtt.Seconds()*1000), int(s.MaxRtt.Seconds()*1000), s.PacketsSent-s.PacketsRecv, s.PacketsSent, s.PacketLoss)
	return
}

func (wt *Which) LocateServer() (*ServerRaw, error) {
	ind := wt.ID - 1 //转化为下标
	switch wt.TYPE {
	case ServerType:
		servers := GetServers()
		if ind < 0 || ind >= len(servers) {
			return nil, errors.New("ID超出下标范围")
		}
		return &servers[ind], nil
	case SubscriptionServerType:
		subscriptions := GetSubscriptions()
		if wt.Sub < 0 || wt.Sub >= len(subscriptions) || ind < 0 || ind >= len(subscriptions[wt.Sub].Servers) {
			return nil, errors.New("ID或Sub超出下标范围")
		}
		return &subscriptions[wt.Sub].Servers[ind], nil
	default:
		return nil, errors.New("LocateServer: 无效的TYPE")
	}
}
