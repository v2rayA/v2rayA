package configure

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

type ServerRaw struct {
	ServerObj serverObj.ServerObj `json:"serverObj"`
	Latency   string              `json:"latency"`
}

type SubscriptionRaw struct {
	Remarks string      `json:"remarks,omitempty"`
	Address string      `json:"address"`
	Status  string      `json:"status"` //update time, error info, etc.
	Servers []ServerRaw `json:"servers"`
	Info    string      `json:"info"` // maybe include some info from provider
	AutoSelect bool     `json:"autoSelect"`
}

func Bytes2SubscriptionRaw(b []byte) (*SubscriptionRaw, error) {
	var s SubscriptionRaw
	rawList := gjson.GetBytes(b, "servers").Array()
	for _, raw := range rawList {
		var obj serverObj.ServerObj
		obj, err := serverObj.New(raw.Get("serverObj.protocol").String())
		if err != nil {
			return nil, err
		}
		s.Servers = append(s.Servers, ServerRaw{ServerObj: obj})
	}
	if err := jsoniter.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	if s.Servers == nil {
		s.Servers = []ServerRaw{}
	}
	return &s, nil
}

func Bytes2ServerRaw(b []byte) (*ServerRaw, error) {
	var s ServerRaw
	var obj serverObj.ServerObj
	protocol := gjson.GetBytes(b, "serverObj.protocol").String()
	if protocol == "" {
		log.Warn("empty protocol, fallback to vmess: %v", gjson.GetBytes(b, "serverObj.ps").String())
		protocol = "vmess"
	}
	obj, err := serverObj.New(protocol)
	if err != nil {
		return nil, err
	}
	s.ServerObj = obj
	if err := jsoniter.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
