<template>
  <div class="modal-card" style="max-width: 520px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $tc("configureServer.title", readonly ? 2 : 1) }}
      </p>
    </header>
    <section ref="section" :class="{ 'modal-card-body': true }">
      <b-tabs v-model="tabChoice" position="is-centered" class="block" type="is-boxed is-twitter same-width-5">
        <b-tab-item label="VMESS">
          <b-field label="Name" label-position="on-border">
            <b-input ref="v2ray_name" v-model="v2ray.ps" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="v2ray_add" v-model="v2ray.add" required placeholder="IP / HOST" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="v2ray_port" v-model="v2ray.port" required :placeholder="$t('configureServer.port')"
              type="number" expanded />
          </b-field>
          <b-field label="ID" label-position="on-border">
            <b-input ref="v2ray_id" v-model="v2ray.id" required placeholder="UserID" expanded />
          </b-field>
          <b-field label="AlterID" label-position="on-border">
            <b-input ref="v2ray_aid" v-model="v2ray.aid" placeholder="AlterID" type="number" min="0" max="65535"
              expanded />
          </b-field>
          <b-field label="Security" label-position="on-border">
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
            </b-select>
          </b-field>
          <b-field v-if="v2ray.tls !== 'none'" label="SNI" label-position="on-border">
            <b-input ref="v2ray_sni" v-model="v2ray.sni" placeholder="SNI" expanded />
          </b-field>
          <b-field v-show="v2ray.tls === 'tls'" label="uTLS fingerprint" label-position="on-border">
            <b-select ref="v2ray_fp" v-model="v2ray.fp" expanded>
              <option value="">empty</option>
              <option value="chrome">chrome</option>
              <option value="firefox">firefox</option>
              <option value="safari">safari</option>
              <option value="ios">ios</option>
              <option value="android">android</option>
              <option value="edge">edge</option>
              <option value="random">random</option>
              <option value="randomized">randomized</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.tls === 'tls'" label="Alpn" label-position="on-border">
            <b-input v-model="v2ray.alpn" placeholder="h3,h2,http/1.1" expanded />
          </b-field>
          <b-field v-show="v2ray.tls !== 'none'" label="Pinned Cert SHA256" label-position="on-border">
            <b-input v-model="v2ray.pinnedPeerCertSha256" :placeholder="$t('pinnedPeerCertSha256')" expanded />
          </b-field>
          <b-field v-show="v2ray.tls !== 'none'" label="Verify Peer Cert By Name" label-position="on-border">
            <b-input v-model="v2ray.verifyPeerCertByName" :placeholder="$t('verifyPeerCertByName')" expanded />
          </b-field>
          <b-field label="Network" label-position="on-border">
            <b-select ref="v2ray_net" v-model="v2ray.net" expanded required @input="handleNetworkChange">
              <option value="tcp">TCP</option>
              <option value="kcp">mKCP</option>
              <option value="ws">WebSocket</option>
              <option value="h2">HTTP/2</option>
              <option value="grpc">gRPC</option>
              <option value="quic">QUIC</option>
              <option value="xhttp">XHTTP</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.net === 'tcp'" label="Type" label-position="on-border">
            <b-select v-model="v2ray.type" expanded>
              <option value="none">{{ $t("configureServer.noObfuscation") }}</option>
              <option value="http">{{ $t("configureServer.httpObfuscation") }}</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.net === 'kcp' || v2ray.net === 'quic'" label="Type" label-position="on-border">
            <b-select v-model="v2ray.type" expanded>
              <option value="none">{{ $t("configureServer.noObfuscation") }}</option>
              <option value="srtp">{{ $t("configureServer.srtpObfuscation") }}</option>
              <option value="utp">{{ $t("configureServer.utpObfuscation") }}</option>
              <option value="wechat-video">{{ $t("configureServer.wechatVideoObfuscation") }}</option>
              <option value="dtls">{{ `${$t("configureServer.dtlsObfuscation")}(${$t("configureServer.forceTLS")})` }}</option>
              <option value="wireguard">{{ $t("configureServer.wireguardObfuscation") }}</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.net === 'ws' || v2ray.net === 'h2' || v2ray.net === 'xhttp' || v2ray.tls === 'tls' || (v2ray.net === 'tcp' && v2ray.type === 'http')" label="Host" label-position="on-border">
            <b-input v-model="v2ray.host" :placeholder="$t('configureServer.hostObfuscation')" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'ws' || v2ray.net === 'h2' || (v2ray.net === 'tcp' && v2ray.type === 'http')" label="Path" label-position="on-border">
            <b-input v-model="v2ray.path" :placeholder="$t('configureServer.pathObfuscation')" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'ws'" label="Max Early Data" label-position="on-border">
            <b-input v-model="v2ray.maxEarlyData" type="number" placeholder="Max Early Data" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'ws'" label="Early Data Header Name" label-position="on-border">
            <b-input v-model="v2ray.earlyDataHeaderName" placeholder="Early Data Header Name" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'mkcp' || v2ray.net === 'kcp'" label="Seed" label-position="on-border">
            <b-input v-model="v2ray.path" :placeholder="$t('configureServer.seedObfuscation')" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'grpc'" label="Service Name" label-position="on-border">
            <b-input ref="v2ray_service_name" v-model="v2ray.path" type="text" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'grpc'" label="MultiMode" label-position="on-border">
            <b-switch v-model="v2ray.multiMode">{{ v2ray.multiMode ? $t('operations.yes') : $t('operations.no') }}</b-switch>
          </b-field>
          <b-field v-show="v2ray.net === 'grpc'" label="Idle Timeout" label-position="on-border">
            <b-input v-model="v2ray.idleTimeout" type="number" placeholder="Idle Timeout (s)" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'grpc'" label="Health Check Timeout" label-position="on-border">
            <b-input v-model="v2ray.healthCheckTimeout" type="number" placeholder="Health Check Timeout (s)" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'grpc'" label="Permit Without Stream" label-position="on-border">
            <b-switch v-model="v2ray.permitWithoutStream">{{ v2ray.permitWithoutStream ? $t('operations.yes') : $t('operations.no') }}</b-switch>
          </b-field>
          <b-field v-show="v2ray.net === 'grpc'" label="Initial Windows Size" label-position="on-border">
            <b-input v-model="v2ray.initialWindowsSize" type="number" placeholder="Initial Windows Size" expanded />
          </b-field>
          <!-- XHTTP fields (VMess) -->
          <b-field v-show="v2ray.net === 'xhttp'" label="Path" label-position="on-border">
            <b-input v-model="v2ray.path" :placeholder="$t('configureServer.pathObfuscation')" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="Mode" label-position="on-border">
            <b-select v-model="v2ray.xhttpMode" expanded>
              <option value="auto">auto</option>
              <option value="packet-up">packet-up</option>
              <option value="stream-up">stream-up</option>
              <option value="stream-one">stream-one</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="Uplink HTTP Method" label-position="on-border">
            <b-select v-model="v2ray.uplinkHTTPMethod" expanded>
              <option value="">default (POST)</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="PATCH">PATCH</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="noGRPCHeader" label-position="on-border">
            <b-switch v-model="v2ray.noGRPCHeader">{{ v2ray.noGRPCHeader ? $t('operations.yes') : $t('operations.no') }}</b-switch>
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="noSSEHeader" label-position="on-border">
            <b-switch v-model="v2ray.noSSEHeader">{{ v2ray.noSSEHeader ? $t('operations.yes') : $t('operations.no') }}</b-switch>
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="scMaxEachPostBytes (From-To)" label-position="on-border">
            <b-input v-model="v2ray.scMaxEachPostBytesFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.scMaxEachPostBytesTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="scMinPostsIntervalMs (From-To)" label-position="on-border">
            <b-input v-model="v2ray.scMinPostsIntervalFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.scMinPostsIntervalTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="scMaxBufferedPosts" label-position="on-border">
            <b-input v-model="v2ray.scMaxBufferedPosts" type="number" placeholder="scMaxBufferedPosts" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="scStreamUpServerSecs (From-To)" label-position="on-border">
            <b-input v-model="v2ray.scStreamUpServerFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.scStreamUpServerTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xPaddingBytes (From-To)" label-position="on-border">
            <b-input v-model="v2ray.xPaddingBytesFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.xPaddingBytesTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xmux maxConcurrency (From-To)" label-position="on-border">
            <b-input v-model="v2ray.xmuxMaxConcurFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.xmuxMaxConcurTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xmux maxConnections (From-To)" label-position="on-border">
            <b-input v-model="v2ray.xmuxMaxConnFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.xmuxMaxConnTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xmux cMaxReuseTimes (From-To)" label-position="on-border">
            <b-input v-model="v2ray.xmuxCMaxReuseFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.xmuxCMaxReuseTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xmux hMaxRequestTimes (From-To)" label-position="on-border">
            <b-input v-model="v2ray.xmuxHMaxReqFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.xmuxHMaxReqTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xmux hMaxReusableSecs (From-To)" label-position="on-border">
            <b-input v-model="v2ray.xmuxHMaxReusableFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.xmuxHMaxReusableTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xmux hKeepAlivePeriod" label-position="on-border">
            <b-input v-model="v2ray.xmuxHKeepAlive" type="number" placeholder="hKeepAlivePeriod" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="Custom Headers" label-position="on-border">
            <div style="width:100%">
              <div v-for="(hdr, idx) in v2ray.xhttpHeaders" :key="idx" style="display:flex;gap:4px;margin-bottom:4px">
                <b-input v-model="hdr.key" placeholder="Header-Name" expanded />
                <b-input v-model="hdr.value" placeholder="value" expanded />
                <b-button type="is-danger is-light" icon-left="delete" size="is-small" @click="v2ray.xhttpHeaders.splice(idx,1)" />
              </div>
              <b-button size="is-small" icon-left="plus" @click="v2ray.xhttpHeaders.push({key:'',value:''})">Add Header</b-button>
            </div>
          </b-field>
          <!-- QUIC -->
          <b-field v-show="v2ray.net === 'quic'" label="QUIC Security" label-position="on-border">
            <b-select v-model="v2ray.quicSecurity" expanded>
              <option value="none">none</option>
              <option value="aes-128-gcm">aes-128-gcm</option>
              <option value="chacha20-poly1305">chacha20-poly1305</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.net === 'quic'" label="Key" label-position="on-border">
            <b-input ref="v2ray_key" v-model="v2ray.key" :placeholder="$t('configureServer.password')" expanded />
          </b-field>
        </b-tab-item>
        <b-tab-item label="VLESS">
          <b-field label="Name" label-position="on-border">
            <b-input v-model="v2ray.ps" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input v-model="v2ray.add" required placeholder="IP / HOST" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input v-model="v2ray.port" required :placeholder="$t('configureServer.port')" type="number" expanded />
          </b-field>
          <b-field label="ID" label-position="on-border">
            <b-input v-model="v2ray.id" required placeholder="UserID" expanded />
          </b-field>
          <b-field v-show="v2ray.type !== 'dtls'" label="TLS" label-position="on-border">
            <b-select v-model="v2ray.tls" expanded @input="handleNetworkChange">
              <option value="none">{{ $t("setting.options.off") }}</option>
              <option value="tls">tls</option>
              <option v-if="variant() === 'xray'" value="reality">reality</option>
              <option v-if="variant() === 'xray'" value="xtls">xtls</option>
            </b-select>
          </b-field>
          <b-field v-if="v2ray.tls !== 'none'" label="SNI" label-position="on-border">
            <b-input v-model="v2ray.sni" placeholder="SNI" expanded />
          </b-field>
          <b-field v-show="v2ray.tls === 'tls' || v2ray.tls === 'reality'" label="uTLS fingerprint" label-position="on-border">
            <b-select v-model="v2ray.fp" expanded>
              <option value="">empty</option>
              <option value="chrome">chrome</option>
              <option value="firefox">firefox</option>
              <option value="safari">safari</option>
              <option value="ios">ios</option>
              <option value="android">android</option>
              <option value="edge">edge</option>
              <option value="random">random</option>
              <option value="randomized">randomized</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.tls === 'tls'" label="Alpn" label-position="on-border">
            <b-input v-model="v2ray.alpn" placeholder="h3,h2,http/1.1" expanded />
          </b-field>
          <b-field v-if="v2ray.tls !== 'none'" label="Flow" label-position="on-border">
            <b-input v-model="v2ray.flow" placeholder="Flow" expanded />
          </b-field>
          <b-field v-show="v2ray.tls === 'reality'" label="Public Key (pbk)" label-position="on-border">
            <b-input v-model="v2ray.pbk" placeholder="Public Key" expanded />
          </b-field>
          <b-field v-show="v2ray.tls === 'reality'" label="Short ID (sid)" label-position="on-border">
            <b-input v-model="v2ray.sid" placeholder="Short ID" expanded />
          </b-field>
          <b-field v-show="v2ray.tls === 'reality'" label="Spider X (spx)" label-position="on-border">
            <b-input v-model="v2ray.spx" placeholder="Spider X" expanded />
          </b-field>
          <b-field v-show="v2ray.tls !== 'none'" label="Pinned Cert SHA256" label-position="on-border">
            <b-input v-model="v2ray.pinnedPeerCertSha256" :placeholder="$t('pinnedPeerCertSha256')" expanded />
          </b-field>
          <b-field v-show="v2ray.tls !== 'none'" label="Verify Peer Cert By Name" label-position="on-border">
            <b-input v-model="v2ray.verifyPeerCertByName" :placeholder="$t('verifyPeerCertByName')" expanded />
          </b-field>
          <b-field label="Network" label-position="on-border">
            <b-select v-model="v2ray.net" expanded required @input="handleNetworkChange">
              <option value="tcp">TCP</option>
              <option value="kcp">mKCP</option>
              <option value="ws">WebSocket</option>
              <option value="h2">HTTP/2</option>
              <option value="grpc">gRPC</option>
              <option value="quic">QUIC</option>
              <option value="xhttp">XHTTP</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.net === 'tcp'" label="Type" label-position="on-border">
            <b-select v-model="v2ray.type" expanded>
              <option value="none">{{ $t("configureServer.noObfuscation") }}</option>
              <option value="http">{{ $t("configureServer.httpObfuscation") }}</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.net === 'kcp' || v2ray.net === 'quic'" label="Type" label-position="on-border">
            <b-select v-model="v2ray.type" expanded>
              <option value="none">{{ $t("configureServer.noObfuscation") }}</option>
              <option value="srtp">{{ $t("configureServer.srtpObfuscation") }}</option>
              <option value="utp">{{ $t("configureServer.utpObfuscation") }}</option>
              <option value="wechat-video">{{ $t("configureServer.wechatVideoObfuscation") }}</option>
              <option value="dtls">{{ `${$t("configureServer.dtlsObfuscation")}(${$t("configureServer.forceTLS")})` }}</option>
              <option value="wireguard">{{ $t("configureServer.wireguardObfuscation") }}</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.net === 'ws' || v2ray.net === 'h2' || v2ray.net === 'xhttp' || v2ray.tls === 'tls' || v2ray.tls === 'reality' || (v2ray.net === 'tcp' && v2ray.type === 'http')" label="Host" label-position="on-border">
            <b-input v-model="v2ray.host" :placeholder="$t('configureServer.hostObfuscation')" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'ws' || v2ray.net === 'h2' || (v2ray.net === 'tcp' && v2ray.type === 'http')" label="Path" label-position="on-border">
            <b-input v-model="v2ray.path" :placeholder="$t('configureServer.pathObfuscation')" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'ws'" label="Max Early Data" label-position="on-border">
            <b-input v-model="v2ray.maxEarlyData" type="number" placeholder="Max Early Data" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'ws'" label="Early Data Header Name" label-position="on-border">
            <b-input v-model="v2ray.earlyDataHeaderName" placeholder="Early Data Header Name" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'mkcp' || v2ray.net === 'kcp'" label="Seed" label-position="on-border">
            <b-input v-model="v2ray.path" :placeholder="$t('configureServer.seedObfuscation')" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'grpc'" label="Service Name" label-position="on-border">
            <b-input v-model="v2ray.path" type="text" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'grpc'" label="MultiMode" label-position="on-border">
            <b-switch v-model="v2ray.multiMode">{{ v2ray.multiMode ? $t('operations.yes') : $t('operations.no') }}</b-switch>
          </b-field>
          <b-field v-show="v2ray.net === 'grpc'" label="Idle Timeout" label-position="on-border">
            <b-input v-model="v2ray.idleTimeout" type="number" placeholder="Idle Timeout (s)" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'grpc'" label="Health Check Timeout" label-position="on-border">
            <b-input v-model="v2ray.healthCheckTimeout" type="number" placeholder="Health Check Timeout (s)" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'grpc'" label="Permit Without Stream" label-position="on-border">
            <b-switch v-model="v2ray.permitWithoutStream">{{ v2ray.permitWithoutStream ? $t('operations.yes') : $t('operations.no') }}</b-switch>
          </b-field>
          <b-field v-show="v2ray.net === 'grpc'" label="Initial Windows Size" label-position="on-border">
            <b-input v-model="v2ray.initialWindowsSize" type="number" placeholder="Initial Windows Size" expanded />
          </b-field>
          <!-- XHTTP fields (VLESS) -->
          <b-field v-show="v2ray.net === 'xhttp'" label="Path" label-position="on-border">
            <b-input v-model="v2ray.path" :placeholder="$t('configureServer.pathObfuscation')" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="Mode" label-position="on-border">
            <b-select v-model="v2ray.xhttpMode" expanded>
              <option value="auto">auto</option>
              <option value="packet-up">packet-up</option>
              <option value="stream-up">stream-up</option>
              <option value="stream-one">stream-one</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="Uplink HTTP Method" label-position="on-border">
            <b-select v-model="v2ray.uplinkHTTPMethod" expanded>
              <option value="">default (POST)</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="PATCH">PATCH</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="noGRPCHeader" label-position="on-border">
            <b-switch v-model="v2ray.noGRPCHeader">{{ v2ray.noGRPCHeader ? $t('operations.yes') : $t('operations.no') }}</b-switch>
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="noSSEHeader" label-position="on-border">
            <b-switch v-model="v2ray.noSSEHeader">{{ v2ray.noSSEHeader ? $t('operations.yes') : $t('operations.no') }}</b-switch>
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="scMaxEachPostBytes (From-To)" label-position="on-border">
            <b-input v-model="v2ray.scMaxEachPostBytesFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.scMaxEachPostBytesTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="scMinPostsIntervalMs (From-To)" label-position="on-border">
            <b-input v-model="v2ray.scMinPostsIntervalFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.scMinPostsIntervalTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="scMaxBufferedPosts" label-position="on-border">
            <b-input v-model="v2ray.scMaxBufferedPosts" type="number" placeholder="scMaxBufferedPosts" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="scStreamUpServerSecs (From-To)" label-position="on-border">
            <b-input v-model="v2ray.scStreamUpServerFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.scStreamUpServerTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xPaddingBytes (From-To)" label-position="on-border">
            <b-input v-model="v2ray.xPaddingBytesFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.xPaddingBytesTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xmux maxConcurrency (From-To)" label-position="on-border">
            <b-input v-model="v2ray.xmuxMaxConcurFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.xmuxMaxConcurTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xmux maxConnections (From-To)" label-position="on-border">
            <b-input v-model="v2ray.xmuxMaxConnFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.xmuxMaxConnTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xmux cMaxReuseTimes (From-To)" label-position="on-border">
            <b-input v-model="v2ray.xmuxCMaxReuseFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.xmuxCMaxReuseTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xmux hMaxRequestTimes (From-To)" label-position="on-border">
            <b-input v-model="v2ray.xmuxHMaxReqFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.xmuxHMaxReqTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xmux hMaxReusableSecs (From-To)" label-position="on-border">
            <b-input v-model="v2ray.xmuxHMaxReusableFrom" type="number" placeholder="From" expanded />
            <b-input v-model="v2ray.xmuxHMaxReusableTo" type="number" placeholder="To" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="xmux hKeepAlivePeriod" label-position="on-border">
            <b-input v-model="v2ray.xmuxHKeepAlive" type="number" placeholder="hKeepAlivePeriod" expanded />
          </b-field>
          <b-field v-show="v2ray.net === 'xhttp'" label="Custom Headers" label-position="on-border">
            <div style="width:100%">
              <div v-for="(hdr, idx) in v2ray.xhttpHeaders" :key="idx" style="display:flex;gap:4px;margin-bottom:4px">
                <b-input v-model="hdr.key" placeholder="Header-Name" expanded />
                <b-input v-model="hdr.value" placeholder="value" expanded />
                <b-button type="is-danger is-light" icon-left="delete" size="is-small" @click="v2ray.xhttpHeaders.splice(idx,1)" />
              </div>
              <b-button size="is-small" icon-left="plus" @click="v2ray.xhttpHeaders.push({key:'',value:''})">Add Header</b-button>
            </div>
          </b-field>
          <!-- QUIC -->
          <b-field v-show="v2ray.net === 'quic'" label="QUIC Security" label-position="on-border">
            <b-select v-model="v2ray.quicSecurity" expanded>
              <option value="none">none</option>
              <option value="aes-128-gcm">aes-128-gcm</option>
              <option value="chacha20-poly1305">chacha20-poly1305</option>
            </b-select>
          </b-field>
          <b-field v-show="v2ray.net === 'quic'" label="Key" label-position="on-border">
            <b-input v-model="v2ray.key" :placeholder="$t('configureServer.password')" expanded />
          </b-field>
        </b-tab-item>
        <b-tab-item label="WireGuard">
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
        </b-tab-item>
        <b-tab-item label="SS">
          <b-field label="Name" label-position="on-border">
            <b-input ref="ss_name" v-model="ss.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="ss_server" v-model="ss.server" required placeholder="IP / HOST" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="ss_port" v-model="ss.port" required :placeholder="$t('configureServer.port')" type="number"
              expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="ss_password" v-model="ss.password" required :placeholder="$t('configureServer.password')"
              expanded />
          </b-field>
          <b-field label="Method" label-position="on-border">
            <b-select ref="ss_method" v-model="ss.method" expanded required>
              <option value="2022-blake3-aes-128-gcm">2022-blake3-aes-128-gcm</option>
              <option value="2022-blake3-aes-256-gcm">2022-blake3-aes-256-gcm</option>
              <option value="2022-blake3-chacha20-poly1305">2022-blake3-chacha20-poly1305</option>
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
            <b-select ref="ss_plugin" v-model="ss.plugin" expanded>
              <option value="">{{ $t("setting.options.off") }}</option>
              <option value="simple-obfs">simple-obfs</option>
              <option value="v2ray-plugin">v2ray-plugin</option>
            </b-select>
          </b-field>
          <b-field v-if="ss.plugin === 'simple-obfs' || ss.plugin === 'v2ray-plugin'" label-position="on-border"
            class="with-icon-alert">
            <template slot="label">
              Impl
              <b-tooltip type="is-dark" :label="$t('setting.messages.ssPluginImpl')" multilined position="is-right">
                <b-icon size="is-samll" icon=" iconfont icon-help-circle-outline" style="
                    position: relative;
                    top: 2px;
                    right: 3px;
                    font-weight: normal;
                  " />
              </b-tooltip>
            </template>
            <b-select ref="ss_plugin_impl" v-model="ss.impl" expanded>
              <option value="">{{ $t("setting.options.default") }}</option>
              <option value="chained">chained</option>
              <option value="transport">transport</option>
            </b-select>
          </b-field>
          <b-field v-show="ss.plugin === 'simple-obfs'" label="Obfs" label-position="on-border">
            <b-select ref="ss_obfs" v-model="ss.obfs" expanded>
              <option value="http">http</option>
              <option value="tls">tls</option>
            </b-select>
          </b-field>
          <b-field v-show="ss.plugin === 'v2ray-plugin'" label="Mode" label-position="on-border">
            <b-select ref="ss_mode" v-model="ss.mode" expanded>
              <option value="websocket">websocket</option>
            </b-select>
          </b-field>
          <b-field v-show="ss.plugin === 'v2ray-plugin'" label="TLS" label-position="on-border">
            <b-select ref="ss_tls" v-model="ss.tls" expanded>
              <option value="">{{ $t("setting.options.off") }}</option>
              <option value="tls">tls</option>
            </b-select>
          </b-field>
          <b-field v-if="(ss.plugin === 'simple-obfs' &&
            (ss.obfs === 'http' || ss.obfs === 'tls')) ||
            ss.plugin === 'v2ray-plugin'
          " label="Host" label-position="on-border">
            <b-input ref="ss_host" v-model="ss.host" placeholder="(optional)" expanded />
          </b-field>
          <b-field v-if="(ss.plugin === 'simple-obfs' && ss.obfs === 'http') ||
            ss.plugin === 'v2ray-plugin'
          " label="Path" label-position="on-border">
            <b-input ref="ss_path" v-model="ss.path" placeholder="/" expanded />
          </b-field>
          <b-field :label="$t('setting.nodeBackend')" label-position="on-border">
            <b-select v-model="ss.backend" expanded>
              <option value="">{{ $t("setting.options.backendSystemDefault") }}</option>
              <option value="v2ray">{{ $t("setting.options.backendV2ray") }}</option>
            </b-select>
          </b-field>
        </b-tab-item>
        <b-tab-item label="SSR">
          <b-field label="Name" label-position="on-border">
            <b-input ref="ssr_name" v-model="ssr.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="ssr_server" v-model="ssr.server" required placeholder="IP / HOST" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="ssr_port" v-model="ssr.port" required :placeholder="$t('configureServer.port')" type="number"
              expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="ssr_password" v-model="ssr.password" required :placeholder="$t('configureServer.password')"
              expanded />
          </b-field>
          <b-field label="Method" label-position="on-border">
            <b-select ref="ssr_method" v-model="ssr.method" expanded required>
              <option value="aes-128-cfb">aes-128-cfb</option>
              <option value="aes-192-cfb">aes-192-cfb</option>
              <option value="aes-256-cfb">aes-256-cfb</option>
              <option value="aes-128-ctr">aes-128-ctr</option>
              <option value="aes-192-ctr">aes-192-ctr</option>
              <option value="aes-256-ctr">aes-256-ctr</option>
              <option value="aes-128-ofb">aes-128-ofb</option>
              <option value="aes-192-ofb">aes-192-ofb</option>
              <option value="aes-256-ofb">aes-256-ofb</option>
              <option value="des-cfb">des-cfb</option>
              <option value="bf-cfb">bf-cfb</option>
              <option value="cast5-cfb">cast5-cfb</option>
              <option value="rc4-md5">rc4-md5</option>
              <option value="chacha20">chacha20</option>
              <option value="chacha20-ietf">chacha20-ietf</option>
              <option value="salsa20">salsa20</option>
              <option value="camellia-128-cfb">camellia-128-cfb</option>
              <option value="camellia-192-cfb">camellia-192-cfb</option>
              <option value="camellia-256-cfb">camellia-256-cfb</option>
              <option value="idea-cfb">idea-cfb</option>
              <option value="rc2-cfb">rc2-cfb</option>
              <option value="seed-cfb">seed-cfb</option>
              <option value="none">none</option>
            </b-select>
          </b-field>
          <b-field label="Protocol" label-position="on-border">
            <b-select ref="ssr_proto" v-model="ssr.proto" expanded required>
              <option value="origin">origin</option>
              <option value="verify_sha1">verify_sha1</option>
              <option value="auth_sha1_v4">auth_sha1_v4</option>
              <option value="auth_aes128_md5">auth_aes128_md5</option>
              <option value="auth_aes128_sha1">auth_aes128_sha1</option>
              <option value="auth_chain_a">auth_chain_a</option>
              <option value="auth_chain_b">auth_chain_b</option>
            </b-select>
          </b-field>
          <b-field v-if="ssr.proto !== 'origin'" label="Protocol Param" label-position="on-border">
            <b-input ref="ssr_protoParam" v-model="ssr.protoParam" placeholder="(optional)" expanded />
          </b-field>
          <b-field label="Obfs" label-position="on-border">
            <b-select ref="ssr_obfs" v-model="ssr.obfs" expanded required>
              <option value="plain">plain</option>
              <option value="http_simple">http_simple</option>
              <option value="http_post">http_post</option>
              <option value="random_head">random_head</option>
              <option value="tls1.2_ticket_auth">tls1.2_ticket_auth</option>
            </b-select>
          </b-field>
          <b-field v-if="ssr.obfs !== 'plain'" label="Obfs Param" label-position="on-border">
            <b-input ref="ssr_obfsParam" v-model="ssr.obfsParam" placeholder="(optional)" expanded />
          </b-field>
        </b-tab-item>
        <b-tab-item label="Trojan">
          <b-field label="Name" label-position="on-border">
            <b-input ref="trojan_name" v-model="trojan.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="trojan_server" v-model="trojan.server" required placeholder="IP / HOST" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="trojan_port" v-model="trojan.port" required :placeholder="$t('configureServer.port')"
              type="number" expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="trojan_password" v-model="trojan.password" required
              :placeholder="$t('configureServer.password')" expanded />
          </b-field>
          <b-field label="Protocol" label-position="on-border">
            <b-select ref="trojan_method" v-model="trojan.method" expanded required>
              <option value="origin">{{ $t("configureServer.origin") }}</option>
              <option value="shadowsocks">shadowsocks</option>
            </b-select>
          </b-field>
          <b-field v-if="trojan.method === 'shadowsocks'" label="Shadowsocks Cipher" label-position="on-border">
            <b-select ref="trojan_ss_cipher" v-model="trojan.ssCipher" expanded required>
              <option value="aes-128-gcm">aes-128-gcm</option>
              <option value="aes-256-gcm">aes-256-gcm</option>
              <option value="chacha20-poly1305">chacha20-poly1305</option>
              <option value="chacha20-ietf-poly1305">
                chacha20-ietf-poly1305
              </option>
            </b-select>
          </b-field>
          <b-field v-if="trojan.method === 'shadowsocks'" label="Shadowsocks Password" label-position="on-border">
            <b-input ref="trojan_ss_password" v-model="trojan.ssPassword" required
              :placeholder="`shadowsocks${$t('configureServer.password')}`" expanded />
          </b-field>
          <b-field label="Pinned Cert SHA256" label-position="on-border">
            <b-input v-model="trojan.pinnedPeerCertSha256" :placeholder="$t('pinnedPeerCertSha256')" expanded />
          </b-field>
          <b-field label="Verify Peer Cert By Name" label-position="on-border">
            <b-input v-model="trojan.verifyPeerCertByName" :placeholder="$t('verifyPeerCertByName')" expanded />
          </b-field>
          <b-field label="SNI(Peer)" label-position="on-border">
            <b-input v-model="trojan.peer" placeholder="SNI(Peer)" expanded />
          </b-field>
          <b-field label="Network" label-position="on-border">
            <b-select ref="trojan_net" v-model="trojan.net" expanded required @input="handleNetworkChange">
              <option value="tcp">TCP</option>
              <option value="kcp">mKCP</option>
              <option value="ws">WebSocket</option>
              <option value="h2">HTTP/2</option>
              <option value="grpc">gRPC</option>
            </b-select>
          </b-field>
          <b-field label="Obfs" label-position="on-border">
            <b-select ref="trojan_obfs" v-model="trojan.obfs" expanded required>
              <option value="none">
                {{ $t("configureServer.noObfuscation") }}
              </option>
              <option value="websocket">websocket</option>
            </b-select>
          </b-field>
          <b-field v-show="trojan.obfs === 'websocket'" label="Websocket Host" label-position="on-border">
            <b-input v-model="trojan.host" expanded />
          </b-field>
          <b-field v-show="trojan.obfs === 'websocket'" label="Websocket Path" label-position="on-border">
            <b-input v-model="trojan.path" placeholder="/" expanded />
          </b-field>
          <b-field v-show="trojan.net === 'ws' || trojan.net === 'h2'" label="Host" label-position="on-border">
            <b-input v-model="trojan.host" :placeholder="$t('configureServer.hostObfuscation')" expanded />
          </b-field>
          <b-field v-show="trojan.tls === 'tls'" label="Alpn" label-position="on-border">
            <b-input v-model="trojan.alpn" placeholder="h2,http/1.1" expanded />
          </b-field>
          <b-field v-show="trojan.net === 'ws' || trojan.net === 'h2'" label="Path" label-position="on-border">
            <b-input v-model="trojan.path" :placeholder="$t('configureServer.pathObfuscation')" expanded />
          </b-field>
          <b-field v-show="trojan.net === 'mkcp' || trojan.net === 'kcp'" label="Seed" label-position="on-border">
            <b-input v-model="trojan.path" :placeholder="$t('configureServer.seedObfuscation')" expanded />
          </b-field>
          <b-field v-show="trojan.net === 'grpc'" label="Service Name" label-position="on-border">
            <b-input ref="trojan_service_name" v-model="trojan.path" type="text" expanded />
          </b-field>
          <b-field :label="$t('setting.nodeBackend')" label-position="on-border">
            <b-select v-model="trojan.backend" expanded>
              <option value="">{{ $t("setting.options.backendSystemDefault") }}</option>
              <option value="v2ray">{{ $t("setting.options.backendV2ray") }}</option>
            </b-select>
          </b-field>
        </b-tab-item>

        <b-tab-item label="Juicity">
          <b-field label="Name" label-position="on-border">
            <b-input ref="juicity_name" v-model="juicity.name" :placeholder="$t('configureServer.servername')"
              expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="juicity_server" v-model="juicity.server" required placeholder="IP / HOST" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="juicity_port" v-model="juicity.port" required :placeholder="$t('configureServer.port')"
              type="number" expanded />
          </b-field>
          <b-field label="UUID" label-position="on-border">
            <b-input ref="juicity_uuid" v-model="juicity.uuid" required placeholder="UUID" expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="juicity_password" v-model="juicity.password" required
              :placeholder="$t('configureServer.password')" expanded />
          </b-field>
          <b-field label="Congestion Control" label-position="on-border">
            <b-select ref="juicity_cc" v-model="juicity.cc" expanded required>
              <option value="bbr">bbr</option>
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
          <b-field label="Name" label-position="on-border">
            <b-input ref="tuic_name" v-model="tuic.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="tuic_server" v-model="tuic.server" required placeholder="IP / HOST" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="tuic_port" v-model="tuic.port" required :placeholder="$t('configureServer.port')"
              type="number" expanded />
          </b-field>
          <b-field label="UUID" label-position="on-border">
            <b-input ref="tuic_uuid" v-model="tuic.uuid" required placeholder="UUID" expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="tuic_password" v-model="tuic.password" required :placeholder="$t('configureServer.password')"
              expanded />
          </b-field>
          <b-field label="Congestion Control" label-position="on-border">
            <b-select ref="tuic_cc" v-model="tuic.cc" expanded required>
              <option value="bbr">bbr</option>
            </b-select>
          </b-field>
          <b-field label="Pinned Cert SHA256" label-position="on-border">
            <b-input v-model="tuic.pinnedPeerCertSha256" :placeholder="$t('pinnedPeerCertSha256')" expanded />
          </b-field>
          <b-field label="Verify Peer Cert By Name" label-position="on-border">
            <b-input v-model="tuic.verifyPeerCertByName" :placeholder="$t('verifyPeerCertByName')" expanded />
          </b-field>
          <b-field label-position="on-border">
            <template slot="label"> DisableSni </template>
            <b-select ref="tuic_disable_sni" v-model="tuic.disableSni" expanded required>
              <option :value="false">{{ $t("operations.no") }}</option>
              <option :value="true">
                {{ $t("operations.yes") }}
              </option>
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
              <option value="quic">quic</option>
            </b-select>
          </b-field>
        </b-tab-item>

        <b-tab-item label="Hysteria2">
          <b-field label="Name" label-position="on-border">
            <b-input ref="hysteria2_name" v-model="hysteria2.name" :placeholder="$t('configureServer.servername')"
              expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="hysteria2_server" v-model="hysteria2.server" required placeholder="IP / HOST" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="hysteria2_port" v-model="hysteria2.port" required :placeholder="$t('configureServer.port')"
              type="number" expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="hysteria2_password" v-model="hysteria2.password" required
              :placeholder="$t('configureServer.password')" expanded />
          </b-field>
          <b-field label="Pinned Cert SHA256" label-position="on-border">
            <b-input v-model="hysteria2.pinnedPeerCertSha256" :placeholder="$t('pinnedPeerCertSha256')" expanded />
          </b-field>
          <b-field label="Verify Peer Cert By Name" label-position="on-border">
            <b-input v-model="hysteria2.verifyPeerCertByName" :placeholder="$t('verifyPeerCertByName')" expanded />
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
        </b-tab-item>

        <b-tab-item label="HTTP">
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
            <b-input ref="http_host" v-model="http.host" required placeholder="IP / HOST" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="http_port" v-model="http.port" required :placeholder="$t('configureServer.port')"
              type="number" expanded />
          </b-field>
          <b-field label="Username" label-position="on-border">
            <b-input ref="http_username" v-model="http.username" :placeholder="$t('configureServer.username')"
              expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="http_password" v-model="http.password" :placeholder="$t('configureServer.password')"
              expanded />
          </b-field>
        </b-tab-item>

        <b-tab-item label="SOCKS5">
          <b-field label="Name" label-position="on-border">
            <b-input ref="socks5_name" v-model="socks5.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="socks5_host" v-model="socks5.host" required placeholder="IP / HOST" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="socks5_port" v-model="socks5.port" required :placeholder="$t('configureServer.port')"
              type="number" expanded />
          </b-field>
          <b-field label="Username" label-position="on-border">
            <b-input ref="socks5_username" v-model="socks5.username" :placeholder="$t('configureServer.username')"
              expanded />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input ref="socks5_password" v-model="socks5.password" :placeholder="$t('configureServer.password')"
              expanded />
          </b-field>
        </b-tab-item>

        <b-tab-item label="AnyTLS">
          <b-field label="Name" label-position="on-border">
            <b-input ref="anytls_name" v-model="anytls.name" :placeholder="$t('configureServer.servername')" expanded />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input ref="anytls_host" v-model="anytls.host" required placeholder="IP / HOST" expanded />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input ref="anytls_port" v-model="anytls.port" required :placeholder="$t('configureServer.port')" type="number" expanded />
          </b-field>
          <b-field label="Auth" label-position="on-border">
            <b-input ref="anytls_auth" v-model="anytls.auth" required placeholder="Authentication Key" expanded />
          </b-field>
          <b-field label="SNI(Peer)" label-position="on-border">
            <b-input ref="anytls_sni" v-model="anytls.sni" placeholder="SNI / Peer (Optional)" expanded />
          </b-field>
          <b-field label="Pinned Cert SHA256" label-position="on-border">
            <b-input v-model="anytls.pinnedPeerCertSha256" :placeholder="$t('pinnedPeerCertSha256')" expanded />
          </b-field>
          <b-field label="Verify Peer Cert By Name" label-position="on-border">
            <b-input v-model="anytls.verifyPeerCertByName" :placeholder="$t('verifyPeerCertByName')" expanded />
          </b-field>
        </b-tab-item>
      </b-tabs>
    </section>
    <footer v-if="!readonly" class="modal-card-foot flex-end">
      <button class="button" type="button" @click="$parent.close()">
        {{ $t("operations.cancel") }}
      </button>
      <button class="button is-primary" @click="handleClickSubmit">
        {{ $t("operations.saveApply") }}
      </button>
    </footer>
  </div>
