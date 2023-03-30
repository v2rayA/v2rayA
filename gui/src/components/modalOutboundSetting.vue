<template>
  <div class="modal-card" style="max-width: 450px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ outbound }} - {{ $t("common.outboundSetting") }}
      </p>
    </header>
    <section class="modal-card-body">
      <b-field label="ProbeURL" label-position="on-border">
        <b-input ref="probe_url" v-model="setting.probeURL" required expanded />
      </b-field>
      <b-field label="ProbeInterval" label-position="on-border">
        <b-input
          ref="probe_interval"
          v-model="setting.probeInterval"
          required
          expanded
        />
      </b-field>
      <b-field label="Type" label-position="on-border">
        <b-select v-model="setting.type" expanded>
          <option value="leastping">
            {{ $t("setting.options.leastPing") }}
          </option>
        </b-select>
      </b-field>
    </section>
    <footer class="modal-card-foot flex-end">
      <button class="button is-danger" type="button" @click="handleClickDelete">
        {{ $t("operations.delete") }}
      </button>
      <div
        style="display: flex; justify-content: flex-end; width: -moz-available"
      >
        <button class="button" @click="$emit('close')">
          {{ $t("operations.cancel") }}
        </button>
        <button class="button is-primary" @click="handleClickSubmit">
          {{ $t("operations.confirm") }}
        </button>
      </div>
    </footer>
  </div>
</template>

<script>
import { handleResponse } from "../assets/js/utils";
import i18n from "@/plugins/i18n";

export default {
  name: "ModalOutboundSetting",
  i18n,
  props: {
    outbound: {
      type: String,
      required: true,
    },
  },
  data: () => ({
    setting: {
      probeURL: "",
      probeInterval: "",
      type: "",
    },
    backendReady: false,
  }),
  created() {
    this.$axios({
      url: apiRoot + "/outbound",
      params: {
        outbound: this.outbound,
      },
    }).then((res) => {
      handleResponse(res, this, () => {
        Object.assign(this.setting, res.data.data.setting);
      });
    });
  },
  methods: {
    handleClickDelete() {
      const that = this;
      this.$buefy.dialog.confirm({
        title: that.$t("delete.title"),
        message: that.$t("outbound.deleteMessage", {
          outboundName: that.outbound,
        }),
        confirmText: that.$t("operations.delete"),
        cancelText: that.$t("operations.cancel"),
        type: "is-danger",
        hasIcon: true,
        icon: " iconfont icon-alert",
        onConfirm: () => {
          that.$emit("delete");
          that.$parent.close();
        },
      });
    },
    handleClickSubmit() {
      let valid = true;
      for (let k in this.$refs) {
        if (!this.$refs.hasOwnProperty(k)) {
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
      const that = this;
      this.$axios({
        url: apiRoot + "/outbound",
        method: "put",
        data: {
          outbound: this.outbound,
          setting: this.setting,
        },
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
      });
    },
  },
};
</script>

<style lang="scss">
.modal-custom-ports .modal-background {
  background-color: rgba(0, 0, 0, 0.6);
}
</style>
