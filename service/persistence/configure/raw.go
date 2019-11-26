package configure

import "V2RayA/model/vmessInfo"

type ServerRaw struct {
	VmessInfo vmessInfo.VmessInfo `json:"vmessInfo"`
}

type SubscriptionRaw struct {
	Address string      `json:"address"`
	Status  string      `json:"status"` //update time, error info, etc.
	Servers []ServerRaw `json:"servers"`
}
