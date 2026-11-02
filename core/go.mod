// This module is the v2rayA merged core (v2raya-core).
// It wraps xray-core (fetched via go mod) and adds MultiObservatory and
// v2ray-compatible gRPC services without modifying xray's source.
//
// SPDX-License-Identifier: MPL-2.0
module github.com/v2rayA/v2raya-core

go 1.27

require (
	anytls v0.0.12
	github.com/daeuniverse/outbound v0.0.0-20250722064253-00c4fbb38759
	github.com/metacubex/mipstack v0.0.0-20260918021735-5e78149cf123
	github.com/miekg/dns v1.1.73
	github.com/sagernet/sing v0.6.0
	github.com/vishvananda/netlink v1.3.1
	// Pseudo-version for Xray-core v26.7.28 (commit 5ca6f4b7). v26 tags are not
	// resolvable via go mod because Xray's module path lacks the /v26 suffix.
	github.com/xtls/xray-core v1.260327.1-0.20261031132019-a4a3dff5b20c
	golang.org/x/sys v0.49.0
	golang.zx2c4.com/wireguard v0.0.0-20250521234502-f333402bd9cb
	golang.zx2c4.com/wireguard/windows v1.1.1
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/andybalholm/brotli v1.2.3 // indirect
	github.com/apernet/quic-go v0.61.1-0.20260806010916-184d081eef3e // indirect
	github.com/awnumar/fastrand v0.0.0-20210315215012-30ee0990fa2d // indirect
	github.com/awnumar/memcall v0.0.0-20190816154910-db5ea08008a3 // indirect
	github.com/awnumar/memguard v0.19.1 // indirect
	github.com/chen3feng/stl4go v0.1.1 // indirect
	github.com/dgryski/go-camellia v0.0.0-20191119043421-69a8a13fb23d // indirect
	github.com/dgryski/go-idea v0.0.0-20170306091226-d2fb45a411fb // indirect
	github.com/dgryski/go-rc2 v0.0.0-20150621095337-8a9021637152 // indirect
	github.com/eknkc/basex v1.0.1 // indirect
	github.com/ghodss/yaml v1.0.1-0.20220118164431-d8423dcdf344 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/google/pprof v0.0.0-20250208200701-d0013a598941 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/huin/goupnp v1.2.0 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/juju/ratelimit v1.0.2 // indirect
	github.com/klauspost/compress v1.20.0 // indirect
	github.com/klauspost/cpuid/v2 v2.4.1 // indirect
	github.com/koron/go-ssdp v0.0.4 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/libp2p/go-nat v1.0.1-0.20250821073202-01afc089f138 // indirect
	github.com/libp2p/go-netroute v0.2.1 // indirect
	github.com/mzz2017/disk-bloom v1.0.1 // indirect
	github.com/olicesx/quic-go v0.0.0-20260910141758-62d80bbebb5b // indirect
	github.com/onsi/ginkgo/v2 v2.22.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pion/dtls/v3 v3.1.5 // indirect
	github.com/pion/logging v0.2.4 // indirect
	github.com/pion/stun/v3 v3.1.7 // indirect
	github.com/pion/transport/v4 v4.1.0 // indirect
	github.com/pires/go-proxyproto v0.16.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/refraction-networking/utls v1.8.3-0.20260301010127-aa6edf4b11af // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/wlynxg/anet v0.0.5 // indirect
	github.com/xtls/reality v0.0.0-20261014101318-83f621788ca8 // indirect
	github.com/yuin/gopher-lua v1.1.2 // indirect
	gitlab.com/yawning/chacha20.git v0.0.0-20230427033715-7877545b1b37 // indirect
	go.uber.org/mock v0.5.2 // indirect
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/crypto/x509roots/fallback v0.0.0-20261020153306-9f2b28d4331a // indirect
	golang.org/x/exp v0.0.0-20250207012021-f9890c6ad9f3 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/tools v0.51.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260706201446-f0a921348800 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	layeh.com/gopher-luar v1.0.11 // indirect
	lukechampine.com/blake3 v1.4.1 // indirect
)

replace anytls v0.0.12 => github.com/anytls/anytls-go v0.0.12

replace github.com/daeuniverse/outbound => github.com/olicesx/outbound v0.0.0-sticky-ip.0.20260917223728-6a44445c9106
