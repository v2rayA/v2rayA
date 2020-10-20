<template>
  <div class="modal-card" style="max-width: 520px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $tc("configureServer.title", readonly ? 2 : 1) }}
      </p>
    </header>
    <section :class="{ 'modal-card-body': true, readonly: readonly }">
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
          <b-field label="Address" label-position="on-border">
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
            v-show="v2ray.protocol !== 'vless'"
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
            <b-select v-model="v2ray.flow" expanded>
              <option value="xtls-rprx-origin">xtls-rprx-origin</option>
              <option value="xtls-rprx-origin-udp443"
                >xtls-rprx-origin-udp443</option
              >
              <option v-if="vlessVersion >= 3" value="xtls-rprx-direct"
                >xtls-rprx-direct</option
              >
              <option v-if="vlessVersion >= 3" value="xtls-rprx-direct-udp443"
                >xtls-rprx-direct-udp443</option
              >
            </b-select>
          </b-field>
          <b-field
            v-show="v2ray.tls === 'tls'"
            label="AllowInsecure"
            label-position="on-border"
          >
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
                v2ray.tls === 'xtls'
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
            v-show="v2ray.net === 'ws' || v2ray.net === 'h2'"
            label="Path"
            label-position="on-border"
          >
            <b-input
              v-model="v2ray.path"
              :placeholder="$t('configureServer.pathObfuscation')"
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
          <b-field label="Address" label-position="on-border">
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
              <option value="camellia-128-cfb">camellia-128-cfb</option>
              <option value="camellia-192-cfb">camellia-192-cfb</option>
              <option value="camellia-256-cfb">camellia-256-cfb</option>
              <option value="cast5-cfb">cast5-cfb</option>
              <option value="chacha20">chacha20</option>
              <option value="chacha20-ietf">chacha20-ietf</option>
              <option value="chacha20-poly1305">chacha20-poly1305</option>
              <option value="chacha20-ietf-poly1305"
                >chacha20-ietf-poly1305</option
              >
              <option value="idea-cfb">idea-cfb</option>
              <option value="rc2-cfb">rc2-cfb</option>
              <option value="rc4-md5">rc4-md5</option>
              <option value="salsa20">salsa20</option>
              <option value="seed-cfb">seed-cfb</option>
              <option value="plain">plain</option>
              <option value="none">none</option>
            </b-select>
          </b-field>
          <b-field label="Obfs" label-position="on-border">
            <b-select ref="ss_obfs" v-model="ss.obfs" expanded required>
              <option value="">{{ $t("setting.options.off") }}</option>
              <option value="http">http</option>
              <option value="tls">tls</option>
            </b-select>
          </b-field>
          <b-field
            v-if="ss.obfs === 'http'"
            label="Path"
            label-position="on-border"
          >
            <b-input ref="ss_path" v-model="ss.path" expanded />
          </b-field>
          <b-field
            v-if="ss.obfs === 'http' || ss.obfs === 'tls'"
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
          <b-field label="Address" label-position="on-border">
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
          <b-field label="Address" label-position="on-border">
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
          <b-field label="Address" label-position="on-border">
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
          <b-field label="Peer" label-position="on-border">
            <b-input
              v-model="trojan.peer"
              :placeholder="`Peer(${$t('common.optional')})`"
              expanded
            />
          </b-field>
          <b-field label="AllowInsecure" label-position="on-border">
            <b-select
              ref="trojan_allow_insecure"
              v-model="trojan.allowInsecure"
              expanded
              required
            >
              <option :value="false">{{ $t("operations.no") }}</option>
              <option :value="true">{{ $t("operations.yes") }}</option>
            </b-select>
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
import { parseURL, generateURL } from "../assets/js/utils";

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
      flow: "xtls-rprx-origin",
      v: "",
      allowInsecure: false,
      protocol: "vmess"
    },
    ss: {
      method: "aes-128-gcm",
      obfs: "",
      path: "/",
      host: "",
      password: "",
      server: "",
      port: "",
      name: "",
      protocol: "ss"
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
      peer: "",
      allowInsecure: false,
      port: "",
      password: "",
      protocol: "trojan"
    },
    tabChoice: 0
  }),
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
            res.data.data.sharingAddress.toLowerCase().startsWith("vmess://")
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
              .startsWith("pingtunnel://")
          ) {
            this.pingtunnel = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 3;
          } else if (
            res.data.data.sharingAddress.toLowerCase().startsWith("trojan://")
          ) {
            this.trojan = this.resolveURL(res.data.data.sharingAddress);
            this.tabChoice = 4;
          }
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
      if (url.toLowerCase().indexOf("vmess://") >= 0) {
        let obj = JSON.parse(
          Base64.decode(url.substring(url.indexOf("://") + 3))
        );
        // console.log(obj);
        obj.ps = decodeURIComponent(obj.ps);
        obj.tls = obj.tls || "none";
        obj.type = obj.type || "none";
        obj.protocol = obj.protocol || "vmess";
        return obj;
      } else if (url.toLowerCase().indexOf("ss://") >= 0) {
        let u = parseURL(url);
        try {
          u.username = Base64.decode(decodeURIComponent(u.username));
        } catch (e) {
          //pass
        }
        let mp = u.username.split(":");
        u.hash = decodeURIComponent(u.hash);
        let obj = {
          method: mp[0],
          password: mp[1],
          server: u.host,
          port: u.port,
          name: u.hash,
          protocol: "ss"
        };
        if (u.params.plugin) {
          u.params.plugin = decodeURIComponent(u.params.plugin);
          const arr = u.params.plugin.split(";");
          for (let i = 1; i < arr.length; i++) {
            //"obfs-local;obfs=tls;obfs-host=4cb6a43103.wns.windows.com"
            const a = arr[i].split("=");
            switch (a[0]) {
              case "obfs":
                obj.obfs = a[1];
                break;
              case "obfs-host":
                obj.host = a[1];
                break;
              case "obfs-path":
                obj.path = a[1];
            }
          }
        } else {
          obj.obfs = "";
        }
        return obj;
      } else if (url.toLowerCase().indexOf("ssr://") >= 0) {
        url = Base64.decode(url.substr(6));
        let arr = url.split("/?");
        let params = arr[1].split("&");
        let m = {};
        for (let param of params) {
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
      } else if (url.toLowerCase().indexOf("pingtunnel://") >= 0) {
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
      } else if (url.toLowerCase().indexOf("trojan://") >= 0) {
        let u = parseURL(url);
        return {
          password: u.username,
          server: u.host,
          port: u.port,
          name: u.hash,
          peer: u.params.peer || "",
          allowInsecure:
            u.params.allowInsecure === true || u.params.allowInsecure === "1",
          protocol: "trojan"
        };
      }
      return null;
    },
    generateURL(srcObj) {
      let obj = {};
      let params = {};
      let s;
      switch (srcObj.protocol) {
        case "vless":
        //FIXME: 临时方案
        // eslint-disable-next-line no-fallthrough
        case "vmess":
          //尽量减少生成的链接长度
          obj = Object.assign({}, srcObj);
          switch (obj.net) {
            case "kcp":
            case "tcp":
              obj.path = "";
              if (obj.tls === "" || obj.tls === "none") {
                obj.host = "";
              }
              break;
            default:
              obj.type = "";
          }
          if (!(obj.protocol === "vless" && obj.tls === "xtls")) {
            delete obj.flow;
          }
          return "vmess://" + Base64.encode(JSON.stringify(obj));
        case "ss":
          /* ss://BASE64(method:password)@server:port#name */
          //TODO: simpleobfs
          s = `ss://${Base64.encode(`${srcObj.method}:${srcObj.password}`)}@${
            srcObj.server
          }:${srcObj.port}/`;
          if (srcObj.obfs !== "") {
            s += `?plugin=${encodeURIComponent(
              `obfs-local;obfs=${srcObj.obfs};obfs-host=${srcObj.host}${
                srcObj.obfs === "http" ? `;obfs-path=${srcObj.path}` : ""
              }`
            )}`;
          }
          s += srcObj.name.length ? `#${Base64.encodeURI(srcObj.name)}` : "";
          return s;

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
          return `pingtunnel://${Base64.encode(
            `${srcObj.server}:${Base64.encodeURI(srcObj.password)}` +
              (srcObj.name.length ? `#${encodeURIComponent(srcObj.name)}` : "")
          )}`;
        case "trojan":
          /* trojan://password@server:port?allowInsecure=1&peer=peer#URIESCAPE(name) */
          params = { allowInsecure: srcObj.allowInsecure };
          if (srcObj.peer !== "") {
            params.peer = srcObj.peer;
          }
          return generateURL({
            protocol: "trojan",
            username: srcObj.password,
            host: srcObj.server,
            port: srcObj.port,
            hash: srcObj.name,
            params
          });
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
        this.v2ray.flow = "xtls-rprx-origin";
      }
    },
    handleClickSubmit() {
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
        let x = this.$refs[k];
        if (!x) {
          continue;
        }
        if (
          x.hasOwnProperty("checkHtml5Validity") &&
          typeof x.checkHtml5Validity === "function" &&
          !x.checkHtml5Validity()
        ) {
          valid = false;
        }
      }
      if (!valid) {
        return;
      }
      let coded = "";
      if (this.tabChoice === 0) {
        coded = this.generateURL(this.v2ray);
      } else if (this.tabChoice === 1) {
        coded = this.generateURL(this.ss);
      } else if (this.tabChoice === 2) {
        coded = this.generateURL(this.ssr);
      } else if (this.tabChoice === 3) {
        coded = this.generateURL(this.pingtunnel);
      } else if (this.tabChoice === 4) {
        coded = this.generateURL(this.trojan);
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
