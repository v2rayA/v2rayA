package configure

import (
	"bytes"
	"encoding/hex"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/v2rayA/v2rayA/db"
	"log"
	"strings"
)

type Configure struct {
	Servers          []*ServerRaw       `json:"servers"`
	Subscriptions    []*SubscriptionRaw `json:"subscriptions"`
	ConnectedServers []*Which           `json:"connectedServers"`
	Setting          *Setting           `json:"setting"`
	Accounts         map[string]string  `json:"accounts"`
	Ports            Ports              `json:"ports"`
	PortWhiteList    PortWhiteList      `json:"portWhiteList"`
	InternalDnsList  *string            `json:"internalDnsList"`
	ExternalDnsList  *string            `json:"externalDnsList"`
	RoutingA         *string            `json:"routingA"`
}

func New() *Configure {
	return &Configure{
		Servers:          make([]*ServerRaw, 0),
		Subscriptions:    make([]*SubscriptionRaw, 0),
		ConnectedServers: make([]*Which, 0),
		Setting:          NewSetting(),
		Accounts:         map[string]string{},
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
	if err := db.BucketClear("system"); err != nil {
		return err
	}
	if err := db.BucketClear("touch"); err != nil {
		return err
	}
	if err := db.BucketClear("accounts"); err != nil {
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
	if err := OverwriteConnects(NewWhiches(cfg.ConnectedServers)); err != nil {
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
func SetInternalDnsList(dnsList *string) (err error) {
	if dnsList == nil {
		return db.Set("system", "internalDnsList", nil)
	}
	return db.Set("system", "internalDnsList", strings.TrimSpace(*dnsList))
}
func SetExternalDnsList(dnsList *string) (err error) {
	if dnsList == nil {
		return db.Set("system", "externalDnsList", nil)
	}
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
func GetConnectedServers() (wts *Whiches) {
	outbounds := GetOutbounds()
	for _, outbound := range outbounds {
		w := GetConnectedServersByOutbound(outbound)
		if w != nil {
			if wts == nil {
				wts = new(Whiches)
			}
			wts.Extend(*w)
		}
	}
	if wts != nil && wts.Len() == 0 {
		wts = nil
	}
	return wts
}
func GetConnectedServersByOutbound(outbound string) *Whiches {
	r := new(Whiches)
	if outbound == "" {
		outbound = "proxy"
	}
	bucket := fmt.Sprintf("outbound.%v", outbound)
	if err := db.Get(bucket, "connectedServers", &r); err != nil {
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
func ClearConnects(outbound string) error {
	if outbound == "" {
		outbound = "proxy"
	}
	return db.Set(fmt.Sprintf("outbound.%v", outbound), "connectedServers", nil)
}
func AddConnect(wt Which) (err error) {
	if wt.Outbound == "" {
		wt.Outbound = "proxy"
	}
	bucket := fmt.Sprintf("outbound.%v", wt.Outbound)
	var wcs Whiches
	_ = db.Get(bucket, "connectedServers", &wcs)
	for _, v := range wcs.Get() {
		if v.EqualTo(wt) {
			return nil
		}
	}
	wcs.Add(wt)
	return db.Set(bucket, "connectedServers", wcs)
}

// OverwriteConnects will replace each outbounds contained in given ws with whiches in the ws
func OverwriteConnects(ws *Whiches) (err error) {
	outWs := make(map[string][]*Which)
	for _, w := range ws.Get() {
		outWs[w.Outbound] = append(outWs[w.Outbound], w)
	}
	for out, ws := range outWs {
		whiches := new(Whiches)
		whiches.Touches = ws
		bucket := fmt.Sprintf("outbound.%v", out)
		if err := db.Set(bucket, "connectedServers", whiches); err != nil {
			return err
		}
	}
	return nil
}

func RemoveConnect(wt Which) (err error) {
	if wt.Outbound == "" {
		wt.Outbound = "proxy"
	}
	bucket := fmt.Sprintf("outbound.%v", wt.Outbound)
	var wcs Whiches
	_ = db.Get(bucket, "connectedServers", &wcs)
	for i, v := range wcs.Touches {
		if v.EqualTo(wt) {
			wcs.Touches = append(wcs.Touches[:i], wcs.Touches[i+1:]...)
			return db.Set(bucket, "connectedServers", wcs)
		}
	}
	return fmt.Errorf("given server cannot be found in database")
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

func SetRunning(running bool) (err error) {
	return db.Set("system", "running", running)
}
func GetRunning() (running bool) {
	_ = db.Get("system", "running", &running)
	return running
}
