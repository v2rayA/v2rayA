package touch

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"net"
	"net/url"
	"strconv"
	"time"
)

/*
Touch是树型结构的前后端通信形式，其结构设计和前端统一。
*/
type SubscriptionStatus string
type Touch struct {
	Servers          []Server           `json:"servers"`
	Subscriptions    []Subscription     `json:"subscriptions"`
	ConnectedServers []*configure.Which `json:"connectedServer"` //冗余一个信息，方便查找
}
type Server struct {
	ID          int                 `json:"id"`
	TYPE        configure.TouchType `json:"_type"`
	Name        string              `json:"name"`
	Address     string              `json:"address"`
	Net         string              `json:"net"`
	PingLatency string              `json:"pingLatency"`
}
type Subscription struct {
	Remarks    string              `json:"remarks,omitempty"`
	ID         int                 `json:"id"`
	TYPE       configure.TouchType `json:"_type"`
	Host       string              `json:"host"`
	Address    string              `json:"address"`
	Status     SubscriptionStatus  `json:"status"`
	Info       string              `json:"info"`
	Servers    []Server            `json:"servers"`
	AutoSelect bool                `json:"autoSelect"`
}

// NewUpdateStatus stamps a subscription update. RFC 3339 carries the zone, so
// a browser in another zone than the service (a container in UTC, a router)
// shows the moment, not one shifted by the difference.
func NewUpdateStatus() SubscriptionStatus {
	return SubscriptionStatus(time.Now().Format(time.RFC3339))
}

/* Mapping []TouchServerRaw to []Server */
func serverRawsToServers(rss []configure.ServerRaw) (ts []Server) {
	ts = make([]Server, len(rss))
	for i, v := range rss {
		var address string
		if v.ServerObj.GetPort() == 0 {
			address = v.ServerObj.GetHostname()
		} else {
			address = net.JoinHostPort(v.ServerObj.GetHostname(), strconv.Itoa(v.ServerObj.GetPort()))
		}
		ts[i] = Server{
			ID:          i + 1,
			Name:        v.ServerObj.GetName(),
			Address:     address,
			Net:         v.ServerObj.ProtoToShow(),
			PingLatency: v.Latency,
		}
	}
	return
}

// GenerateTouch generates a touch from database
func GenerateTouch() (t Touch) {
	servers := configure.GetServers()
	t.Servers = serverRawsToServers(servers)
	subscriptions := configure.GetSubscriptions()
	t.Subscriptions = make([]Subscription, len(subscriptions))
	for i, v := range subscriptions {
		u, err := url.Parse(v.Address)
		if err != nil {
			// it may is OOCv1
			tmp := make(map[string]string)
			_ = jsoniter.Unmarshal([]byte(v.Address), &tmp)
			if addr, ok := tmp["baseUrl"]; !ok {
				log.Warn("%v", err)
				continue
			} else {
				u, err = url.Parse(addr)
				if err != nil {
					log.Warn("%v", err)
					continue
				}
			}
		}
		t.Subscriptions[i] = Subscription{
			Remarks:    v.Remarks,
			ID:         i + 1,
			Host:       u.Host,
			Address:    v.Address,
			Status:     SubscriptionStatus(v.Status),
			Servers:    serverRawsToServers(v.Servers),
			Info:       v.Info,
			AutoSelect: v.AutoSelect,
		}
	}
	t.ConnectedServers = configure.GetConnectedServers().ToWhiches()
	markSelected(t.ConnectedServers, configure.LocatorOf(servers, subscriptions))
	//补充TYPE
	for i := range t.Subscriptions {
		t.Subscriptions[i].TYPE = configure.SubscriptionType
		for j := range t.Subscriptions[i].Servers {
			t.Subscriptions[i].Servers[j].TYPE = configure.SubscriptionServerType
		}
	}
	for i := range t.Servers {
		t.Servers[i].TYPE = configure.ServerType
	}
	return
}

// markSelected flags the member each group routes through alone, when its
// setting names one that is still a member.
func markSelected(connected []*configure.Which, loc *configure.Locator) {
	selected := make(map[string]string)
	for _, w := range connected {
		if _, ok := selected[w.Outbound]; !ok {
			selected[w.Outbound] = configure.GetOutboundSetting(w.Outbound).Selected
		}
		link := selected[w.Outbound]
		if link == "" {
			continue
		}
		sr, err := loc.Locate(&w.NodeRef)
		if err != nil || sr.ServerObj == nil {
			continue
		}
		w.Selected = sr.ServerObj.ExportToURL() == link
	}
}
