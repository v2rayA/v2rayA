<template>
  <div
    ref="modal"
    class="modal-card modal-routinga"
    style="width: 100%; height: 2000px; margin: auto"
  >
    <header class="modal-card-head">
      <p class="modal-card-title">RoutingA</p>
    </header>
    <section class="modal-card-body rules">
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
    handleClickManual() {
      window.open("https://github.com/v2rayA/v2rayA/wiki/RoutingA", "_blank");
    },
    handleClickSubmit() {
      this.$axios({
        url: apiRoot + "/routingA",
        method: "put",
        data: {
          routingA: this.routingA,
        },
      }).then((res) => {
        handleResponse(res, this, () => {
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
