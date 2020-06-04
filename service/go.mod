module v2rayA

go 1.13

require (
	github.com/DeanThompson/ginpprof v0.0.0-20190408063150-3be636683586
	github.com/beevik/ntp v0.3.0
	github.com/cakturk/go-netstat v0.0.0-20190620190123-a633b9c55b1a
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/cors v1.3.0
	github.com/gin-gonic/gin v1.4.0
	github.com/gocarina/gocsv v0.0.0-20200302151839-87c60d755c58
	github.com/golang/protobuf v1.4.0
	github.com/google/gopacket v1.1.17
	github.com/gookit/color v1.2.0
	github.com/json-iterator/go v1.1.9
	github.com/matoous/go-nanoid v0.0.0-20200226125206-b0a1054fe39d
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/muhammadmuzzammil1998/jsonc v0.0.0-20190906142622-1265e9b150c6
	github.com/mzz2017/go-engine v0.0.0-20200509094339-b56921189229
	github.com/mzz2017/shadowsocksR v0.0.0-20200126130347-721f53a7b15a
	github.com/nadoo/glider v0.9.3
	github.com/pkg/errors v0.9.1
	github.com/stevenroose/gonfig v0.1.4
	github.com/tidwall/gjson v1.3.5
	github.com/tidwall/sjson v1.0.4
	golang.org/x/crypto v0.0.0-20200406173513-056763e48d71 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/sys v0.0.0-20200413165638-669c56c373c4
	gopkg.in/yaml.v2 v2.2.7 // indirect
	v2ray.com/core v4.19.1+incompatible
)

// Replace dependency modules with local developing copy
// use `go list -m all` to confirm the final module used
//replace github.com/mzz2017/shadowsocksR => ../../shadowsocksR
//replace github.com/mzz2017/go-engine => ../../go-engine
