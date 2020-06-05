package configure

import (
	"bytes"
	"encoding/hex"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/xujiajun/nutsdb"
	"log"
	"strings"
	"v2ray.com/core/common/errors"
	"v2rayA/core/ipforward"
	"v2rayA/db"
)

type Configure struct {
	Servers         []*ServerRaw       `json:"servers"`
	Subscriptions   []*SubscriptionRaw `json:"subscriptions"`
	ConnectedServer *Which             `json:"connectedServer"` //冗余一个信息，方便查找
	Setting         *Setting           `json:"setting"`
	Accounts        map[string]string  `json:"accounts"`
	Ports           Ports              `json:"ports"`
	PortWhiteList   PortWhiteList      `json:"portWhiteList"`
	DohList         string             `json:"dohlist"`
	CustomPac       CustomPac          `json:"customPac"`
	RoutingA        *string            `json:"routingA"`
}

func New() *Configure {
	return &Configure{
		Servers:         make([]*ServerRaw, 0),
		Subscriptions:   make([]*SubscriptionRaw, 0),
		ConnectedServer: nil,
		Setting:         NewSetting(),
		Ports: Ports{
			Socks5:      20170,
			Http:        20171,
			HttpWithPac: 20172,
		},
		Accounts: map[string]string{},
	}
}
func decode(b []byte) (result []byte) {
	arr := bytes.Split(b, []byte(`\u00`))
	b2 := make([]byte, 0)
	result = arr[0]
	for i := 1; i <= len(arr); i++ {
		if i < len(arr) {
			b2 = append(b2, arr[i][0], arr[i][1])
		}
		if i == len(arr) || len(arr[i]) > 2 {
			dst := make([]byte, hex.DecodedLen(len(b2)))
			_, _ = hex.Decode(dst, b2)
			b2 = b2[:0]
			result = append(result, dst...)
			if i < len(arr) {
				result = append(result, arr[i][2:]...)
			}
		}
	}
	return
}

func SetConfigure(cfg *Configure) error {
	if err := db.BucketClear("system"); errors.Cause(err) != nutsdb.ErrBucketEmpty {
		return err
	}
	if err := db.BucketClear("touch"); errors.Cause(err) != nutsdb.ErrBucketEmpty {
		return err
	}
	if err := db.BucketClear("accounts"); errors.Cause(err) != nutsdb.ErrBucketEmpty {
		return err
	}
	for username, password := range cfg.Accounts {
		if err := SetAccount(username, password); err != nil {
			return err
		}
	}
	if err := AppendServers(cfg.Servers); err != nil {
		return err
	}
	if err := AppendSubscriptions(cfg.Subscriptions); err != nil {
		return err
	}
	if err := SetRoutingA(cfg.RoutingA); err != nil {
		return err
	}
	if err := SetDohList(&cfg.DohList); err != nil {
		return err
	}
	if err := SetCustomPac(&cfg.CustomPac); err != nil {
		return err
	}
	if err := SetPortWhiteList(&cfg.PortWhiteList); err != nil {
		return err
	}
	if err := SetConnect(cfg.ConnectedServer); err != nil {
		return err
	}
	if err := SetSetting(cfg.Setting); err != nil {
		return err
	}
	if err := SetPorts(&cfg.Ports); err != nil {
		return err
	}
	return nil
}

func RemoveSubscriptions(indexes []int) (err error) {
	//TODO: separate the SubscriptionRaw and []ServerRaw
	return db.ListRemove("touch", "subscriptions", indexes)
}

func RemoveServers(indexes []int) (err error) {
	//TODO: separate the SubscriptionRaw and []ServerRaw
	return db.ListRemove("touch", "servers", indexes)
}
func SetServer(index int, server *ServerRaw) (err error) {
	return db.ListSet("touch", "servers", index, server)
}
func SetSubscription(index int, subscription *SubscriptionRaw) (err error) {
	return db.ListSet("touch", "subscriptions", index, subscription)
}
func SetSetting(setting *Setting) (err error) {
	return db.Set("system", "setting", setting)
}
func SetPorts(ports *Ports) (err error) {
	return db.Set("system", "ports", ports)
}
func SetPortWhiteList(portWhiteList *PortWhiteList) (err error) {
	return db.Set("system", "portWhiteList", portWhiteList.Compressed())
}
func SetDohList(dohList *string) (err error) {
	return db.Set("system", "dohList", strings.TrimSpace(*dohList))
}
func SetCustomPac(customPac *CustomPac) (err error) {
	return db.Set("system", "customPac", customPac)
}
func SetRoutingA(routingA *string) (err error) {
	return db.Set("system", "routingA", routingA)
}

