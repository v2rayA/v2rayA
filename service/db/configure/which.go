package configure

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Whiches struct {
	Touches        []*Which `json:"touches"`
	sort.Interface `json:"-"`
}

func (ws *Whiches) Len() int {
	if ws == nil {
		return 0
	}
	return len(ws.Touches)
}

func (ws *Whiches) Less(i, j int) bool {
	quantifyType := map[TouchType]int{
		ServerType:             0,
		SubscriptionType:       1,
		SubscriptionServerType: 2,
	}
	// serverType has higher priority
	if ws.Touches[i].TYPE != ws.Touches[j].TYPE {
		return quantifyType[ws.Touches[i].TYPE] < quantifyType[ws.Touches[j].TYPE]
	}
	// they are both server
	if ws.Touches[i].TYPE == ServerType {
		return ws.Touches[i].ID < ws.Touches[j].ID
	}
	// they are both subscriptionServer, but sub indexes are different
	if ws.Touches[i].Sub != ws.Touches[j].Sub {
		return ws.Touches[i].Sub < ws.Touches[j].Sub
	}
	// they are both subscriptionServer, and sub indexes are the same
	return ws.Touches[i].ID < ws.Touches[j].ID
}

func (ws *Whiches) Swap(i, j int) {
	ws.Touches[i], ws.Touches[j] = ws.Touches[j], ws.Touches[i]
}

/*
Sort whiches, first sort by type, and then by index.

Sorting rules: server < subscription

small index < large index
*/
func (ws *Whiches) Sort() {
	sort.Sort(ws)
}

/*
Sort whiches, first sort by type, and then by index.

Sorting rules: server < subscription

small index > large index
*/
func (ws *Whiches) SortSameTypeReverse() {
	sort.Sort(ws)
	var typ TouchType
	var begin = 0
	var i int
	for i = 0; i < len(ws.Touches); i++ {
		if typ == "" {
			typ = ws.Touches[0].TYPE
		}
		if ws.Touches[i].TYPE != typ {
			for j := 0; j < (i-begin)/2; j++ {
				ws.Touches[begin+j], ws.Touches[i-j-1] = ws.Touches[i-j-1], ws.Touches[begin+j]
			}
			begin = i
		}
	}
	if begin < len(ws.Touches)-1 {
		for j := 0; j < (i-begin)/2; j++ {
			ws.Touches[begin+j], ws.Touches[i-j-1] = ws.Touches[i-j-1], ws.Touches[begin+j]
		}
	}
}

func (ws *Whiches) Get() []*Which {
	if ws == nil {
		return nil
	}
	return ws.Touches
}

func (ws *Whiches) Add(which Which) {
	ws.Touches = append(ws.Touches, &which)
}

func (ws *Whiches) Extend(which Whiches) {
	ws.Touches = append(ws.Touches, which.Touches...)
}

func NewWhiches(wt []*Which) *Whiches {
	ws := new(Whiches)
	var theCopy = make([]*Which, len(wt))
	copy(theCopy, wt)
	ws.Touches = theCopy
	return ws
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
	TYPE     TouchType `json:"_type"`                 //Server还是Subscription
	ID       int       `json:"id"`                    //代表某个subscription或某个server的ID是多少, 从1开始. 如果是SubscriptionServer, 代表这个server在该Subscription中的ID
	Sub      int       `json:"sub"`                   //仅当TYPE为SubscriptionServer时有效, 代表Subscription的下标, 从0开始.
	Latency  string    `json:"pingLatency,omitempty"` //历史遗留问题，前后端通信还是使用pingLatency这个名字，该值仅作为ping的返回值
	Link     string    //optional
	Outbound string    `json:"outbound"`
}

