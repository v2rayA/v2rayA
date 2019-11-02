package nodeData

import (
	"V2RayA/models/touch"
	"V2RayA/models/vmessInfo"
)

/*
对VmessInfo更高层次的抽象，加入了对应的config配置
*/
type NodeData struct {
	VmessInfo vmessInfo.VmessInfo `json:"vmessInfo"`
	Config    string              `json:"config"`
}

func (nd *NodeData) ToTouchServerRaw() (tsr touch.TouchServerRaw) {
	tsr.VmessInfo = nd.VmessInfo
	tsr.Connected = false
	return
}
