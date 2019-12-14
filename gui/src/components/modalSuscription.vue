<template>
  <div class="modal-card" style="max-width: 400px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">修改订阅</p>
    </header>
    <section class="modal-card-body">
      <b-field label="REMARKS">
        <b-input v-model="which.remarks" placeholder="别名" />
      </b-field>
    </section>
    <footer class="modal-card-foot flex-end">
      <button class="button" type="button" @click="$parent.close()">
        取消
      </button>
      <button class="button is-primary" @click="handleClickSubmit">
        保存
      </button>
    </footer>
  </div>
</template>

<script>
import { isVersionGreaterEqual } from "@/assets/js/utils";
import { SnackbarProgrammatic } from "../plugins/buefy";

export default {
  name: "ModalSubscription",
  props: {
    which: {
      type: Object,
      default() {
        return null;
      }
    }
  },
  mounted() {
    if (!isVersionGreaterEqual(localStorage["version"], "0.5.0")) {
      this.$buefy.snackbar.open({
        message: "修改订阅别名需要V2RayA版本高于0.5.0",
        type: "is-warning",
        queue: false,
        position: "is-top",
        duration: 3000,
        actionText: "查看帮助",
        onAction: () => {
          window.open(
            "https://github.com/mzz2017/V2RayA#%E4%BD%BF%E7%94%A8",
            "_blank"
          );
        }
      });
    }
  },
  methods: {
    handleClickSubmit() {
      this.$emit("submit", this.which);
    }
  }
};
</script>

<style lang="scss">
.is-twitter .is-active a {
  color: #4099ff !important;
}
.readonly {
  pointer-events: none;
}
.same-width-5 li {
  width: 5em;
}
</style>
