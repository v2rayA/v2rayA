module github.com/v2rayA/v2rayA

go 1.16

require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/beevik/ntp v0.3.0
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
	github.com/muhammadmuzzammil1998/jsonc v0.0.0-20201229145248-615b0916ca38
	github.com/mzz2017/go-engine v0.0.0-20200509094339-b56921189229
	github.com/pkg/errors v0.9.1
	github.com/shadowsocks/go-shadowsocks2 v0.1.5-0.20210421162817-acdbac05f5a5
	github.com/shirou/gopsutil v3.21.8+incompatible // indirect
	github.com/shirou/gopsutil/v3 v3.21.8
	github.com/stevenroose/gonfig v0.1.5
	github.com/tidwall/gjson v1.7.5
	github.com/v2fly/v2ray-core/v4 v4.42.2-0.20210928173456-a9979057dcaa
	github.com/v2rayA/RoutingA v1.0.0
	github.com/v2rayA/beego/v2 v2.0.4
	github.com/v2rayA/go-uci v0.0.0-20210907104827-4cf744297b41
	github.com/v2rayA/shadowsocksR v1.0.3
	github.com/xujiajun/nutsdb v0.5.0
	golang.org/x/net v0.0.0-20210903162142-ad29c8ab022f
	golang.org/x/sys v0.0.0-20210903071746-97244b99971b
	google.golang.org/grpc v1.40.0
)

// Replace dependency modules with local developing copy
// use `go list -m all` to confirm the final module used
//replace github.com/v2rayA/shadowsocksR => ../../shadowsocksR
//replace github.com/mzz2017/go-engine => ../../go-engine
//replace github.com/v2rayA/beego/v2 => ../../beego

// windows/arm64 support
replace github.com/go-ole/go-ole => github.com/go-ole/go-ole v0.0.0-20210915003542-8b1f7f90f6b1

replace github.com/shirou/gopsutil/v3 => github.com/shirou/gopsutil v0.0.0-20210919144451-80d5b574053f