</template>

<script>
import { generateURL, handleResponse, parseURL } from "@/assets/js/utils";
import { Base64 } from "js-base64";

export default {
  name: "ModalServer",
  props: {
    which: {
      type: Object,
      default() {
        return null;
      },
    },
    readonly: {
      type: Boolean,
      default: false,
    },
  },
  data: () => ({
    vlessVersion: 0,
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
      pinnedPeerCertSha256: "",
      verifyPeerCertByName: "",
      protocol: "vmess",
      key: "none",
      xhttpMode: "auto",
      xhttpHeaders: [],
      noGRPCHeader: false,
      noSSEHeader: false,
      uplinkHTTPMethod: "",
      scMaxEachPostBytesFrom: "",
      scMaxEachPostBytesTo: "",
      scMinPostsIntervalFrom: "",
      scMinPostsIntervalTo: "",
      scMaxBufferedPosts: "",
      scStreamUpServerFrom: "",
      scStreamUpServerTo: "",
      xPaddingBytesFrom: "",
      xPaddingBytesTo: "",
      xmuxMaxConcurFrom: "",
      xmuxMaxConcurTo: "",
      xmuxMaxConnFrom: "",
      xmuxMaxConnTo: "",
      xmuxCMaxReuseFrom: "",
      xmuxCMaxReuseTo: "",
      xmuxHMaxReqFrom: "",
      xmuxHMaxReqTo: "",
      xmuxHMaxReusableFrom: "",
      xmuxHMaxReusableTo: "",
      xmuxHKeepAlive: "",
      maxEarlyData: "",
      earlyDataHeaderName: "",
      multiMode: false,
      idleTimeout: "",
      healthCheckTimeout: "",
      permitWithoutStream: false,
      initialWindowsSize: "",
    },
    ss: {
      method: "2022-blake3-aes-128-gcm",
      plugin: "",
      obfs: "http",
      tls: "",
      path: "/",
      mode: "websocket",
      host: "",
      password: "",
      server: "",
      port: "",
      name: "",
      protocol: "ss",
      impl: "",
      backend: "",
    },
    ssr: {
      method: "aes-128-cfb",
      password: "",
      server: "",
      port: "",
      name: "",
      proto: "origin",
      protoParam: "",
      obfs: "plain",
      obfsParam: "",
      protocol: "ssr",
    },
    trojan: {
      name: "",
      server: "",
      peer: "" /* tls sni */,
      host: "" /* websocket host */,
      path: "" /* websocket path */,
      pinnedPeerCertSha256: "",
      verifyPeerCertByName: "",
      port: "",
      password: "",
      method: "origin" /* shadowsocks */,
      ssCipher: "aes-128-gcm",
      ssPassword: "",
      net: "tcp",
      obfs: "none" /* websocket */,
      protocol: "trojan",
      backend: "",
    },
    juicity: {
      name: "",
      server: "",
      port: "",
      sni: "",
      cc: "bbr",
      uuid: "",
      password: "",
      pinnedCertchainSha256: "",
      protocol: "juicity",
    },
    tuic: {
      name: "",
      server: "",
      port: "",
      sni: "",
      cc: "bbr",
      uuid: "",
      password: "",
      pinnedPeerCertSha256: "",
      verifyPeerCertByName: "",
      disableSni: false,
      alpn: "h3",
      udpRelayMode: "native",
      protocol: "tuic",
    },
    hysteria2: {
      name: "",
      server: "",
      port: "",
      password: "",
      sni: "",
      obfs: "none",
      obfsPassword: "",
      pinnedPeerCertSha256: "",
      verifyPeerCertByName: "",
      protocol: "hysteria2",
    },
    http: {
      username: "",
      password: "",
      host: "",
      port: "",
      protocol: "http",
      name: "",
    },
    socks5: {
      username: "",
      password: "",
      host: "",
      port: "",
      protocol: "socks5",
      name: "",
    },
    anytls: {
      name: "",
      host: "",
      port: "",
      auth: "",
      sni: "",
      pinnedPeerCertSha256: "",
      verifyPeerCertByName: "",
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
    },
    tabChoice: 0,
  }),
  mounted() {
    if (this.which !== null) {
      this.$axios({
        url: apiRoot + "/sharingAddress",
        method: "get",
        params: {
          touch: this.which,
        },
      }).then((res) => {
        handleResponse(res, this, () => {
          if (
            res.data.data.sharingAddress.toLowerCase().startsWith("vmess://")
          ) {
            this.v2ray = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 0;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("vless://")
          ) {
            this.v2ray = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 1;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("wireguard://")
          ) {
            this.wireguard = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 2;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("ss://")
          ) {
            this.ss = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 3;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("ssr://")
          ) {
            this.ssr = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 4;
          } else if (
            res.data.data.sharingAddress
              .toLowerCase()
              .startsWith("trojan://") ||
            res.data.data.sharingAddress
              .toLowerCase()
              .startsWith("trojan-go://")
          ) {
            this.trojan = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 5;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("juicity://")
          ) {
            this.juicity = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 6;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("tuic://")
          ) {
            this.tuic = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 7;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("hysteria2://") ||
            res.data.data.sharingAddress.toLowerCase().startsWith("hy2://")
          ) {
            this.hysteria2 = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 8;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("http://") ||
            res.data.data.sharingAddress.toLowerCase().startsWith("https://")
          ) {
            this.http = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 9;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("socks5://")
          ) {
            this.socks5 = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 10;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("anytls://")
          ) {
            this.anytls = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 11;
          }
          this.$nextTick(() => {
            if (this.readonly) {
              this.$refs.section
                .querySelectorAll("input, textarea")
                .forEach((x) => (x.readOnly = "readOnly"));
              this.$refs.section.querySelectorAll("select").forEach((x) => {
                const text = x.querySelector(
                  `option[value="${x.value}"]`
                ).textContent;
                console.log(x.value, text);
                x.outerHTML = `<input type="text" class="input" readonly="readonly" value="${text}">`;
              });
            }
          });
        });
      });
    }
  },
  watch: {
    tabChoice(val) {
      if (val === 0) this.v2ray.protocol = "vmess";
      if (val === 1) this.v2ray.protocol = "vless";
    },
  },
  methods: {
    variant() {
      return localStorage["variant"]?.toLowerCase() || "v2ray";
    },
    handleV2rayProtocolSwitch() {
      // protocol is now driven by tab selection
    },
    resolveURL(url) {
      if (url.toLowerCase().startsWith("vmess://")) {
        let obj = JSON.parse(
          Base64.decode(url.substring(url.indexOf("://") + 3))
        );
        // console.log(obj);
        obj.ps = decodeURIComponent(obj.ps);
        obj.tls = obj.tls || "none";
        obj.type = obj.type || "none";
        obj.scy = obj.scy || "auto";
        obj.protocol = obj.protocol || "vmess";
        return obj;
      } else if (url.toLowerCase().startsWith("vless://")) {
        let u = parseURL(url);
        const o = {
          ps: decodeURIComponent(u.hash),
          add: u.host,
          port: u.port,
          id: decodeURIComponent(u.username),
          flow: u.params.flow || "",
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
          pinnedPeerCertSha256: u.params.pinnedPeerCertSha256 || "",
          verifyPeerCertByName: u.params.verifyPeerCertByName || "",
          key: u.params.key,
          xhttpMode: u.params.xhttpMode || "auto",
          noGRPCHeader: u.params.noGRPCHeader === "true",
          noSSEHeader: u.params.noSSEHeader === "true",
          uplinkHTTPMethod: u.params.uplinkHTTPMethod || "",
          scMaxEachPostBytesFrom: u.params.scMaxEachPostBytesFrom || "",
          scMaxEachPostBytesTo: u.params.scMaxEachPostBytesTo || "",
          scMinPostsIntervalFrom: u.params.scMinPostsIntervalFrom || "",
          scMinPostsIntervalTo: u.params.scMinPostsIntervalTo || "",
          scMaxBufferedPosts: u.params.scMaxBufferedPosts || "",
          scStreamUpServerFrom: u.params.scStreamUpServerFrom || "",
          scStreamUpServerTo: u.params.scStreamUpServerTo || "",
          xPaddingBytesFrom: u.params.xPaddingBytesFrom || "",
          xPaddingBytesTo: u.params.xPaddingBytesTo || "",
          xmuxMaxConcurFrom: u.params.xmuxMaxConcurFrom || "",
          xmuxMaxConcurTo: u.params.xmuxMaxConcurTo || "",
          xmuxMaxConnFrom: u.params.xmuxMaxConnFrom || "",
          xmuxMaxConnTo: u.params.xmuxMaxConnTo || "",
          xmuxCMaxReuseFrom: u.params.xmuxCMaxReuseFrom || "",
          xmuxCMaxReuseTo: u.params.xmuxCMaxReuseTo || "",
          xmuxHMaxReqFrom: u.params.xmuxHMaxReqFrom || "",
          xmuxHMaxReqTo: u.params.xmuxHMaxReqTo || "",
          xmuxHMaxReusableFrom: u.params.xmuxHMaxReusableFrom || "",
          xmuxHMaxReusableTo: u.params.xmuxHMaxReusableTo || "",
          xmuxHKeepAlive: u.params.xmuxHKeepAlive || "",
          xhttpHeaders: (() => {
            try {
              const raw = u.params.xhttpHeaders;
              if (!raw) return [];
              const obj = JSON.parse(raw);
              return Object.entries(obj).map(([key, value]) => ({ key, value }));
            } catch (_) { return []; }
          })(),
          maxEarlyData: u.params.maxEarlyData || "",
          earlyDataHeaderName: u.params.earlyDataHeaderName || "",
          multiMode: u.params.multiMode === "true" || u.params.multiMode === "1",
          idleTimeout: u.params.idleTimeout || "",
          healthCheckTimeout: u.params.healthCheckTimeout || "",
          permitWithoutStream: u.params.permitWithoutStream === "true" || u.params.permitWithoutStream === "1",
          initialWindowsSize: u.params.initialWindowsSize || "",
          protocol: "vless",
        };
        if (o.alpn !== "") {
          o.alpn = decodeURIComponent(o.alpn);
        }
        if (o.net === "mkcp" || o.net === "kcp") {
          o.path = u.params.seed;
        }
        console.log(o);
        return o;
      } else if (url.toLowerCase().startsWith("ss://")) {
        let u = parseURL(url);
        let userinfo = u.username;
        // Handle SIP002 format: ss://BASE64URL@host:port vs legacy ss://BASE64
        let method = "", password = "";
        try {
          let decoded = Base64.decode(userinfo);
          let idx = decoded.indexOf(":");
          if (idx > -1) {
            method = decoded.substring(0, idx);
            password = decoded.substring(idx + 1);
          }
        } catch (e) {
          method = userinfo;
        }
        const ssPlugin = u.params.plugin || "";
        return {
          method: method,
          password: password,
          server: u.host,
          port: u.port,
          name: decodeURIComponent(u.hash || ""),
          plugin: ssPlugin.split(";")[0] || "",
          plugin_opts: ssPlugin.split(";").slice(1).join(";") || "",
          protocol: "ss",
          backend: u.params["v2raya-backend"] || "",
        };
      } else if (url.toLowerCase().startsWith("ssr://")) {
        url = Base64.decode(url.substr(6));
        let arr = url.split("/?");
        let query = arr[1].split("&");
        let m = {};
        for (let param of query) {
          let [key, val] = param.split("=", 2);
          val = Base64.decode(val);
          m[key] = val;
        }
        let pre = arr[0].split(":");
        if (pre.length > 6) {
          //如果长度多于6，说明host中包含字符:，重新合并前几个分组到host去
          pre[pre.length - 6] = pre.slice(0, pre.length - 5).join(":");
          pre = pre.slice(pre.length - 6);
        }
        pre[5] = Base64.decode(pre[5]);
        return {
          method: pre[3],
          password: pre[5],
          server: pre[0],
          port: pre[1],
          name: m["remarks"],
          proto: pre[2],
          protoParam: m["protoparam"],
          obfs: pre[4],
          obfsParam: m["obfsparam"],
          protocol: "ssr",
        };
      } else if (
        url.toLowerCase().startsWith("trojan://") ||
        url.toLowerCase().startsWith("trojan-go://")
      ) {
        let u = parseURL(url);
        const o = {
          password: decodeURIComponent(u.username),
          server: u.host,
          port: u.port,
          name: decodeURIComponent(u.hash),
          peer: u.params.peer || u.params.sni || "",
          pinnedPeerCertSha256: u.params.pinnedPeerCertSha256 || "",
          verifyPeerCertByName: u.params.verifyPeerCertByName || "",
          method: "origin",
          net: u.params.type || "tcp",
          obfs: "none",
          ssCipher: "2022-blake3-aes-128-gcm",
          path: u.params.path || u.params.serviceName || "",
          protocol: "trojan",
          backend: u.params["v2raya-backend"] || "",
        };
        if (url.toLowerCase().startsWith("trojan-go://")) {
          console.log(u.params.encryption);
          if (u.params.encryption?.startsWith("ss;")) {
            o.method = "shadowsocks";
            const fields = u.params.encryption.split(";");
            o.ssCipher = fields[1];
            o.ssPassword = fields[2];
          }
          if (u.params.type === "ws") {
            o.obfs = "websocket";
            o.host = u.params.host || "";
            o.path = u.params.path || "/";
          }
        }
        return o;
      } else if (url.toLowerCase().startsWith("juicity://")) {
        let u = parseURL(url);
        return {
          name: decodeURIComponent(u.hash),
          uuid: decodeURIComponent(u.username),
          password: decodeURIComponent(u.password),
          server: u.host,
          port: u.port,
          sni: u.params.sni || "",
          pinnedCertchainSha256: u.params.pinned_certchain_sha256 || "",
          cc: u.params.congestion_control || "bbr",
          protocol: "juicity",
        };
      } else if (url.toLowerCase().startsWith("tuic://")) {
        let u = parseURL(url);
        return {
          name: decodeURIComponent(u.hash),
          uuid: decodeURIComponent(u.username),
          password: decodeURIComponent(u.password),
          server: u.host,
          port: u.port,
          sni: u.params.sni || "",
          pinnedPeerCertSha256: u.params.pinnedPeerCertSha256 || u.params.pinned_peer_cert_sha256 || "",
          verifyPeerCertByName: u.params.verifyPeerCertByName || u.params.verify_peer_cert_by_name || "",
          disableSni:
            u.params.disable_sni === "true" || u.params.disable_sni === "1",
          alpn: u.params.alpn,
          cc: u.params.congestion_control || "bbr",
          udpRelayMode: u.params.udp_relay_mode || "native",
          protocol: "tuic",
        };
      } else if (
        url.toLowerCase().startsWith("hysteria2://") ||
        url.toLowerCase().startsWith("hy2://")
      ) {
        let u = parseURL(url);
        return {
          name: decodeURIComponent(u.hash),
          password: decodeURIComponent(u.username),
          server: u.host,
          port: u.port,
          sni: u.params.sni || "",
          pinnedPeerCertSha256: u.params.pinnedPeerCertSha256 || u.params.pinned_peer_cert_sha256 || "",
          verifyPeerCertByName: u.params.verifyPeerCertByName || u.params.verify_peer_cert_by_name || "",
          obfs: u.params.obfs || "none",
          obfsPassword: u.params["obfs-password"] || "",
          protocol: "hysteria2",
        };
      } else if (
        url.toLowerCase().startsWith("http://") ||
        url.toLowerCase().startsWith("https://")
      ) {
        let u = parseURL(url);
        return {
          username: decodeURIComponent(u.username),
          password: decodeURIComponent(u.password),
          host: u.host,
          port: u.port,
          protocol: u.protocol,
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
        return {
          name: decodeURIComponent(u.hash),
          host: u.host,
          port: u.port,
          auth: auth,
          sni: sni,
          pinnedPeerCertSha256: u.params.pinnedPeerCertSha256 || u.params.pinned_peer_cert_sha256 || "",
          verifyPeerCertByName: u.params.verifyPeerCertByName || u.params.verify_peer_cert_by_name || "",
          protocol: "anytls",
        };
      } else if (url.toLowerCase().startsWith("wireguard://")) {
        let u = parseURL(url);
        return {
          name: decodeURIComponent(u.hash),
          address: u.host,
          port: u.port,
          publicKey: decodeURIComponent(u.username),
          privateKey: u.params.privateKey || "",
          localAddress: u.params.localAddress || "",
          dns: u.params.dns || "",
          mtu: u.params.mtu || "",
          allowedIPs: u.params.allowedIPs || "",
          persistentKeepalive: u.params.persistentKeepalive || "",
          preSharedKey: u.params.preSharedKey || "",
          endpoint: u.params.endpoint || "",
        };
      }
      return null;
    },
    generateURL(srcObj) {
      let query = {};
      let obj = {};
      let tmp;
      switch (srcObj.protocol) {
        case "vless":
          // todo: support reality and xhttp
          // https://github.com/XTLS/Xray-core/discussions/716
          query = {
            type: srcObj.net,
            flow: srcObj.flow || "",
            security: srcObj.tls,
            fp: srcObj.fp || "",
            path: srcObj.path,
            host: srcObj.host,
            headerType: srcObj.type,
            sni: srcObj.sni,
            pinnedPeerCertSha256: srcObj.pinnedPeerCertSha256,
            verifyPeerCertByName: srcObj.verifyPeerCertByName,
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
          if (srcObj.net === "mkcp" || srcObj.net === "kcp") {
            query.seed = srcObj.path;
          }
          if (srcObj.net === "quic") {
            query.key = srcObj.key;
            query.quicSecurity = srcObj.quicSecurity;
          }
          if (srcObj.net === "xhttp") {
            query.xhttpMode = srcObj.xhttpMode;
            if (srcObj.noGRPCHeader) query.noGRPCHeader = "true";
            if (srcObj.noSSEHeader) query.noSSEHeader = "true";
            if (srcObj.uplinkHTTPMethod) query.uplinkHTTPMethod = srcObj.uplinkHTTPMethod;
            if (srcObj.scMaxEachPostBytesFrom) query.scMaxEachPostBytesFrom = srcObj.scMaxEachPostBytesFrom;
            if (srcObj.scMaxEachPostBytesTo) query.scMaxEachPostBytesTo = srcObj.scMaxEachPostBytesTo;
            if (srcObj.scMinPostsIntervalFrom) query.scMinPostsIntervalFrom = srcObj.scMinPostsIntervalFrom;
            if (srcObj.scMinPostsIntervalTo) query.scMinPostsIntervalTo = srcObj.scMinPostsIntervalTo;
            if (srcObj.scMaxBufferedPosts) query.scMaxBufferedPosts = srcObj.scMaxBufferedPosts;
            if (srcObj.scStreamUpServerFrom) query.scStreamUpServerFrom = srcObj.scStreamUpServerFrom;
            if (srcObj.scStreamUpServerTo) query.scStreamUpServerTo = srcObj.scStreamUpServerTo;
            if (srcObj.xPaddingBytesFrom) query.xPaddingBytesFrom = srcObj.xPaddingBytesFrom;
            if (srcObj.xPaddingBytesTo) query.xPaddingBytesTo = srcObj.xPaddingBytesTo;
            if (srcObj.xmuxMaxConcurFrom) query.xmuxMaxConcurFrom = srcObj.xmuxMaxConcurFrom;
            if (srcObj.xmuxMaxConcurTo) query.xmuxMaxConcurTo = srcObj.xmuxMaxConcurTo;
            if (srcObj.xmuxMaxConnFrom) query.xmuxMaxConnFrom = srcObj.xmuxMaxConnFrom;
            if (srcObj.xmuxMaxConnTo) query.xmuxMaxConnTo = srcObj.xmuxMaxConnTo;
            if (srcObj.xmuxCMaxReuseFrom) query.xmuxCMaxReuseFrom = srcObj.xmuxCMaxReuseFrom;
            if (srcObj.xmuxCMaxReuseTo) query.xmuxCMaxReuseTo = srcObj.xmuxCMaxReuseTo;
            if (srcObj.xmuxHMaxReqFrom) query.xmuxHMaxReqFrom = srcObj.xmuxHMaxReqFrom;
            if (srcObj.xmuxHMaxReqTo) query.xmuxHMaxReqTo = srcObj.xmuxHMaxReqTo;
            if (srcObj.xmuxHMaxReusableFrom) query.xmuxHMaxReusableFrom = srcObj.xmuxHMaxReusableFrom;
            if (srcObj.xmuxHMaxReusableTo) query.xmuxHMaxReusableTo = srcObj.xmuxHMaxReusableTo;
            if (srcObj.xmuxHKeepAlive) query.xmuxHKeepAlive = srcObj.xmuxHKeepAlive;
            if (srcObj.xhttpHeaders && srcObj.xhttpHeaders.length > 0) {
              const hdrsObj = {};
              srcObj.xhttpHeaders.forEach(h => { if (h.key) hdrsObj[h.key] = h.value; });
              if (Object.keys(hdrsObj).length > 0) query.xhttpHeaders = JSON.stringify(hdrsObj);
            }
          }
          if (query.security == "reality") {
            query.pbk = srcObj.pbk;
            query.sid = srcObj.sid;
            query.spx = srcObj.spx;
          }
          return generateURL({
            protocol: "vless",
            username: srcObj.id,
            host: srcObj.add,
            port: srcObj.port,
            hash: srcObj.ps,
            params: query,
          });
        case "vmess":
          //https://github.com/2dust/v2rayN/wiki/%E5%88%86%E4%BA%AB%E9%93%BE%E6%8E%A5%E6%A0%BC%E5%BC%8F%E8%AF%B4%E6%98%8E(ver-2)
          obj = Object.assign({}, srcObj);
          switch (obj.net) {
            case "kcp":
            case "tcp":
            case "quic":
              break;
            default:
              obj.type = "";
          }
          switch (obj.net) {
            case "ws":
            case "h2":
            case "http":
            case "quic":
            case "grpc":
            case "kcp":
            case "mkcp":
              break;
            default:
              if (obj.net === "tcp" && obj.type === "http") {
                break;
              }
              obj.path = "";
          }
          return "vmess://" + Base64.encode(JSON.stringify(obj));
        case "ss":
          /* ss://BASE64(method:password)@server:port#name */
          tmp = `ss://${Base64.encode(`${srcObj.method}:${srcObj.password}`)}@${srcObj.server
            }:${srcObj.port}/`;
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
          tmp += srcObj.name.length
            ? `#${encodeURIComponent(srcObj.name)}`
            : "";
          return tmp;

        case "ssr":
          /* ssr://server:port:proto:method:obfs:URLBASE64(password)/?remarks=URLBASE64(remarks)&protoparam=URLBASE64(protoparam)&obfsparam=URLBASE64(obfsparam)) */
          return `ssr://${Base64.encode(
            `${srcObj.server}:${srcObj.port}:${srcObj.proto}:${srcObj.method}:${srcObj.obfs
            }:${Base64.encodeURI(srcObj.password)}/?remarks=${Base64.encodeURI(
              srcObj.name
            )}&protoparam=${Base64.encodeURI(
              srcObj.protoParam
            )}&obfsparam=${Base64.encodeURI(srcObj.obfsParam)}`
          )}`;
        case "trojan":
          /* trojan://password@server:port?allowInsecure=1&sni=sni#URIESCAPE(name) */
          query = {
            type: srcObj.net,
            pinnedPeerCertSha256: srcObj.pinnedPeerCertSha256,
            verifyPeerCertByName: srcObj.verifyPeerCertByName,
          };
          if (srcObj.peer !== "") {
            query.sni = srcObj.peer;
          }
          tmp = "trojan";
          if (srcObj.method !== "origin" || srcObj.obfs !== "none") {
            tmp = "trojan-go";
            query.type = srcObj.obfs === "none" ? "original" : "ws";
            if (srcObj.method === "shadowsocks") {
              query.encryption = `ss;${srcObj.ssCipher};${srcObj.ssPassword}`;
            }
            if (query.type === "ws") {
              query.host = srcObj.host || "";
              query.path = srcObj.path || "/";
            }
          }

          if (srcObj.alpn !== "") {
            query.alpn = srcObj.alpn;
          }
          if (srcObj.net === "grpc") {
            query.serviceName = srcObj.path;
          }
          if (srcObj.net === "mkcp" || srcObj.net === "kcp") {
            query.seed = srcObj.path;
          }
          return generateURL({
            protocol: tmp,
            username: srcObj.password,
            host: srcObj.server,
            port: srcObj.port,
            hash: srcObj.name,
            params: query,
          });
        case "juicity":
          query = {
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
            pinned_peer_cert_sha256: srcObj.pinnedPeerCertSha256,
            verify_peer_cert_by_name: srcObj.verifyPeerCertByName,
            congestion_control: srcObj.cc,
            disable_sni: srcObj.disableSni,
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
            pinned_peer_cert_sha256: srcObj.pinnedPeerCertSha256,
            verify_peer_cert_by_name: srcObj.verifyPeerCertByName,
          };
          if (srcObj.sni !== "") {
            query.sni = srcObj.sni;
          }
          if (srcObj.obfs !== "none") {
            query.obfs = srcObj.obfs;
            query["obfs-password"] = srcObj.obfsPassword;
          }
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
          if (srcObj.sni) {
            query.peer = srcObj.sni;
          }
          if (srcObj.pinnedPeerCertSha256) {
            query.pinnedPeerCertSha256 = srcObj.pinnedPeerCertSha256;
          }
          if (srcObj.verifyPeerCertByName) {
            query.verifyPeerCertByName = srcObj.verifyPeerCertByName;
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
          if (srcObj.privateKey) {
            query.privateKey = srcObj.privateKey;
          }
          if (srcObj.localAddress) {
            query.localAddress = srcObj.localAddress;
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
            username: srcObj.publicKey,
            host: srcObj.address,
            port: srcObj.port,
            hash: srcObj.name,
            params: query,
          });
      }
      return null;
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
    async handleClickSubmit() {
      let valid = true;
      for (let k in this.$refs) {
        if (!this.$refs.hasOwnProperty(k)) {
          continue;
        }
        if (this.tabChoice === 0 && !k.startsWith("v2ray_")) {
          continue;
        }
        if (this.tabChoice === 1 && !k.startsWith("v2ray_")) {
          continue;
        }
        if (this.tabChoice === 2 && !k.startsWith("wireguard_")) {
          continue;
        }
        if (this.tabChoice === 3 && !k.startsWith("ss_")) {
          continue;
        }
        if (this.tabChoice === 4 && !k.startsWith("ssr_")) {
          continue;
        }
        if (this.tabChoice === 5 && !k.startsWith("trojan_")) {
          continue;
        }
        if (this.tabChoice === 6 && !k.startsWith("juicity_")) {
          continue;
        }
        if (this.tabChoice === 7 && !k.startsWith("tuic_")) {
          continue;
        }
        if (this.tabChoice === 8 && !k.startsWith("hysteria2_")) {
          continue;
        }
        if (this.tabChoice === 9 && !k.startsWith("http_")) {
          continue;
        }
        if (this.tabChoice === 10 && !k.startsWith("socks5_")) {
          continue;
        }
        if (this.tabChoice === 11 && !k.startsWith("anytls_")) {
          continue;
        }
        let x = this.$refs[k];
        if (!x) {
          continue;
        }
        if (
          x.$el.offsetParent && // is visible
          x.hasOwnProperty("checkHtml5Validity") &&
          typeof x.checkHtml5Validity === "function" &&
          !x.checkHtml5Validity()
        ) {
          console.error("validate failed", x);
          valid = false;
        }
      }
      if (!valid) {
        return;
      }
      let coded = "";
      // 0: vmess, 1: vless, 2: wireguard, 3: ss, 4: ssr, 5: trojan, 6: juicity, 7: tuic, 8: hysteria2, 9: http, 10: socks5, 11: anytls
      if (this.tabChoice === 0) {
        coded = this.generateURL(this.v2ray);
      } else if (this.tabChoice === 1) {
        coded = this.generateURL(this.v2ray);
      } else if (this.tabChoice === 2) {
        // wireguard://address:port?key=value#name
        coded = this.generateURL(this.wireguard);
      } else if (this.tabChoice === 3) {
        // ss://BASE64(method:password)@server:port?plugin=...&v2raya-backend=...#name
        const { method, password, server, port, name, plugin, plugin_opts, backend } = this.ss;
        let userinfo = btoa(`${method}:${password}`);
        let params = [];
        if (plugin) {
          params.push(`plugin=${encodeURIComponent(plugin + (plugin_opts ? `;${plugin_opts}` : ""))}`);
        }
        if (backend) {
          params.push(`v2raya-backend=${encodeURIComponent(backend)}`);
        }
        let url = `ss://${userinfo}@${server}:${port}`;
        if (params.length) url += `?${params.join("&")}`;
        if (name) url += `#${encodeURIComponent(name)}`;
        coded = url;
      } else if (this.tabChoice === 4) {
        // ssr://server:port:proto:method:obfs:base64(password)/?remarks=base64(remarks)
        const { server, port, proto, method, obfs, password, name, protoParam, obfsParam } = this.ssr;
        let pwdB64 = btoa(password);
        let remarksB64 = name ? btoa(name) : "";
        let protoParamB64 = protoParam ? btoa(protoParam) : "";
        let obfsParamB64 = obfsParam ? btoa(obfsParam) : "";
        let url = `ssr://${btoa(`${server}:${port}:${proto}:${method}:${obfs}:${pwdB64}/?remarks=${remarksB64}&protoparam=${protoParamB64}&obfsparam=${obfsParamB64}`)}`;
        coded = url;
      } else if (this.tabChoice === 5) {
        // trojan://password@server:port?pinnedPeerCertSha256=&verifyPeerCertByName=&sni=sni&v2raya-backend=...#name
        const { password, server, port, pinnedPeerCertSha256, verifyPeerCertByName, peer, name, backend } = this.trojan;
        let params = [];
        if (pinnedPeerCertSha256) params.push("pinnedPeerCertSha256=" + encodeURIComponent(pinnedPeerCertSha256));
        if (verifyPeerCertByName) params.push("verifyPeerCertByName=" + encodeURIComponent(verifyPeerCertByName));
        if (peer) params.push(`sni=${encodeURIComponent(peer)}`);
        if (backend) params.push(`v2raya-backend=${encodeURIComponent(backend)}`);
        let url = `trojan://${encodeURIComponent(password)}@${server}:${port}`;
        if (params.length) url += `?${params.join("&")}`;
        if (name) url += `#${encodeURIComponent(name)}`;
        coded = url;
      } else if (this.tabChoice === 6) {
        // juicity://uuid:password@server:port?cc=xxx#name
        const { uuid, password, server, port, cc, sni, name } = this.juicity;
        let params = [];
        if (cc) params.push(`congestion_control=${encodeURIComponent(cc)}`);
        if (sni) params.push(`sni=${encodeURIComponent(sni)}`);
        let url = `juicity://${uuid}:${password}@${server}:${port}`;
        if (params.length) url += `?${params.join("&")}`;
        if (name) url += `#${encodeURIComponent(name)}`;
        coded = url;
      } else if (this.tabChoice === 7) {
        // tuic://uuid:password@server:port?pinned_peer_cert_sha256=&verify_peer_cert_by_name=&cc=xxx#name
        const { uuid, password, server, port, pinnedPeerCertSha256, verifyPeerCertByName, cc, sni, name } = this.tuic;
        let params = [];
        if (pinnedPeerCertSha256) params.push("pinned_peer_cert_sha256=" + encodeURIComponent(pinnedPeerCertSha256));
        if (verifyPeerCertByName) params.push("verify_peer_cert_by_name=" + encodeURIComponent(verifyPeerCertByName));
        if (cc) params.push(`congestion_control=${encodeURIComponent(cc)}`);
        if (sni) params.push(`sni=${encodeURIComponent(sni)}`);
        let url = `tuic://${uuid}:${password}@${server}:${port}`;
        if (params.length) url += `?${params.join("&")}`;
        if (name) url += `#${encodeURIComponent(name)}`;
        coded = url;
      } else if (this.tabChoice === 8) {
        // hysteria2://password@server:port?pinned_peer_cert_sha256=&verify_peer_cert_by_name=&obfs=xxx#name
        const { password, server, port, pinnedPeerCertSha256, verifyPeerCertByName, obfs, obfsPassword, sni, name } = this.hysteria2;
        let params = [];
        if (pinnedPeerCertSha256) params.push("pinned_peer_cert_sha256=" + encodeURIComponent(pinnedPeerCertSha256));
        if (verifyPeerCertByName) params.push("verify_peer_cert_by_name=" + encodeURIComponent(verifyPeerCertByName));
        if (obfs) params.push(`obfs=${encodeURIComponent(obfs)}`);
        if (obfsPassword) params.push(`obfs-password=${encodeURIComponent(obfsPassword)}`);
        if (sni) params.push(`sni=${encodeURIComponent(sni)}`);
        let url = `hysteria2://${encodeURIComponent(password)}@${server}:${port}`;
        if (params.length) url += `?${params.join("&")}`;
        if (name) url += `#${encodeURIComponent(name)}`;
        coded = url;
      } else if (this.tabChoice === 9) {
        // http(s)://username:password@server:port#name
        const { protocol, username, password, host, port, name } = this.http;
        let url = `${protocol}://`;
        if (username && password) url += `${encodeURIComponent(username)}:${encodeURIComponent(password)}@`;
        url += `${host}:${port}`;
        if (name) url += `#${encodeURIComponent(name)}`;
        coded = url;
      } else if (this.tabChoice === 10) {
        // socks5://username:password@server:port#name
        const { username, password, host, port, name } = this.socks5;
        let url = `socks5://`;
        if (username && password) url += `${encodeURIComponent(username)}:${encodeURIComponent(password)}@`;
        url += `${host}:${port}`;
        if (name) url += `#${encodeURIComponent(name)}`;
        coded = url;
      } else if (this.tabChoice === 11) {
        // anytls://auth@host:port?peer=sni&pinnedPeerCertSha256=&verifyPeerCertByName=#name
        const { auth, host, port, sni, pinnedPeerCertSha256, verifyPeerCertByName, name } = this.anytls;
        let params = [];
        if (sni) params.push(`peer=${encodeURIComponent(sni)}`);
        if (pinnedPeerCertSha256) params.push("pinnedPeerCertSha256=" + encodeURIComponent(pinnedPeerCertSha256));
        if (verifyPeerCertByName) params.push("verifyPeerCertByName=" + encodeURIComponent(verifyPeerCertByName));
        let url = `anytls://${encodeURIComponent(auth)}@${host}:${port}`;
        if (params.length) url += `?${params.join("&")}`;
        if (name) url += `#${encodeURIComponent(name)}`;
        coded = url;
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

.same-width-5 li {
  min-width: 5em;
  width: unset !important;
}
</style>
