package net

type Listen struct {
	Listen                    string `json:"listen,omitempty"`
	ListenPort                int    `json:"listen_port,omitempty"`
	TcpFastOpen               bool   `json:"tcp_fast_open,omitempty"`
	TcpMultiPath              bool   `json:"tcp_multi_path,omitempty"`
	UdpFragment               bool   `json:"udp_fragment,omitempty"`
	UdpTimeout                int    `json:"udp_timeout,omitempty"`
	Detour                    string `json:"detour,omitempty"`
	Sniff                     bool   `json:"sniff,omitempty"`
	SniffOverrideDestination  bool   `json:"sniff_override_destination,omitempty"`
	SniffTimeout              string `json:"sniff_timeout,omitempty"`
	DomainStrategy            string `json:"domain_strategy,omitempty"`
	UdpDisableDomainUnmapping bool   `json:"udp_disable_domain_unmapping,omitempty"`
}
