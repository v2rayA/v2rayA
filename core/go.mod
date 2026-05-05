// This module is the v2rayA merged core (v2raya-core).
// It wraps xray-core (as a git submodule) and adds MultiObservatory and
// v2ray-compatible gRPC services without modifying xray's source.
//
// SPDX-License-Identifier: MPL-2.0
module github.com/v2rayA/v2raya-core

go 1.26

require (
	anytls v0.0.12
	github.com/daeuniverse/softwind v0.0.0-20231230065827-eed67f20d2c1
	github.com/sagernet/sing v0.5.1
	github.com/xtls/xray-core v0.0.0-local
	google.golang.org/grpc v1.79.3
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/andybalholm/brotli v1.0.6 // indirect
	github.com/apernet/quic-go v0.59.1-0.20260217092621-db4786c77a22 // indirect
	github.com/chen3feng/stl4go v0.1.1 // indirect
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/dgryski/go-camellia v0.0.0-20191119043421-69a8a13fb23d // indirect
	github.com/dgryski/go-idea v0.0.0-20170306091226-d2fb45a411fb // indirect
	github.com/dgryski/go-rc2 v0.0.0-20150621095337-8a9021637152 // indirect
	github.com/eknkc/basex v1.0.1 // indirect
	github.com/ghodss/yaml v1.0.1-0.20220118164431-d8423dcdf344 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/pprof v0.0.0-20230705174524-200ffdc848b8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/juju/ratelimit v1.0.2 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/miekg/dns v1.1.72 // indirect
	github.com/mzz2017/disk-bloom v1.0.1 // indirect
	github.com/mzz2017/quic-go v0.0.0-20231230054300-5221ce9164a3 // indirect
	github.com/onsi/ginkgo/v2 v2.11.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pires/go-proxyproto v0.11.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/qtls-go1-20 v0.4.1 // indirect
	github.com/refraction-networking/utls v1.8.3-0.20260301010127-aa6edf4b11af // indirect
	github.com/sagernet/sing-shadowsocks v0.2.7 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/vishvananda/netlink v1.3.1 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/xtls/reality v0.0.0-20260322125925-9234c772ba8f // indirect
	gitlab.com/yawning/chacha20.git v0.0.0-20230427033715-7877545b1b37 // indirect
	go.uber.org/mock v0.5.2 // indirect
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/mod v0.33.0 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	golang.org/x/tools v0.42.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	golang.zx2c4.com/wireguard v0.0.0-20250521234502-f333402bd9cb // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gvisor.dev/gvisor v0.0.0-20260122175437-89a5d21be8f0 // indirect
	lukechampine.com/blake3 v1.4.1 // indirect
)

replace github.com/xtls/xray-core => ./xray

replace anytls v0.0.12 => github.com/anytls/anytls-go v0.0.12
