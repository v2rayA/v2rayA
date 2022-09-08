module github.com/v2rayA/v2rayA

go 1.17

require (
	github.com/adrg/xdg v0.4.0
	github.com/beevik/ntp v0.3.0
	github.com/devfeel/mapper v0.7.5
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.1
	github.com/golang/protobuf v1.5.2
	github.com/google/gopacket v1.1.19
	github.com/gorilla/websocket v1.4.2
	github.com/json-iterator/go v1.1.12
	github.com/matoous/go-nanoid v1.5.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/muhammadmuzzammil1998/jsonc v0.0.0-20201229145248-615b0916ca38
	github.com/mzz2017/go-engine v0.0.0-20200509094339-b56921189229
	github.com/pkg/errors v0.9.1
	github.com/shadowsocks/go-shadowsocks2 v0.1.5
	github.com/shirou/gopsutil/v3 v3.21.11
	github.com/stevenroose/gonfig v0.1.5
	github.com/tidwall/gjson v1.10.2
	github.com/tidwall/sjson v1.2.3
	github.com/v2fly/v2ray-core/v4 v4.42.2-0.20210928173456-a9979057dcaa
	github.com/v2rayA/RoutingA v1.0.2
	github.com/v2rayA/beego/v2 v2.0.7
	github.com/v2rayA/shadowsocksR v1.0.4
	github.com/v2rayA/v2ray-lib v0.0.0-20211227083129-d4f59fbf62b8
	github.com/vearutop/statigz v1.1.7
	go.etcd.io/bbolt v1.3.6
	golang.org/x/net v0.0.0-20210903162142-ad29c8ab022f
	golang.org/x/sys v0.0.0-20211025201205-69cdffdb9359
	google.golang.org/grpc v1.40.0
)

require (
	github.com/dgryski/go-camellia v0.0.0-20191119043421-69a8a13fb23d // indirect
	github.com/dgryski/go-idea v0.0.0-20170306091226-d2fb45a411fb // indirect
	github.com/dgryski/go-rc2 v0.0.0-20150621095337-8a9021637152 // indirect
	github.com/djherbis/times v1.5.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.9.0 // indirect
	github.com/gocarina/gocsv v0.0.0-20210408192840-02d7211d929d // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/pires/go-proxyproto v0.6.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/riobard/go-bloom v0.0.0-20200614022211-cdc8013cb5b3 // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/ugorji/go/codec v1.1.7 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	gitlab.com/yawning/chacha20.git v0.0.0-20190903091407-6d1cb28dc72c // indirect
	go4.org/intern v0.0.0-20211027215823-ae77deb06f29 // indirect
	go4.org/unsafe/assume-no-moving-gc v0.0.0-20211027215541-db492cf91b37 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	inet.af/netaddr v0.0.0-20211027220019-c74959edd3b6 // indirect
)

// Replace dependency modules with local developing copy
// use `go list -m all` to confirm the final module used
//replace github.com/v2rayA/shadowsocksR => ../../shadowsocksR
//replace github.com/mzz2017/go-engine => ../../go-engine
//replace github.com/v2rayA/beego/v2 => D:\beego

replace go4.org/unsafe/assume-no-moving-gc => go4.org/unsafe/assume-no-moving-gc v0.0.0-20220617031537-928513b29760
