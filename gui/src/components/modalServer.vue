<template>
  <div class="modal-card" style="max-width: 520px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">节点配置</p>
    </header>
    <section :class="{ 'modal-card-body': true, readonly: readonly }">
      <b-tabs
        v-model="tabChoice"
        position="is-centered"
        class="block"
        type="is-boxed is-twitter same-width-5"
      >
        <b-tab-item label="VMESS">
          <b-field label="Name" label-position="on-border">
            <b-input
              ref="vmess_name"
              v-model="vmess.ps"
              placeholder="节点名称"
              expanded
            />
          </b-field>
          <b-field label="Address" label-position="on-border">
            <b-input
              ref="vmess_add"
              v-model="vmess.add"
              required
              placeholder="IP / HOST"
              expanded
            />
          </b-field>
          <b-field label="Port" label-position="on-border">
            <b-input
              ref="vmess_port"
              v-model="vmess.port"
              required
              placeholder="端口号"
              type="number"
              expanded
            />
          </b-field>
          <b-field label="ID" label-position="on-border">
            <b-input
              ref="vmess_id"
              v-model="vmess.id"
              required
              placeholder="UserID"
              expanded
            />
          </b-field>
          <b-field label="AID" label-position="on-border">
            <b-input
              ref="vmess_aid"
              v-model="vmess.aid"
              placeholder="AlterID"
              type="number"
              min="0"
              max="65535"
              expanded
            />
          </b-field>
          <b-field
            v-show="vmess.type !== 'dtls'"
            label="TLS"
            label-position="on-border"
          >
            <b-select v-model="vmess.tls" expanded @input="handleNetworkChange">
              <option value="none">关闭</option>
              <option value="tls">开启</option>
            </b-select>
          </b-field>
          <b-field label="Network" label-position="on-border">
            <b-select
              ref="vmess_net"
              v-model="vmess.net"
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
            v-show="vmess.net === 'tcp'"
            label="Type"
            label-position="on-border"
          >
            <b-select v-model="vmess.type" expanded>
              <option value="none">不伪装</option>
              <option value="http">伪装http</option>
            </b-select>
          </b-field>
          <b-field
            v-show="vmess.net === 'kcp'"
            label="Type"
            label-position="on-border"
          >
            <b-select v-model="vmess.type" expanded>
              <option value="none">不伪装</option>
              <option value="srtp">伪装视频通话(srtp)</option>
              <option value="utp">伪装BT下载(uTP)</option>
              <option value="wechat-video">伪装微信视频通话</option>
              <option value="dtls">伪装DTLS1.2数据包(将强制开启TLS)</option>
              <option value="wireguard">伪装WireGuard数据包</option>
            </b-select>
          </b-field>
          <b-field
            v-show="vmess.net === 'ws' || vmess.net === 'h2'"
            label="Host"
            label-position="on-border"
          >
            <b-input
              v-model="vmess.host"
              placeholder="伪装域名(host)"
              expanded
            />
          </b-field>
          <b-field
            v-show="vmess.net === 'ws' || vmess.net === 'h2'"
            label="Path"
            label-position="on-border"
          >
            <b-input v-model="vmess.path" placeholder="路径(path)" expanded />
          </b-field>
        </b-tab-item>
        <b-tab-item label="SS">
          <b-field label="Name" label-position="on-border">
            <b-input
              ref="ss_name"
              v-model="ss.name"
              placeholder="节点名称"
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
              placeholder="端口号"
              type="number"
              expanded
            />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input
              ref="ss_password"
              v-model="ss.password"
              required
              placeholder="密码"
              expanded
            />
          </b-field>
          <b-field label="Method" label-position="on-border">
            <b-select ref="ss_method" v-model="ss.method" expanded required>
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
            </b-select>
          </b-field>
        </b-tab-item>
        <b-tab-item label="SSR">
          <b-field label="Name" label-position="on-border">
            <b-input
              ref="ssr_name"
              v-model="ssr.name"
              placeholder="节点名称"
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
              placeholder="端口号"
              type="number"
              expanded
            />
          </b-field>
          <b-field label="Password" label-position="on-border">
            <b-input
              ref="ssr_password"
              v-model="ssr.password"
              required
              placeholder="密码"
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
            </b-select>
          </b-field>
          <b-field label="Protocol" label-position="on-border">
            <b-select ref="ssr_proto" v-model="ssr.proto" expanded required>
              <option value="origin">origin</option>
              <option value="verify_sha1">verify_sha1</option>
              <option value="auth_sha1_v4">auth_sha1_v4</option>
              <option value="auth_aes128_md5">auth_aes128_md5</option>
              <option value="auth_aes128_sha1">auth_aes128_sha1</option>
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
      </b-tabs>
    </section>
    <footer v-if="!readonly" class="modal-card-foot flex-end">
      <button class="button" type="button" @click="$parent.close()">
        取消
      </button>
      <button class="button is-primary" @click="handleClickSubmit">
        保存
      </button>
    </footer>
  </div>
