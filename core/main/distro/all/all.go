// Package all loads all v2raya-core features for the merged binary.
// It is based on xray-core's main/distro/all/all.go with the following changes:
//   - multiobservatory and v2ray-compatible command features from hint/ are added
//   - xray's main/json loader is replaced by our hint/conf loader
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package all

import (
	// Mandatory xray features.
	_ "github.com/xtls/xray-core/app/dispatcher"
	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"

	// Commander and its services.
	_ "github.com/xtls/xray-core/app/commander"
	_ "github.com/xtls/xray-core/app/log/command"
	_ "github.com/xtls/xray-core/app/proxyman/command"
	_ "github.com/xtls/xray-core/app/stats/command"

	// Observatory command (xray native gRPC path).
	_ "github.com/xtls/xray-core/app/observatory/command"

	// Other optional xray features.
	_ "github.com/xtls/xray-core/app/dns"
	_ "github.com/xtls/xray-core/app/dns/fakedns"
	_ "github.com/xtls/xray-core/app/log"
	_ "github.com/xtls/xray-core/app/metrics"
	_ "github.com/xtls/xray-core/app/policy"
	_ "github.com/xtls/xray-core/app/reverse"
	_ "github.com/xtls/xray-core/app/router"
	_ "github.com/xtls/xray-core/app/stats"

	// Fix dependency cycle.
	_ "github.com/xtls/xray-core/transport/internet/tagged/taggedimpl"

	// Observatory feature (single group).
	_ "github.com/xtls/xray-core/app/observatory"

	// Inbound/outbound proxies.
	_ "github.com/xtls/xray-core/proxy/blackhole"
	_ "github.com/xtls/xray-core/proxy/dns"
	_ "github.com/xtls/xray-core/proxy/dokodemo"
	_ "github.com/xtls/xray-core/proxy/freedom"
	_ "github.com/xtls/xray-core/proxy/http"
	_ "github.com/xtls/xray-core/proxy/loopback"
	_ "github.com/xtls/xray-core/proxy/shadowsocks"
	_ "github.com/xtls/xray-core/proxy/socks"
	_ "github.com/xtls/xray-core/proxy/trojan"
	_ "github.com/xtls/xray-core/proxy/vless/inbound"
	_ "github.com/xtls/xray-core/proxy/vless/outbound"
	_ "github.com/xtls/xray-core/proxy/vmess/inbound"
	_ "github.com/xtls/xray-core/proxy/vmess/outbound"
	_ "github.com/xtls/xray-core/proxy/wireguard"

	// Transports.
	_ "github.com/xtls/xray-core/transport/internet/grpc"
	_ "github.com/xtls/xray-core/transport/internet/httpupgrade"
	_ "github.com/xtls/xray-core/transport/internet/kcp"
	_ "github.com/xtls/xray-core/transport/internet/reality"
	_ "github.com/xtls/xray-core/transport/internet/splithttp"
	_ "github.com/xtls/xray-core/transport/internet/tcp"
	_ "github.com/xtls/xray-core/transport/internet/tls"
	_ "github.com/xtls/xray-core/transport/internet/udp"
	_ "github.com/xtls/xray-core/transport/internet/websocket"

	// Transport headers.
	_ "github.com/xtls/xray-core/transport/internet/headers/http"
	_ "github.com/xtls/xray-core/transport/internet/headers/noop"

	// TOML and YAML config loaders (JSON is replaced by hint/conf below).
	_ "github.com/xtls/xray-core/main/toml"
	_ "github.com/xtls/xray-core/main/yaml"

	// Config loader from file or http(s).
	_ "github.com/xtls/xray-core/main/confloader/external"

	// xray sub-commands.
	_ "github.com/xtls/xray-core/main/commands/all"

	// v2raya-core extensions:
	//   - MultiObservatory: multi-group observatory backed by xray observatory
	//   - command: v2ray-compatible observatory gRPC service (native v2ray service name)
	//   - hint/conf: JSON config loader that supports multiObservatory field
	//     (replaces github.com/xtls/xray-core/main/json)
	//   - hint/proxy/anytls: native anytls outbound protocol handler
	//   - hint/proxy/hysteria2: native hysteria2 outbound protocol handler
	//   - hint/proxy/juicity: native juicity outbound protocol handler
	_ "github.com/v2rayA/v2raya-core/hint/app/observatory/command"
	_ "github.com/v2rayA/v2raya-core/hint/app/observatory/multiobservatory"
	_ "github.com/v2rayA/v2raya-core/hint/conf"
	_ "github.com/v2rayA/v2raya-core/hint/proxy/anytls"
	_ "github.com/v2rayA/v2raya-core/hint/proxy/hysteria2"
	_ "github.com/v2rayA/v2raya-core/hint/proxy/juicity"
)
