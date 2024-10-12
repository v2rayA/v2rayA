package configure

type Ports struct {
	Socks5        int     `json:"socks5"`
	Http          int     `json:"http"`
	Socks5WithPac int     `json:"socks5WithPac"`
	HttpWithPac   int     `json:"httpWithPac"`
	Vmess         int     `json:"vmess"`
	Api           ApiPort `json:"api"`
}

type ApiPort struct {
	Port     int      `json:"port"`
	Services []string `json:"services"`
}
