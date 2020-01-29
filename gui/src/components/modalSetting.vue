<template>
  <div class="modal-card" style="max-width: 520px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">设置</p>
    </header>
    <section class="modal-card-body rules">
      <b-field
        label="GFWList"
        horizontal
        custom-class="modal-setting-label"
        style="position: relative"
        >最新：
        <a
          href="https://github.com/ToutyRater/V2Ray-SiteDAT/blob/master/geofiles/h2y.dat"
          target="_blank"
          class="is-link"
          >{{ remoteGFWListVersion }}</a
        >本地：
        <b-tooltip
          v-if="dayjs(localGFWListVersion).isAfter(dayjs(remoteGFWListVersion))"
          label="该时间是指本地文件最后修改时间，因此可能会领先最新版本"
          position="is-bottom"
          type="is-danger"
          dashed
          multilined
          animated
        >
          {{ localGFWListVersion ? localGFWListVersion : "无" }}
        </b-tooltip>
        <span v-else>{{
          localGFWListVersion ? localGFWListVersion : "无"
        }}</span>
        <b-button
          size="is-small"
          type="is-text"
          style="position: relative;top:-2px;text-decoration:none"
          @click="handleClickUpdateGFWList"
          >更新
        </b-button>
      </b-field>
      <b-field
        v-if="customPacFileVersion"
        label="自定义规则"
        horizontal
        custom-class="modal-setting-label"
        >最后更新时间： <span>{{ customPacFileVersion }}</span>
        <b-button
          size="is-small"
          type="is-text"
          style="position: relative;top:-2px;text-decoration:none"
          >更新
        </b-button>
      </b-field>
      <hr class="dropdown-divider" style="margin: 1.25rem 0 1.25rem" />
      <b-field
        v-show="transparentValid"
        label-position="on-border"
        class="with-icon-alert"
      >
        <template slot="label">
          全局透明代理
          <b-tooltip
            type="is-dark"
            label="全局代理开启后，任何TCP、UDP流量均会经过V2Ray，此时PAC端口的配置将被覆盖。另外，如需作为网关使得连接本机的其他主机也享受代理，请勾选“开启IP转发”。注：本机docker不会走代理。"
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
        <b-select v-model="transparent" expanded>
          <option value="close">关闭</option>
          <option value="proxy">代理模式</option>
          <option value="whitelist">大陆白名单(Recommend)</option>
          <option value="gfwlist">GFWList</option>
        </b-select>
        <template v-if="transparent !== 'close'">
          <b-button
            style="border-radius: 0;z-index: 2;"
            @click="handleClickPortWhiteList"
          >
            端口白名单
          </b-button>
          <b-checkbox-button
            v-model="ipforward"
            :native-value="true"
            style="position:relative;left:-1px;"
            >开启IP转发
          </b-checkbox-button>
        </template>
      </b-field>
      <b-field label="PAC模式" label-position="on-border">
        <b-select v-model="pacMode" expanded style="flex-shrink: 0">
          <option value="whitelist">大陆白名单(Recommend)</option>
          <option value="gfwlist">GFWList</option>
          <option value="custom">自定义PAC（待开发）</option>
        </b-select>
        <template v-if="pacMode === 'custom'">
          <b-input
            v-model="customPac.url"
            placeholder="SiteDAT file URL"
            custom-class="no-shadow"
          />
          <b-button
            v-if="pacMode === 'custom'"
            type="is-primary"
            style="margin-left:0;border-bottom-left-radius: 0;border-top-left-radius: 0;color:rgba(0,0,0,0.75)"
            outlined
            @click="handleClickConfigurePac"
            >配置
          </b-button>
        </template>
      </b-field>
      <b-field
        v-show="
          (transparent === 'close' && pacMode === 'gfwlist') ||
            transparent === 'gfwlist'
        "
        label="自动更新GFWList"
        label-position="on-border"
      >
        <b-select v-model="pacAutoUpdateMode" expanded>
          <option value="none">不自动更新PAC文件</option>
          <option value="auto_update">服务端启动时更新PAC文件</option>
        </b-select>
      </b-field>
      <b-field label="自动更新订阅" label-position="on-border">
        <b-select v-model="subscriptionAutoUpdateMode" expanded>
          <option value="none">不自动更新订阅</option>
          <option value="auto_update">服务端启动时更新订阅</option>
        </b-select>
      </b-field>
      <b-field
        v-if="transparent === 'close'"
        label="解析订阅链接/更新时优先使用"
        label-position="on-border"
      >
        <b-select v-model="proxyModeWhenSubscribe" expanded>
          <option value="direct">直连模式</option>
          <option value="pac">PAC模式</option>
          <option value="proxy">代理模式</option>
        </b-select>
      </b-field>
      <b-field
        v-show="showDnsForward"
        label="转发DNS查询"
        label-position="on-border"
      >
        <template slot="label">
          转发DNS查询
          <b-tooltip
            type="is-dark"
            label="转发DNS查询可以有效规避DNS污染，但有可能会降低网页打开速度，请视情况开启。"
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
        <b-select v-model="dnsforward" expanded>
          <option value="no">关闭</option>
          <option value="yes">启用</option>
        </b-select>
      </b-field>
      <b-field label-position="on-border">
        <template slot="label">
          TCPFastOpen
          <b-tooltip
            type="is-dark"
            label="简化TCP握手流程以加速建立连接，可能会增加封包的特征。当前仅支持vmess节点。"
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
        <b-select v-model="tcpFastOpen" expanded>
          <option value="default">保持系统默认</option>
          <option value="yes">启用</option>
          <option value="no">禁用</option>
        </b-select>
      </b-field>
      <b-field label-position="on-border" class="with-icon-alert">
        <template slot="label">
          多路复用
          <b-tooltip
            type="is-dark"
            label="复用TCP连接以减少握手延迟，但会影响吞吐量大的使用场景，如观看视频、下载、测速。当前仅支持vmess节点。"
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
        <b-select v-model="muxOn" expanded style="flex: 1">
          <option value="no">关闭</option>
          <option value="yes">启用</option>
        </b-select>
        <cus-b-input
          v-if="muxOn === 'yes'"
          ref="muxinput"
          v-model="mux"
          placeholder="最大并发连接数"
          custom-class="no-shadow"
          type="number"
          min="1"
          max="1024"
          validation-icon=" iconfont icon-alert"
          style="flex: 1"
        />
      </b-field>
      <!--      <b-field label="SERVER列表" label-position="on-border">-->
      <!--        <b-select v-model="serverListMode" expanded>-->
      <!--          <option value="noSubscription">仅显示非订阅节点</option>-->
      <!--          <option value="all">显示全部节点</option>-->
      <!--        </b-select>-->
      <!--      </b-field>-->
    </section>
    <footer class="modal-card-foot flex-end">
      <button
        class="button footer-absolute-left"
        type="button"
        @click="$emit('clickPorts')"
      >
        地址与端口
      </button>
      <button class="button" type="button" @click="$parent.close()">
        取消
      </button>
      <button class="button is-primary" @click="handleClickSubmit">
        保存设置
      </button>
    </footer>
  </div>
