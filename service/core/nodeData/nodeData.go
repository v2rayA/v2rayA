package nodeData

import (
	"v2rayA/core/vmessInfo"
	"v2rayA/db/configure"
)

/*
对VmessInfo更高层次的抽象，加入了对应的config配置
*/
type NodeData struct {
	VmessInfo vmessInfo.VmessInfo `json:"vmessInfo"`
	//Config    string              `json:"config"`
}

func (nd *NodeData) ToServerRaw() (tsr *configure.ServerRaw) {
	tsr = new(configure.ServerRaw)
	tsr.VmessInfo = nd.VmessInfo
	return
}
