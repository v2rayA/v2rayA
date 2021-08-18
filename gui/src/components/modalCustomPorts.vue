<template>
  <div class="modal-card" style="max-width: 450px;margin:auto">
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
          :label="$t('customAddressPort.portVlessGrpc')"
          label-position="on-border"
        >
          <b-input
            v-model="table.vlessGrpc"
            placeholder="0"
            type="number"
            min="0"
            required
          ></b-input>
        </b-field>
        <b-message
          v-if="table.vlessGrpc > 0 && table.vlessGrpcLink"
          type="is-info"
          style="font-size:13px"
          class="after-line-dot5"
        >
          <p>
            {{ $t("customAddressPort.portVlessGrpcLink") }}:
            <code>{{ table.vlessGrpcLink }}</code>
          </p>
        </b-message>
        <b-message
          type="is-info"
          style="font-size:13px"
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
import { handleResponse } from "../assets/js/utils";
import i18n from "@/plugins/i18n";

export default {
  name: "ModalCustomPorts",
  i18n,
  data: () => ({
    table: {
      backendAddress: "http://localhost:2017",
      socks5: "20170",
      http: "20171",
      httpWithPac: "20172",
      vlessGrpc: "0",
      vlessGrpcLink: ""
    },
    backendReady: false
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
    }
  },
  created() {
    this.table.backendAddress = localStorage["backendAddress"];
    this.$axios({
      url: apiRoot + "/ports"
    }).then(res => {
      handleResponse(res, this, () => {
        this.backendReady = true;
        Object.assign(this.table, res.data.data);
      });
    });
  },
  methods: {
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
            httpWithPac: parseInt(this.table.httpWithPac),
            vlessGrpc: parseInt(this.table.vlessGrpc)
          }
        }).then(res => {
          handleResponse(res, this, () => {
            if (res.data.data?.vlessGrpcLink) {
              this.$buefy.dialog.confirm({
                title: `${this.$t("customAddressPort.portVlessGrpcLink")}`,
                message: res.data.data.vlessGrpcLink,
                size: "is-small",
                closeOnConfirm: true
              });
            }
            localStorage["backendAddress"] = backendAddress;
            this.$emit("close");
          });
        });
      } else {
        this.$axios({
          url: backendAddress + "/api/version"
        }).then(() => {
          localStorage["backendAddress"] = backendAddress;
          this.$emit("close");
          this.$remount();
        });
      }
    }
  }
};
</script>

<style lang="scss">
.modal-custom-ports .modal-background {
  background-color: rgba(0, 0, 0, 0.6);
}
</style>
