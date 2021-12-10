module github.com/v2rayA/v2rayA

go 1.16

require (
	github.com/beevik/ntp v0.3.0
	github.com/boltdb/bolt v1.3.1
	github.com/devfeel/mapper v0.7.5
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/gzip v0.0.3
	github.com/gin-gonic/gin v1.7.1
	github.com/gocarina/gocsv v0.0.0-20210408192840-02d7211d929d // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/gopacket v1.1.19
	github.com/gorilla/websocket v1.4.2
	github.com/json-iterator/go v1.1.11
	github.com/matoous/go-nanoid v1.5.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/muhammadmuzzammil1998/jsonc v0.0.0-20201229145248-615b0916ca38
	github.com/mzz2017/go-engine v0.0.0-20200509094339-b56921189229
	github.com/pkg/errors v0.9.1
	github.com/shadowsocks/go-shadowsocks2 v0.1.5-0.20210421162817-acdbac05f5a5
	github.com/shirou/gopsutil/v3 v3.21.11
	github.com/stevenroose/gonfig v0.1.5
	github.com/tidwall/gjson v1.10.2
	github.com/tidwall/sjson v1.2.3
	github.com/v2fly/v2ray-core/v4 v4.42.1
	github.com/v2rayA/RoutingA v1.0.0
	github.com/v2rayA/beego/v2 v2.0.4
	github.com/v2rayA/go-uci v0.0.0-20210907104827-4cf744297b41
	github.com/v2rayA/shadowsocksR v1.0.3
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d
	golang.org/x/sys v0.0.0-20211013075003-97ac67df715c
	google.golang.org/grpc v1.40.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

// Replace dependency modules with local developing copy
// use `go list -m all` to confirm the final module used
//replace github.com/v2rayA/shadowsocksR => ../../shadowsocksR
//replace github.com/mzz2017/go-engine => ../../go-engine
//replace github.com/v2rayA/beego/v2 => ../../beego

replace github.com/boltdb/bolt => github.com/go-gitea/bolt v0.0.0-20170420010917-ccd680d8c1a0
