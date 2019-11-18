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
          {{ localGFWListVersion ? localGFWListVersion : "无" }} </b-tooltip
        ><span v-else>{{
          localGFWListVersion ? localGFWListVersion : "无"
        }}</span
        ><b-button
          size="is-small"
          type="is-text"
          style="position: relative;top:-2px;text-decoration:none"
          @click="handleClickUpdateGFWList"
          >更新</b-button
        ></b-field
      >
      <b-field
        v-if="customPacFileVersion"
        label="自定义规则"
        horizontal
        custom-class="modal-setting-label"
        >最后更新时间： <span>{{ customPacFileVersion }}</span
        ><b-button
          size="is-small"
          type="is-text"
          style="position: relative;top:-2px;text-decoration:none"
          >更新</b-button
        ></b-field
      >
      <hr class="dropdown-divider" style="margin: 1.25rem 0 1.25rem" />
      <b-field
        v-if="!dockerMode"
        label-position="on-border"
        class="with-icon-alert"
      >
        <template slot="label">
          透明全局代理
          <b-tooltip
            type="is-dark"
            label="全局代理开启后，任何TCP、UDP流量均会经过V2Ray，此时PAC端口的配置将被覆盖。另外，如需作为网关使得连接本机的其他主机也享受代理，请勾选“开启IP转发”。"
            multilined
            position="is-right"
          >
            <b-icon
              size="is-small"
              icon=" iconfont icon-help-circle-outline"
              style="position:relative;top:2px;right:3px;font-weight:normal"
            ></b-icon>
          </b-tooltip>
        </template>
        <b-select v-model="transparent" expanded>
          <option value="close">关闭</option>
          <option value="proxy">代理模式</option>
          <option value="whitelist">大陆白名单</option>
          <option value="gfwlist">GFWList</option>
        </b-select>
        <b-checkbox-button
          v-if="transparent !== 'close'"
          v-model="ipforward"
          :native-value="true"
          >开启IP转发</b-checkbox-button
        >
      </b-field>
      <b-field
        v-show="transparent === 'close'"
        label="PAC模式"
        label-position="on-border"
      >
        <b-select v-model="pacMode" expanded style="flex-shrink: 0">
          <option value="whitelist">大陆白名单</option>
          <option value="gfwlist">GFWList</option>
          <option value="custom">自定义PAC（待开发）</option>
        </b-select>
        <template v-if="pacMode === 'custom'">
          <b-input
            v-model="customPac.url"
            placeholder="SiteDAT file URL"
            custom-class="no-shadow"
          ></b-input>
          <b-button
            v-if="pacMode === 'custom'"
            type="is-primary"
            style="margin-left:0;border-bottom-left-radius: 0;border-top-left-radius: 0;color:rgba(0,0,0,0.75)"
            outlined
            @click="handleClickConfigurePac"
            >配置</b-button
          >
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
      <b-field label="解析订阅链接/更新时优先使用" label-position="on-border">
        <b-select v-model="proxyModeWhenSubscribe" expanded>
          <option value="direct">直连模式</option>
          <option value="pac">PAC模式</option>
          <option value="proxy">代理模式</option>
        </b-select>
      </b-field>
      <b-field label-position="on-border">
        <template slot="label">
          TCPFastOpen
          <b-tooltip
            type="is-dark"
            label="简化TCP握手流程以加速建立连接，可能会增加封包的特征。"
            multilined
            position="is-right"
          >
            <b-icon
              size="is-small"
              icon=" iconfont icon-help-circle-outline"
              style="position:relative;top:2px;right:3px;font-weight:normal"
            ></b-icon>
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
            label="复用TCP连接以减少握手延迟，但会影响吞吐量大的使用场景，如观看视频、下载、测速。"
            multilined
            position="is-right"
          >
            <b-icon
              size="is-small"
              icon=" iconfont icon-help-circle-outline"
              style="position:relative;top:2px;right:3px;font-weight:normal"
            ></b-icon>
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
        ></cus-b-input>
      </b-field>
      <!--      <b-field label="SERVER列表" label-position="on-border">-->
      <!--        <b-select v-model="serverListMode" expanded>-->
      <!--          <option value="noSubscription">仅显示非订阅节点</option>-->
      <!--          <option value="all">显示全部节点</option>-->
      <!--        </b-select>-->
      <!--      </b-field>-->
    </section>
    <footer class="modal-card-foot flex-end">
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
import { handleResponse } from "@/assets/js/utils";
import dayjs from "dayjs";
import ModalConfigurePac from "@/components/ModalConfigurePac";
import CusBInput from "./input/Input.vue";

export default {
  name: "ModalSetting",
  components: { CusBInput },
  data: () => ({
    proxyModeWhenSubscribe: "direct",
    tcpFastOpen: "default",
    muxOn: "no",
    mux: "8",
    transparent: "close",
    ipforward: false,
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
    customPacFileVersion: "checking..."
  }),
  computed: {
    dockerMode() {
      return window.localStorage["docker"] === "true";
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
          ipforward: this.ipforward
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
</style>
