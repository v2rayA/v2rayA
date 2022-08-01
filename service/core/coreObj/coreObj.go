package coreObj

type APIObject struct {
	Tag      string   `json:"tag"`
	Services []string `json:"services"`
}
type Observatory struct {
	SubjectSelector []string `json:"subjectSelector"`
	ProbeURL        string   `json:"probeURL,omitempty"`
	ProbeInterval   string   `json:"probeInterval,omitempty"`
}
type Balancer struct {
	Tag      string           `json:"tag"`
	Selector []string         `json:"selector"`
	Strategy BalancerStrategy `json:"strategy"`
}
type BalancerStrategy struct {
	Type string `json:"type"`
}
type FakeDns struct {
	IpPool   string `json:"ipPool"`
	PoolSize int    `json:"poolSize"`
}
type RoutingRule struct {
	Type        string   `json:"type"`
	OutboundTag string   `json:"outboundTag,omitempty"`
	BalancerTag string   `json:"balancerTag,omitempty"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Network     string   `json:"network,omitempty"`
	Port        string   `json:"port,omitempty"`
	SourcePort  string   `json:"sourcePort,omitempty"`
	Protocol    []string `json:"protocol,omitempty"`
	Source      []string `json:"source,omitempty"`
	User        []string `json:"user,omitempty"`
}
type Log struct {
	Access   string `json:"access"`
	Error    string `json:"error"`
	Loglevel string `json:"loglevel"`
}
type Sniffing struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride,omitempty"`
	MetadataOnly bool     `json:"metadataOnly"`
}
type Inbound struct {
	Port           int              `json:"port"`
	Protocol       string           `json:"protocol"`
	Listen         string           `json:"listen,omitempty"`
	Sniffing       Sniffing         `json:"sniffing,omitempty"`
	Settings       *InboundSettings `json:"settings,omitempty"`
	StreamSettings *StreamSettings  `json:"streamSettings"`
	Tag            string           `json:"tag,omitempty"`
}
type InboundSettings struct {
	Auth           string      `json:"auth,omitempty"`
	UDP            bool        `json:"udp,omitempty"`
	IP             interface{} `json:"ip,omitempty"`
	Accounts       []Account   `json:"accounts,omitempty"`
	Clients        interface{} `json:"clients,omitempty"`
	Decryption     string      `json:"decryption,omitempty"`
	Network        string      `json:"network,omitempty"`
	UserLevel      int         `json:"userLevel,omitempty"`
	Address        string      `json:"address,omitempty"`
	Port           int         `json:"port,omitempty"`
	FollowRedirect bool        `json:"followRedirect,omitempty"`
}
type VlessClient struct {
	Id    string `json:"id"`
	Level int    `json:"level,omitempty"`
	Email string `json:"email,omitempty"`
}
type Account struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}
type User struct {
	ID         string `json:"id"`
	AlterID    int    `json:"alterId,omitempty"`
	Encryption string `json:"encryption,omitempty"`
	Flow       string `json:"flow,omitempty"`
	Security   string `json:"security,omitempty"`
}
type Vnext struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Users   []User `json:"users"`
}
type Server struct {
	Network  string         `json:"network,omitempty"`
	Address  string         `json:"address,omitempty"`
	Method   string         `json:"method,omitempty"`
	Ota      bool           `json:"ota,omitempty"`
	Password string         `json:"password,omitempty"`
	Port     int            `json:"port,omitempty"`
	Users    []OutboundUser `json:"users,omitempty"`
}
type Settings struct {
	Vnext          interface{} `json:"vnext,omitempty"`
	Servers        interface{} `json:"servers,omitempty"`
	DomainStrategy string      `json:"domainStrategy,omitempty"`
	Port           int         `json:"port,omitempty"`
	Address        string      `json:"address,omitempty"`
	Network        string      `json:"network,omitempty"`
	Redirect       string      `json:"redirect,omitempty"`
	UserLevel      *int        `json:"userLevel,omitempty"`
}
type TLSSettings struct {
	AllowInsecure                    bool          `json:"allowInsecure"`
	ServerName                       interface{}   `json:"serverName,omitempty"`
	Alpn                             []string      `json:"alpn,omitempty"`
	PinnedPeerCertificateChainSha256 string        `json:"pinnedPeerCertificateChainSha256,omitempty"`
	Certificates                     []Certificate `json:"certificates,omitempty"`
}
type Certificate struct {
	CertificateFile string `json:"certificateFile"`
	KeyFile         string `json:"keyFile"`
}
type Headers struct {
	Host string `json:"Host"`
}
type WsSettings struct {
	Path    string  `json:"path"`
	Headers Headers `json:"headers"`
}
type StreamSettings struct {
	Network      string        `json:"network,omitempty"`
	Security     string        `json:"security,omitempty"`
	TLSSettings  *TLSSettings  `json:"tlsSettings,omitempty"`
	XTLSSettings *TLSSettings  `json:"xtlsSettings,omitempty"`
	TCPSettings  *TCPSettings  `json:"tcpSettings,omitempty"`
	KcpSettings  *KcpSettings  `json:"kcpSettings,omitempty"`
	WsSettings   *WsSettings   `json:"wsSettings,omitempty"`
	HTTPSettings *HttpSettings `json:"httpSettings,omitempty"`
	GrpcSettings *GrpcSettings `json:"grpcSettings,omitempty"`
	Sockopt      *Sockopt      `json:"sockopt,omitempty"`
}
type GrpcSettings struct {
	ServiceName string `json:"serviceName"`
}
type Sockopt struct {
	Mark        *int    `json:"mark,omitempty"`
	Tos         *int    `json:"tos,omitempty"`
	TCPFastOpen *bool   `json:"tcpFastOpen,omitempty"`
	Tproxy      *string `json:"tproxy,omitempty"`
}
type Mux struct {
	Enabled     bool `json:"enabled"`
	Concurrency int  `json:"concurrency"`
}
type OutboundObject struct {
	Tag            string          `json:"tag"`
	Protocol       string          `json:"protocol"`
	Settings       Settings        `json:"settings,omitempty"`
	StreamSettings *StreamSettings `json:"streamSettings,omitempty"`
	ProxySettings  *ProxySettings  `json:"proxySettings,omitempty"`
	Mux            *Mux            `json:"mux,omitempty"`
	Balancers      []string        `json:"-"`
}
type OutboundUser struct {
	User  string `json:"user"`
	Pass  string `json:"pass"`
	Level int    `json:"level,omitempty"`
}
type ProxySettings struct {
	Tag string `json:"tag,omitempty"`
}
type TCPSettings struct {
	ConnectionReuse bool      `json:"connectionReuse"`
	Header          TCPHeader `json:"header"`
}
type HTTPRequest struct {
	Version string        `json:"version"`
	Method  string        `json:"method"`
	Path    []string      `json:"path"`
	Headers HTTPReqHeader `json:"headers"`
}
type HTTPReqHeader struct {
	Host           []string `json:"Host"`
	UserAgent      []string `json:"User-Agent"`
	AcceptEncoding []string `json:"Accept-Encoding"`
	Connection     []string `json:"Connection"`
	Pragma         string   `json:"Pragma"`
}
type HTTPResponse struct {
	Version string         `json:"version"`
	Status  string         `json:"status"`
	Reason  string         `json:"reason"`
	Headers HTTPRespHeader `json:"headers"`
}
type HTTPRespHeader struct {
	ContentType      []string `json:"Content-Type"`
	TransferEncoding []string `json:"Transfer-Encoding"`
	Connection       []string `json:"Connection"`
	Pragma           string   `json:"Pragma"`
}
type TCPHeader struct {
	Type     string       `json:"type"`
	Request  HTTPRequest  `json:"request"`
	Response HTTPResponse `json:"response"`
}
type KcpSettings struct {
	Mtu              int       `json:"mtu"`
	Tti              int       `json:"tti"`
	UplinkCapacity   int       `json:"uplinkCapacity"`
	DownlinkCapacity int       `json:"downlinkCapacity"`
	Congestion       bool      `json:"congestion"`
	ReadBufferSize   int       `json:"readBufferSize"`
	WriteBufferSize  int       `json:"writeBufferSize"`
	Header           KcpHeader `json:"header"`
	Seed             string    `json:"seed,omitempty"`
}
type KcpHeader struct {
	Type     string      `json:"type"`
	Request  interface{} `json:"request"`
	Response interface{} `json:"response"`
}
type HttpSettings struct {
	Path   string   `json:"path"`
	Host   []string `json:"host,omitempty"`
	Method string   `json:"method,omitempty"`
}
type Hosts map[string]interface{}

