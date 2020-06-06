package configure

import "github.com/mzz2017/v2rayA/core/vmessInfo"

type ServerRaw struct {
	VmessInfo vmessInfo.VmessInfo `json:"vmessInfo"`
}

type SubscriptionRaw struct {
	Remarks string      `json:"remarks,omitempty"`
	Address string      `json:"address"`
	Status  string      `json:"status"` //update time, error info, etc.
	Servers []ServerRaw `json:"servers"`
}
