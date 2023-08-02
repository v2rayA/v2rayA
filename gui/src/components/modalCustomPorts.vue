<template>
  <div class="modal-card" style="max-width: 450px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $t("customAddressPort.title") }}
      </p>
    </header>
    <section class="modal-card-body">
      <b-field
        :label="$t('customAddressPort.serviceAddress')"
        label-position="on-border"
      >
        <b-input
          ref="backendAddress"
          v-model="table.backendAddress"
          placeholder="http://localhost:2017"
          pattern="https?://.+(:\d+)?"
        >
          >
        </b-input>
      </b-field>
      <template v-if="backendReady && dockerMode === false && !addressChanged">
        <b-field
          :label="$t('customAddressPort.portSocks5')"
          label-position="on-border"
        >
          <b-input
            v-model="table.socks5"
            placeholder="20170"
            type="number"
            min="0"
            required
          ></b-input>
        </b-field>
        <b-field
          :label="$t('customAddressPort.portHttp')"
          label-position="on-border"
        >
          <b-input
            v-model="table.http"
            placeholder="20171"
            type="number"
            min="0"
            required
          ></b-input>
        </b-field>
        <b-field
          :label="$t('customAddressPort.portSocks5WithPac')"
          label-position="on-border"
        >
          <b-input
            v-model="table.socks5WithPac"
            placeholder="0"
            type="number"
            min="0"
            required
          ></b-input>
        </b-field>
        <b-field
          :label="$t('customAddressPort.portHttpWithPac')"
          label-position="on-border"
        >
          <b-input
            v-model="table.httpWithPac"
            placeholder="20172"
            type="number"
            min="0"
            required
          ></b-input>
        </b-field>
        <b-field
          :label="$t('customAddressPort.portVmess')"
          label-position="on-border"
        >
          <b-input
            v-model="table.vmess"
            placeholder="0"
            type="number"
            min="0"
            required
          ></b-input>
        </b-field>
        <b-message
          v-if="table.vmess > 0 && table.vmessLink"
          type="is-info"
          style="font-size: 13px"
          class="after-line-dot5"
        >
          <b-button
            type="is-link is-info is-outlined"
            @click="handleClickShowVmessLink"
          >
            {{ $t("customAddressPort.portVmessLink") }}
          </b-button>
        </b-message>
        <b-message
          type="is-info"
          style="font-size: 13px"
          class="after-line-dot5"
        >
          <p
            v-show="!dockerMode"
            v-html="$t('customAddressPort.messages.0')"
          ></p>
          <p
            v-show="dockerMode"
            v-html="$t('customAddressPort.messages.1')"
          ></p>
          <p
            v-show="dockerMode"
            v-html="$t('customAddressPort.messages.2')"
          ></p>
          <p v-html="$t('customAddressPort.messages.3')"></p>
        </b-message>
      </template>
    </section>
    <footer class="modal-card-foot flex-end">
      <button class="button" @click="$emit('close')">
        {{ $t("operations.cancel") }}
      </button>
      <button class="button is-primary" @click="handleClickSubmit">
        {{ $t("operations.confirm") }}
      </button>
    </footer>
  </div>
</template>

<script>
import { handleResponse } from "@/assets/js/utils";
import i18n from "@/plugins/i18n";
import ModalSharing from "@/components/modalSharing";
import CONST from "@/assets/js/const";

export default {
  name: "ModalCustomPorts",
  i18n,
  data: () => ({
    table: {
      backendAddress: "http://localhost:2017",
      socks5: "20170",
      http: "20171",
      socks5WithPac: "0",
      httpWithPac: "20172",
      vmess: "0",
      vmessLink: "",
    },
    backendReady: false,
  }),
  computed: {
    dockerMode() {
      return window.localStorage["docker"] === "true";
    },
    addressChanged() {
      let backendAddress = this.table.backendAddress;
      if (backendAddress.endsWith("/")) {
        backendAddress = backendAddress.substr(0, backendAddress.length - 1);
      }
      return backendAddress + "/api" !== apiRoot;
    },
  },
  created() {
    this.table.backendAddress = localStorage["backendAddress"];
    this.$axios({
      url: apiRoot + "/ports",
    }).then((res) => {
      handleResponse(res, this, () => {
        this.backendReady = true;
        Object.assign(this.table, res.data.data);
      });
    });
  },
  methods: {
    handleClickShowVmessLink() {
      if (this.table.vmessLink) {
        this.$parent.close();
        this.$buefy.modal.open({
          width: 500,
          component: ModalSharing,
          props: {
            title: this.$t("customAddressPort.portVmessLink"),
            sharingAddress: this.table.vmessLink,
            shortDesc: "VMess | v2rayA",
            type: CONST.ServerType,
          },
        });
      } else {
        this.$buefy.toast.open({
          message: "no vmessLink found",
          type: "is-warning",
          position: "is-top",
          queue: false,
          duration: 5000,
        });
      }
    },
    handleClickSubmit() {
      if (!this.$refs.backendAddress.checkHtml5Validity()) {
        return;
      }
      //去除末位'/'
      let backendAddress = this.table.backendAddress;
      if (backendAddress.endsWith("/")) {
        backendAddress = backendAddress.substr(0, backendAddress.length - 1);
      }
      //当前服务端是否正常工作
      if (this.backendReady && !this.addressChanged) {
        this.$axios({
          url: backendAddress + "/api/ports",
          method: "put",
          data: {
            socks5: parseInt(this.table.socks5),
            http: parseInt(this.table.http),
            socks5WithPac: parseInt(this.table.socks5WithPac),
            httpWithPac: parseInt(this.table.httpWithPac),
            vmess: parseInt(this.table.vmess),
          },
        }).then((res) => {
          handleResponse(res, this, () => {
            if (res.data.data?.vmessLink) {
              this.$buefy.modal.open({
                width: 500,
                component: ModalSharing,
                props: {
                  title: this.$t("customAddressPort.portVmessLink"),
                  sharingAddress: res.data.data.vmessLink,
                  shortDesc: "VMess | v2rayA",
                  type: CONST.ServerType,
                },
              });
            }
            localStorage["backendAddress"] = backendAddress;
            this.$emit("close");
          });
        });
      } else {
        this.$axios({
          url: backendAddress + "/api/version",
        }).then(() => {
          localStorage["backendAddress"] = backendAddress;
          this.$emit("close");
          this.$remount();
        });
      }
    },
  },
};
</script>

<style lang="scss">
.modal-custom-ports .modal-background {
  background-color: rgba(0, 0, 0, 0.6);
}
</style>
