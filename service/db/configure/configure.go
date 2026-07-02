package configure

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// isJSONFieldExists 检查原始 JSON 中是否存在指定字段。
// 用于在迁移场景中判断字段是否由旧配置显式设置。
func isJSONFieldExists(raw []byte, field string) bool {
	return gjson.GetBytes(raw, field).Exists()
}

type Configure struct {
	Servers             []*ServerRaw        `json:"servers"`
	Subscriptions       []*SubscriptionRaw  `json:"subscriptions"`
	ConnectedServers    []*Which            `json:"connectedServers"`
	Setting             *Setting            `json:"setting"`
	Accounts            map[string]string   `json:"accounts"`
	Ports               Ports               `json:"ports"`
	DnsRules            []DnsRule           `json:"dnsRules"`
	RoutingA            *string             `json:"routingA"`
	DomainsExcluded     *string             `json:"domainsExcluded"`
	TproxyWhiteIpGroups TproxyWhiteIpGroups `json:"tproxyWhiteIpGroups"`
}

func New() *Configure {
	return &Configure{
		Servers:          make([]*ServerRaw, 0),
		Subscriptions:    make([]*SubscriptionRaw, 0),
		ConnectedServers: make([]*Which, 0),
		Setting:          NewSetting(),
		Accounts:         map[string]string{},
		Ports: Ports{
			Socks5:        20170,
			Socks5WithPac: 0,
			Http:          20171,
			HttpWithPac:   20172,
			Vmess:         0,
		},
		RoutingA:        nil,
		DomainsExcluded: nil,
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
	if cfg.DnsRules != nil {
		if err := SetDnsRules(cfg.DnsRules); err != nil {
			return err
		}
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
	return db.ListRemove("touch", "subscriptions", indexes)
}

func RemoveServers(indexes []int) (err error) {
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
			t, e := Bytes2ServerRaw(b)
			if e != nil {
				log.Warn("GetServers: %v", e)
				continue
			}
			r = append(r, *t)
		}
	}
	return r
}

func GetSubscriptions() []SubscriptionRaw {
	r := make([]SubscriptionRaw, 0)
	raw, err := db.ListGetAll("touch", "subscriptions")
	if err == nil {
		for _, b := range raw {
			t, e := Bytes2SubscriptionRaw(b)
			if e != nil {
				log.Warn("%v", e)
				continue
			}
			r = append(r, *t)
		}
	}
	return r
}
func GetSubscription(index int) *SubscriptionRaw {
	b, err := db.ListGet("touch", "subscriptions", index)
	if err != nil {
		return nil
	}
	s, err := Bytes2SubscriptionRaw(b)
	if err != nil {
		log.Warn("%v", err)
		return nil
	}
	return s
}