</template>

<script>
import { handleResponse, isIntranet } from "@/assets/js/utils";
import dayjs from "dayjs";
import ModalConfigurePac from "@/components/modalConfigurePac";
import CusBInput from "./input/Input.vue";
import { isVersionGreaterEqual, parseURL } from "../assets/js/utils";
import BButton from "buefy/src/components/button/Button";
import BSelect from "buefy/src/components/select/Select";
import BCheckboxButton from "buefy/src/components/checkbox/CheckboxButton";
import modalPortWhiteList from "@/components/modalPortWhiteList";

export default {
  name: "ModalSetting",
  components: { BCheckboxButton, BSelect, BButton, CusBInput },
  data: () => ({
    proxyModeWhenSubscribe: "direct",
    tcpFastOpen: "default",
    muxOn: "no",
    mux: "8",
    transparent: "close",
    ipforward: false,
    dnsforward: "no",
    pacAutoUpdateMode: "none",
    subscriptionAutoUpdateMode: "none",
    customSiteDAT: {},
    pacMode: "whitelist",
    customPac: {
      url: "",
      defaultProxyMode: "direct",
      routingRules: []
    },
    showClockPicker: true,
    serverListMode: "noSubscription",
    remoteGFWListVersion: "checking...",
    localGFWListVersion: "checking...",
    customPacFileVersion: "checking...",
    showDnsForward: false
  }),
  computed: {
    dockerMode() {
      return window.localStorage["docker"] === "true";
    },
    v2rayaPort() {
      let U = parseURL(apiRoot);
      let port = U.port;
      if (!port) {
        port =
          U.protocol === "http" ? "80" : U.protocol === "https" ? "443" : "";
      }
      return port;
    },
    transparentValid() {
      let val = localStorage["transparentValid"];
      return (
        val === "undefined" || //最早版本, 无法判断是否valid, 就默认valid了
        val === "true" || //boolean版本
        val === "yes" //最新string版本
      );
    }
  },
  created() {
    this.$axios({
      url: apiRoot + "/remoteGFWListVersion"
    }).then(res => {
      handleResponse(res, this, () => {
        this.remoteGFWListVersion = res.data.data.remoteGFWListVersion;
      });
    });
    this.$axios({
      url: apiRoot + "/setting"
    }).then(res => {
      handleResponse(res, this, () => {
        Object.assign(this, res.data.data.setting);
        delete res.data.data["setting"];
        Object.assign(this, res.data.data);
        this.subscriptionAutoUpdateTime = new Date(
          this.subscriptionAutoUpdateTime
        );
        this.pacAutoUpdateTime = new Date(this.pacAutoUpdateTime);
        this.showDnsForward = isVersionGreaterEqual(
          localStorage["version"],
          "0.6.1"
        );
      });
    });
    //白名单有没有项，没有就post一下
    this.$axios({
      url: apiRoot + "/portWhiteList"
    }).then(res => {
      handleResponse(res, this, () => {
        if (res.data.data.tcp === null && res.data.data.udp === null) {
          this.$axios({
            url: apiRoot + "/portWhiteList",
            method: "post",
            data: {
              requestPort: this.v2rayaPort
            }
          });
        }
      });
    });
  },
  methods: {
    dayjs() {
      return dayjs.apply(this, arguments);
    },
    handleClickUpdateGFWList() {
      this.$axios({
        url: apiRoot + "/gfwList",
        method: "put"
      }).then(res => {
        handleResponse(res, this, () => {
          this.localGFWListVersion = res.data.data.localGFWListVersion;
          this.$buefy.toast.open({
            message: "更新成功",
            type: "is-warning",
            position: "is-top",
            duration: 5000
          });
        });
      });
    },
    requestUpdateSetting() {
      this.$axios({
        url: apiRoot + "/setting",
        method: "put",
        data: {
          proxyModeWhenSubscribe: this.proxyModeWhenSubscribe,
          pacAutoUpdateMode: this.pacAutoUpdateMode,
          pacAutoUpdateTime: this.pacAutoUpdateTime
            ? this.pacAutoUpdateTime.getTime()
            : 0,
          subscriptionAutoUpdateMode: this.subscriptionAutoUpdateMode,
          subscriptionAutoUpdateTime: this.subscriptionAutoUpdateTime
            ? this.subscriptionAutoUpdateTime.getTime()
            : 0,
          customPac: this.customPac,
          pacMode: this.pacMode,
          tcpFastOpen: this.tcpFastOpen,
          muxOn: this.muxOn,
          mux: parseInt(this.mux),
          transparent: this.transparent,
          ipforward: this.ipforward,
          dnsforward: this.dnsforward
        }
      }).then(res => {
        handleResponse(res, this, () => {
          this.$buefy.toast.open({
            message: res.data.code,
            type: "is-primary",
            position: "is-top"
          });
          this.$parent.close();
        });
      });
    },
    handleClickSubmit() {
      if (this.muxOn === "yes" && !this.$refs.muxinput.checkHtml5Validity()) {
        return;
      }
      if (this.pacMode === "custom" && this.customPac.url.length <= 0) {
        this.$buefy.toast.open({
          message: "自定义PAC模式下，SiteDAT file URL不能为空",
          type: "is-warning",
          position: "is-top",
          duration: 3000
        });
        return;
      } else if (
        this.pacMode === "custom" &&
        this.customPac.routingRules.length <= 0
      ) {
        this.$buefy.toast.open({
          message: "您还没有配置PAC路由呢",
          type: "is-warning",
          position: "is-top",
          duration: 3000
        });
        return;
      }
      if (this.transparent !== "close" && !isIntranet(apiRoot)) {
        let U = parseURL(apiRoot);
        let port = U.port;
        if (!port) {
          port =
            U.protocol === "http" ? "80" : U.protocol === "https" ? "443" : "";
        }
        this.$axios({
          url: apiRoot + "/portWhiteList"
        })
          .then(res => {
            handleResponse(res, this, () => {
              this.$buefy.dialog.confirm({
                title: "提示",
                message: `<div class=""><p>您正在对不同子网下的机器设置透明代理，请确认不走代理的出方向端口。</p>
              <p>当前设置的端口白名单为：</p>
              <p>TCP: ${res.data.data.tcp.join(", ")}</p>
              <p>UDP: ${res.data.data.udp.join(", ")}</p>`,
                cancelText: "取消",
                confirmText: "确认无误",
                type: "is-danger",
                onConfirm: () => this.requestUpdateSetting()
              });
            });
          })
          .catch(() => {
            //可能是服务端是老旧版本，没这个接口
            this.requestUpdateSetting();
          });
      } else {
        this.requestUpdateSetting();
      }
    },
    handleClickConfigurePac() {
      const that = this;
      this.$buefy.modal.open({
        parent: this,
        component: ModalConfigurePac,
        hasModalCard: true,
        canCancel: false,
        props: {
          customPac: this.customPac
        },
        events: {
          submit(val) {
            that.customPac = val;
          }
        }
      });
    },
    handleClickPortWhiteList() {
      this.$buefy.modal.open({
        parent: this,
        component: modalPortWhiteList,
        hasModalCard: true,
        canCancel: true
      });
    }
  }
};
</script>

<style lang="scss">
.rules {
  height: 390px;
  overflow-x: hidden;
}

.flex-end {
  justify-content: flex-end !important;
}

.modal-setting-label {
  width: 7em;
  padding: 0 !important;
  text-align: left !important;
}

.modal-setting-clockpicker {
  .background {
    display: unset;
    background-color: rgba(10, 10, 10, 0.6) !important;
  }

  .dropdown-menu {
    position: fixed;
    top: 50% !important;
    left: 50% !important;
    transform: translate3d(-50%, -50%, 0);
    right: unset !important;
    z-index: 50 !important;
  }
}

//让"更新"按钮右对齐
.rules .field.is-horizontal .field-body .field:last-child {
  text-align: right;
}

.no-shadow {
  box-shadow: none !important;
}

.with-icon-alert {
  p.help {
    position: absolute;
    bottom: -18px;
    right: 0;
  }

  .icon-alert {
    font-size: 18px;
  }
}

.control:first-of-type:not(:last-of-type) .select select {
  border-radius: 4px 0 0 4px !important;
}

.footer-absolute-left {
  position: absolute;
  left: 20px;
}
</style>
