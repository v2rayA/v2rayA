module github.com/v2rayA/v2rayA

go 1.16

require (
	github.com/beevik/ntp v0.3.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.1
	github.com/go-playground/validator/v10 v10.5.0 // indirect
	github.com/gocarina/gocsv v0.0.0-20210408192840-02d7211d929d // indirect
	github.com/golang/mock v1.5.0 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.5
	github.com/google/gopacket v1.1.19
	github.com/google/uuid v1.2.0 // indirect
	github.com/gookit/color v1.4.2
	github.com/gorilla/websocket v1.4.2
	github.com/json-iterator/go v1.1.11
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/matoous/go-nanoid v1.5.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/muhammadmuzzammil1998/jsonc v0.0.0-20201229145248-615b0916ca38
	github.com/mzz2017/go-engine v0.0.0-20200509094339-b56921189229
	github.com/pelletier/go-toml v1.9.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/refraction-networking/utls v0.0.0-20201210053706-2179f286686b // indirect
	github.com/shadowsocks/go-shadowsocks2 v0.1.5-0.20210421162817-acdbac05f5a5
	github.com/stevenroose/gonfig v0.1.5
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/tidwall/gjson v1.7.5
	github.com/ugorji/go v1.2.5 // indirect
	github.com/v2rayA/routingA v0.0.0-20201204065601-aef348ea7aa1
	github.com/v2rayA/shadowsocksR v1.0.3
	github.com/xujiajun/nutsdb v0.6.0
	go.starlark.net v0.0.0-20210429133630-0c63ff3779a6 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/net v0.0.0-20210428140749-89ef3d95e781
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887
	google.golang.org/genproto v0.0.0-20210429181445-86c259c2b4ab // indirect
	google.golang.org/grpc v1.37.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	v2ray.com/core v4.19.1+incompatible
)

// Replace dependency modules with local developing copy
// use `go list -m all` to confirm the final module used
//replace github.com/v2rayA/shadowsocksR => ../../shadowsocksR
//replace github.com/mzz2017/go-engine => ../../go-engine
replace v2ray.com/core => github.com/v2ray/v2ray-core v0.0.0-20200603100350-6b5d2fed91c0
