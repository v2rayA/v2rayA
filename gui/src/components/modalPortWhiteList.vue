<template>
  <div class="modal-card" style="max-width: 450px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $t("egressPortWhitelist.title") }}
      </p>
    </header>
    <section class="modal-card-body">
      <b-message type="is-info" class="after-line-dot5">
        <p>
          <b>{{ $t("egressPortWhitelist.messages.0") }}</b>
        </p>
        <p>{{ $t("egressPortWhitelist.messages.1") }}</p>
        <p>{{ $t("egressPortWhitelist.messages.2", { v2rayaPort }) }}</p>
        <p>
          <b>{{ $t("egressPortWhitelist.messages.3") }}</b>
        </p>
        <p>{{ $t("egressPortWhitelist.messages.4") }}</p>
      </b-message>
      <b-field :label="$t('egressPortWhitelist.tcpPortWhitelist')">
        <b-taginput
          v-model="tcp"
          :before-adding="beforeAdding"
          icon=" iconfont icon-label"
        >
        </b-taginput> </b-field
      ><b-field :label="$t('egressPortWhitelist.udpPortWhitelist')">
        <b-taginput
          v-model="udp"
          :before-adding="beforeAdding"
          icon=" iconfont icon-label"
        >
        </b-taginput>
      </b-field>
    </section>
    <footer class="modal-card-foot flex-end">
      <button class="button" @click="$emit('close')">
        {{ $t("operations.cancel") }}
      </button>
      <button class="button is-primary" @click="handleClickSubmit">
        {{ $t("operations.save") }}
      </button>
    </footer>
  </div>
</template>

<script>
import { handleResponse, parseURL } from "@/assets/js/utils";

export default {
  name: "ModalPortWhiteList",
  data: () => ({
    tcp: [],
    udp: [],
  }),
  computed: {
    v2rayaPort() {
      let U = parseURL(apiRoot);
      let port = U.port;
      if (!port) {
        port =
          U.protocol === "http" ? "80" : U.protocol === "https" ? "443" : "";
      }
      return port;
    },
  },
  created() {
    this.$axios({
      url: apiRoot + "/portWhiteList",
    }).then((res) => {
      handleResponse(res, this, () => {
        if (res.data.data.tcp && res.data.data.tcp.length > 0) {
          this.tcp = res.data.data.tcp;
        }
        if (res.data.data.udp && res.data.data.udp.length > 0) {
          this.udp = res.data.data.udp;
        }
      });
    });
  },
  methods: {
    handleClickSubmit() {
      this.$axios({
        url: apiRoot + "/portWhiteList",
        method: "put",
        data: {
          tcp: this.tcp,
          udp: this.udp,
        },
      }).then((res) => {
        handleResponse(res, this, () => {
          this.$emit("close");
        });
      });
    },
    beforeAdding(tag) {
      return /^\d+$/.test(tag) || /^\d+:\d+$/.test(tag);
    },
  },
};
</script>

<style lang="scss" scoped>
.after-line-dot5 {
  font-size: 14px;
  p {
    font-size: 14px;
  }
}
</style>
