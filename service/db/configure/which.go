package configure

import (
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"time"
)

type Whiches struct {
	Touches        []*Which `json:"touches"`
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

func (ws *Whiches) Get() []*Which {
	return ws.Touches
}

func (ws *Whiches) Add(which Which) {
	ws.Touches = append(ws.Touches, &which)
}

func (ws *Whiches) Set(wt []*Which) {
	ws.Touches = wt
}

/*去重，并做下标范围检测，只保留符合下标范围的项*/
func (ws *Whiches) GetNonDuplicated() (w []*Which) {
	ts := make(map[Which]struct{})
	//下标范围检测，并利用map的key值无重复特性去重
	for i := range ws.Touches {
		ind := ws.Touches[i].ID - 1
		v := *ws.Touches[i]
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
	w = make([]*Which, 0)
	for k := range ts {
		t := k
		w = append(w, &t)
	}
	return
}

type Which struct {
	TYPE    TouchType `json:"_type"`                 //Server还是Subscription
	ID      int       `json:"id"`                    //代表某个subscription或某个server的ID是多少, 从1开始. 如果是SubscriptionServer, 代表这个server在该Subscription中的ID
	Sub     int       `json:"sub"`                   //仅当TYPE为SubscriptionServer时有效, 代表Subscription的下标, 从0开始.
	Latency string    `json:"pingLatency,omitempty"` //历史遗留问题，前后端通信还是使用pingLatency这个名字
	Link    string    //optional
}

func (w *Which) Ping(timeout time.Duration) (err error) {
	if w.TYPE == SubscriptionType {
		return newError("you cannot ping a subscription")
	}
	tsr, err := w.LocateServer()
	if err != nil {
		return
	}
	//BEGIN
	host := tsr.VmessInfo.Add
	if net.ParseIP(host) == nil {
		var hosts []string
		hosts, err = net.LookupHost(host)
		if err != nil || len(hosts) <= 0 {
			if err != nil {
				w.Latency = err.Error()
			} else {
				w.Latency = "querying dns failed: " + host
			}
			return
		}
		host = hosts[0]
	}
	t := time.Now()
	conn, e := net.DialTimeout("tcp", net.JoinHostPort(host, tsr.VmessInfo.Port), timeout)
	if e == nil || (strings.Contains(e.Error(), "refuse")) {
		if e == nil {
			_ = conn.Close()
		}
		//log.Println(host+":"+tsr.VmessInfo.Port, e)
		w.Latency = fmt.Sprintf("%.0fms", time.Since(t).Seconds()*1000)
	} else {
		log.Println(e)
		w.Latency = "TIMEOUT"
	}
	return
}

func (w *Which) LocateServer() (sr *ServerRaw, err error) {
	ind := w.ID - 1 //转化为下标
	switch w.TYPE {
	case ServerType:
		servers := GetServers()
		if ind < 0 || ind >= len(servers) {
			return nil, newError("LocateServer: ID exceed range")
		}
		return &servers[ind], nil
	case SubscriptionServerType:
		subscriptions := GetSubscriptions()
		if w.Sub < 0 || w.Sub >= len(subscriptions) || ind < 0 || ind >= len(subscriptions[w.Sub].Servers) {
			return nil, newError("LocateServer: ID or Sub exceed range")
		}
		return &subscriptions[w.Sub].Servers[ind], nil
	default:
		return nil, newError("LocateServer: invalid TYPE")
	}
}

func (ws *Whiches) FillLinks() (err error) {
	servers := GetServers()
	subscriptions := GetSubscriptions()
	for _, w := range ws.Touches {
		ind := w.ID - 1 //转化为下标
		switch w.TYPE {
		case ServerType:
			if ind < 0 || ind >= len(servers) {
				return newError("LocateServer: ID exceed range")
			}
			w.Link = servers[ind].VmessInfo.ExportToURL()
		case SubscriptionServerType:
			if w.Sub < 0 || w.Sub >= len(subscriptions) || ind < 0 || ind >= len(subscriptions[w.Sub].Servers) {
				return newError("LocateServer: ID or Sub exceed range")
			}
			w.Link = subscriptions[w.Sub].Servers[ind].VmessInfo.ExportToURL()
		default:
			return newError("LocateServer: invalid TYPE")
		}
	}
	return nil
}

func (ws *Whiches) Save() (err error) {
	whiches := ws.GetNonDuplicated()
	var (
		serverIndexes       = make(map[int]*Which)
		subscriptionIndexes = make(map[int]map[int]*Which)
	)
	for _, which := range whiches {
		ind := which.ID - 1 // to index
		switch which.TYPE {
		case ServerType:
			serverIndexes[ind] = which
		case SubscriptionServerType:
			if _, ok := subscriptionIndexes[which.Sub]; !ok {
				subscriptionIndexes[which.Sub] = make(map[int]*Which)
			}
			subscriptionIndexes[which.Sub][ind] = which
		default:
		}
	}
	// TODO:
	return nil
}
