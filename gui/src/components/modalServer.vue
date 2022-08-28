<template>
  <div class="modal-card" style="max-width: 520px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $tc("configureServer.title", readonly ? 2 : 1) }}
      </p>
    </header>
    <section ref="section" :class="{ 'modal-card-body': true }">
      <b-tabs
        v-model="tabChoice"
        position="is-centered"
        class="block"
        type="is-boxed is-twitter same-width-5"
      >
        <b-tab-item label="V2RAY">
          <b-field
            v-show="showVLess"
            label="Protocol"
            label-position="on-border"
          >
            <b-select
              v-model="v2ray.protocol"
              expanded
              @input="handleV2rayProtocolSwitch"
            >
              <option value="vmess">VMESS</option>
              <option value="vless">VLESS</option>
            </b-select>
          </b-field>
          <b-field label="Name" label-position="on-border">
            <b-input
              ref="v2ray_name"
              v-model="v2ray.ps"
              :placeholder="$t('configureServer.servername')"
              expanded
            />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input
              ref="v2ray_add"
              v-model="v2ray.add"
              required
              placeholder="IP / HOST"
              expanded
            />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input
              ref="v2ray_port"
              v-model="v2ray.port"
              required
              :placeholder="$t('configureServer.port')"
              type="number"
              expanded
            />
          </b-field>
          <b-field label="ID" label-position="on-border">
            <b-input
              ref="v2ray_id"
              v-model="v2ray.id"
              required
              placeholder="UserID"
              expanded
            />
          </b-field>
          <b-field
            v-if="v2ray.protocol === 'vmess'"
            label="AlterID"
            label-position="on-border"
          >
            <b-input
              ref="v2ray_aid"
              v-model="v2ray.aid"
              placeholder="AlterID"
              type="number"
              min="0"
              max="65535"
              required
              expanded
            />
          </b-field>
          <b-field
            v-if="v2ray.protocol === 'vmess'"
            label="Security"
            label-position="on-border"
          >
            <b-select v-model="v2ray.scy" expanded>
              <option value="auto">Auto</option>
              <option value="aes-128-gcm">aes-128-gcm</option>
              <option value="chacha20-poly1305">chacha20-poly1305</option>
              <option value="none">none</option>
              <option value="zero">zero</option>
            </b-select>
          </b-field>
          <b-field
            v-show="v2ray.type !== 'dtls'"
            label="TLS"
            label-position="on-border"
          >
            <b-select v-model="v2ray.tls" expanded @input="handleNetworkChange">
              <option value="none">{{ $t("setting.options.off") }}</option>
              <option value="tls">tls</option>
              <option
                v-if="v2ray.protocol === 'vless' && vlessVersion >= 2"
                value="xtls"
                >xtls</option
              >
            </b-select>
          </b-field>
          <b-field
            v-show="v2ray.tls === 'xtls'"
            label="Flow"
            label-position="on-border"
          >
            <b-autocomplete
              v-model="v2ray.flow"
              open-on-focus
              placeholder="xtls-rprx-direct"
              :data="filteredDataArray"
              @select="option => (flowSelected = option)"
            >
            </b-autocomplete>
          </b-field>
          <b-field v-show="v2ray.tls !== 'none'" label-position="on-border">
            <template slot="label">
              AllowInsecure
              <b-tooltip
                v-show="v2ray.protocol === 'vless'"
                type="is-dark"
                :label="$t('server.messages.notRecommend', { name: 'VLESS' })"
                multilined
                position="is-right"
              >
                <b-icon
                  size="is-small"
                  icon=" iconfont icon-help-circle-outline"
                  style="position:relative;top:2px;right:3px;font-weight:normal"
                />
              </b-tooltip>
            </template>
            <b-select
              ref="v2ray_allow_insecure"
              v-model="v2ray.allowInsecure"
              expanded
              required
            >
              <option :value="false">{{ $t("operations.no") }}</option>
              <option :value="true">{{ $t("operations.yes") }}</option>
            </b-select>
          </b-field>
          <b-field label="Network" label-position="on-border">
            <b-select
              ref="v2ray_net"
              v-model="v2ray.net"
              expanded
              required
              @input="handleNetworkChange"
            >
              <option value="tcp">TCP</option>
              <option value="kcp">mKCP</option>
              <option value="ws">WebSocket</option>
              <option value="h2">HTTP/2</option>
              <option value="grpc">gRPC</option>
            </b-select>
          </b-field>
          <b-field
            v-show="v2ray.net === 'tcp'"
            label="Type"
            label-position="on-border"
          >
            <b-select v-model="v2ray.type" expanded>
              <option value="none"
                >{{ $t("configureServer.noObfuscation") }}
              </option>
              <option value="http"
                >{{ $t("configureServer.httpObfuscation") }}
              </option>
            </b-select>
          </b-field>
          <b-field
            v-show="v2ray.net === 'kcp'"
            label="Type"
            label-position="on-border"
          >
            <b-select v-model="v2ray.type" expanded>
              <option value="none"
                >{{ $t("configureServer.noObfuscation") }}
              </option>
              <option value="srtp"
                >{{ $t("configureServer.srtpObfuscation") }}
              </option>
              <option value="utp"
                >{{ $t("configureServer.utpObfuscation") }}
              </option>
              <option value="wechat-video"
                >{{ $t("configureServer.wechatVideoObfuscation") }}
              </option>
              <option value="dtls"
                >{{
                  `${$t("configureServer.dtlsObfuscation")}(${$t(
                    "configureServer.forceTLS"
                  )})`
                }}
              </option>
              <option value="wireguard"
                >{{ $t("configureServer.wireguardObfuscation") }}
              </option>
            </b-select>
          </b-field>
          <b-field
            v-show="
              v2ray.net === 'ws' ||
                v2ray.net === 'h2' ||
                v2ray.tls === 'tls' ||
                v2ray.tls === 'xtls' ||
                (v2ray.net === 'tcp' && v2ray.type === 'http')
            "
            label="Host"
            label-position="on-border"
          >
            <b-input
              v-model="v2ray.host"
              :placeholder="$t('configureServer.hostObfuscation')"
              expanded
            />
          </b-field>
          <b-field
            v-show="v2ray.tls === 'tls' || v2ray.tls === 'xtls'"
            label="Alpn"
            label-position="on-border"
          >
            <b-input v-model="v2ray.alpn" placeholder="h2,http/1.1" expanded />
          </b-field>
          <b-field
            v-show="
              v2ray.net === 'ws' ||
                v2ray.net === 'h2' ||
                (v2ray.net === 'tcp' && v2ray.type === 'http')
            "
            label="Path"
            label-position="on-border"
          >
            <b-input
              v-model="v2ray.path"
              :placeholder="$t('configureServer.pathObfuscation')"
              expanded
            />
          </b-field>
          <b-field
            v-show="v2ray.net === 'mkcp' || v2ray.net === 'kcp'"
            label="Seed"
            label-position="on-border"
          >
            <b-input
              v-model="v2ray.path"
              :placeholder="$t('configureServer.seedObfuscation')"
              expanded
            />
          </b-field>
          <b-field
            v-show="v2ray.net === 'grpc'"
            label="ServiceName"
            label-position="on-border"
          >
            <b-input
              ref="v2ray_service_name"
              v-model="v2ray.path"
              type="text"
              expanded
            />
          </b-field>
        </b-tab-item>
        <b-tab-item label="SS">
          <b-field label="Name" label-position="on-border">
            <b-input
              ref="ss_name"
              v-model="ss.name"
              :placeholder="$t('configureServer.servername')"
              expanded
            />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input
              ref="ss_server"
              v-model="ss.server"
              required
              placeholder="IP / HOST"
              expanded
            />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input
              ref="ss_port"
              v-model="ss.port"
              required
              :placeholder="$t('configureServer.port')"
              type="number"
              expanded
            />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input
              ref="ss_password"
              v-model="ss.password"
              required
              :placeholder="$t('configureServer.password')"
              expanded
            />
          </b-field>
          <b-field label="Method" label-position="on-border">
            <b-select ref="ss_method" v-model="ss.method" expanded required>
              <option value="aes-128-gcm">aes-128-gcm</option>
              <option value="aes-256-gcm">aes-256-gcm</option>
              <option value="chacha20-poly1305">chacha20-poly1305</option>
              <option value="chacha20-ietf-poly1305"
                >chacha20-ietf-poly1305</option
              >
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
          <b-field
            v-if="ss.plugin === 'simple-obfs' || ss.plugin === 'v2ray-plugin'"
            label-position="on-border"
            class="with-icon-alert"
          >
            <template slot="label">
              Impl
              <b-tooltip
                type="is-dark"
                :label="$t('setting.messages.ssPluginImpl')"
                multilined
                position="is-right"
              >
                <b-icon
                  size="is-samll"
                  icon=" iconfont icon-help-circle-outline"
                  style="position:relative;top:2px;right:3px;font-weight:normal"
                />
              </b-tooltip>
            </template>
            <b-select ref="ss_plugin_impl" v-model="ss.impl" expanded>
              <option value="">{{ $t("setting.options.default") }}</option>
              <option value="chained">chained</option>
              <option value="transport">transport</option>
            </b-select>
          </b-field>
          <b-field
            v-show="ss.plugin === 'simple-obfs'"
            label="Obfs"
            label-position="on-border"
          >
            <b-select ref="ss_obfs" v-model="ss.obfs" expanded>
              <option value="http">http</option>
              <option value="tls">tls</option>
            </b-select>
          </b-field>
          <b-field
            v-show="ss.plugin === 'v2ray-plugin'"
            label="Mode"
            label-position="on-border"
          >
            <b-select ref="ss_mode" v-model="ss.mode" expanded>
              <option value="websocket">websocket</option>
            </b-select>
          </b-field>
          <b-field
            v-show="ss.plugin === 'v2ray-plugin'"
            label="TLS"
            label-position="on-border"
          >
            <b-select ref="ss_tls" v-model="ss.tls" expanded>
              <option value="">{{ $t("setting.options.off") }}</option>
              <option value="tls">tls</option>
            </b-select>
          </b-field>
          <b-field
            v-if="
              (ss.plugin === 'simple-obfs' &&
                (ss.obfs === 'http' || ss.obfs === 'tls')) ||
                ss.plugin === 'v2ray-plugin'
            "
            label="Host"
            label-position="on-border"
          >
            <b-input
              ref="ss_host"
              v-model="ss.host"
              placeholder="(optional)"
              expanded
            />
          </b-field>
          <b-field
            v-if="
              (ss.plugin === 'simple-obfs' && ss.obfs === 'http') ||
                ss.plugin === 'v2ray-plugin'
            "
            label="Path"
            label-position="on-border"
          >
            <b-input ref="ss_path" v-model="ss.path" placeholder="/" expanded />
          </b-field>
        </b-tab-item>
        <b-tab-item label="SSR">
          <b-field label="Name" label-position="on-border">
            <b-input
              ref="ssr_name"
              v-model="ssr.name"
              :placeholder="$t('configureServer.servername')"
              expanded
            />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input
              ref="ssr_server"
              v-model="ssr.server"
              required
              placeholder="IP / HOST"
              expanded
            />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input
              ref="ssr_port"
              v-model="ssr.port"
              required
              :placeholder="$t('configureServer.port')"
              type="number"
              expanded
            />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input
              ref="ssr_password"
              v-model="ssr.password"
              required
              :placeholder="$t('configureServer.password')"
              expanded
            />
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
          <b-field
            v-if="ssr.proto !== 'origin'"
            label="Protocol Param"
            label-position="on-border"
          >
            <b-input
              ref="ssr_protoParam"
              v-model="ssr.protoParam"
              placeholder="(optional)"
              expanded
            />
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
          <b-field
            v-if="ssr.obfs !== 'plain'"
            label="Obfs Param"
            label-position="on-border"
          >
            <b-input
              ref="ssr_obfsParam"
              v-model="ssr.obfsParam"
              placeholder="(optional)"
              expanded
            />
          </b-field>
        </b-tab-item>
        <b-tab-item label="PingTunnel">
          <b-field label="Name" label-position="on-border">
            <b-input
              ref="pingtunnel_name"
              v-model="pingtunnel.name"
              :placeholder="$t('configureServer.servername')"
              expanded
            />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input
              ref="pingtunnel_server"
              v-model="pingtunnel.server"
              required
              placeholder="IP / HOST"
              expanded
            />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input
              ref="pingtunnel_password"
              v-model="pingtunnel.password"
              required
              :placeholder="$t('configureServer.password')"
              type="number"
              expanded
            />
          </b-field>
        </b-tab-item>
        <b-tab-item label="Trojan">
          <b-field label="Name" label-position="on-border">
            <b-input
              ref="trojan_name"
              v-model="trojan.name"
              :placeholder="$t('configureServer.servername')"
              expanded
            />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input
              ref="trojan_server"
              v-model="trojan.server"
              required
              placeholder="IP / HOST"
              expanded
            />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input
              ref="trojan_port"
              v-model="trojan.port"
              required
              :placeholder="$t('configureServer.port')"
              type="number"
              expanded
            />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input
              ref="trojan_password"
              v-model="trojan.password"
              required
              :placeholder="$t('configureServer.password')"
              expanded
            />
          </b-field>
          <b-field label="Protocol" label-position="on-border">
            <b-select
              ref="trojan_method"
              v-model="trojan.method"
              expanded
              required
            >
              <option value="origin">{{ $t("configureServer.origin") }}</option>
              <option value="shadowsocks">shadowsocks</option>
            </b-select>
          </b-field>
          <b-field
            v-if="trojan.method === 'shadowsocks'"
            label="Shadowsocks Cipher"
            label-position="on-border"
          >
            <b-select
              ref="trojan_ss_cipher"
              v-model="trojan.ssCipher"
              expanded
              required
            >
              <option value="aes-128-gcm">aes-128-gcm</option>
              <option value="aes-256-gcm">aes-256-gcm</option>
              <option value="chacha20-poly1305">chacha20-poly1305</option>
              <option value="chacha20-ietf-poly1305"
                >chacha20-ietf-poly1305</option
              >
            </b-select>
          </b-field>
          <b-field
            v-if="trojan.method === 'shadowsocks'"
            label="Shadowsocks Password"
            label-position="on-border"
          >
            <b-input
              ref="trojan_ss_password"
              v-model="trojan.ssPassword"
              required
              :placeholder="`shadowsocks${$t('configureServer.password')}`"
              expanded
            />
          </b-field>
          <b-field label-position="on-border">
            <template slot="label">
              AllowInsecure
              <b-tooltip
                v-show="trojan.method !== 'origin' || trojan.obfs !== 'none'"
                type="is-dark"
                :label="
                  $t('server.messages.notAllowInsecure', { name: 'Trojan-Go' })
                "
                multilined
                position="is-right"
              >
                <b-icon
                  size="is-small"
                  icon=" iconfont icon-help-circle-outline"
                  style="position:relative;top:2px;right:3px;font-weight:normal"
                />
              </b-tooltip>
            </template>
            <b-select
              ref="trojan_allow_insecure"
              v-model="trojan.allowInsecure"
              expanded
              required
            >
              <option :value="false">{{ $t("operations.no") }}</option>
              <option
                :value="true"
                :disabled="trojan.method !== 'origin' || trojan.obfs !== 'none'"
                >{{ $t("operations.yes") }}</option
              >
            </b-select>
          </b-field>
          <b-field label="SNI(Peer)" label-position="on-border">
            <b-input v-model="trojan.peer" placeholder="SNI(Peer)" expanded />
          </b-field>
          <b-field label="Obfs" label-position="on-border">
            <b-select ref="trojan_obfs" v-model="trojan.obfs" expanded required>
              <option value="none">{{
                $t("configureServer.noObfuscation")
              }}</option>
              <option value="websocket">websocket</option>
            </b-select>
          </b-field>
          <b-field
            v-show="trojan.obfs === 'websocket'"
            label="Websocket Host"
            label-position="on-border"
          >
            <b-input v-model="trojan.host" expanded />
          </b-field>
          <b-field
            v-show="trojan.obfs === 'websocket'"
            label="Websocket Path"
            label-position="on-border"
          >
            <b-input v-model="trojan.path" placeholder="/" expanded />
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
            <b-input
              ref="http_name"
              v-model="http.name"
              :placeholder="$t('configureServer.servername')"
              expanded
            />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input
              ref="http_host"
              v-model="http.host"
              required
              placeholder="IP / HOST"
              expanded
            />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input
              ref="http_port"
              v-model="http.port"
              required
              :placeholder="$t('configureServer.port')"
              type="number"
              expanded
            />
          </b-field>
          <b-field label="Username" label-position="on-border">
            <b-input
              ref="http_username"
              v-model="http.username"
              :placeholder="$t('configureServer.username')"
              expanded
            />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input
              ref="http_password"
              v-model="http.password"
              :placeholder="$t('configureServer.password')"
              expanded
            />
          </b-field>
        </b-tab-item>

        <b-tab-item label="SOCKS5">
          <b-field label="Name" label-position="on-border">
            <b-input
              ref="socks5_name"
              v-model="socks5.name"
              :placeholder="$t('configureServer.servername')"
              expanded
            />
          </b-field>
          <b-field label="Host" label-position="on-border">
            <b-input
              ref="socks5_host"
              v-model="socks5.host"
              required
              placeholder="IP / HOST"
              expanded
            />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input
              ref="socks5_port"
              v-model="socks5.port"
              required
              :placeholder="$t('configureServer.port')"
              type="number"
              expanded
            />
          </b-field>
          <b-field label="Username" label-position="on-border">
            <b-input
              ref="socks5_username"
              v-model="socks5.username"
              :placeholder="$t('configureServer.username')"
              expanded
            />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input
              ref="socks5_password"
              v-model="socks5.password"
              :placeholder="$t('configureServer.password')"
              expanded
            />
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
import { handleResponse } from "@/assets/js/utils";
import { Base64 } from "js-base64";
import { parseURL, generateURL } from "@/assets/js/utils";

