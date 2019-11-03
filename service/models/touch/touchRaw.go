package touch

import (
	"V2RayA/models/vmessInfo"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"sync"
	"time"
)

/*TouchRaw是配置文件存储形式*/
type TouchRaw struct {
	Servers         []TouchServerRaw  `json:"servers"`
	Subscriptions   []SubscriptionRaw `json:"subscriptions"`
	mutex           sync.Mutex        `json:"-"`
	ConnectedServer *WhichTouch       `json:"connectedServer"` //冗余一个信息，方便查找
	Setting         *Setting          `json:"setting"`
}

type Setting struct {
	PacMode                    PacMode                `json:"pacMode"`
	CustomPac                  CustomPac              `json:"customPac"`
	ProxyModeWhenSubscribe     ProxyModeWhenSubscribe `json:"proxyModeWhenSubscribe"`
	PacAutoUpdateMode          AutoUpdateMode         `json:"pacAutoUpdateMode"`
	PacAutoUpdateTime          int                    `json:"pacAutoUpdateTime"` //时间戳
	SubscriptionAutoUpdateMode AutoUpdateMode         `json:"subscriptionAutoUpdateMode"`
	SubscriptionAutoUpdateTime int                    `json:"subscriptionAutoUpdateTime"` //时间戳
}

func NewSetting() (setting *Setting) {
	return &Setting{
		PacMode: WhitelistMode,
		CustomPac: CustomPac{
			URL:              "",
			DefaultProxyMode: DefaultDirectMode,
			RoutingRules:     []RoutingRule{},
		},
		ProxyModeWhenSubscribe:     ProxyModeDirect,
		PacAutoUpdateMode:          DoNotUpdatePac,
		PacAutoUpdateTime:          int(21 * time.Hour / time.Millisecond), //凌晨5点
		SubscriptionAutoUpdateMode: DoNotUpdatePac,
		SubscriptionAutoUpdateTime: int(21 * time.Hour / time.Millisecond), //凌晨5点
	}

}

type CustomPac struct {
	URL              string                  `json:"url"`              //SiteDAT文件的URL
	DefaultProxyMode RoutingDefaultProxyMode `json:"defaultProxyMode"` //默认路由规则, proxy还是direct
	RoutingRules     []RoutingRule           `json:"routingRules"`
}

//v2rayTmpl.RoutingRule的前端友好版本
type RoutingRule struct {
	Tags      []string     `json:"tags"`      //SiteDAT文件的标签
	MatchType PacMatchType `json:"matchType"` //是domain匹配还是ip匹配
	RuleType  PacRuleType  `json:"ruleType"`  //在名单上的项进行直连、代理还是拦截
}

type TouchServerRaw struct {
	VmessInfo vmessInfo.VmessInfo `json:"vmessInfo"`
	Connected bool                `json:"connected,omitempty"`
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
		if v.VmessInfo.Protocol == "" {
			v.VmessInfo.Protocol = "vmess"
		}
		ts[i] = TouchServer{
			ID:        i + 1,
			Name:      v.VmessInfo.Ps,
			Address:   v.VmessInfo.Add + ":" + v.VmessInfo.Port,
			Net:       fmt.Sprintf("%v(%v)", v.VmessInfo.Protocol, v.VmessInfo.Net),
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

func (tr *TouchRaw) ReadFromFile(path string) (err error) {
	b, err := ioutil.ReadFile(path)
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
	case SubscriptionServerType:
		if wt.Sub < 0 || wt.Sub >= len(tr.Subscriptions) || ind < 0 || ind >= len(tr.Subscriptions[wt.Sub].Servers) {
			return nil, errors.New("ID或Sub超出下标范围")
		}
		return &tr.Subscriptions[wt.Sub].Servers[ind], nil
	default:
		return nil, errors.New("LocateServer: 无效的TYPE")
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
