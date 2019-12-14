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
