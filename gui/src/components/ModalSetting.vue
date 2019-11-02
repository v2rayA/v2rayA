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
      <b-field label="PAC模式使用" label-position="on-border">
        <b-select v-model="pacMode" expanded style="flex-shrink: 0">
          <option value="whitelist">大陆白名单</option>
          <option value="gfwlist">GFWList</option>
          <option value="custom">自定义PAC</option>
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
            style="margin-left:0px;border-bottom-left-radius: 0;border-top-left-radius: 0;color:rgba(0,0,0,0.75)"
            outlined
            @click="handleClickConfigurePac"
            >配置</b-button
          >
        </template>
      </b-field>
      <b-field label="(待开发)每日自动更新PAC文件" label-position="on-border">
        <b-select v-model="pacAutoUpdateMode" expanded>
          <option value="none">关闭自动更新</option>
          <option value="auto_update">每日自动更新PAC文件</option>
        </b-select>
        <b-clockpicker
          v-if="pacAutoUpdateMode !== 'none'"
          v-model="pacAutoUpdateTime"
          placeholder="选择每日更新时间点"
          icon=" iconfont icon-clock2"
          hour-format="24"
          position="is-bottom-left"
          class="modal-setting-clockpicker"
        >
        </b-clockpicker>
      </b-field>
      <b-field label="(待开发)每日自动更新订阅" label-position="on-border">
        <b-select v-model="subscriptionAutoUpdateMode" expanded>
          <option value="none">关闭自动更新</option>
          <option value="auto_update">每日自动更新订阅</option>
        </b-select>
        <b-clockpicker
          v-if="subscriptionAutoUpdateMode !== 'none'"
          v-model="subscriptionAutoUpdateTime"
          placeholder="选择每日更新时间点"
          icon=" iconfont icon-clock2"
          hour-format="24"
          position="is-bottom-left"
          class="modal-setting-clockpicker"
        >
        </b-clockpicker>
      </b-field>
      <b-field label="解析订阅链接/更新时优先使用" label-position="on-border">
        <b-select v-model="proxyModeWhenSubscribe" expanded>
          <option value="direct">直连模式</option>
          <option value="pac">PAC模式</option>
          <option value="proxy">代理模式</option>
        </b-select>
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

export default {
  name: "ModalSetting",
  data: () => ({
    proxyModeWhenSubscribe: "direct",
    pacAutoUpdateMode: "none",
    pacAutoUpdateTime: null,
    subscriptionAutoUpdateMode: "none",
    subscriptionAutoUpdateTime: null,
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
        });
      });
    },
    handleClickSubmit() {
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
          pacMode: this.pacMode
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
</style>
