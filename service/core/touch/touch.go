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
	Remarks string              `json:"remarks,omitempty"`
	ID      int                 `json:"id"`
	TYPE    configure.TouchType `json:"_type"`
	Host    string              `json:"host"`
	Status  SubscriptionStatus  `json:"status"`
	Info    string              `json:"info"`
	Servers []Server            `json:"servers"`
}

func NewUpdateStatus() SubscriptionStatus {
	return SubscriptionStatus(time.Now().Local().Format("2006-1-2 15:04:05"))
}

/* Mapping []TouchServerRaw to []Server */
func serverRawsToServers(rss []configure.ServerRawV2) (ts []Server) {
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
	t.Servers = serverRawsToServers(configure.GetServersV2())
	subscriptions := configure.GetSubscriptionsV2()
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
			Remarks: v.Remarks,
			ID:      i + 1,
			Host:    u.Host,
			Status:  SubscriptionStatus(v.Status),
			Servers: serverRawsToServers(v.Servers),
			Info:    v.Info,
		}
	}
	t.ConnectedServers = configure.GetConnectedServers().Get()
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
