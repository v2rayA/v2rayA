package config

type Param struct {
	Address string `id:"address" short:"a" default:"0.0.0.0" desc:"listening address"`
	Port    string `id:"port" short:"p" default:"8080" desc:"listening port"`
}

func GetServiceConfig() *Param {
	return &param
}