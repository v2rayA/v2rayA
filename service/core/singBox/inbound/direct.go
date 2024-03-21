package inbound

type Direct struct {
	Network         string `json:"network,omitempty"`
	OverrideAddress string `json:"override_address,omitempty"`
	OverridePort    string `json:"override_port,omitempty"`
}

func (i Direct) inbound() {}
