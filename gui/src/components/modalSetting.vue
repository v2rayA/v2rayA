<template>
  <div class="modal-card modal-setting" style="max-width: 800px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">{{ $t("common.setting") }}</p>
    </header>
    <section class="modal-card-body rules">
      <b-field
        label="GFWList"
        horizontal
        custom-class="modal-setting-label"
        style="position: relative"
        ><span>{{ $t("common.latest") }}:</span>
        <a
          href="https://github.com/v2rayA/dist-v2ray-rules-dat/releases"
          target="_blank"
          class="is-link"
          >{{ remoteGFWListVersion }}</a
        ><span>{{ $t("common.local") }}:</span>
        <b-tooltip
          v-if="dayjs(localGFWListVersion).isAfter(dayjs(remoteGFWListVersion))"
          :label="$t('setting.messages.gfwlist')"
          position="is-bottom"
          type="is-danger"
          dashed
          multilined
          animated
        >
          {{ localGFWListVersion ? localGFWListVersion : $t("none") }}
        </b-tooltip>
        <span v-else>{{
          localGFWListVersion ? localGFWListVersion : $t("none")
        }}</span>
        <b-button
          size="is-small"
          type="is-text"
          style="
            position: relative;
            top: -2px;
            text-decoration: none;
            font-weight: bold;
          "
          @click="handleClickUpdateGFWList"
          >{{ $t("operations.update") }}
        </b-button>
      </b-field>
      <hr class="dropdown-divider" style="margin: 1.25rem 0 1.25rem" />
      <b-field label-position="on-border" class="with-icon-alert">
        <template slot="label">
          {{ $t("setting.transparentProxy") }}
          <b-tooltip
            type="is-dark"
            :label="$t('setting.messages.transparentProxy')"
            multilined
            position="is-right"
          >
            <b-icon
              size="is-small"
              icon=" iconfont icon-help-circle-outline"
              style="
                position: relative;
                top: 2px;
                right: 3px;
                font-weight: normal;
              "
            />
          </b-tooltip>
        </template>
        <b-select v-model="transparent" expanded>
          <option value="close">{{ $t("setting.options.off") }}</option>
          <option value="proxy">
            {{ $t("setting.options.on") }}: {{ $t("setting.options.global") }}
          </option>
          <option value="whitelist">
            {{ $t("setting.options.on") }}:
            {{ $t("setting.options.whitelistCn") }}
          </option>
          <option value="gfwlist">
            {{ $t("setting.options.on") }}: {{ $t("setting.options.gfwlist") }}
          </option>
          <option value="pac">
            {{ $t("setting.options.on") }}:
            {{ $t("setting.options.sameAsPacMode") }}
          </option>
        </b-select>
        <b-checkbox-button
          v-show="!lite"
          v-model="ipforward"
          :native-value="true"
          style="position: relative; left: -1px"
          >{{ $t("setting.ipForwardOn") }}
        </b-checkbox-button>
        <b-checkbox-button
          v-model="portSharing"
          :native-value="true"
          style="position: relative; left: -1px"
          >{{ $t("setting.portSharingOn") }}
        </b-checkbox-button>
      </b-field>

      <b-field v-show="transparent !== 'close'" label-position="on-border">
        <template slot="label">
          {{ $t("setting.transparentType") }}
          <b-tooltip
            type="is-dark"
            multilined
            :label="$t('setting.messages.transparentType')"
            position="is-right"
          >
            <b-icon
              size="is-small"
              icon=" iconfont icon-help-circle-outline"
              style="
                position: relative;
                top: 2px;
                right: 3px;
                font-weight: normal;
              "
            />
          </b-tooltip>
        </template>
        <b-select v-model="transparentType" expanded class="left-border">
          <option v-show="!lite" value="redirect">redirect</option>
          <option v-show="!lite" value="tproxy">tproxy</option>
          <option value="system_proxy">system proxy</option>
        </b-select>
      </b-field>
      <b-field label-position="on-border">
        <template slot="label">
          {{ $t("setting.pacMode") }}
          <b-tooltip
            type="is-dark"
            :label="$t('setting.messages.pacMode')"
            multilined
            position="is-right"
          >
            <b-icon
              size="is-small"
              icon=" iconfont icon-help-circle-outline"
              style="
                position: relative;
                top: 2px;
                right: 3px;
                font-weight: normal;
              "
            />
          </b-tooltip>
        </template>
        <b-select v-model="pacMode" expanded style="flex-shrink: 0">
          <option value="whitelist">
            {{ $t("setting.options.whitelistCn") }}
          </option>
          <option value="gfwlist">{{ $t("setting.options.gfwlist") }}</option>
          <!--          <option value="custom">{{-->
          <!--            $t("setting.options.customRouting")-->
          <!--          }}</option>-->
          <option value="routingA">RoutingA</option>
        </b-select>
        <template v-if="pacMode === 'custom'">
          <b-button
            type="is-primary"
            style="
              margin-left: 0;
              border-bottom-left-radius: 0;
              border-top-left-radius: 0;
              color: rgba(0, 0, 0, 0.75);
            "
            outlined
            @click="handleClickConfigurePac"
            >{{ $t("operations.configure") }}
          </b-button>
        </template>
        <template v-if="pacMode === 'routingA'">
          <b-button
            style="
              margin-left: 0;
              border-bottom-left-radius: 0;
              border-top-left-radius: 0;
              color: rgba(0, 0, 0, 0.75);
            "
            outlined
            @click="handleClickConfigureRoutingA"
            >{{ $t("operations.configure") }}
          </b-button>
        </template>
        <p></p>
      </b-field>
      <b-field label-position="on-border">
        <template slot="label">
          {{ $t("setting.preventDnsSpoofing") }}
          <b-tooltip
            type="is-dark"
            :label="$t('setting.messages.preventDnsSpoofing')"
            multilined
            position="is-right"
          >
            <b-icon
              size="is-small"
              icon=" iconfont icon-help-circle-outline"
              style="
                position: relative;
                top: 2px;
                right: 3px;
                font-weight: normal;
              "
            />
          </b-tooltip>
        </template>
        <b-select v-model="antipollution" expanded class="left-border">
          <option value="closed">{{ $t("setting.options.closed") }}</option>
          <option value="none">
            {{ $t("setting.options.antiDnsHijack") }}
          </option>
          <option value="dnsforward">
            {{ $t("setting.options.forwardDnsRequest") }}
          </option>
          <option value="doh">{{ $t("setting.options.doh") }}</option>
          <option value="advanced">{{ $t("setting.options.advanced") }}</option>
        </b-select>
        <b-button
          v-if="antipollution === 'advanced'"
          :class="{
            'right-extra-button': antipollution === 'closed',
            'no-border-radius': antipollution !== 'closed',
          }"
          @click="handleClickDnsSetting"
        >
          {{ $t("operations.configure") }}
        </b-button>
        <p></p>
      </b-field>
      <b-field v-show="showSpecialMode" label-position="on-border">
        <template slot="label">
          {{ $t("setting.specialMode") }}
          <b-tooltip
            type="is-dark"
            multilined
            :label="$t('setting.messages.specialMode')"
            position="is-right"
          >
            <b-icon
              size="is-small"
              icon=" iconfont icon-help-circle-outline"
              style="
                position: relative;
                top: 2px;
                right: 3px;
                font-weight: normal;
              "
            />
          </b-tooltip>
        </template>
        <b-select v-model="specialMode" expanded class="left-border">
          <option value="none">{{ $t("setting.options.closed") }}</option>
          <option value="supervisor">supervisor</option>
          <option v-show="antipollution !== 'closed'" value="fakedns">
            fakedns
          </option>
        </b-select>
      </b-field>
      <b-field label-position="on-border">
        <template slot="label">
          TCPFastOpen
          <b-tooltip
            type="is-dark"
            :label="$t('setting.messages.tcpFastOpen')"
            multilined
            position="is-right"
          >
            <b-icon
              size="is-small"
              icon=" iconfont icon-help-circle-outline"
              style="
                position: relative;
                top: 2px;
                right: 3px;
                font-weight: normal;
              "
            />
          </b-tooltip>
        </template>
        <b-select v-model="tcpFastOpen" expanded>
          <option value="default">{{ $t("setting.options.default") }}</option>
          <option value="yes">{{ $t("setting.options.on") }}</option>
          <option value="no">{{ $t("setting.options.off") }}</option>
        </b-select>
      </b-field>
      <b-field label-position="on-border" class="with-icon-alert">
        <template slot="label">
          {{ $t("setting.mux") }}
          <b-tooltip
            type="is-dark"
            :label="$t('setting.messages.mux')"
            multilined
            position="is-right"
          >
            <b-icon
              size="is-small"
              icon=" iconfont icon-help-circle-outline"
              style="
                position: relative;
                top: 2px;
                right: 3px;
                font-weight: normal;
              "
            />
          </b-tooltip>
        </template>
        <b-select v-model="muxOn" expanded style="flex: 1">
          <option value="no">{{ $t("setting.options.off") }}</option>
          <option value="yes">{{ $t("setting.options.on") }}</option>
        </b-select>
        <cus-b-input
          v-if="muxOn === 'yes'"
          ref="muxinput"
          v-model="mux"
          :placeholder="$t('setting.concurrency')"
          custom-class="no-shadow"
          type="number"
          min="1"
          max="1024"
          validation-icon=" iconfont icon-alert"
          style="flex: 1"
        />
      </b-field>
      <b-field
        v-show="pacMode === 'gfwlist' || transparent === 'gfwlist'"
        :label="$t('setting.autoUpdateGfwlist')"
        label-position="on-border"
      >
        <b-select v-model="pacAutoUpdateMode" expanded>
          <option value="none">{{ $t("setting.options.off") }}</option>
          <option value="auto_update">
            {{ $t("setting.options.updateGfwlistWhenStart") }}
          </option>
          <option value="auto_update_at_intervals">
            {{ $t("setting.options.updateGfwlistAtIntervals") }}
          </option>
        </b-select>
        <cus-b-input
          v-if="pacAutoUpdateMode === 'auto_update_at_intervals'"
          ref="autoUpdatePacInput"
          v-model="pacAutoUpdateIntervalHour"
          custom-class="no-shadow"
          type="number"
          min="1"
          validation-icon=" iconfont icon-alert"
          style="flex: 1"
        />
      </b-field>
      <b-field :label="$t('setting.autoUpdateSub')" label-position="on-border">
        <b-select v-model="subscriptionAutoUpdateMode" expanded>
          <option value="none">{{ $t("setting.options.off") }}</option>
          <option value="auto_update">
            {{ $t("setting.options.updateSubWhenStart") }}
          </option>
          <option value="auto_update_at_intervals">
            {{ $t("setting.options.updateSubAtIntervals") }}
          </option>
        </b-select>
        <cus-b-input
          v-if="subscriptionAutoUpdateMode === 'auto_update_at_intervals'"
          ref="autoUpdateSubInput"
          v-model="subscriptionAutoUpdateIntervalHour"
          custom-class="no-shadow"
          type="number"
          min="1"
          validation-icon=" iconfont icon-alert"
          style="flex: 1"
        />
      </b-field>
      <b-field
        :label="$t('setting.preferModeWhenUpdate')"
        label-position="on-border"
      >
        <b-select v-model="proxyModeWhenSubscribe" expanded>
          <option value="direct">
            {{
              transparent === "close" || lite
                ? $t("setting.options.direct")
                : $t("setting.options.dependTransparentMode")
            }}
          </option>
          <option value="proxy">{{ $t("setting.options.global") }}</option>
          <option value="pac">{{ $t("setting.options.pac") }}</option>
        </b-select>
      </b-field>
    </section>
    <footer class="modal-card-foot flex-end">
      <button
        class="button footer-absolute-left"
        type="button"
        @click="$emit('clickPorts')"
      >
        {{ $t("customAddressPort.title") }}
      </button>
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
import dayjs from "dayjs";
import ModalCustomRouting from "@/components/modalCustomRouting";
import ModalCustomRoutingA from "@/components/modalCustomRoutingA";
import CusBInput from "./input/Input.vue";
import { parseURL, toInt } from "@/assets/js/utils";
import BButton from "buefy/src/components/button/Button";
import BSelect from "buefy/src/components/select/Select";
import BCheckboxButton from "buefy/src/components/checkbox/CheckboxButton";
import modalDnsSetting from "./modalDnsSetting";
import axios from "../plugins/axios";
import { waitingConnected } from "@/assets/js/networkInspect";

