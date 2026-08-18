<template>
  <div
    ref="modal"
    class="modal-card modal-routinga"
    style="width: 100%; height: 100%; margin: auto"
  >
    <header class="modal-card-head">
      <p class="modal-card-title">RoutingA</p>
    </header>
    <section class="modal-card-body rules">
      <!-- Deprecation warning for inbound definitions -->
      <b-message
        v-if="hasInboundDef"
        type="is-warning"
        size="is-small"
        :active="true"
        closable
        @close="hasInboundDef = false"
      >
        {{ $t("routingA.inboundDeprecated") }}
      </b-message>
      <b-input
        v-model="routingA"
        type="textarea"
        class="full-min-height"
        custom-class="full-min-height horizon-scroll code-font"
        :placeholder="$t('routingA.messages.0')"
        autocomplete="off"
        autocorrect="off"
        autocapitalize="off"
        spellcheck="false"
        @input="checkInboundDef"
      />
    </section>
    <footer class="modal-card-foot">
      <div
        style="
          position: relative;
          display: flex;
          justify-content: flex-end;
          width: 100%;
        "
      >
        <button class="button btn-left" @click="handleClickManual">
          {{ $t("operations.helpManual") }}
        </button>
        <button class="button" type="button" @click="$parent.close()">
          {{ $t("operations.cancel") }}
        </button>
        <button class="button is-primary" @click="handleClickSubmit">
          {{ $t("operations.save") }}
        </button>
      </div>
    </footer>
  </div>
</template>

<script>
import { handleResponse } from "@/assets/js/utils";

export default {
  name: "ModalCustomRoutingA",
  data: () => ({
    routingA: "",
    hasInboundDef: false,
  }),
  mounted() {
    this.$axios({
      url: apiRoot + "/routingA",
    })
      .then((res) => {
        handleResponse(
          res,
          this,
          () => {
            this.routingA = res.data.data.routingA;
            this.checkInboundDef();
          },
          () => {
            this.$parent.close();
          }
        );
      })
      .catch(() => {
        this.$parent.close();
      });
    this.initRoutingAAnimationContentStyle();
  },
  methods: {
    initRoutingAAnimationContentStyle() {
      let e = this.$refs.modal;
      while (e && !/\banimation-content\b/.test(e.className)) {
        e = e.parentElement;
      }
      if (e) {
        e.className = e.className.replace(
          "animation-content",
          "routinga-animation-content"
        );
      }
    },
    checkInboundDef() {
      // Check if the RoutingA text contains inbound definitions (deprecated feature)
      const lines = (this.routingA || "").split("\n");
      this.hasInboundDef = lines.some(
        (line) =>
          line.trim().startsWith("inbound(") || line.trim().startsWith("inbound (")
      );
    },
    handleClickManual() {
      window.open("https://github.com/v2rayA/v2rayA/wiki/RoutingA", "_blank");
    },
    handleClickSubmit() {
      // If inbound definitions exist, show a confirmation dialog
      if (this.hasInboundDef) {
        this.$buefy.dialog.confirm({
          message: this.$t("routingA.inboundDeprecatedConfirm"),
          type: "is-warning",
          confirmText: this.$t("operations.save"),
          cancelText: this.$t("operations.cancel"),
          onConfirm: () => {
            this.submitRoutingA();
          },
        });
      } else {
        this.submitRoutingA();
      }
    },
    submitRoutingA() {
      this.$axios({
        url: apiRoot + "/routingA",
        method: "put",
        data: {
          routingA: this.routingA,
        },
      }).then((res) => {
        handleResponse(res, this, () => {
          // Show warning from server if any
          if (res.data.data && res.data.data.warning) {
            this.$buefy.toast.open({
              message: res.data.data.warning,
              type: "is-warning",
              position: "is-top",
              duration: 8000,
              queue: false,
            });
          }
          this.$parent.close();
        });
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.btn-left {
  position: absolute;
  left: 0;
  top: 0;
}
</style>
<style lang="scss">
.full-min-height {
  height: 100%;
  max-height: unset !important;
}

.horizon-scroll {
  overflow-x: auto;
  white-space: pre;
  line-height: 1.6em;
}
</style>
