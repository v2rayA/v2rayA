package configure

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

type NodeRef struct {
	TYPE     TouchType `json:"_type"`
	ID       int       `json:"id"`
	Sub      int       `json:"sub"`
	Outbound string    `json:"outbound"`
}

type NodeRefs struct {
	Touches []*NodeRef `json:"touches"`
}

func (rs *NodeRefs) Len() int {
	if rs == nil {
		return 0
	}
	return len(rs.Touches)
}

func (rs *NodeRefs) Get() []*NodeRef {
	if rs == nil {
		return nil
	}
	return rs.Touches
}

func (rs *NodeRefs) Add(ref NodeRef) {
	rs.Touches = append(rs.Touches, &ref)
}

func (rs *NodeRefs) Extend(refs NodeRefs) {
	rs.Touches = append(rs.Touches, refs.Touches...)
}

func NewNodeRefs(refs []*NodeRef) *NodeRefs {
	result := &NodeRefs{Touches: make([]*NodeRef, len(refs))}
	copy(result.Touches, refs)
	return result
}

func (rs *NodeRefs) ToWhiches() []*Which {
	if rs == nil {
		return nil
	}
	whiches := make([]*Which, len(rs.Touches))
	for i, ref := range rs.Touches {
		whiches[i] = &Which{NodeRef: *ref}
	}
	return whiches
}

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
			typ = ws.Touches[i].TYPE
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
	ts := make(map[NodeRef]Which)
	//下标范围检测，并利用map的key值无重复特性去重
	for i := range ws.Touches {
		ind := ws.Touches[i].ID - 1
		v := *ws.Touches[i]
		switch v.TYPE {
		case SubscriptionType:
			if ind >= 0 && ind < GetLenSubscriptions() {
				ts[v.NodeRef] = v
			}
		case ServerType:
			if ind >= 0 && ind < GetLenServers() {
				ts[v.NodeRef] = v
			}
		case SubscriptionServerType:
			if v.Sub >= 0 && v.Sub < GetLenSubscriptions() && ind >= 0 && ind < GetLenSubscriptionServers(v.Sub) {
				ts[v.NodeRef] = v
			}
		}
	}
	//还原回slice
	w = make([]*Which, 0)
	for _, v := range ts {
		t := v
		w = append(w, &t)
	}
	return
}

type Which struct {
	NodeRef
	Latency string `json:"pingLatency,omitempty"` //历史遗留问题，前后端通信还是使用pingLatency这个名字，该值仅作为ping的返回值
	Link    string //optional
	// Selected marks, in a touch, the member the group routes through alone.
	Selected bool `json:"selected,omitempty"`
}

func (w Which) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		TYPE     TouchType `json:"_type"`
		ID       int       `json:"id"`
		Sub      int       `json:"sub"`
		Latency  string    `json:"pingLatency,omitempty"`
		Link     string    `json:"Link"`
		Outbound string    `json:"outbound"`
		Selected bool      `json:"selected,omitempty"`
	}{w.TYPE, w.ID, w.Sub, w.Latency, w.Link, w.Outbound, w.Selected})
}

func (w *NodeRef) EqualTo(another NodeRef) (ok bool) {
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
func (w *Which) Ping(loc *Locator, timeout time.Duration) (err error) {
	if w.TYPE == SubscriptionType {
		return fmt.Errorf("you cannot ping a subscription")
	}
	tsr, err := loc.Locate(&w.NodeRef)
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
	if e == nil {
		_ = conn.Close()
		w.Latency = fmt.Sprintf("%.0fms", time.Since(t).Seconds()*1000)
	} else {
		log.Debug("Ping: %v", e)
		w.Latency = "TIMEOUT"
	}
	return
}

// Locator resolves whiches against one read of each server list. Reading and
// parsing every stored node per lookup made a loop over n nodes cost n reads,
// which a latency test over a large subscription turned into seconds.
type Locator struct {
	serversOnce       sync.Once
	servers           []ServerRaw
	subscriptionsOnce sync.Once
	subscriptions     []SubscriptionRaw
}

// NewLocator reads each list on its first use; a Locator is safe for
// concurrent lookups.
func NewLocator() *Locator {
	return &Locator{}
}

// LocatorOf resolves against lists the caller has already read.
func LocatorOf(servers []ServerRaw, subscriptions []SubscriptionRaw) *Locator {
	l := &Locator{servers: servers, subscriptions: subscriptions}
	l.serversOnce.Do(func() {})
	l.subscriptionsOnce.Do(func() {})
	return l
}

func (l *Locator) Servers() []ServerRaw {
	l.serversOnce.Do(func() { l.servers = GetServers() })
	return l.servers
}

func (l *Locator) Subscriptions() []SubscriptionRaw {
	l.subscriptionsOnce.Do(func() { l.subscriptions = GetSubscriptions() })
	return l.subscriptions
}

func (l *Locator) Locate(w *NodeRef) (sr *ServerRaw, err error) {
	ind := w.ID - 1 //转化为下标
	switch w.TYPE {
	case ServerType:
		servers := l.Servers()
		if ind < 0 || ind >= len(servers) {
			return nil, common.Coded("SERVER_NOT_FOUND", fmt.Errorf("server #%d does not exist (there are %d servers); reload the page", w.ID, len(servers)), map[string]interface{}{
				"id":    w.ID,
				"count": len(servers),
			})
		}
		return &servers[ind], nil
	case SubscriptionServerType:
		subscriptions := l.Subscriptions()
		if w.Sub < 0 || w.Sub >= len(subscriptions) || ind < 0 || ind >= len(subscriptions[w.Sub].Servers) {
			return nil, common.Coded("SUBSCRIPTION_SERVER_NOT_FOUND", fmt.Errorf("server #%d of subscription #%d does not exist; reload the page", w.ID, w.Sub+1), map[string]interface{}{
				"id":  w.ID,
				"sub": w.Sub + 1,
			})
		}
		return &subscriptions[w.Sub].Servers[ind], nil
	default:
		return nil, common.Coded("UNKNOWN_ITEM_TYPE", fmt.Errorf("unknown item type %q; expected %q or %q", w.TYPE, ServerType, SubscriptionServerType), map[string]interface{}{"type": string(w.TYPE)})
	}
}

// LocateServerRaw reads the lists for this one lookup; loops use a Locator.
func (w *NodeRef) LocateServerRaw() (sr *ServerRaw, err error) {
	return NewLocator().Locate(w)
}

func (ws *Whiches) FillLinks() (err error) {
	loc := NewLocator()
	for _, w := range ws.Touches {
		sr, err := loc.Locate(&w.NodeRef)
		if err != nil {
			return err
		}
		w.Link = sr.ServerObj.ExportToURL()
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
	loc := NewLocator()
	for index, which := range serverIndexes {
		sRaw, err := loc.Locate(&which.NodeRef)
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
		subRaw := GetSubscription(subIndex)
		for index, which := range serverIndexes {
			subRaw.Servers[index].Latency = which.Latency
		}
		if err := SetSubscription(subIndex, subRaw); err != nil {
			return err
		}
	}
	return nil
}
