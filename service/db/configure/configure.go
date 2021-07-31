package configure

import (
	"bytes"
	"encoding/hex"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/v2rayA/v2rayA/common/errors"
	"github.com/v2rayA/v2rayA/db"
	"github.com/xujiajun/nutsdb"
	"log"
	"strings"
)

type Configure struct {
	Servers                    []*ServerRaw       `json:"servers"`
	Subscriptions              []*SubscriptionRaw `json:"subscriptions"`
	ConnectedServer            *Which             `json:"connectedServer"` //冗余一个信息，方便查找
	AdditionalConnectedServers []Which            `json:"additionalConnectedServers"`
	Setting                    *Setting           `json:"setting"`
	Accounts                   map[string]string  `json:"accounts"`
	Ports                      Ports              `json:"ports"`
	PortWhiteList              PortWhiteList      `json:"portWhiteList"`
	InternalDnsList            *string            `json:"internalDnsList"`
	ExternalDnsList            *string            `json:"externalDnsList"`
	RoutingA                   *string            `json:"routingA"`
}

func New() *Configure {
	return &Configure{
		Servers:                    make([]*ServerRaw, 0),
		Subscriptions:              make([]*SubscriptionRaw, 0),
		ConnectedServer:            nil,
		AdditionalConnectedServers: nil,
		Setting:                    NewSetting(),
		Accounts:                   map[string]string{},
		Ports: Ports{
			Socks5:      20170,
			Http:        20171,
			HttpWithPac: 20172,
		},
		PortWhiteList:   PortWhiteList{},
		InternalDnsList: nil,
		ExternalDnsList: nil,
		RoutingA:        nil,
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
	if err := SetInternalDnsList(cfg.InternalDnsList); err != nil {
		return err
	}
	if err := SetExternalDnsList(cfg.ExternalDnsList); err != nil {
		return err
	}
	if err := SetPortWhiteList(&cfg.PortWhiteList); err != nil {
		return err
	}
	if err := SetConnect(cfg.ConnectedServer); err != nil {
		return err
	}
	for _, w := range cfg.AdditionalConnectedServers {
		if err := SetConnect(&w); err != nil {
			return err
		}
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
func SetInternalDnsList(dnsList *string) (err error) {
	return db.Set("system", "internalDnsList", strings.TrimSpace(*dnsList))
}
func SetExternalDnsList(dnsList *string) (err error) {
	return db.Set("system", "externalDnsList", strings.TrimSpace(*dnsList))
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
	if r == nil {
		r = NewSetting()
		_ = db.Set("system", "setting", r)
	}
	if r.SpecialMode == "" {
		r.SpecialMode = SpecialModeNone
	}
	if r.TransparentType == "" {
		r.TransparentType = TransparentRedirect
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
func GetExternalDnsListNotNil() (list []string) {
	r := new(string)
	_ = db.Get("system", "externalDnsList", &r)
	list = strings.Split(strings.TrimSpace(*r), "\n")
	if len(list) == 1 && list[0] == "" {
		return []string{}
	}
	return
}
func GetInternalDnsListNotNil() (list []string) {
	r := new(string)
	_ = db.Get("system", "internalDnsList", &r)
	if len(strings.TrimSpace(*r)) == 0 {
		*r = `119.29.29.29 -> direct
https://doh.alidns.com/dns-query -> direct
tcp://dns.opendns.com:5353 -> proxy`
	}
	list = strings.Split(strings.TrimSpace(*r), "\n")
	if len(list) == 1 && list[0] == "" {
		return []string{}
	}
	return
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
	if r == "" {
		return RoutingATemplate
	}
	return
}
func GetConnectedServers() (wts []*Which) {
	outbounds := GetOutbounds()
	for _, outbound := range outbounds {
		w := GetConnectedServer(outbound)
		if w != nil {
			wts = append(wts, w)
		}
	}
	if len(wts) == 0 {
		wts = nil
	}
	return wts
}
func GetConnectedServer(outbound string) *Which {
	r := new(Which)
	if outbound == "" || outbound == "proxy" {
		if err := db.Get("touch", "connectedServer", &r); err != nil {
			return nil
		}
		if r != nil {
			r.Outbound = "proxy"
		}
		return r
	} else {
		if err := db.Get(fmt.Sprintf("outbound.%v", outbound), "connectedServer", &r); err != nil {
			return nil
		}
		return r
	}
}

func GetLenSubscriptions() int {
	l, err := db.ListLen("touch", "subscriptions")
	if err != nil {
		panic(err)
	}
	return l
}
func GetLenSubscriptionServers(index int) int {
	b, err := db.ListGetRaw("touch", "subscriptions", index)
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

func ClearConnected(outbound string) error {
	if outbound == "" || outbound == "proxy" {
		return db.Set("touch", "connectedServer", nil)
	} else {
		return db.Set(fmt.Sprintf("outbound.%v", outbound), "connectedServer", nil)
	}
}

func SetConnect(wt *Which) (err error) {
	if wt == nil || wt.Outbound == "" || wt.Outbound == "proxy" {
		return db.Set("touch", "connectedServer", wt)
	} else {
		return db.Set(fmt.Sprintf("outbound.%v", wt.Outbound), "connectedServer", wt)
	}
}

func GetOutbounds() (outbounds []string) {
	outbounds = append(outbounds, "proxy")
	members, _ := db.StringSetGetAll("outbounds", "names")
	for _, m := range members {
		outbounds = append(outbounds, m)
	}
	return
}

func AddOutbound(outbound string) (err error) {
	if outbound == "proxy" ||
		outbound == "direct" ||
		outbound == "block" {
		return fmt.Errorf("cannot add %v as the outbound name", outbound)
	}
	return db.SetAdd("outbounds", "names", outbound)
}

func RemoveOutbound(outbound string) (err error) {
	if err = db.BucketClear(fmt.Sprintf("outbound.%v", outbound)); err != nil {
		return err
	}
	return db.SetRemove("outbounds", "names", outbound)
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
