package models

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"sync"
)

/*TouchRaw是配置文件存储形式*/
type TouchRaw struct {
	Servers         []TouchServerRaw  `json:"servers"`
	Subscriptions   []SubscriptionRaw `json:"subscriptions"`
	mutex           sync.Mutex        `json:"-"`
	ConnectedServer *WhichTouch       `json:"connectedServer"` //冗余一个信息，方便查找
}

type TouchServerRaw struct {
	VmessInfo VmessInfo `json:"vmessInfo"`
	Connected bool      `json:"connected,omitempty"`
}

type SubscriptionRaw struct {
	Address string             `json:"address"`
	Status  SubscriptionStatus `json:"status"` //update time, error info, etc.
	Servers []TouchServerRaw   `json:"servers"`
}

/* 将[]TouchServerRaw映射到[]TouchServer */
func serverRawsToServers(rss []TouchServerRaw) (ts []TouchServer) {
	ts = make([]TouchServer, len(rss))
	for i, v := range rss {
		ts[i] = TouchServer{
			ID:        i + 1,
			Name:      v.VmessInfo.Ps,
			Address:   v.VmessInfo.Add + ":" + v.VmessInfo.Port,
			Net:       v.VmessInfo.Net,
			Connected: v.Connected,
		}
	}
	return
}

/* 根据TouchRaw创建一个Touch */
func (tr *TouchRaw) ToTouch() (t Touch) {
	t.Servers = serverRawsToServers(tr.Servers)
	t.Subscriptions = make([]Subscription, len(tr.Subscriptions))
	for i, v := range tr.Subscriptions {
		u, err := url.Parse(v.Address)
		if err != nil {
			continue
		}
		t.Subscriptions[i] = Subscription{
			ID:      i + 1,
			Host:    u.Host,
			Status:  v.Status,
			Servers: serverRawsToServers(v.Servers),
		}
	}
	t.ConnectedServer = tr.ConnectedServer
	return
}

func (tr *TouchRaw) ReadFromFile() (err error) {
	b, err := ioutil.ReadFile(".tr")
	if err != nil {
		return
	}
	return json.Unmarshal(b, tr)
}

func (tr *TouchRaw) WriteToFile() (err error) {
	if tr == nil {
		return
	}
	b, _ := json.Marshal(*tr)
	return ioutil.WriteFile(".tr", b, os.FileMode(0600))
}

func (tr *TouchRaw) Lock() {
	tr.mutex.Lock()
}

func (tr *TouchRaw) Unlock() {
	tr.mutex.Unlock()
}

func (tr *TouchRaw) LocateServer(wt *WhichTouch) (*TouchServerRaw, error) {
	if wt == nil {
		return nil, errors.New("参数为nil")
	}
	ind := wt.ID - 1 //转化为下标
	switch wt.TYPE {
	case ServerType:
		if ind < 0 || ind >= len(tr.Servers) {
			return nil, errors.New("ID超出下标范围")
		}
		return &tr.Servers[ind], nil
	case SubscriptionType:
		if wt.Sub < 0 || wt.Sub >= len(tr.Subscriptions) || ind < 0 || ind >= len(tr.Subscriptions[wt.Sub].Servers) {
			return nil, errors.New("ID或Sub超出下标范围")
		}
		return &tr.Subscriptions[wt.Sub].Servers[ind], nil
	default:
		return nil, errors.New("无效的TYPE")
	}
}

/*既不停止v2ray服务，也不进行写配置文件的操作*/
func (tr *TouchRaw) SetDisConnect() {
	if tr.ConnectedServer == nil {
		return
	}
	tsr, _ := tr.LocateServer(tr.ConnectedServer)
	tsr.Connected = false
	tr.ConnectedServer = nil
}

/*既不启动v2ray服务，也不进行写配置文件的操作*/
func (tr *TouchRaw) SetConnect(wt *WhichTouch) (err error) {
	tsr, err := tr.LocateServer(wt)
	if err != nil {
		return
	}
	tsr.Connected = true
	tr.ConnectedServer = wt
	return
}
