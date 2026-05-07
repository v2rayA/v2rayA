<template>
  <div class="modal-card">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $tc("configureServer.title", readonly ? 2 : 1) }}
      </p>
    </header>
    <section ref="section" :class="{ 'modal-card-body': true }">
      <b-tabs v-model="tabChoice" class="block" type="is-boxed is-twitter" vertical>
        <b-tab-item label="V2RAY">
          <div v-if="tabChoice === 0">
            <b-field label="Protocol" label-position="on-border">
              <b-select v-model="v2ray.protocol" expanded @input="handleV2rayProtocolSwitch">
                <option value="vmess">VMESS</option>
                <option value="vless">VLESS</option>
              </b-select>
            </b-field>
            <b-field label="Name" label-position="on-border">
              <b-input ref="v2ray_name" v-model="v2ray.ps" :placeholder="$t('configureServer.servername')" expanded />
            </b-field>
            <b-field label="Host" label-position="on-border">
              <b-input ref="v2ray_add" required placeholder="IP / HOST" v-model="v2ray.add" expanded />
            </b-field>
            <b-field label="Port" label-position="on-border">
              <b-input ref="v2ray_port" required :placeholder="$t('configureServer.port')" type="number" v-model="v2ray.port"
                expanded />
            </b-field>
            <b-field label="ID" label-position="on-border">
              <b-input ref="v2ray_id" required placeholder="UserID" v-model="v2ray.id" expanded />
            </b-field>
            <b-field v-if="v2ray.protocol === 'vmess'" label="AlterID" label-position="on-border">
              <b-input ref="v2ray_aid" placeholder="AlterID" type="number" min="0" max="65535" v-model="v2ray.aid"
                expanded />
            </b-field>
            <b-field v-if="v2ray.protocol === 'vmess'" label="Security" label-position="on-border">
              <b-select v-model="v2ray.scy" expanded required>
                <option value="auto">Auto</option>
                <option value="aes-256-gcm">aes-256-gcm</option>
                <option value="aes-128-gcm">aes-128-gcm</option>
                <option value="chacha20-poly1305">chacha20-poly1305</option>
                <option value="xchacha20-poly1305">xchacha20-poly1305</option>
                <option value="none">none</option>
                <option value="zero">zero</option>
              </b-select>
            </b-field>
            <b-field v-show="v2ray.type !== 'dtls'" label="TLS" label-position="on-border">
              <b-select v-model="v2ray.tls" expanded @input="handleNetworkChange">
                <option value="none">{{ $t("setting.options.off") }}</option>
                <option value="tls">tls</option>
                <option v-if="variant() === 'xray'" value="reality"> reality </option>
                <option v-if="variant() === 'xray'" value="xtls">xtls</option>
              </b-select>
            </b-field>
            <b-field v-if="v2ray.tls !== 'none'" label="SNI" label-position="on-border">
              <b-input ref="v2ray_sni" v-model="v2ray.sni" placeholder="SNI" expanded />
            </b-field>
            <b-field v-show="v2ray.tls === 'tls' || v2ray.tls === 'reality'" label="uTLS fingerprint"
              label-position="on-border">
              <b-select ref="v2ray_fp" v-model="v2ray.fp" expanded>
                <option value="">empty</option>
                <option value="chrome">chrome</option>
                <option value="firefox">firefox</option>
                <option value="safari">safari</option>
                <option value="ios">ios</option>
                <option value="android">android</option>
                <option value="edge">edge</option>
                <option value="360">360</option>
                <option value="qq">qq</option>
                <option value="random">random</option>
                <option value="randomized">randomized</option>
              </b-select>
            </b-field>
            <b-field v-if="v2ray.protocol === 'vless' && v2ray.tls !== 'none'" label="Encryption" label-position="on-border">
              <b-input ref="v2ray_encryption" v-model="v2ray.scy" placeholder="none" expanded />
            </b-field>
            <b-field v-if="v2ray.protocol === 'vless' && v2ray.tls !== 'none'" label="Flow" label-position="on-border">
              <b-input ref="v2ray_flow" v-model="v2ray.flow" placeholder="Flow" expanded />
            </b-field>
            <b-field v-show="v2ray.tls === 'reality'" label="pbk" label-position="on-border">
              <b-input v-model="v2ray.pbk" placeholder="pbk" expanded />
            </b-field>
            <b-field v-show="v2ray.tls === 'reality'" label="sid" label-position="on-border">
              <b-input v-model="v2ray.sid" placeholder="sid" expanded />
            </b-field>
            <b-field v-show="v2ray.tls === 'reality'" label="spx" label-position="on-border">
              <b-input v-model="v2ray.spx" placeholder="spx" expanded />
            </b-field>
            <b-field v-show="v2ray.tls !== 'none'" label="Alpn" label-position="on-border">
              <b-input v-model="v2ray.alpn" placeholder="h3,h2,http/1.1" expanded />
            </b-field>
            <b-field label-position="on-border">
              <template slot="label"> AllowInsecure </template>
              <b-tooltip v-show="v2ray.protocol === 'vless'" type="is-dark"
                :label="$t('server.messages.notRecommend', { name: 'VLESS' })" multilined position="is-right">
                <b-icon size="is-small" icon=" iconfont icon-help-circle-outline"
                  style="position: relative; top: 2px; right: 3px; font-weight: normal" />
              </b-tooltip>
              <b-select v-model="v2ray.allowInsecure" expanded required>
                <option :value="false">{{ $t("operations.no") }}</option>
                <option :value="true">{{ $t("operations.yes") }}</option>
              </b-select>
            </b-field>
            <b-field label="Network" label-position="on-border">
              <b-select v-model="v2ray.net" expanded @input="handleNetworkChange">
                <option value="tcp">TCP</option>
                <option value="kcp">mKCP</option>
                <option value="ws">WebSocket</option>
                <option value="h2">HTTP/2</option>
                <option value="quic">QUIC</option>
                <option value="grpc">gRPC</option>
                <option v-if="variant() === 'xray'" value="xhttp">xhttp</option>
              </b-select>
            </b-field>
            <b-field v-show="v2ray.net === 'tcp'" label="Header Type" label-position="on-border">
              <b-select v-model="v2ray.type" expanded>
                <option value="none">None</option>
                <option value="http">HTTP</option>
              </b-select>
            </b-field>
            <b-field v-show="v2ray.net === 'kcp' || v2ray.net === 'quic'" label="Header Type" label-position="on-border">
              <b-select v-model="v2ray.type" expanded>
                <option value="none">None</option>
                <option value="srtp">SRTP</option>
                <option value="utp">UTP</option>
                <option value="wechat-video">Wechat-Video</option>
                <option value="dtls">DTLS</option>
                <option value="wireguard">Wireguard</option>
              </b-select>
            </b-field>
            <b-field
              v-show="v2ray.net === 'ws' || v2ray.net === 'h2' || (v2ray.net === 'tcp' && v2ray.type === 'http') || v2ray.net === 'xhttp'"
              label="Path" label-position="on-border">
              <b-input v-model="v2ray.path" placeholder="/" expanded />
            </b-field>
            <b-field
              v-show="v2ray.net === 'ws' || v2ray.net === 'h2' || (v2ray.net === 'tcp' && v2ray.type === 'http') || v2ray.net === 'xhttp'"
              label="Host" label-position="on-border">
              <b-input v-model="v2ray.host" placeholder="Host" expanded />
            </b-field>
            <b-field
              v-show="v2ray.net === 'ws'"
              label="Max Early Data"
              label-position="on-border"
            >
              <b-input
                v-model="v2ray.maxEarlyData"
                type="number"
                placeholder="Max Early Data"
                expanded
              />
            </b-field>
            <b-field
              v-show="v2ray.net === 'ws'"
              label="Early Data Header Name"
              label-position="on-border"
            >
              <b-input
                v-model="v2ray.earlyDataHeaderName"
                placeholder="Early Data Header Name"
                expanded
              />
            </b-field>
            <b-field v-show="v2ray.net === 'grpc'" label="Service Name" label-position="on-border">
              <b-input ref="v2ray_service_name" v-model="v2ray.path" type="text" expanded />
            </b-field>
            <b-field
              v-show="v2ray.net === 'grpc'"
              label="MultiMode"
              label-position="on-border"
            >
              <b-switch v-model="v2ray.multiMode">
                {{ v2ray.multiMode ? $t('operations.yes') : $t('operations.no') }}
              </b-switch>
            </b-field>
            <b-field
              v-show="v2ray.net === 'grpc'"
              label="Idle Timeout"
              label-position="on-border"
            >
              <b-input
                v-model="v2ray.idleTimeout"
                type="number"
                placeholder="Idle Timeout (s)"
                expanded
              />
            </b-field>
            <b-field
              v-show="v2ray.net === 'grpc'"
              label="Health Check Timeout"
              label-position="on-border"
            >
              <b-input
                v-model="v2ray.healthCheckTimeout"
                type="number"
                placeholder="Health Check Timeout (s)"
                expanded
              />
            </b-field>
            <b-field
              v-show="v2ray.net === 'grpc'"
              label="Permit Without Stream"
              label-position="on-border"
            >
              <b-switch v-model="v2ray.permitWithoutStream">
                {{ v2ray.permitWithoutStream ? $t('operations.yes') : $t('operations.no') }}
              </b-switch>
            </b-field>
            <b-field
              v-show="v2ray.net === 'grpc'"
              label="Initial Windows Size"
              label-position="on-border"
            >
              <b-input
                v-model="v2ray.initialWindowsSize"
                type="number"
                placeholder="Initial Windows Size"
                expanded
              />
            </b-field>
            <b-field v-show="v2ray.net === 'quic'" label="Key" label-position="on-border">
              <b-input v-model="v2ray.key" placeholder="key" expanded />
            </b-field>
            <b-field v-show="v2ray.net === 'quic'" label="Security" label-position="on-border">
              <b-select v-model="v2ray.quicSecurity" expanded>
                <option value="none">none</option>
                <option value="aes-128-gcm">aes-128-gcm</option>
                <option value="chacha20-poly1305">chacha20-poly1305</option>
              </b-select>
            </b-field>
            <b-field v-show="v2ray.net === 'xhttp'" label="xhttp Mode" label-position="on-border">
              <b-select v-model="v2ray.xhttpMode" expanded>
                <option value="auto">auto</option>
                <option value="download">download</option>
                <option value="streaming">streaming</option>
                <option value="packet">packet</option>
                <option value="packet-up">packet-up</option>
                <option value="stream-up">stream-up</option>
                <option value="stream-one">stream-one</option>
              </b-select>
            </b-field>
            <b-field v-show="v2ray.net === 'xhttp' && v2ray.xhttpMode === 'packet'" label="xhttp RawJson"
              label-position="on-border">
              <b-input v-model="v2ray.xhttpRawJson" type="textarea" placeholder='{"scy": "chacha20-poly1305"}' expanded />
            </b-field>
          </div>
        </b-tab-item>

        <b-tab-item label="WireGuard">
          <div v-if="tabChoice === 1">
          <b-field label="Name" label-position="on-border">
            <b-input ref="wireguard_name" v-model="wireguard.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Address" label-position="on-border">
            <b-input ref="wireguard_address" v-model="wireguard.address" required placeholder="IP / HOST" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="wireguard_port" v-model="wireguard.port" required :placeholder="$t('configureServer.port')" type="number" expanded />
          </b-field>
          <b-field label="Public Key" label-position="on-border">
            <b-input ref="wireguard_public_key" v-model="wireguard.publicKey" required placeholder="Public Key" expanded />
          </b-field>
          <b-field label="Private Key" label-position="on-border">
            <b-input ref="wireguard_private_key" v-model="wireguard.privateKey" required placeholder="Private Key" expanded />
          </b-field>
          <b-field label="Address (Local)" label-position="on-border">
            <b-input ref="wireguard_local_address" v-model="wireguard.localAddress" placeholder="CIDR, e.g. 10.0.0.1/24" expanded />
          </b-field>
          <b-field label="DNS" label-position="on-border">
            <b-input ref="wireguard_dns" v-model="wireguard.dns" placeholder="DNS Server" expanded />
          </b-field>
          <b-field label="MTU" label-position="on-border">
            <b-input ref="wireguard_mtu" v-model="wireguard.mtu" type="number" placeholder="MTU" expanded />
          </b-field>
          <b-field label="Allowed IPs" label-position="on-border">
            <b-input ref="wireguard_allowed_ips" v-model="wireguard.allowedIPs" placeholder="0.0.0.0/0, ::/0" expanded />
          </b-field>
          <b-field label="Persistent Keepalive" label-position="on-border">
            <b-input ref="wireguard_persistent_keepalive" v-model="wireguard.persistentKeepalive" type="number" placeholder="Persistent Keepalive (s)" expanded />
          </b-field>
          <b-field label="Pre-shared Key" label-position="on-border">
            <b-input ref="wireguard_pre_shared_key" v-model="wireguard.preSharedKey" placeholder="Pre-shared Key" expanded />
          </b-field>
          <b-field label="Endpoint" label-position="on-border">
            <b-input ref="wireguard_endpoint" v-model="wireguard.endpoint" placeholder="Endpoint (optional, default same as Address:Port)" expanded />
          </b-field>
          </div>
        </b-tab-item>

        <b-tab-item label="SS">
          <div v-if="tabChoice === 2">
          <b-field label="Name" label-position="on-border">
            <b-input ref="ss_name" v-model="ss.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="ss_server" required placeholder="IP / HOST" v-model="ss.server" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="ss_port" required :placeholder="$t('configureServer.port')" type="number" v-model="ss.port"
              expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="ss_password" required :placeholder="$t('configureServer.password')" v-model="ss.password"
              expanded />
          </b-field>
          <b-field label="Method" label-position="on-border">
            <b-select v-model="ss.method" expanded required>
              <option value="2022-blake3-aes-128-gcm">
                2022-blake3-aes-128-gcm
              </option>
              <option value="2022-blake3-aes-256-gcm">
                2022-blake3-aes-256-gcm
              </option>
              <option value="2022-blake3-chacha20-poly1305">
                2022-blake3-chacha20-poly1305
              </option>
              <option value="aes-128-gcm">aes-128-gcm</option>
              <option value="aes-256-gcm">aes-256-gcm</option>
              <option value="chacha20-poly1305">chacha20-poly1305</option>
              <option value="chacha20-ietf-poly1305">
                chacha20-ietf-poly1305
              </option>
              <option value="plain">plain</option>
              <option value="none">none</option>
            </b-select>
          </b-field>
          <b-field label="Plugin" label-position="on-border">
            <b-select v-model="ss.plugin" expanded required>
              <option value="">Off</option>
              <option value="simple-obfs">simple-obfs</option>
              <option value="v2ray-plugin">v2ray-plugin</option>
            </b-select>
          </b-field>
          <b-field v-show="ss.plugin !== ''" label="Plugin Opts" label-position="on-border">
            <b-input v-model="ss.plugin_opts" placeholder="obfs=http;obfs-host=www.baidu.com" expanded />
          </b-field>
          <b-field :label="$t('setting.nodeBackend')" label-position="on-border">
            <b-select v-model="ss.backend" expanded>
              <option value="">{{ $t("setting.options.backendSystemDefault") }}</option>
              <option value="v2ray">{{ $t("setting.options.backendV2ray") }}</option>
            </b-select>
          </b-field>
        </b-tab-item>

        <b-tab-item label="SSR">
          <div v-if="tabChoice === 3">
          <b-field label="Name" label-position="on-border">
            <b-input ref="ssr_name" v-model="ssr.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="ssr_server" required placeholder="IP / HOST" v-model="ssr.server" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="ssr_port" required :placeholder="$t('configureServer.port')" type="number" v-model="ssr.port"
              expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="ssr_password" required :placeholder="$t('configureServer.password')" v-model="ssr.password"
              expanded />
          </b-field>
          <b-field label="Method" label-position="on-border">
            <b-select v-model="ssr.method" expanded required>
              <option value="aes-128-ctr">aes-128-ctr</option>
              <option value="aes-192-ctr">aes-192-ctr</option>
              <option value="aes-256-ctr">aes-256-ctr</option>
              <option value="aes-128-cfb">aes-128-cfb</option>
              <option value="aes-192-cfb">aes-192-cfb</option>
              <option value="aes-256-cfb">aes-256-cfb</option>
              <option value="rc4-md5">rc4-md5</option>
              <option value="chacha20-ietf">chacha20-ietf</option>
              <option value="chacha20">chacha20</option>
              <option value="salsa20">salsa20</option>
              <option value="none">none</option>
            </b-select>
          </b-field>
          <b-field label="Protocol" label-position="on-border">
            <b-select v-model="ssr.proto" expanded required>
              <option value="origin">origin</option>
              <option value="auth_sha1_v4">auth_sha1_v4</option>
              <option value="auth_aes128_md5">auth_aes128_md5</option>
              <option value="auth_aes128_sha1">auth_aes128_sha1</option>
              <option value="auth_chain_a">auth_chain_a</option>
              <option value="auth_chain_b">auth_chain_b</option>
            </b-select>
          </b-field>
          <b-field label="Protocol Param" label-position="on-border">
            <b-input v-model="ssr.protoParam" expanded />
          </b-field>
          <b-field label="Obfs" label-position="on-border">
            <b-select v-model="ssr.obfs" expanded required>
              <option value="plain">plain</option>
              <option value="http_simple">http_simple</option>
              <option value="http_post">http_post</option>
              <option value="tls1.2_ticket_auth">tls1.2_ticket_auth</option>
            </b-select>
          </b-field>
          <b-field label="Obfs Param" label-position="on-border">
            <b-input v-model="ssr.obfsParam" expanded />
          </b-field>
        </b-tab-item>

        <b-tab-item label="Trojan">
          <div v-if="tabChoice === 4">
          <b-field label="Name" label-position="on-border">
            <b-input ref="trojan_name" v-model="trojan.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="trojan_server" required placeholder="IP / HOST" v-model="trojan.server" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="trojan_port" required :placeholder="$t('configureServer.port')" type="number"
              v-model="trojan.port" expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="trojan_password" required :placeholder="$t('configureServer.password')"
              v-model="trojan.password" expanded />
          </b-field>
          <b-field label-position="on-border">
            <template slot="label"> AllowInsecure </template>
            <b-tooltip v-show="trojan.method !== 'origin' || trojan.obfs !== 'none'" type="is-dark"
              :label="$t('server.messages.notAllowInsecure', { name: 'Trojan-Go' })" multilined position="is-right">
              <b-icon size="is-small" icon=" iconfont icon-help-circle-outline"
                style="position: relative; top: 2px; right: 3px; font-weight: normal" />
            </b-tooltip>
            <b-select ref="trojan_allow_insecure" v-model="trojan.allowInsecure" expanded required>
              <option :value="false">{{ $t("operations.no") }}</option>
              <option :value="true"> {{ $t("operations.yes") }} </option>
            </b-select>
          </b-field>
          <b-field label="SNI(Peer)" label-position="on-border">
            <b-input v-model="trojan.peer" placeholder="SNI(Peer)" expanded />
          </b-field>
          <b-field :label="$t('setting.nodeBackend')" label-position="on-border">
            <b-select v-model="trojan.backend" expanded>
              <option value="">{{ $t("setting.options.backendSystemDefault") }}</option>
              <option value="v2ray">{{ $t("setting.options.backendV2ray") }}</option>
            </b-select>
          </b-field>
        </b-tab-item>

        <b-tab-item label="Juicity">
          <div v-if="tabChoice === 5">
          <b-field label="Name" label-position="on-border">
            <b-input ref="juicity_name" v-model="juicity.name" :placeholder="$t('configureServer.servername')"
              expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="juicity_server" required placeholder="IP / HOST" v-model="juicity.server" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="juicity_port" required :placeholder="$t('configureServer.port')" type="number"
              v-model="juicity.port" expanded />
          </b-field>
          <b-field label="UUID" label-position="on-border">
            <b-input ref="juicity_uuid" required placeholder="UUID" v-model="juicity.uuid" expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="juicity_password" required :placeholder="$t('configureServer.password')"
              v-model="juicity.password" expanded />
          </b-field>
          <b-field label="Congestion Control" label-position="on-border">
            <b-select ref="juicity_cc" v-model="juicity.cc" expanded required>
              <option value="bbr">bbr</option>
            </b-select>
          </b-field>
          <b-field label-position="on-border">
            <template slot="label"> AllowInsecure </template>
            <b-select ref="juicity_allow_insecure" v-model="juicity.allowInsecure" expanded required>
              <option :value="false">{{ $t("operations.no") }}</option>
              <option :value="true"> {{ $t("operations.yes") }} </option>
            </b-select>
          </b-field>
          <b-field label="SNI" label-position="on-border">
            <b-input v-model="juicity.sni" placeholder="SNI" expanded />
          </b-field>
          <b-field label="Pinned Cert Chain Sha256" label-position="on-border">
            <b-input v-model="juicity.pinnedCertchainSha256" :placeholder="$t('configureServer.pinnedCertchainSha256')"
              expanded />
          </b-field>
        </b-tab-item>

        <b-tab-item label="Tuic">
          <div v-if="tabChoice === 6">
          <b-field label="Name" label-position="on-border">
            <b-input ref="tuic_name" v-model="tuic.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="tuic_server" required placeholder="IP / HOST" v-model="tuic.server" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="tuic_port" required :placeholder="$t('configureServer.port')" type="number" v-model="tuic.port"
              expanded />
          </b-field>
          <b-field label="UUID" label-position="on-border">
            <b-input ref="tuic_uuid" required placeholder="UUID" v-model="tuic.uuid" expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="tuic_password" required :placeholder="$t('configureServer.password')" v-model="tuic.password"
              expanded />
          </b-field>
          <b-field label="Congestion Control" label-position="on-border">
            <b-select ref="tuic_cc" v-model="tuic.cc" expanded required>
              <option value="bbr">bbr</option>
            </b-select>
          </b-field>
          <b-field label-position="on-border">
            <template slot="label"> AllowInsecure </template>
            <b-select v-if="tuic.disableSni === false" ref="tuic_allow_insecure" v-model="tuic.allowInsecure" expanded
              required>
              <option :value="false">{{ $t("operations.no") }}</option>
              <option :value="true"> {{ $t("operations.yes") }} </option>
            </b-select>
          </b-field>
          <b-field label-position="on-border">
            <template slot="label"> DisableSni </template>
            <b-select ref="tuic_disable_sni" v-model="tuic.disableSni" expanded required>
              <option :value="false">{{ $t("operations.no") }}</option>
              <option :value="true"> {{ $t("operations.yes") }} </option>
            </b-select>
          </b-field>
          <b-field v-if="tuic.disableSni === false" label="SNI" label-position="on-border">
            <b-input v-model="tuic.sni" placeholder="SNI" expanded />
          </b-field>
          <b-field label="ALPN" label-position="on-border">
            <b-input v-model="tuic.alpn" placeholder="h3" expanded />
          </b-field>
          <b-field label-position="on-border">
            <template slot="label"> UDP relay mode </template>
            <b-select ref="tuic_udp_relay_mode" v-model="tuic.udpRelayMode" expanded required>
              <option value="native">native</option>
          </div>
              <option value="quic">quic</option>
            </b-select>
          </b-field>
        </b-tab-item>

        <b-tab-item label="Hysteria2">
          <div v-if="tabChoice === 7">
          <b-field label="Name" label-position="on-border">
            <b-input ref="hysteria2_name" v-model="hysteria2.name" :placeholder="$t('configureServer.servername')"
              expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="hysteria2_server" required placeholder="IP / HOST" v-model="hysteria2.server" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="hysteria2_port" required :placeholder="$t('configureServer.port')" type="number"
              v-model="hysteria2.port" expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="hysteria2_password" required :placeholder="$t('configureServer.password')"
              v-model="hysteria2.password" expanded />
          </b-field>
          <b-field label-position="on-border">
            <template slot="label"> AllowInsecure </template>
            <b-select ref="hysteria2_allow_insecure" v-model="hysteria2.allowInsecure" expanded required>
              <option :value="false">{{ $t("operations.no") }}</option>
              <option :value="true"> {{ $t("operations.yes") }} </option>
            </b-select>
          </b-field>
          <b-field label="SNI" label-position="on-border">
            <b-input v-model="hysteria2.sni" placeholder="SNI" expanded />
          </b-field>
          <b-field label="Obfs" label-position="on-border">
            <b-select v-model="hysteria2.obfs" expanded required>
              <option value="none">none</option>
              <option value="salamander">salamander</option>
            </b-select>
          </b-field>
          <b-field v-if="hysteria2.obfs !== 'none'" label="Obfs Password" label-position="on-border">
            <b-input v-model="hysteria2.obfsPassword" placeholder="Obfs Password" expanded />
          </b-field>
          <b-field label="Up Mbps" label-position="on-border">
            <b-input v-model="hysteria2.up" placeholder="e.g. 100" expanded />
          </b-field>
          <b-field label="Down Mbps" label-position="on-border">
            <b-input v-model="hysteria2.down" placeholder="e.g. 100" expanded />
          </b-field>
          <b-field label="Congestion" label-position="on-border">
            <b-select v-model="hysteria2.congestion" expanded>
              <option value="">default</option>
              <option value="bbr">bbr</option>
              <option value="cubic">cubic</option>
            </b-select>
          </b-field>
          <b-field label-position="on-border">
            <template slot="label">
              FinalMask
              <b-tooltip type="is-dark" :label="$t('server.messages.hysteria2FinalMaskInfo')" multilined position="is-right">
                <b-icon size="is-small" icon=" iconfont icon-help-circle-outline"
                  style="position: relative; top: 2px; right: 3px; font-weight: normal" />
              </b-tooltip>
            </template>
            <b-checkbox v-model="hysteria2.finalMask">
              Use native Xray implementation (requires Xray-core v26.1.23+)
          </div>
            </b-checkbox>
          </b-field>
        </b-tab-item>

        <b-tab-item label="HTTP">
          <div v-if="tabChoice === 8">
          <b-field label="Protocol" label-position="on-border">
            <b-select v-model="http.protocol" expanded>
              <option value="http">HTTP</option>
              <option value="https">HTTPS</option>
            </b-select>
          </b-field>
          <b-field label="Name" label-position="on-border">
            <b-input ref="http_name" v-model="http.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="http_host" required placeholder="IP / HOST" v-model="http.host" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="http_port" required :placeholder="$t('configureServer.port')" type="number" v-model="http.port"
              expanded />
          </b-field>
          <b-field label="Username" label-position="on-border">
            <b-input ref="http_username" v-model="http.username" :placeholder="$t('configureServer.username')"
              expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="http_password" v-model="http.password" :placeholder="$t('configureServer.password')"
              expanded />
          </div>
          </b-field>
        </b-tab-item>
        <b-tab-item label="SOCKS5">
          <div v-if="tabChoice === 9">
          <b-field label="Name" label-position="on-border">
            <b-input ref="socks5_name" v-model="socks5.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="socks5_host" required placeholder="IP / HOST" v-model="socks5.host" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="socks5_port" required :placeholder="$t('configureServer.port')" type="number" v-model="socks5.port"
              expanded />
          </b-field>
          <b-field label="Username" label-position="on-border">
            <b-input ref="socks5_username" v-model="socks5.username" :placeholder="$t('configureServer.username')"
              expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="socks5_password" v-model="socks5.password" :placeholder="$t('configureServer.password')"
              expanded />
          </div>
          </b-field>
        </b-tab-item>
        <b-tab-item label="AnyTLS">
          <div v-if="tabChoice === 10">
          <b-field label="Name" label-position="on-border">
            <b-input ref="anytls_name" v-model="anytls.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="anytls_host" required placeholder="IP / HOST" v-model="anytls.host" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="anytls_port" required :placeholder="$t('configureServer.port')" type="number" v-model="anytls.port"
              expanded />
          </b-field>
          <b-field label="Auth" label-position="on-border">
            <b-input ref="anytls_auth" required placeholder="Authentication Key" v-model="anytls.auth" expanded />
          </b-field>
          <b-field label="SNI(Peer)" label-position="on-border">
            <b-input ref="anytls_sni" placeholder="SNI / Peer (Optional)" v-model="anytls.sni" expanded />
          </b-field>
          <b-field label-position="on-border">
            <template slot="label"> AllowInsecure </template>
            <b-select ref="anytls_allow_insecure" v-model="anytls.allowInsecure" expanded required>
              <option :value="false">{{ $t("operations.no") }}</option>
          </div>
              <option :value="true">{{ $t("operations.yes") }}</option>
            </b-select>
          </b-field>
        </b-tab-item>
      </b-tabs>
    </section>
    <footer class="modal-card-foot flex-end">
      <button class="button" type="button" @click="$parent.close()">
        {{ $t("operations.cancel") }}
      </button>
      <button v-if="!readonly" class="button is-primary" @click="handleClickSubmit">
        {{ $t("operations.saveApply") }}
      </button>
    </footer>
  </div>
