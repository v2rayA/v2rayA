module github.com/v2rayA/v2rayA

go 1.26

require (
	github.com/adrg/xdg v0.4.0
	github.com/beevik/ntp v0.3.0
	github.com/devfeel/mapper v0.7.5
	github.com/gin-contrib/cors v1.6.0
	github.com/gin-gonic/gin v1.9.1
	github.com/go-leo/slicex v1.0.14
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/gorilla/websocket v1.5.0
	github.com/json-iterator/go v1.1.12
	github.com/matoous/go-nanoid v1.5.0
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/muhammadmuzzammil1998/jsonc v0.0.0-20201229145248-615b0916ca38
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil/v3 v3.21.11
	github.com/stevenroose/gonfig v0.1.5
	github.com/tidwall/gjson v1.10.2
	github.com/v2fly/v2ray-core/v5 v5.7.0
	github.com/v2rayA/RoutingA v1.0.2
	github.com/v2rayA/beego/v2 v2.0.7
	github.com/v2rayA/v2rayA-lib4 v0.0.0-20230812094818-595f87cb2a49
	github.com/vearutop/statigz v1.1.7
	go.etcd.io/bbolt v1.4.3
	golang.org/x/crypto v0.46.0
	golang.org/x/net v0.48.0
	golang.org/x/sys v0.39.0
	google.golang.org/grpc v1.57.1
	google.golang.org/protobuf v1.36.1
	gopkg.in/yaml.v3 v3.0.1
	modernc.org/sqlite v1.29.5
)

require (
	github.com/andybalholm/brotli v1.0.6 // indirect
	github.com/bytedance/sonic v1.11.2 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
	github.com/chenzhuoyu/iasm v0.9.1 // indirect
	github.com/djherbis/times v1.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.19.0 // indirect
	github.com/gocarina/gocsv v0.0.0-20210408192840-02d7211d929d // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/pprof v0.0.0-20250208200701-d0013a598941 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/miekg/dns v1.1.72 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/onsi/ginkgo/v2 v2.22.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.1.1 // indirect
	github.com/pires/go-proxyproto v0.7.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/quic-go/qtls-go1-20 v0.3.3 // indirect
	github.com/refraction-networking/utls v1.8.2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/sjson v1.2.3 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba // indirect
	golang.org/x/arch v0.7.0 // indirect
	golang.org/x/exp v0.0.0-20250207012021-f9890c6ad9f3 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230807174057-1744710a1577 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	modernc.org/gc/v3 v3.0.0-20240107210532-573471604cb6 // indirect
	modernc.org/libc v1.41.0 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.7.2 // indirect
	modernc.org/strutil v1.2.0 // indirect
	modernc.org/token v1.1.0 // indirect
)

// Replace dependency modules with local developing copy
// use `go list -m all` to confirm the final module used
//replace github.com/v2rayA/shadowsocksR => ../../shadowsocksR
//replace github.com/mzz2017/go-engine => ../../go-engine
//replace github.com/v2rayA/beego/v2 => D:\beego

// replace github.com/daeuniverse/softwind => ../../softwind

replace github.com/v2fly/v2ray-core/v5 => github.com/v2rayA/v2ray-core/v5 v5.0.0-20230812170925-960565fa0686
