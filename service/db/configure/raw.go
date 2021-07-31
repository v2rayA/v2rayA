package configure

import "github.com/v2rayA/v2rayA/core/vmessInfo"

type ServerRaw struct {
	VmessInfo vmessInfo.VmessInfo `json:"vmessInfo"`
	Latency   string              `json:"latency"`
}

type SubscriptionRaw struct {
	Remarks string      `json:"remarks,omitempty"`
	Address string      `json:"address"`
	Status  string      `json:"status"` //update time, error info, etc.
	Servers []ServerRaw `json:"servers"`
	Info    string      `json:"info"` // maybe include some info from provider
}
