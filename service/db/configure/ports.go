package configure

type Ports struct {
	Socks5      int `json:"socks5"`
	Http        int `json:"http"`
	HttpWithPac int `json:"httpWithPac"`
	VlessGrpc   int `json:"vlessGrpc"`
}