export default {
  name: "ModalServer",
  props: {
    which: {
      type: Object,
      default() {
        return null;
      }
    },
    readonly: {
      type: Boolean,
      default: false
    }
  },
  data: () => ({
    showVLess: false,
    vlessVersion: 0,
    v2ray: {
      ps: "",
      add: "",
      port: "",
      id: "",
      aid: "",
      net: "tcp",
      type: "none",
      host: "",
      path: "",
      tls: "none",
      flow: "xtls-rprx-direct",
      alpn: "",
      scy: "",
      v: "",
      allowInsecure: false,
      protocol: "vmess"
    },
    ss: {
      method: "aes-128-gcm",
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
      impl: ""
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
      protocol: "ssr"
    },
    pingtunnel: {
      name: "",
      server: "",
      password: "",
      protocol: "pingtunnel"
    },
    trojan: {
      name: "",
      server: "",
      peer: "" /* tls sni */,
      host: "" /* websocket host */,
      path: "" /* websocket path */,
      allowInsecure: false,
      port: "",
      password: "",
      method: "origin" /* shadowsocks */,
      ssCipher: "aes-128-gcm",
      ssPassword: "",
      obfs: "none" /* websocket */,
      protocol: "trojan"
    },
    http: {
      username: "",
      password: "",
      host: "",
      port: "",
      protocol: "http",
      name: ""
    },
    socks5: {
      username: "",
      password: "",
      host: "",
      port: "",
      protocol: "socks5",
      name: ""
    },
    tabChoice: 0,
    presetFlows: [
      "xtls-rprx-direct",
      "xtls-rprx-direct-udp443",
      "xtls-rprx-splice",
      "xtls-rprx-splice-udp443"
    ],
    flowSelected: null
  }),
  computed: {
    filteredDataArray() {
      return this.presetFlows.filter(option => {
        return option.toString().indexOf(this.v2ray.flow) >= 0;
      });
    }
  },
  mounted() {
    if (localStorage["vlessValid"] === "true") {
      this.showVLess = true;
      this.vlessVersion = 1;
    } else {
      const t = parseInt(localStorage["vlessValid"]);
      if (!isNaN(t) && t > 0) {
        this.showVLess = true;
        this.vlessVersion = t;
      }
    }
    if (this.which !== null) {
      this.$axios({
        url: apiRoot + "/sharingAddress",
        method: "get",
        params: {
          touch: this.which
        }
      }).then(res => {
        handleResponse(res, this, () => {
          if (
            res.data.data.sharingAddress.toLowerCase().startsWith("vmess://") ||
            res.data.data.sharingAddress.toLowerCase().startsWith("vless://")
          ) {
            this.v2ray = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 0;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("ss://")
          ) {
            this.ss = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 1;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("ssr://")
          ) {
            this.ssr = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 2;
          } else if (
            res.data.data.sharingAddress
              .toLowerCase()
              .startsWith("pingtunnel://") ||
            res.data.data.sharingAddress
              .toLowerCase()
              .startsWith("ping-tunnel://")
          ) {
            this.pingtunnel = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 3;
          } else if (
            res.data.data.sharingAddress
              .toLowerCase()
              .startsWith("trojan://") ||
            res.data.data.sharingAddress
              .toLowerCase()
              .startsWith("trojan-go://")
          ) {
            this.trojan = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 4;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("http://") ||
            res.data.data.sharingAddress.toLowerCase().startsWith("https://")
          ) {
            this.http = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 5;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("socks5://")
          ) {
            this.socks5 = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 6;
          }
          this.$nextTick(() => {
            if (this.readonly) {
              this.$refs.section
                .querySelectorAll("input, textarea")
                .forEach(x => (x.readOnly = "readOnly"));
              this.$refs.section.querySelectorAll("select").forEach(x => {
                const text = x.querySelector(`option[value="${x.value}"]`)
                  .textContent;
                console.log(x.value, text);
                x.outerHTML = `<input type="text" class="input" readonly="readonly" value="${text}">`;
              });
            }
          });
        });
      });
    }
  },
  methods: {
    handleV2rayProtocolSwitch() {
      if (this.v2ray.tls === "xtls" && this.v2ray.protocol === "vmess") {
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
          net: u.params.type || "tcp",
          type: u.params.headerType || "none",
          host: u.params.sni || u.params.host || "",
          path: u.params.path || u.params.serviceName || "",
          alpn: u.params.alpn || "",
          tls: u.params.security || "none",
          flow: u.params.flow || "xtls-rprx-direct",
          allowInsecure: u.params.allowInsecure || false,
          protocol: "vless"
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
        let mp;
        if (!u.password) {
          try {
            u.username = Base64.decode(decodeURIComponent(u.username));
            mp = u.username.split(":");
            if (mp.length > 2) {
              mp[1] = mp.slice(1).join(":");
              mp = mp.slice(0, 2);
            }
          } catch (e) {
            //pass
          }
        } else {
          mp = [u.username, u.password];
        }
        u.hash = decodeURIComponent(u.hash);
        let obj = {
          method: mp[0],
          password: mp[1],
          server: u.host,
          port: u.port,
          name: u.hash,
          obfs: "http",
          plugin: "",
          protocol: "ss",
          impl: ""
        };
        if (u.params.plugin) {
          u.params.plugin = decodeURIComponent(u.params.plugin);
          const arr = u.params.plugin.split(";");
          obj.plugin = arr[0];
          switch (obj.plugin) {
            case "obfs-local":
            case "simpleobfs":
              obj.plugin = "simple-obfs";
              break;
            case "v2ray-plugin":
              obj.tls = "";
              obj.mode = "websocket";
              break;
          }
          for (let i = 1; i < arr.length; i++) {
            //"obfs-local;obfs=tls;obfs-host=4cb6a43103.wns.windows.com"
            const a = arr[i].split("=");
            switch (a[0]) {
              case "obfs":
                obj.obfs = a[1];
                break;
              case "host":
              case "obfs-host":
                obj.host = a[1];
                break;
              case "path":
              case "obfs-path":
                obj.path = a[1];
                break;
              case "mode":
                obj.mode = a[1];
                break;
              case "tls":
                obj.tls = "tls";
                break;
              case "impl":
                obj.impl = a[1];
            }
          }
        }
        return obj;
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
          protocol: "ssr"
        };
      } else if (url.toLowerCase().startsWith("pingtunnel://")) {
        let u = url.substr(13);
        u = Base64.decode(u);
        const regexp = /(.+):(.+)#(.*)/;
        let arr = regexp.exec(u);
        return {
          server: arr[1],
          password: Base64.decode(arr[2]),
          name: decodeURIComponent(arr[3]),
          protocol: "pingtunnel"
        };
      } else if (url.toLowerCase().startsWith("ping-tunnel://")) {
        let u = parseURL(url);
        return {
          server: u.host,
          password: decodeURIComponent(u.username),
          name: decodeURIComponent(u.hash),
          protocol: "pingtunnel"
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
          allowInsecure:
            u.params.allowInsecure === true || u.params.allowInsecure === "1",
          method: "origin",
          obfs: "none",
          ssCipher: "aes-128-gcm",
          protocol: "trojan"
        };
        if (url.toLowerCase().startsWith("" + "")) {
          console.log(u.params.encryption);
          if (u.params.encryption?.startsWith("ss;")) {
            o.method = "shadowsocks";
            const fields = u.params.encryption.split(";");
            o.ssCipher = fields[1];
            o.ssPassword = fields[2];
          }
          const obfsMap = {
            original: "none",
            "": "none",
            ws: "websocket"
          };
          o.obfs = obfsMap[u.params.type || ""];
          if (o.obfs === "ws") {
            o.obfs = "websocket";
          }
          if (o.obfs === "websocket") {
            o.host = u.params.host || "";
            o.path = u.params.path || "/";
          }
        }
        return o;
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
          name: decodeURIComponent(u.hash)
        };
      } else if (url.toLowerCase().startsWith("socks5://")) {
        let u = parseURL(url);
        return {
          username: decodeURIComponent(u.username),
          password: decodeURIComponent(u.password),
          host: u.host,
          port: u.port,
          protocol: u.protocol,
          name: decodeURIComponent(u.hash)
        };
      }
      return null;
    },
    generateURL(srcObj) {
      let obj = {};
      let query = {};
      let tmp;
      switch (srcObj.protocol) {
        case "vless":
          // https://github.com/XTLS/Xray-core/discussions/716
          query = {
            type: srcObj.net,
            security: srcObj.tls,
            path: srcObj.path,
            host: srcObj.host,
            headerType: srcObj.type,
            sni: srcObj.host,
            flow: srcObj.flow,
            allowInsecure: srcObj.allowInsecure
          };
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
            protocol: "vless",
            username: srcObj.id,
            host: srcObj.add,
            port: srcObj.port,
            hash: srcObj.ps,
            params: query
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
          if (!(obj.protocol === "vless" && obj.tls === "xtls")) {
            delete obj.flow;
          }
          return "vmess://" + Base64.encode(JSON.stringify(obj));
        case "ss":
          /* ss://BASE64(method:password)@server:port#name */
          tmp = `ss://${Base64.encode(`${srcObj.method}:${srcObj.password}`)}@${
            srcObj.server
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
            `${srcObj.server}:${srcObj.port}:${srcObj.proto}:${srcObj.method}:${
              srcObj.obfs
            }:${Base64.encodeURI(srcObj.password)}/?remarks=${Base64.encodeURI(
              srcObj.name
            )}&protoparam=${Base64.encodeURI(
              srcObj.protoParam
            )}&obfsparam=${Base64.encodeURI(srcObj.obfsParam)}`
          )}`;
        case "pingtunnel":
          return generateURL({
            protocol: "ping-tunnel",
            username: srcObj.password,
            host: srcObj.server,
            hash: srcObj.name
          });
        case "trojan":
          /* trojan://password@server:port?allowInsecure=1&sni=sni#URIESCAPE(name) */
          query = {
            allowInsecure: srcObj.allowInsecure
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
            delete query.allowInsecure;
          }
          return generateURL({
            protocol: tmp,
            username: srcObj.password,
            host: srcObj.server,
            port: srcObj.port,
            hash: srcObj.name,
            params: query
          });
        case "http":
        case "https":
          tmp = {
            protocol: srcObj.protocol + "-proxy",
            host: srcObj.host,
            port: srcObj.port,
            hash: srcObj.name
          };
          if (srcObj.username && srcObj.password) {
            Object.assign(tmp, {
              username: srcObj.username,
              password: srcObj.password
            });
          }
          return generateURL(tmp);
        case "socks5":
          tmp = {
            protocol: "socks5",
            host: srcObj.host,
            port: srcObj.port,
            hash: srcObj.name
          };
          if (srcObj.username && srcObj.password) {
            Object.assign(tmp, {
              username: srcObj.username,
              password: srcObj.password
            });
          }
          return generateURL(tmp);
      }
      return null;
    },
    handleNetworkChange() {
      this.v2ray.type = "none";
      if (this.v2ray.tls === "xtls" && this.v2ray.net === "ws") {
        this.$buefy.toast.open({
          message: this.$t("setting.messages.xtlsNotWithWs"),
          type: "is-warning",
          position: "is-top",
          queue: false,
          duration: 5000
        });
        this.$nextTick(() => {
          this.v2ray.tls = "tls";
        });
      } else if (this.v2ray.tls === "xtls" && !this.v2ray.flow) {
        this.v2ray.flow = this.presetFlows[0];
      } else if (this.v2ray.tls === "none" && this.v2ray.net === "grpc") {
        this.$buefy.toast.open({
          message: this.$t("setting.messages.grpcShouldWithTls"),
          type: "is-warning",
          position: "is-top",
          queue: false,
          duration: 5000
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
        if (this.tabChoice === 1 && !k.startsWith("ss_")) {
          continue;
        }
        if (this.tabChoice === 2 && !k.startsWith("ssr_")) {
          continue;
        }
        if (this.tabChoice === 3 && !k.startsWith("pingtunnel_")) {
          continue;
        }
        if (this.tabChoice === 4 && !k.startsWith("trojan_")) {
          continue;
        }
        if (this.tabChoice === 5 && !k.startsWith("http_")) {
          continue;
        }
        if (this.tabChoice === 6 && !k.startsWith("socks5_")) {
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
      if (this.tabChoice === 0) {
        if (
          this.v2ray.allowInsecure === true || // sometimes bool, sometimes string
          this.v2ray.allowInsecure === "true"
        ) {
          const { result } = await this.$buefy.dialog.confirm({
            title: this.$t("InSecureConfirm.title"),
            message: this.$t("InSecureConfirm.message"),
            confirmText: this.$t("InSecureConfirm.confirm"),
            cancelText: this.$t("InSecureConfirm.cancel"),
            type: "is-danger",
            hasIcon: true,
            onConfirm: () => true,
            onCancel: () => false
          });
          if (!result) {
            return;
          }
        }
        coded = this.generateURL(this.v2ray);
      } else if (this.tabChoice === 1) {
        coded = this.generateURL(this.ss);
      } else if (this.tabChoice === 2) {
        coded = this.generateURL(this.ssr);
      } else if (this.tabChoice === 3) {
        coded = this.generateURL(this.pingtunnel);
      } else if (this.tabChoice === 4) {
        coded = this.generateURL(this.trojan);
      } else if (this.tabChoice === 5) {
        coded = this.generateURL(this.http);
      } else if (this.tabChoice === 6) {
        coded = this.generateURL(this.socks5);
      }
      this.$emit("submit", coded);
    }
  }
};
</script>

<style lang="scss">
.is-twitter .is-active a {
  color: #4099ff !important;
}

.readonly {
  pointer-events: none;
}

.same-width-5 li {
  min-width: 5em;
  width: unset !important;
}
</style>