func GetSettingNotNil() *Setting {
	r := new(Setting)
	b, e := db.GetRaw("system", "setting")
	if e == nil {
		_ = jsoniter.Unmarshal(b, r)
	}
	_ = common.FillEmpty(r, NewSetting())
	// Restore TproxyExcludedInterfaces from DB if user explicitly cleared it
	// FillEmpty replaces empty strings with defaults, but we need to preserve
	// user's intent to clear this field
	if e == nil && b != nil {
		var raw Setting
		if err := jsoniter.Unmarshal(b, &raw); err == nil {
			// If user explicitly saved an empty string, keep it empty
			r.TproxyExcludedInterfaces = raw.TproxyExcludedInterfaces
		}
	}
	if r.TransparentType == "" {
		r.TransparentType = TransparentRedirect
	}
	// 执行新 DNS 模块配置迁移（处理旧配置升级场景）
	MigrateSetting(r)
	// 处理 DNS 缓存布尔字段默认值：common.FillEmpty 跳过布尔字段，
	// 因此如果旧配置中缺少这些字段，它们会保持 false。
	if e == nil && b != nil {
		hasDNSListenAddr := isJSONFieldExists(b, "dnsListenAddr")
		hasDNSCacheEnabled := isJSONFieldExists(b, "dnsCacheEnabled")
		hasDNSPrefetch := isJSONFieldExists(b, "dnsPrefetch")
		hasDNSNegativeCache := isJSONFieldExists(b, "dnsNegativeCache")

		if !hasDNSListenAddr && !hasDNSCacheEnabled && !hasDNSPrefetch && !hasDNSNegativeCache {
			r.DnsCacheEnabled = true
			r.DnsPrefetch = true
			r.DnsNegativeCache = true
		}
	} else if e != nil || b == nil {
		r.DnsCacheEnabled = true
		r.DnsPrefetch = true
		r.DnsNegativeCache = true
	}
	return r
}
func GetPortsNotNil() *Ports {
	p := new(Ports)
	_ = db.Get("system", "ports", &p)
	if p == nil {
		p = new(Ports)
		p.Socks5 = 20170
		p.Http = 20171
		p.Socks5WithPac = 0
		p.HttpWithPac = 20172
		p.Vmess = 0
		p.Api = ApiPort{Port: 0}
	}
	return p
}
func GetCustomPacNotNil() *CustomPac {
	r := new(CustomPac)
	_ = db.Get("system", "customPac", r)
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
func SetDomainsExcluded(domains string) (err error) {
	return db.Set("system", "domainsExcluded", domains)
}
func SetTproxyWhiteIpGroups(countryCodes []string, customIps []string) (err error) {
	return db.Set("system", "tproxyWhiteIpGroups", TproxyWhiteIpGroups{
		CountryCodes: countryCodes,
		CustomIps:    customIps,
	})
}
func GetDomainsExcluded() (r string) {
	db.Get("system", "domainsExcluded", &r)
	return r
}
func GetTproxyWhiteIpGroups() (r TproxyWhiteIpGroups) {
	db.Get("system", "tproxyWhiteIpGroups", &r)
	return r
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
	if err := db.Get(bucket, "connectedServers", r); err != nil {
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
	b, err := db.ListGet("touch", "subscriptions", index)
	if err != nil {
		log.Fatal("GetLenSubscriptionServers: %v", err)
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
	// Normalize Outbound field of existing entries for consistent comparison
	for _, v := range wcs.Get() {
		if v.Outbound == "" {
			v.Outbound = "proxy"
		}
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
	// Normalize Outbound field of existing entries for consistent comparison
	for _, v := range wcs.Touches {
		if v.Outbound == "" {
			v.Outbound = "proxy"
		}
	}
	for i, v := range wcs.Touches {
		if v.EqualTo(wt) {
			wcs.Touches = append(wcs.Touches[:i], wcs.Touches[i+1:]...)
			return db.Set(bucket, "connectedServers", wcs)
		}
	}
	return fmt.Errorf("given server cannot be found in database")
}

func GetOutbounds() (outbounds []string) {
	// keep order
	members, _ := db.StringSetGetAll("outbounds", "names")
	for _, m := range members {
		if m != "proxy" {
			outbounds = append(outbounds, m)
		}
	}
	sort.Strings(outbounds)
	// "proxy" is always the first outbound.
	outbounds = append([]string{"proxy"}, outbounds...)
	return
}

// InitDefaultOutbound ensures the default "proxy" outbound group exists in the database.
// It should be called during system initialization.
func InitDefaultOutbound() error {
	members, err := db.StringSetGetAll("outbounds", "names")
	if err != nil {
		// bucket doesn't exist yet, create it with the default
		return db.SetAdd("outbounds", "names", DefaultOutboundName)
	}
	for _, m := range members {
		if m == DefaultOutboundName {
			return nil // already exists
		}
	}
	return db.SetAdd("outbounds", "names", DefaultOutboundName)
}

func AddOutbound(outbound string) (err error) {
	if outbound == "proxy" ||
		outbound == "direct" ||
		outbound == "block" {
		return fmt.Errorf("cannot add %v as the outbound name", outbound)
	}
	if err = db.SetAdd("outbounds", "names", outbound); err != nil {
		return err
	}
	// Apply default OutboundSetting for the new outbound group
	return SetOutboundSetting(outbound, DefaultOutboundSetting())
}

func SetOutboundSetting(outbound string, setting OutboundSetting) (err error) {
	if _, err := time.ParseDuration(setting.ProbeInterval); err != nil {
		return err
	}
	return db.Set(fmt.Sprintf("outbound.%v", outbound), "setting", setting)
}

func GetOutboundSetting(outbound string) (setting OutboundSetting) {
	err := db.Get(fmt.Sprintf("outbound.%v", outbound), "setting", &setting)
	if err != nil {
		return DefaultOutboundSetting()
	}
	return setting
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
	accounts, err := GetAccounts()
	if err != nil {
		return err
	}
	for _, account := range accounts {
		if err = db.Delete("accounts", account[0]); err != nil {
			return err
		}
	}
	return nil
}
func ExistsAccount(username string) bool {
	pwd, err := GetPasswordOfAccount(username)
	return err == nil && isAccountPasswordHash(pwd)
}

func GetPasswordOfAccount(username string) (pwd string, err error) {
	err = db.Get("accounts", username, &pwd)
	if err == nil && !isAccountPasswordHash(pwd) {
		return "", fmt.Errorf("account not found")
	}
	return
}

func GetAccounts() (accounts [][2]string, err error) {
	unames, err := db.GetBucketKeys("accounts")
	if err != nil {
		return nil, err
	}
	for _, uname := range unames {
		var passwd string
		err = db.Get("accounts", uname, &passwd)
		if err != nil || !isAccountPasswordHash(passwd) {
			continue
		}
		accounts = append(accounts, [2]string{uname, passwd})
	}
	return accounts, nil
}

func HasAnyAccounts() bool {
	accounts, err := GetAccounts()
	return err == nil && len(accounts) > 0
}

func isAccountPasswordHash(passwordHash string) bool {
	if len(passwordHash) == 32 {
		_, err := hex.DecodeString(passwordHash)
		return err == nil
	}
	return strings.HasPrefix(passwordHash, "$2a$") ||
		strings.HasPrefix(passwordHash, "$2b$") ||
		strings.HasPrefix(passwordHash, "$2y$")
}

func SetRunning(running bool) (err error) {
	return db.Set("system", "running", running)
}
func GetRunning() (running bool) {
	_ = db.Get("system", "running", &running)
	return running
}
