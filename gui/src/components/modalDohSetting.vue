<template>
  <div class="modal-card" style="max-width: 450px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $t("doh.title") }}
      </p>
    </header>
    <section class="modal-card-body">
      <b-message type="is-info" class="after-line-dot5">
        <p>{{ $t("doh.messages.0") }}</p>
        <p>{{ $t("doh.messages.1") }}</p>
        <p>{{ $t("doh.messages.2") }}</p>
        <p>{{ $t("doh.messages.3") }}</p>
        <p v-html="$t('doh.messages.4')" />
        <p v-html="$t('doh.messages.5')" />
      </b-message>
      <b-field :label="$t('doh.dohPriorityList')">
        <b-input v-model="dohlist" type="textarea" />
      </b-field>
      <b-message type="is-danger" class="after-line-dot5">
        <p>{{ $t("doh.messages.6") }}</p>
      </b-message>
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
import { handleResponse } from "@/assets/js/utils";

export default {
  name: "ModalDohSetting",
  data: () => ({
    dohlist: ""
  }),
  created() {
    this.$axios({
      url: apiRoot + "/dohList"
    }).then(res => {
      handleResponse(res, this, () => {
        if (res.data.data.dohlist) {
          let dohlist = res.data.data.dohlist;
          dohlist.trim();
          let arr = dohlist.split("\n");
          if (arr.length > 0) {
            this.dohlist = dohlist;
          }
        }
      });
    });
  },
  methods: {
    handleClickSubmit() {
      this.$axios({
        url: apiRoot + "/dohList",
        method: "put",
        data: {
          dohlist: this.dohlist
        }
      }).then(res => {
        handleResponse(res, this, () => {
          this.$emit("close");
        });
      });
    },
    beforeAdding(tag) {
      return /^\d+$/.test(tag) || /^\d+:\d+$/.test(tag);
    }
  }
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