export default {
  name: "ModalSetting",
  components: { BCheckboxButton, BSelect, BButton, CusBInput },
  data: () => ({
    proxyModeWhenSubscribe: "direct",
    tcpFastOpen: "default",
    muxOn: "no",
    mux: "8",
    transparent: "close",
    transparentType: "redirect",
    ipforward: false,
    portSharing: false,
    dnsForceMode: false,
    dnsforward: "no",
    antipollution: "none",
    specialMode: "none",
    pacAutoUpdateMode: "none",
    pacAutoUpdateIntervalHour: 0,
    subscriptionAutoUpdateMode: "none",
    subscriptionAutoUpdateIntervalHour: 0,
    customSiteDAT: {},
    pacMode: "whitelist",
    showClockPicker: true,
    serverListMode: "noSubscription",
    remoteGFWListVersion: "checking...",
    localGFWListVersion: "checking...",
    showSpecialMode: true,
  }),
  computed: {
    lite() {
      return (
        window.localStorage["lite"] && parseInt(window.localStorage["lite"]) > 0
      );
    },
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
      return toInt(port);
    },
  },
  watch: {
    antipollution(val) {
      if (val === "closed" && this.specialMode === "fakedns") {
        this.specialMode = "none";
      }
    },
  },
  created() {
    this.$axios({
      url: apiRoot + "/remoteGFWListVersion",
    }).then((res) => {
      handleResponse(res, this, () => {
        this.remoteGFWListVersion = res.data.data.remoteGFWListVersion;
      });
    });
    this.$axios({
      url: apiRoot + "/setting",
    }).then((res) => {
      handleResponse(res, this, () => {
        Object.assign(this, res.data.data.setting);
        delete res.data.data["setting"];
        Object.assign(this, res.data.data);
        this.subscriptionAutoUpdateTime = new Date(
          this.subscriptionAutoUpdateTime
        );
        this.pacAutoUpdateTime = new Date(this.pacAutoUpdateTime);
        if (this.lite) {
          this.transparentType = "system_proxy";
          this.showSpecialMode = false;
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
        method: "put",
        timeout: 0,
      }).then((res) => {
        handleResponse(res, this, () => {
          this.localGFWListVersion = res.data.data.localGFWListVersion;
          this.$buefy.toast.open({
            message: this.$t("common.success"),
            type: "is-warning",
            position: "is-top",
            duration: 5000,
            queue: false,
          });
        });
      });
    },
    requestUpdateSetting() {
      let cancel;
      waitingConnected(
        this.$axios({
          url: apiRoot + "/setting",
          method: "put",
          data: {
            proxyModeWhenSubscribe: this.proxyModeWhenSubscribe,
            pacAutoUpdateMode: this.pacAutoUpdateMode,
            pacAutoUpdateIntervalHour: parseInt(this.pacAutoUpdateIntervalHour),
            subscriptionAutoUpdateMode: this.subscriptionAutoUpdateMode,
            subscriptionAutoUpdateIntervalHour: parseInt(
              this.subscriptionAutoUpdateIntervalHour
            ),
            pacMode: this.pacMode,
            tcpFastOpen: this.tcpFastOpen,
            muxOn: this.muxOn,
            mux: parseInt(this.mux),
            transparent: this.transparent,
            transparentType: this.transparentType,
            ipforward: this.ipforward,
            portSharing: this.portSharing,
            dnsforward: this.antipollution === "dnsforward" ? "yes" : "no", //版本兼容
            antipollution: this.antipollution,
            specialMode: this.specialMode,
          },
          cancelToken: new axios.CancelToken(function executor(c) {
            cancel = c;
          }),
        }).then((res) => {
          handleResponse(res, this, () => {
            this.$buefy.toast.open({
              message: res.data.code,
              type: "is-primary",
              position: "is-top",
              queue: false,
            });
            this.$parent.close();
          });
          if (
            res.data.code !== "SUCCESS" &&
            res.data.message.indexOf("invalid config") >= 0
          ) {
            // FIXME: tricky
            this.$parent.$parent.runningState.running =
              this.$t("common.notRunning");
          }
        }),
        3 * 1000,
        cancel
      );
    },
    handleClickSubmit() {
      if (this.muxOn === "yes" && !this.$refs.muxinput.checkHtml5Validity()) {
        return;
      }
      if (
        this.subscriptionAutoUpdateMode === "auto_update_at_intervals" &&
        !this.$refs.autoUpdateSubInput.checkHtml5Validity()
      ) {
        return;
      }
      if (
        this.pacAutoUpdateMode === "auto_update_at_intervals" &&
        !this.$refs.autoUpdatePacInput.checkHtml5Validity()
      ) {
        return;
      }
      console.log(apiRoot);
      this.requestUpdateSetting();
    },
    handleClickConfigurePac() {
      this.$buefy.modal.open({
        parent: this,
        component: ModalCustomRouting,
        hasModalCard: true,
        canCancel: true,
      });
    },
    handleClickConfigureRoutingA() {
      this.$buefy.modal.open({
        parent: this,
        component: ModalCustomRoutingA,
        hasModalCard: true,
        canCancel: true,
      });
    },
    handleClickDnsSetting() {
      this.$buefy.modal.open({
        parent: this,
        component: modalDnsSetting,
        hasModalCard: true,
        canCancel: true,
      });
    },
  },
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
.left-border select {
  border-radius: 4px 0 0 4px !important;
}
.right-extra-button {
  border-radius: 0 4px 4px 0;
}
.no-border-radius {
  border-radius: 0;
}
.modal-setting {
  .b-checkbox.checkbox {
    margin-right: 0;
  }
}
</style>
