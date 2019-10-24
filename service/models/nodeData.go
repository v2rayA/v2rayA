package models

/*
对VmessInfo更高层次的抽象，加入了对应的config配置
*/
type NodeData struct {
	VmessInfo VmessInfo `json:"vmessInfo"`
	Config    string    `json:"config"`
}

func (nd *NodeData) ToTouchServerRaw() (tsr TouchServerRaw) {
	tsr.VmessInfo = nd.VmessInfo
	tsr.Connected = false
	return
}