func (w *Which) EqualTo(another Which) (ok bool) {
	switch w.TYPE {
	case SubscriptionServerType:
		return another.TYPE == w.TYPE &&
			another.Sub == w.Sub &&
			another.ID == w.ID &&
			another.Outbound == w.Outbound
	case ServerType:
		return another.TYPE == w.TYPE &&
			another.ID == w.ID &&
			another.Outbound == w.Outbound
	case SubscriptionType:
		return another.TYPE == w.TYPE &&
			another.ID == w.ID
	default:
		return false
	}
}
func (w *Which) Ping(timeout time.Duration) (err error) {
	if w.TYPE == SubscriptionType {
		return fmt.Errorf("you cannot ping a subscription")
	}
	tsr, err := w.LocateServerRaw()
	if err != nil {
		return
	}
	//BEGIN
	host := tsr.ServerObj.GetHostname()
	if net.ParseIP(host) == nil {
		var hosts []string
		hosts, err = resolv.LookupHost(host)
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
	conn, e := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(tsr.ServerObj.GetPort())), timeout)
	if e == nil || (strings.Contains(e.Error(), "refuse")) {
		if e == nil {
			_ = conn.Close()
		}
		//log.Println(host+":"+tsr.VmessInfo.Port, e)
		w.Latency = fmt.Sprintf("%.0fms", time.Since(t).Seconds()*1000)
	} else {
		log.Debug("Ping: %v", e)
		w.Latency = "TIMEOUT"
	}
	return
}

func (w *Which) LocateServerRaw() (sr *ServerRawV2, err error) {
	ind := w.ID - 1 //转化为下标
	switch w.TYPE {
	case ServerType:
		servers := GetServersV2()
		if ind < 0 || ind >= len(servers) {
			return nil, fmt.Errorf("LocateServerRaw: ID exceed range")
		}
		return &servers[ind], nil
	case SubscriptionServerType:
		subscriptions := GetSubscriptionsV2()
		if w.Sub < 0 || w.Sub >= len(subscriptions) || ind < 0 || ind >= len(subscriptions[w.Sub].Servers) {
			return nil, fmt.Errorf("LocateServerRaw: ID or Sub exceed range")
		}
		return &subscriptions[w.Sub].Servers[ind], nil
	default:
		return nil, fmt.Errorf("LocateServerRaw: invalid TYPE")
	}
}

func (ws *Whiches) FillLinks() (err error) {
	servers := GetServersV2()
	subscriptions := GetSubscriptionsV2()
	for _, w := range ws.Touches {
		ind := w.ID - 1 //转化为下标
		switch w.TYPE {
		case ServerType:
			if ind < 0 || ind >= len(servers) {
				return fmt.Errorf("LocateServerRaw: ID exceed range")
			}
			w.Link = servers[ind].ServerObj.ExportToURL()
		case SubscriptionServerType:
			if w.Sub < 0 || w.Sub >= len(subscriptions) || ind < 0 || ind >= len(subscriptions[w.Sub].Servers) {
				return fmt.Errorf("LocateServerRaw: ID or Sub exceed range")
			}
			w.Link = subscriptions[w.Sub].Servers[ind].ServerObj.ExportToURL()
		default:
			return fmt.Errorf("LocateServerRaw: invalid TYPE")
		}
	}
	return nil
}

func (ws *Whiches) SaveLatencies() (err error) {
	whiches := ws.GetNonDuplicated()
	var (
		serverIndexes       = make(map[int]*Which)
		subscriptionIndexes = make(map[int]map[int]*Which)
	)
	// deduplicate
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
	// set servers
	for index, which := range serverIndexes {
		sRaw, err := which.LocateServerRaw()
		if err != nil {
			return err
		}
		sRaw.Latency = which.Latency
		if err := SetServer(index, sRaw); err != nil {
			return err
		}
	}
	// set subscriptions
	for subIndex, serverIndexes := range subscriptionIndexes {
		subRaw := GetSubscriptionV2(subIndex)
		for index, which := range serverIndexes {
			subRaw.Servers[index].Latency = which.Latency
		}
		if err := SetSubscription(subIndex, subRaw); err != nil {
			return err
		}
	}
	return nil
}
