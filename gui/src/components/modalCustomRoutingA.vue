<template>
  <div
    class="modal-card modal-configure-pac"
    style="max-width: 550px;height:700px;margin:auto"
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
      />
    </section>
    <footer class="modal-card-foot">
      <div
        style="position:relative;display:flex;justify-content:flex-end;width:100%"
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
    routingA: ""
  }),
  mounted() {
    this.$axios({
      url: apiRoot + "/routingA"
    })
      .then(res => {
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
  },
  methods: {
    handleClickManual() {
      window.open("https://github.com/mzz2017/V2RayA/wiki/RoutingA", "_blank");
    },
    handleClickSubmit() {
      this.$axios({
        url: apiRoot + "/routingA",
        method: "put",
        data: {
          routingA: this.routingA
        }
      }).then(res => {
        handleResponse(res, this, () => {
          this.$parent.close();
        });
      });
    }
  }
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
}

.horizon-scroll {
  overflow-x: auto;
  white-space: nowrap;
}

.code-font {
  font-family: Monaco, Menlo, Consolas, Courier New, monospace;
}
</style>