type DNS struct {
	Hosts           Hosts         `json:"hosts,omitempty"`
	Servers         []interface{} `json:"servers"`
	ClientIp        string        `json:"clientIp,omitempty"`
	Tag             string        `json:"tag,omitempty"`
	DisableFallback *bool         `json:"disableFallback,omitempty"`
	QueryStrategy   string        `json:"queryStrategy,omitempty"`
}
type DnsServer struct {
	Address      string   `json:"address"`
	Port         int      `json:"port,omitempty"`
	Domains      []string `json:"domains,omitempty"`
	SkipFallback bool     `json:"skipFallback,omitempty"`
}
type Policy struct {
	Levels struct {
		Num0 struct {
			Handshake         int  `json:"handshake,omitempty"`
			ConnIdle          int  `json:"connIdle,omitempty"`
			UplinkOnly        int  `json:"uplinkOnly,omitempty"`
			DownlinkOnly      int  `json:"downlinkOnly,omitempty"`
			StatsUserUplink   bool `json:"statsUserUplink,omitempty"`
			StatsUserDownlink bool `json:"statsUserDownlink,omitempty"`
			BufferSize        int  `json:"bufferSize,omitempty"`
		} `json:"0"`
	} `json:"levels"`
	System struct {
		StatsInboundUplink   bool `json:"statsInboundUplink,omitempty"`
		StatsInboundDownlink bool `json:"statsInboundDownlink,omitempty"`
	} `json:"system"`
}
