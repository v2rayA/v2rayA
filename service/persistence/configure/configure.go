package configure

import (
	"V2RayA/persistence"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type Configure struct {
	Servers         []TouchServerRaw  `json:"servers"`
	Subscriptions   []SubscriptionRaw `json:"subscriptions"`
	ConnectedServer *Which            `json:"connectedServer"` //冗余一个信息，方便查找
	Setting         *Setting          `json:"setting"`
}

func New() *Configure {
	return &Configure{
		Servers:         make([]TouchServerRaw, 0),
		Subscriptions:   make([]SubscriptionRaw, 0),
		ConnectedServer: nil,
		Setting:         NewSetting(),
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
func setSomething(path string, val interface{}) (err error) {
	b, _ := json.Marshal(val)
	_, err = persistence.DoAndSave("json.set", "V2RayA.configure", path, b)
	return
}
func SetConfigure(cfg *Configure) error {
	return setSomething(".", cfg)
}
func SetSubscription(index int, subscription *SubscriptionRaw) (err error) {
	return setSomething(fmt.Sprintf("subscriptions[%d]", index), subscription)
}
func SetSetting(setting *Setting) (err error) {
	return setSomething("setting", setting)
}

func appendSomething(path string, val interface{}) (err error) {
	b, _ := json.Marshal(val)
	_, err = persistence.DoAndSave("json.arrappend", "V2RayA.configure", path, b)
	return
}
func AppendServer(server *TouchServerRaw) (err error) {
	return appendSomething("servers", server)
}
func AppendSubscription(subscription *SubscriptionRaw) (err error) {
	return appendSomething("subscriptions", subscription)
}

func getSomething(path string, v interface{}) {
	re, err := persistence.Do("json.get", "V2RayA.configure", path)
	if err != nil || re == nil {
		v = nil
		return
	}
	_ = json.Unmarshal(decode(re.([]byte)), v)
	return
}
func IsConfigureExists() bool {
	var v interface{}
	getSomething(".", &v)
	switch v.(type) {
	case nil:
		return false
	}
	return true
}
func GetServers() []TouchServerRaw {
	r := make([]TouchServerRaw, 0)
	getSomething("servers", &r)
	return r
}
func GetSubscriptions() []SubscriptionRaw {
	r := make([]SubscriptionRaw, 0)
	getSomething("subscriptions", &r)
	return r
}
func GetSetting() *Setting {
	r := new(Setting)
	getSomething("setting", &r)
	return r
}
func GetConnectedServer() *Which {
	r := new(Which)
	getSomething("connectedServer", &r)
	return r
}

func getLenSomething(path string) int {
	re, err := persistence.Do("json.arrlen", "V2RayA.configure", path)
	if err != nil {
		return 0
	}
	return int(re.(int64))
}
func GetLenSubscriptions() int {
	return getLenSomething("subscriptions")
}
func GetLenSubscriptionServers(sub int) int {
	return getLenSomething(fmt.Sprintf("subscriptions[%d].servers", sub))
}
func GetLenServers() int {
	return getLenSomething("servers")
}

func removeSomethingFromArray(path string, index int) error {
	_, err := persistence.DoAndSave("json.arrpop", "V2RayA.configure", path, index)
	if err != nil {
		return err
	}
	return nil
}
func RemoveSubscription(index int) error {
	err := removeSomethingFromArray("subscriptions", index)
	return err
}
func RemoveServer(index int) error {
	err := removeSomethingFromArray("servers", index)
	return err
}

/*不会停止v2ray.service*/
func ClearConnected() error {
	return SetConnect(nil)
}

/*不会启动v2ray.service*/
func SetConnect(wt *Which) (err error) {
	return setSomething("connectedServer", wt)
}
