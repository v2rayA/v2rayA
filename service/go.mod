module github.com/v2rayA/v2rayA

go 1.13

require (
	github.com/beevik/ntp v0.3.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/cors v1.3.0
	github.com/gin-gonic/gin v1.4.0
	github.com/golang/protobuf v1.4.0
	github.com/google/gopacket v1.1.17
	github.com/gookit/color v1.2.0
	github.com/json-iterator/go v1.1.9
	github.com/matoous/go-nanoid v0.0.0-20200226125206-b0a1054fe39d
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/muhammadmuzzammil1998/jsonc v0.0.0-20190906142622-1265e9b150c6
	github.com/mzz2017/go-engine v0.0.0-20200509094339-b56921189229
	github.com/nadoo/glider v0.13.0
	github.com/pkg/errors v0.9.1
	github.com/stevenroose/gonfig v0.1.4
	github.com/tidwall/gjson v1.3.5
	github.com/v2rayA/routingA v0.0.0-20201204065601-aef348ea7aa1
	github.com/v2rayA/shadowsocksR v1.0.2
	github.com/xujiajun/nutsdb v0.5.0
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	golang.org/x/sys v0.0.0-20201202213521-69691e467435
	gopkg.in/yaml.v2 v2.2.7 // indirect
	v2ray.com/core v4.19.1+incompatible
)

// Replace dependency modules with local developing copy
// use `go list -m all` to confirm the final module used
//replace github.com/v2rayA/shadowsocksR => ../../shadowsocksR
//replace github.com/mzz2017/go-engine => ../../go-engine
replace v2ray.com/core => github.com/v2ray/v2ray-core v0.0.0-20200603100350-6b5d2fed91c0
