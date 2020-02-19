package configure

import (
	"V2RayA/global"
	"V2RayA/model/ipforward"
	"V2RayA/persistence"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Configure struct {
	Servers         []ServerRaw       `json:"servers"`
	Subscriptions   []SubscriptionRaw `json:"subscriptions"`
	ConnectedServer *Which            `json:"connectedServer"` //冗余一个信息，方便查找
	Setting         *Setting          `json:"setting"`
	Accounts        map[string]string `json:"accounts"`
	CustomPorts     Ports             `json:"ports"`
	PortWhiteList   PortWhiteList     `json:"portWhiteList"`
	DohList         string            `json:"dohlist"`
	CustomPac       CustomPac         `json:"customPac"`
}

func New() *Configure {
	return &Configure{
		Servers:         make([]ServerRaw, 0),
		Subscriptions:   make([]SubscriptionRaw, 0),
		ConnectedServer: nil,
		Setting:         NewSetting(),
		CustomPorts: Ports{
			Socks5:      20170,
			Http:        20171,
			HttpWithPac: 20172,
		},
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
	return persistence.Set(".", cfg)
}
func SetSubscriptions(subscriptions []SubscriptionRaw) (err error) {
	return persistence.Set("subscriptions", subscriptions)
}
func SetServers(servers []ServerRaw) (err error) {
	return persistence.Set("servers", servers)
}
func SetServer(index int, server *ServerRaw) (err error) {
	return persistence.Set(fmt.Sprintf("servers.%d", index), server)
}
func SetSubscription(index int, subscription *SubscriptionRaw) (err error) {
	return persistence.Set(fmt.Sprintf("subscriptions.%d", index), subscription)
}
func SetSetting(setting *Setting) (err error) {
	return persistence.Set("setting", setting)
}
func SetTransparent(transparent TransparentMode) (err error) {
	return persistence.Set("setting.transparent", transparent)
}
func SetPorts(ports *Ports) (err error) {
	return persistence.Set("ports", ports)
}
func SetPortWhiteList(portWhiteList *PortWhiteList) (err error) {
	return persistence.Set("portWhiteList", portWhiteList.Compressed())
}
func SetDohList(dohList *string) (err error) {
	return persistence.Set("dohList", strings.TrimSpace(*dohList))
}
func SetCustomPac(customPac *CustomPac) (err error) {
	return persistence.Set("customPac", customPac)
}

func AppendServer(server *ServerRaw) (err error) {
	return persistence.Append("servers", server)
}
func AppendSubscription(subscription *SubscriptionRaw) (err error) {
	return persistence.Append("subscriptions", subscription)
}

func IsConfigureNotExists() bool {
	f, err := os.OpenFile(global.GetEnvironmentConfig().Config, os.O_RDONLY, os.ModeAppend)
	if err != nil {
		return os.IsNotExist(err)
	}
	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(f)
	return !(err == nil && n > 0)
}
func GetServers() []ServerRaw {
	r := make([]ServerRaw, 0)
	_ = persistence.Get("servers", &r)
	return r
}
func GetSubscriptions() []SubscriptionRaw {
	r := make([]SubscriptionRaw, 0)
	_ = persistence.Get("subscriptions", &r)
	return r
}
func GetSubscription(id int) *SubscriptionRaw {
	s := new(SubscriptionRaw)
	err := persistence.Get(fmt.Sprintf("subscriptions.%d", id), &s)
	if err != nil {
		return nil
	}
	return s
}
func GetSettingNotNil() *Setting {
	r := new(Setting)
	_ = persistence.Get("setting", &r)
	r.IpForward = ipforward.IsIpForwardOn() //永远用真实值
	if r.AntiPollution == "" {
		r.AntiPollution = AntipollutionNone
	}
	return r
}
func GetPorts() *Ports {
	r := new(Ports)
	err := persistence.Get("ports", &r)
	if err != nil {
		return nil
	}
	return r
}
func GetPortWhiteListNotNil() *PortWhiteList {
	r := new(PortWhiteList)
	_ = persistence.Get("portWhiteList", &r)
	return r
}
func GetDohListNotNil() *string {
	r := new(string)
	_ = persistence.Get("dohList", &r)
	if len(strings.TrimSpace(*r)) == 0 {
		*r = `https://dns.rubyfish.cn/dns-query
https://i.233py.com/dns-query`
	}
	return r
}
func GetCustomPacNotNil() *CustomPac {
	r := new(CustomPac)
	_ = persistence.Get("customPac", &r)
	if r.DefaultProxyMode == "" {
		r = &CustomPac{
			DefaultProxyMode: DefaultProxyMode,
			RoutingRules:     []RoutingRule{},
		}
	}
	return r
}
func GetConnectedServer() *Which {
	r := new(Which)
	err := persistence.Get("connectedServer", &r)
	if err != nil {
		return nil
	}
	return r
}

func GetLenSubscriptions() int {
	l, err := persistence.GetArrayLen("subscriptions")
	if err != nil {
		panic(err)
	}
	return l
}
func GetLenSubscriptionServers(sub int) int {
	l, err := persistence.GetArrayLen(fmt.Sprintf("subscriptions.%d.servers", sub))
	if err != nil {
		panic(err)
	}
	return l
}
func GetLenServers() int {
	l, err := persistence.GetArrayLen("servers")
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
	return persistence.Set("connectedServer", wt)
}

func SetAccount(username, password string) (err error) {
	path := fmt.Sprintf("accounts.%s", username)
	return persistence.Set(path, password)
}
func ExistsAccount(username string) bool {
	return persistence.Exists(fmt.Sprintf("accounts.%s", username))
}

func GetPasswordOfAccount(username string) (pwd string, err error) {
	path := fmt.Sprintf("accounts.%s", username)
	if !persistence.Exists(path) {
		return "", errors.New("用户名不存在")
	}
	err = persistence.Get(path, &pwd)
	return
}

func HasAnyAccounts() bool {
	l, err := persistence.GetObjectLen("accounts")
	return err == nil && l > 0
}