func AppendServers(server []*ServerRaw) (err error) {
	return db.ListAppend("touch", "servers", server)
}
func AppendSubscriptions(subscription []*SubscriptionRaw) (err error) {
	return db.ListAppend("touch", "subscriptions", subscription)
}

func IsConfigureNotExists() bool {
	l, err := db.GetBucketLen("system")
	return err != nil || l == 0
}
func GetServers() []ServerRaw {
	r := make([]ServerRaw, 0)
	raw, err := db.ListGetAll("touch", "servers")
	if err == nil {
		for _, b := range raw {
			var t ServerRaw
			e := jsoniter.Unmarshal(b, &t)
			if e != nil {
				continue
			}
			r = append(r, t)
		}
	}
	return r
}
func GetSubscriptions() []SubscriptionRaw {
	r := make([]SubscriptionRaw, 0)
	raw, err := db.ListGetAll("touch", "subscriptions")
	if err == nil {
		for _, b := range raw {
			var t SubscriptionRaw
			e := jsoniter.Unmarshal(b, &t)
			if e != nil {
				continue
			}
			r = append(r, t)
		}
	}
	return r
}
func GetSubscription(index int) *SubscriptionRaw {
	s := new(SubscriptionRaw)
	err := db.ListGet("touch", "subscriptions", index, &s)
	if err != nil {
		return nil
	}
	return s
}
func GetSettingNotNil() *Setting {
	r := new(Setting)
	_ = db.Get("system", "setting", &r)
	r.IpForward = ipforward.IsIpForwardOn() //永远用真实值
	if r.AntiPollution == "" {
		r.AntiPollution = AntipollutionNone
	}
	return r
}
func GetPorts() *Ports {
	r := new(Ports)
	err := db.Get("system", "ports", &r)
	if err != nil {
		return nil
	}
	return r
}
func GetPortWhiteListNotNil() *PortWhiteList {
	r := new(PortWhiteList)
	_ = db.Get("system", "portWhiteList", &r)
	return r
}
func GetDohListNotNil() *string {
	r := new(string)
	_ = db.Get("system", "dohList", &r)
	if len(strings.TrimSpace(*r)) == 0 {
		*r = `https://doh.opendns.com/dns-query
https://dns.alidns.com/dns-query`
	}
	return r
}
func GetCustomPacNotNil() *CustomPac {
	r := new(CustomPac)
	_ = db.Get("system", "customPac", &r)
	if r.DefaultProxyMode == "" {
		r = &CustomPac{
			DefaultProxyMode: DefaultProxyMode,
			RoutingRules:     []RoutingRule{},
		}
	}
	return r
}
func GetRoutingA() (r string) {
	_ = db.Get("system", "routingA", &r)
	return
}
func GetConnectedServer() *Which {
	r := new(Which)
	err := db.Get("touch", "connectedServer", &r)
	if err != nil {
		return nil
	}
	return r
}

func GetLenSubscriptions() int {
	l, err := db.ListLen("touch", "subscriptions")
	if err != nil {
		panic(err)
	}
	return l
}
func GetLenSubscriptionServers(index int) int {
	b, err := db.GetRaw("touch", "subscriptions")
	if err != nil {
		log.Fatal(err)
	}
	return len(gjson.GetBytes(b, "servers").Array())
}
func GetLenServers() int {
	l, err := db.ListLen("touch", "servers")
	if err != nil {
		panic(err)
	}
	return l
}

/*不会停止v2ray.service*/
func ClearConnected() error {
	return SetConnect(nil)
}

/*不会启动v2ray.service*/
func SetConnect(wt *Which) (err error) {
	return db.Set("touch", "connectedServer", wt)
}

func SetAccount(username, password string) (err error) {
	return db.Set("accounts", username, password)
}
func ResetAccounts() (err error) {
	return db.BucketClear("accounts")
}
func ExistsAccount(username string) bool {
	return db.Exists("accounts", username)
}

func GetPasswordOfAccount(username string) (pwd string, err error) {
	err = db.Get("accounts", username, &pwd)
	return
}

func HasAnyAccounts() bool {
	l, err := db.GetBucketLen("accounts")
	return err == nil && l > 0
}