</template>

<script>
import { handleResponse } from "@/assets/js/utils";
import { Base64 } from "js-base64";

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
    vmess: {
      ps: "",
      add: "",
      port: "",
      id: "",
      aid: "0",
      net: "tcp",
      type: "none",
      host: "",
      path: "",
      tls: "none",
      v: "",
      protocol: "vmess"
    },
    ss: {
      method: "aes-128-cfb",
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
    tabChoice: 0
  }),
  mounted() {
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
            this.vmess = this.resolveURL(res.data.data.sharingAddress);
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
          }
        });
      });
    }
  },
  methods: {
    resolveURL(url) {
      if (url.toLowerCase().indexOf("vmess://") >= 0) {
        let obj = JSON.parse(
          Base64.decode(url.substring(url.indexOf("://") + 3))
        );
        obj.ps = unescape(obj.ps);
        obj.tls = obj.tls || "none";
        obj.type = obj.type || "none";
        obj.protocol = "vmess";
        return obj;
      } else if (url.toLowerCase().indexOf("ss://") >= 0) {
        const regexp = /ss:\/\/(.+)@(.+):(.+)#(.*)/;
        let arr = regexp.exec(url);
        arr[1] = Base64.decode(arr[1]);
        let mp = arr[1].split(":");
        return {
          method: mp[0],
          password: mp[1],
          server: arr[2],
          port: arr[3],
          name: arr[4],
          protocol: "ss"
        };
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
      }
      return null;
    },
    generateURL(srcObj) {
      let obj = {};
      switch (srcObj.protocol) {
        case "vmess":
          //尽量减少生成的链接长度
          obj = Object.assign({}, srcObj);
          switch (obj.net) {
            case "kcp":
            case "tcp":
              obj.path = "";
              obj.host = "";
              break;
            default:
              obj.type = "";
          }
          return "vmess://" + Base64.encode(JSON.stringify(obj));
        case "ss":
          /* ss://BASE64(method:password)@server:port#name */
          return `ss://${Base64.encode(
            `${srcObj.method}:${srcObj.password}`
          )}@${srcObj.server}:${srcObj.port}#${srcObj.name}`;
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
      }
      return null;
    },
    handleNetworkChange() {
      this.vmess.type = "none";
    },
    handleClickSubmit() {
      let valid = true;
      for (let k in this.$refs) {
        if (!this.$refs.hasOwnProperty(k)) {
          continue;
        }
        if (this.tabChoice === 0 && !k.startsWith("vmess_")) {
          continue;
        }
        if (this.tabChoice === 1 && !k.startsWith("ss_")) {
          continue;
        }
        if (this.tabChoice === 2 && !k.startsWith("ssr_")) {
          continue;
        }
        let x = this.$refs[k];
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
        coded = this.generateURL(this.vmess);
      } else if (this.tabChoice === 1) {
        coded = this.generateURL(this.ss);
      } else if (this.tabChoice === 2) {
        coded = this.generateURL(this.ssr);
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
  width: 5em;
}
</style>