</template>

<script>
import { bracketIfIPv6, generateURL, handleResponse, parseURL } from "@/assets/js/utils";
import { Base64 } from "js-base64";
import { Decoder } from "@nuintun/qrcode";

export default {
  name: "ModalServer",
  props: {
    readonly: {
      type: Boolean,
      default: false,
    },
    which: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      tabChoice: 0,
      v2ray: {
        ps: "",
        add: "",
        port: "",
        id: "",
        flow: "",
        aid: "",
        net: "tcp",
        type: "none",
        host: "",
        path: "",
        tls: "none",
        quicSecurity: "none",
        fp: "",
        pbk: "",
        sid: "",
        spx: "",
        alpn: "",
        scy: "auto",
        v: "",
        allowInsecure: false,
        protocol: "vmess",
        key: "none",
        xhttpMode: "auto",
        xhttpRawJson: "",
        maxEarlyData: "",
        earlyDataHeaderName: "",
        multiMode: false,
        idleTimeout: "",
        healthCheckTimeout: "",
        permitWithoutStream: false,
        initialWindowsSize: "",
      },
      ss: {
        name: "",
        server: "",
        port: "",
        password: "",
        method: "chacha20-ietf-poly1305",
        plugin: "",
        plugin_opts: "",
        protocol: "ss",
        backend: "",
      },
      ssr: {
        server: "",
        port: "",
        proto: "origin",
        method: "aes-128-ctr",
        obfs: "plain",
        password: "",
        name: "",
        protoParam: "",
        obfsParam: "",
        protocol: "ssr",
      },
      trojan: {
        password: "",
        server: "",
        port: "",
        allowInsecure: false,
        peer: "",
        name: "",
        protocol: "trojan",
        backend: "",
      },
      juicity: {
        uuid: "",
        password: "",
        server: "",
        port: "",
        cc: "bbr",
        allowInsecure: false,
        sni: "",
        pinnedCertchainSha256: "",
        name: "",
        protocol: "juicity",
      },
      tuic: {
        uuid: "",
        password: "",
        server: "",
        port: "",
        cc: "bbr",
        allowInsecure: false,
        disableSni: false,
        sni: "",
        alpn: "h3",
        udpRelayMode: "native",
        name: "",
        protocol: "tuic",
      },
      hysteria2: {
        password: "",
        server: "",
        port: "",
        allowInsecure: false,
        obfs: "none",
        obfsPassword: "",
        sni: "",
        up: "",
        down: "",
        congestion: "",
        finalMask: false,
        name: "",
        protocol: "hysteria2",
      },
      http: {
        protocol: "http",
        name: "",
        host: "",
        port: "",
        username: "",
        password: "",
      },
      socks5: {
        name: "",
        host: "",
        port: "",
        username: "",
        password: "",
      },
      anytls: {
        auth: "",
        host: "",
        port: "",
        sni: "",
        allowInsecure: false,
        name: "",
        protocol: "anytls",
      },
      wireguard: {
        name: "",
        address: "",
        port: "",
        publicKey: "",
        privateKey: "",
        localAddress: "",
        dns: "",
        mtu: "",
        allowedIPs: "",
        persistentKeepalive: "",
        preSharedKey: "",
        endpoint: "",
        protocol: "wireguard",
      },
    };
  },
  mounted() {
    document
      .querySelector("#QRCodeImport")
      .addEventListener("change", this.handleFileChange, false);
    if (this.which) {
      this.$axios({
        url: apiRoot + "/sharingAddress",
        method: "get",
        params: {
          touch: {
            id: this.which.id,
            _type: this.which._type,
            sub: this.which.sub,
          },
        },
      }).then((res) => {
        handleResponse(res, this, () => {
          if (res.data.data.sharingAddress.toLowerCase().startsWith("vmess://")) {
            this.v2ray = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 0;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("vless://")
          ) {
            this.v2ray = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 0;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("wireguard://")
          ) {
            this.wireguard = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 1;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("ss://")
          ) {
            this.ss = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 2;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("ssr://")
          ) {
            this.ssr = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 3;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("trojan://")
          ) {
            this.trojan = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 4;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("juicity://")
          ) {
            this.juicity = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 5;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("tuic://")
          ) {
            this.tuic = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 6;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("hysteria2://") ||
            res.data.data.sharingAddress.toLowerCase().startsWith("hy2://")
          ) {
            this.hysteria2 = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 7;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("http://") ||
            res.data.data.sharingAddress.toLowerCase().startsWith("https://") ||
            res.data.data.sharingAddress.toLowerCase().startsWith("http-proxy://") ||
            res.data.data.sharingAddress.toLowerCase().startsWith("https-proxy://")
          ) {
            this.http = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 8;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("socks5://")
          ) {
            this.socks5 = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 9;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("anytls://")
          ) {
            this.anytls = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 10;
          }
          this.$nextTick(() => {
            if (this.readonly) {
              this.$refs.section
                .querySelectorAll("input, textarea")
                .forEach((x) => (x.disabled = true));
              this.$refs.section
                .querySelectorAll("select")
                .forEach((x) => (x.parentNode.className += " is-disabled"));
            }
          });
        });
      });
    }
  },
  methods: {
    variant() {
      return localStorage["variant"]?.toLowerCase() || "v2ray";
    },
    handleV2rayProtocolSwitch() {
      if (this.v2ray.protocol === "vless" && this.v2ray.scy === "auto") {
        this.v2ray.scy = "none";
      }
    },
    handleNetworkChange() {
      this.v2ray.type = "none";
      if (this.v2ray.tls === "none" && this.v2ray.net === "grpc") {
        this.$buefy.toast.open({
          message: this.$t("setting.messages.grpcShouldWithTls"),
          type: "is-warning",
          position: "is-top",
          queue: false,
          duration: 5000,
        });
        this.$nextTick(() => {
          this.v2ray.tls = "tls";
        });
      }
    },
    resolveURL(url) {
      if (url.toLowerCase().startsWith("vmess://")) {
        let obj = JSON.parse(
          Base64.decode(url.substring(url.indexOf("://") + 3))
        );
        console.log(obj);
        return {
          ps: obj.ps,
          add: obj.add,
          port: obj.port,
          id: obj.id,
          aid: obj.aid,
          scy: obj.scy,
          net: obj.net,
          type: obj.type,
          host: obj.host,
          path: obj.path,
          tls: obj.tls,
          allowInsecure: obj.allowInsecure || false,
          key: obj.key,
          quicSecurity: obj.quicSecurity,
          xhttpMode: obj.xhttpMode || "auto",
          xhttpRawJson: obj.xhttpRawJson || "",
          protocol: "vmess",
        };
      } else if (url.toLowerCase().startsWith("vless://")) {
        let u = parseURL(url);
        let o = {
          ps: decodeURIComponent(u.hash),
          add: u.host,
          port: u.port,
          id: u.username ? decodeURIComponent(u.username) : "",
          net: u.params.type || "tcp",
          type: u.params.headerType || "none",
          host: u.params.host || "",
          path: u.params.path || u.params.serviceName || "",
          alpn: u.params.alpn || "",
          sni: u.params.sni || "",
          tls: u.params.security || "none",
          quicSecurity: u.params.quicSecurity || "none",
          fp: u.params.fp || "",
          pbk: u.params.pbk || "",
          sid: u.params.sid || "",
          spx: u.params.spx || "",
          allowInsecure: u.params.allowInsecure === "true",
          flow: u.params.flow || u.params.flows || "",
          scy: u.params.encryption || "none",
          key: u.params.key,
          xhttpMode: u.params.xhttpMode || "auto",
          xhttpRawJson: u.params.xhttpRawJson || "",
          maxEarlyData: u.params.maxEarlyData || "",
          earlyDataHeaderName: u.params.earlyDataHeaderName || "",
          multiMode: u.params.multiMode === "true" || u.params.multiMode === "1",
          idleTimeout: u.params.idleTimeout || "",
          healthCheckTimeout: u.params.healthCheckTimeout || "",
          permitWithoutStream: u.params.permitWithoutStream === "true" || u.params.permitWithoutStream === "1",
          initialWindowsSize: u.params.initialWindowsSize || "",
          protocol: "vless",
        };
        if (o.net === "mkcp" || o.net === "kcp") {
          o.path = u.params.seed;
        }
        console.log(o);
        return o;
      } else if (url.toLowerCase().startsWith("ss://")) {
        let u = parseURL(url);
        let method, password, server, port, name, plugin;
        name = u.hash;
        plugin = u.params.plugin || "";
        if (u.username) {
          // SIP002
          let parts = Base64.decode(decodeURIComponent(u.username)).split(":");
          method = parts[0];
          password = parts.slice(1).join(":");
          server = u.host;
          port = u.port;
        } else {
          // Legacy
          let t = url.substring(url.indexOf("://") + 3);
          if (t.indexOf("#") > -1) t = t.substring(0, t.indexOf("#"));
          let decoded = Base64.decode(t);
          let parts = decoded.split("@");
          let methodAndPassword = parts[0].split(":");
          method = methodAndPassword[0];
          password = methodAndPassword[1];
          let serverAndPort = parts[1].split(":");
          server = serverAndPort[0];
          port = serverAndPort[1];
        }
        return {
          method,
          password,
          server,
          port,
          name,
          plugin,
          protocol: "ss",
          backend: u.params["v2raya-backend"] || "",
        };
      } else if (url.toLowerCase().startsWith("ssr://")) {
        let t = url.substring(6);
        if (t.indexOf("#") > -1) t = t.substring(0, t.indexOf("#"));
        url = Base64.decode(t);
        let arr = url.split("/?");
        let pre = arr[0].split(":");
        if (pre.length > 6) {
          //如果长度多于6，说明host中包含字符:，重新合并前几个分组到host去
          pre[pre.length - 6] = pre.slice(0, pre.length - 5).join(":");
          pre = pre.slice(pre.length - 6);
        }
        let query = {};
        if (arr[1]) {
          arr[1].split("&").forEach((x) => {
            let a = x.split("=");
            if (a.length > 1 && a[1]) {
              query[a[0]] = Base64.decode(a[1]);
            } else {
              query[a[0]] = "";
            }
          });
        }
        return {
          server: pre[0],
          port: pre[1],
          proto: pre[2],
          method: pre[3],
          obfs: pre[4],
          password: Base64.decode(pre[5]),
          name: query.remarks,
          protoParam: query.protoparam,
          obfsParam: query.obfsparam,
          protocol: "ssr",
        };
      } else if (url.toLowerCase().startsWith("trojan://")) {
        let u = parseURL(url);
        return {
          password: decodeURIComponent(u.username),
          server: u.host,
          port: u.port,
          allowInsecure: u.params.allowInsecure === "1",
          peer: u.params.sni || "",
          name: decodeURIComponent(u.hash),
          protocol: "trojan",
          backend: u.params["v2raya-backend"] || "",
        };
      } else if (url.toLowerCase().startsWith("juicity://")) {
        let u = parseURL(url);
        let password = decodeURIComponent(u.password);
        let uuid = decodeURIComponent(u.username);
        return {
          uuid: uuid,
          password: password,
          server: u.host,
          port: u.port,
          cc: u.params.congestion_control || "bbr",
          allowInsecure: u.params.allow_insecure === "1",
          sni: u.params.sni || "",
          pinnedCertchainSha256: u.params.pinned_certchain_sha256 || "",
          name: decodeURIComponent(u.hash),
          protocol: "juicity",
        };
      } else if (url.toLowerCase().startsWith("tuic://")) {
        let u = parseURL(url);
        let password = decodeURIComponent(u.password);
        let uuid = decodeURIComponent(u.username);
        return {
          uuid: uuid,
          password: password,
          server: u.host,
          port: u.port,
          cc: u.params.congestion_control || "bbr",
          allowInsecure: u.params.allow_insecure === "1",
          disableSni: u.params.disable_sni === "1",
          sni: u.params.sni || "",
          alpn: u.params.alpn || "h3",
          udpRelayMode: u.params.udp_relay_mode || "native",
          name: decodeURIComponent(u.hash),
          protocol: "tuic",
        };
      } else if (
        url.toLowerCase().startsWith("hysteria2://") ||
        url.toLowerCase().startsWith("hy2://")
      ) {
        let u = parseURL(url);
        return {
          password: decodeURIComponent(u.username),
          server: u.host,
          port: u.port,
          allowInsecure: u.params.insecure === "1",
          obfs: u.params.obfs || "none",
          obfsPassword: u.params["obfs-password"] || "",
          sni: u.params.sni || "",
          up: u.params.upmbps || "",
          down: u.params.downmbps || "",
          congestion: u.params.congestion || "",
          finalMask: u.params.finalmask === "1",
          name: decodeURIComponent(u.hash),
          protocol: "hysteria2",
        };
      } else if (
        url.toLowerCase().startsWith("http://") ||
        url.toLowerCase().startsWith("https://") ||
        url.toLowerCase().startsWith("http-proxy://") ||
        url.toLowerCase().startsWith("https-proxy://")
      ) {
        let u = parseURL(url);
        return {
          username: decodeURIComponent(u.username),
          password: decodeURIComponent(u.password),
          host: u.host,
          port: u.port,
          protocol: u.protocol.replace("-proxy", ""),
          name: decodeURIComponent(u.hash),
        };
      } else if (url.toLowerCase().startsWith("socks5://")) {
        let u = parseURL(url);
        return {
          username: decodeURIComponent(u.username),
          password: decodeURIComponent(u.password),
          host: u.host,
          port: u.port,
          protocol: u.protocol,
          name: decodeURIComponent(u.hash),
        };
      } else if (url.toLowerCase().startsWith("anytls://")) {
        let u = parseURL(url);
        let auth = u.username ? decodeURIComponent(u.username) : "";
        let sni = u.params.peer || u.params.sni || "";
        let allowInsecure = u.params.insecure === "1";
        return {
          name: decodeURIComponent(u.hash),
          host: u.host,
          port: u.port,
          auth: auth,
          sni: sni,
          allowInsecure: allowInsecure,
          protocol: "anytls",
        };
      } else if (url.toLowerCase().startsWith("wireguard://")) {
        let u = parseURL(url);
        return {
          name: decodeURIComponent(u.hash),
          address: u.host,
          port: u.port,
          privateKey: decodeURIComponent(u.username),
          publicKey: u.params.publicKey || "",
          localAddress: u.params.address || u.params.localAddress || "",
          dns: u.params.dns || "",
          mtu: u.params.mtu || "",
          allowedIPs: u.params.allowedIPs || "",
          persistentKeepalive: u.params.persistentKeepalive || "",
          preSharedKey: u.params.preSharedKey || "",
          endpoint: u.params.endpoint || "",
          protocol: "wireguard",
        };
      }
      return null;
    },
    generateURL(srcObj) {
      let obj = {};
      let query = {};
      let tmp = "";
      switch (srcObj.protocol) {
        case "vmess":
          obj = {
            v: "2",
            ps: srcObj.ps,
            add: srcObj.add,
            port: srcObj.port,
            id: srcObj.id,
            aid: srcObj.aid,
            scy: srcObj.scy,
            net: srcObj.net,
            type: srcObj.type,
            host: srcObj.host,
            path: srcObj.path,
            tls: srcObj.tls,
            allowInsecure: srcObj.allowInsecure,
            key: srcObj.key,
            quicSecurity: srcObj.quicSecurity,
            xhttpMode: srcObj.xhttpMode,
            xhttpRawJson: srcObj.xhttpRawJson,
          };
          return "vmess://" + Base64.encode(JSON.stringify(obj));
        case "vless":
          query = {
            type: srcObj.net,
            flow: srcObj.flow || "",
            security: srcObj.tls,
            fp: srcObj.fp || "",
            path: srcObj.path,
            host: srcObj.host,
            headerType: srcObj.type,
            sni: srcObj.sni,
            allowInsecure: srcObj.allowInsecure,
          };
          if (srcObj.alpn !== "") {
            query.alpn = srcObj.alpn;
          }
          if (srcObj.net === "ws") {
            if (srcObj.maxEarlyData) {
              query.maxEarlyData = srcObj.maxEarlyData;
            }
            if (srcObj.earlyDataHeaderName) {
              query.earlyDataHeaderName = srcObj.earlyDataHeaderName;
            }
          }
          if (srcObj.net === "grpc") {
            query.serviceName = srcObj.path;
            if (srcObj.multiMode) {
              query.multiMode = srcObj.multiMode;
            }
            if (srcObj.idleTimeout) {
              query.idleTimeout = srcObj.idleTimeout;
            }
            if (srcObj.healthCheckTimeout) {
              query.healthCheckTimeout = srcObj.healthCheckTimeout;
            }
            if (srcObj.permitWithoutStream) {
              query.permitWithoutStream = srcObj.permitWithoutStream;
            }
            if (srcObj.initialWindowsSize) {
              query.initialWindowsSize = srcObj.initialWindowsSize;
            }
          }
          if (srcObj.tls === "reality") {
            query.pbk = srcObj.pbk;
            query.sid = srcObj.sid;
            query.spx = srcObj.spx || "/";
          }
          if (srcObj.net === "mkcp" || srcObj.net === "kcp") {
            query.seed = srcObj.path;
          }
          if (srcObj.net === "quic") {
            query.key = srcObj.key;
            query.quicSecurity = srcObj.quicSecurity;
          }
          if (srcObj.net === "xhttp") {
            query.xhttpMode = srcObj.xhttpMode;
            if (srcObj.xhttpMode === "packet") {
              query.xhttpRawJson = srcObj.xhttpRawJson;
            }
          }
          return generateURL({
            protocol: "vless",
            username: srcObj.id,
            host: srcObj.add,
            port: srcObj.port,
            hash: srcObj.ps,
            params: query,
          });
        case "ss":
          /* ss://BASE64(method:password)@server:port#name */
          tmp = `ss://${Base64.encode(`${srcObj.method}:${srcObj.password}`)}@${bracketIfIPv6(srcObj.server)}:${srcObj.port}`;
          if (srcObj.plugin) {
            const plugin = [srcObj.plugin];
            if (srcObj.plugin === "v2ray-plugin") {
              if (srcObj.tls) {
                plugin.push("tls");
              }
              if (srcObj.mode !== "websocket") {
                plugin.push("mode=" + srcObj.mode);
              }
              if (srcObj.host) {
                plugin.push("host=" + srcObj.host);
              }
              if (srcObj.path) {
                if (!srcObj.path.startsWith("/")) {
                  srcObj.path = "/" + srcObj.path;
                }
                plugin.push("path=" + srcObj.path);
              }
              if (srcObj.impl) {
                plugin.push("impl=" + srcObj.impl);
              }
            } else {
              plugin.push("obfs=" + srcObj.obfs);
              plugin.push("obfs-host=" + srcObj.host);
              if (srcObj.obfs === "http") {
                plugin.push("obfs-path=" + srcObj.path);
              }
              if (srcObj.impl) {
                plugin.push("impl=" + srcObj.impl);
              }
            }
            tmp += `?plugin=${encodeURIComponent(plugin.join(";"))}`;
          }
          if (srcObj.backend) {
            tmp += (tmp.includes("?") ? "&" : "?") + `v2raya-backend=${encodeURIComponent(srcObj.backend)}`;
          }
          tmp += srcObj.name.length ? `#${encodeURIComponent(srcObj.name)}` : "";
          return tmp;

        case "ssr":
          /* ssr://server:port:proto:method:obfs:URLBASE64(password)/?remarks=URLBASE64(remarks)&protoparam=URLBASE64(protoparam)&obfsparam=URLBASE64(obfsparam)) */
          return `ssr://${Base64.encode(
            `${bracketIfIPv6(srcObj.server)}:${srcObj.port}:${srcObj.proto}:${srcObj.method}:${srcObj.obfs
            }:${Base64.encodeURI(srcObj.password)}/?remarks=${Base64.encodeURI(
              srcObj.name
            )}&protoparam=${Base64.encodeURI(
              srcObj.protoParam
            )}&obfsparam=${Base64.encodeURI(srcObj.obfsParam)}`
          )}`;
        case "trojan":
          /* trojan://password@server:port?allowInsecure=1&sni=sni#URIESCAPE(name) */
          query = {
            type: srcObj.net || "tcp",
            allowInsecure: srcObj.allowInsecure ? "1" : "0",
          };
          if (srcObj.peer !== "") {
            query.sni = srcObj.peer;
          }
          if (srcObj.backend) {
            query["v2raya-backend"] = srcObj.backend;
          }
          return generateURL({
            protocol: "trojan",
            username: srcObj.password,
            host: srcObj.server,
            port: srcObj.port,
            hash: srcObj.name,
            params: query,
          });
        case "juicity":
          query = {
            allow_insecure: srcObj.allowInsecure ? "1" : "0",
            congestion_control: srcObj.cc,
          };
          if (srcObj.sni !== "") {
            query.sni = srcObj.sni;
          }
          if (srcObj.pinnedCertchainSha256 !== "") {
            query.pinned_certchain_sha256 = srcObj.pinnedCertchainSha256;
          }
          return generateURL({
            protocol: "juicity",
            username: srcObj.uuid,
            password: srcObj.password,
            host: srcObj.server,
            port: srcObj.port,
            hash: srcObj.name,
            params: query,
          });
        case "tuic":
          query = {
            allow_insecure: srcObj.allowInsecure ? "1" : "0",
            congestion_control: srcObj.cc,
            disable_sni: srcObj.disableSni ? "1" : "0",
            alpn: srcObj.alpn,
            udp_relay_mode: srcObj.udpRelayMode,
          };
          if (srcObj.sni !== "") {
            query.sni = srcObj.sni;
          }
          return generateURL({
            protocol: "tuic",
            username: srcObj.uuid,
            password: srcObj.password,
            host: srcObj.server,
            port: srcObj.port,
            hash: srcObj.ps || srcObj.name,
            params: query,
          });
        case "hysteria2":
          query = {
            insecure: srcObj.allowInsecure ? "1" : "0",
          };
          if (srcObj.sni !== "") {
            query.sni = srcObj.sni;
          }
          if (srcObj.obfs !== "none") {
            query.obfs = srcObj.obfs;
            query["obfs-password"] = srcObj.obfsPassword;
          }
          if (srcObj.up) query.upmbps = srcObj.up;
          if (srcObj.down) query.downmbps = srcObj.down;
          if (srcObj.congestion) query.congestion = srcObj.congestion;
          if (srcObj.finalMask) query.finalmask = "1";
          return generateURL({
            protocol: "hysteria2",
            username: srcObj.password,
            host: srcObj.server,
            port: srcObj.port,
            hash: srcObj.ps || srcObj.name,
            params: query,
          });
        case "http":
        case "https":
          tmp = {
            protocol: srcObj.protocol + "-proxy",
            host: srcObj.host,
            port: srcObj.port,
            hash: srcObj.name,
          };
          if (srcObj.username && srcObj.password) {
            Object.assign(tmp, {
              username: srcObj.username,
              password: srcObj.password,
            });
          }
          return generateURL(tmp);
        case "socks5":
          tmp = {
            protocol: "socks5",
            host: srcObj.host,
            port: srcObj.port,
            hash: srcObj.name,
          };
          if (srcObj.username && srcObj.password) {
            Object.assign(tmp, {
              username: srcObj.username,
              password: srcObj.password,
            });
          }
          return generateURL(tmp);
        case "anytls":
          query = {};
          if (srcObj.sni) {
            query.peer = srcObj.sni;
          }
          if (srcObj.allowInsecure) {
            query.insecure = "1";
          }
          return generateURL({
            protocol: "anytls",
            username: srcObj.auth,
            host: srcObj.host,
            port: srcObj.port,
            hash: srcObj.name,
            params: query,
          });
        case "wireguard":
          query = {};
          if (srcObj.publicKey) {
            query.publicKey = srcObj.publicKey;
          }
          if (srcObj.localAddress) {
            query.address = srcObj.localAddress;
          }
          if (srcObj.dns) {
            query.dns = srcObj.dns;
          }
          if (srcObj.mtu) {
            query.mtu = srcObj.mtu;
          }
          if (srcObj.allowedIPs) {
            query.allowedIPs = srcObj.allowedIPs;
          }
          if (srcObj.persistentKeepalive) {
            query.persistentKeepalive = srcObj.persistentKeepalive;
          }
          if (srcObj.preSharedKey) {
            query.preSharedKey = srcObj.preSharedKey;
          }
          if (srcObj.endpoint) {
            query.endpoint = srcObj.endpoint;
          }
          return generateURL({
            protocol: "wireguard",
            username: srcObj.privateKey,
            host: srcObj.address,
            port: srcObj.port,
            hash: srcObj.name,
            params: query,
          });
      }
      return null;
    },
    handleFileChange(e) {
      const file = e.target.files[0];
      let elem = document.querySelector("#QRCodeImport");
      // eslint-disable-next-line no-self-assign
      elem.outerHTML = elem.outerHTML;
      this.$nextTick(() => {
        document
          .querySelector("#QRCodeImport")
          .addEventListener("change", this.handleFileChange, false);
      });
      // console.log(file);
      if (!file.type.match(/image\/.*/)) {
        this.$buefy.toast.open({
          message: this.$t("import.qrcodeError"),
          type: "is-warning",
          position: "is-top",
          queue: false,
        });
        return;
      }
      const reader = new FileReader();
      reader.onload = (e) => {
        // target.result 该属性表示目标对象的DataURL
        // console.log(e.target.result);
        const file = e.target.result;
        const qrcode = new Decoder();
        qrcode
          .scan(file)
          .then((result) => {
            console.log(result);
            this.resolveURL(result.data);
          })
          .catch((error) => {
            console.error(error);
            this.$buefy.toast.open({
              message: this.$t("import.qrcodeError"),
              type: "is-warning",
              position: "is-top",
              queue: false,
            });
          });
      };
      reader.readAsDataURL(file);
    },
    async handleClickSubmit() {
      let valid = true;
      const prefixMap = {
        0: "v2ray_",
        1: "wireguard_",
        2: "ss_",
        3: "ssr_",
        4: "trojan_",
        5: "juicity_",
        6: "tuic_",
        7: "hysteria2_",
        8: "http_",
        9: "socks5_",
        10: "anytls_",
      };
      const currentPrefix = prefixMap[this.tabChoice];
      for (let k in this.$refs) {
        if (!this.$refs.hasOwnProperty(k)) continue;
        if (!k.startsWith(currentPrefix)) continue;

        let x = this.$refs[k];
        if (!x) continue;
        let el = x.$el || x;
        if (
          el.offsetParent &&
          x.hasOwnProperty("checkHtml5Validity") &&
          typeof x.checkHtml5Validity === "function" &&
          !x.checkHtml5Validity()
        ) {
          console.error("validate failed", k, x);
          valid = false;
        }
      }
      if (!valid) {
        return;
      }
      let coded = "";
      if (this.tabChoice === 0) {
        const { allowInsecure, protocol } = this.v2ray;
        if (allowInsecure) {
          const result = await new Promise((resolve) => {
            this.$buefy.dialog.confirm({
              title: this.$t("InSecureConfirm.title"),
              message: this.$t("InSecureConfirm.message"),
              confirmText: this.$t("InSecureConfirm.confirm"),
              cancelText: this.$t("InSecureConfirm.cancel"),
              type: "is-danger",
              hasIcon: true,
              onConfirm: () => resolve(true),
              onCancel: () => resolve(false),
            });
          });
          if (!result) {
            return;
          }
        }
        coded = this.generateURL(this.v2ray);
      } else if (this.tabChoice === 1) {
        coded = this.generateURL(this.wireguard);
      } else if (this.tabChoice === 2) {
        coded = this.generateURL(this.ss);
      } else if (this.tabChoice === 3) {
        coded = this.generateURL(this.ssr);
      } else if (this.tabChoice === 4) {
        coded = this.generateURL(this.trojan);
      } else if (this.tabChoice === 5) {
        coded = this.generateURL(this.juicity);
      } else if (this.tabChoice === 6) {
        coded = this.generateURL(this.tuic);
      } else if (this.tabChoice === 7) {
        coded = this.generateURL(this.hysteria2);
      } else if (this.tabChoice === 8) {
        coded = this.generateURL(this.http);
      } else if (this.tabChoice === 9) {
        coded = this.generateURL(this.socks5);
      } else if (this.tabChoice === 10) {
        coded = this.generateURL(this.anytls);
      }
      this.$emit("submit", coded);
    },
  },
};
</script>

<style lang="scss">
.is-twitter .is-active a {
  color: #4099ff !important;
}

.modal-card {
  max-width: 720px;
  width: 100%;
}

.b-tabs.is-vertical {
  display: flex !important;
  flex-direction: row !important;
  align-items: stretch;

  .tabs {
    flex: 0 0 100px !important;
    min-width: 100px !important;
    max-width: 100px !important;

    ul {
      width: 100%;
    }

    li {
      font-size: 0.85rem;

      a {
        padding: 0.5em 0.2em !important;
        justify-content: flex-start;
      }
    }
  }

  .tab-content {
    flex: 1 1 0 !important;
    /* Allow to shrink to zero */
    min-width: 0 !important;
    padding: 0.5rem 0.2rem 0.5rem 0.5rem !important;
    overflow-x: hidden;

    .control,
    .field,
    input,
    select {
      max-width: 100% !important;
      min-width: 0 !important;
    }
  }
}

@media screen and (max-width: 768px) {
  .b-tabs.is-vertical {
    .tabs {
      flex: 0 0 80px !important;
      min-width: 80px !important;
      max-width: 80px !important;
    }
  }
}
</style>
