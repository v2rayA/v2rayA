package vmessInfo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
)

type VmessInfo struct {
	Ps       string `json:"ps"`
	Add      string `json:"add"`
	Port     string `json:"port"`
	ID       string `json:"id"`
	Aid      string `json:"aid"`
	Net      string `json:"net"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Path     string `json:"path"`
	TLS      string `json:"tls"`
	V        string `json:"v"`
	Protocol string `json:"protocol"`
}

func (v *VmessInfo) ExportToURL() string {
	switch v.Protocol {
	case "", "vmess":
		//去除info中的protocol，减少URL体积
		it := reflect.TypeOf(*v)
		iv := reflect.ValueOf(*v)
		m := make(map[string]interface{})
		for i := 0; i < it.NumField(); i++ {
			f := it.Field(i)
			chKey := f.Tag.Get("json")
			if chKey == "protocol" { //不转换protocol
				continue
			}
			m[chKey] = iv.FieldByName(f.Name).Interface()
		}
		b, _ := json.Marshal(m)
		return "vmess://" + base64.StdEncoding.EncodeToString(b)
	case "shadowsocks":
		/* ss://BASE64(method:password)@server:port#name */
		return fmt.Sprintf(
			"ss://%v@%v:%v#%v",
			base64.StdEncoding.EncodeToString([]byte(v.Net+":"+v.ID)),
			v.Add,
			v.Port,
			v.Ps,
		)
	}
	return ""
}
