<template>
  <div class="modal-card" style="max-width: 450px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $t("dns.title") }}
      </p>
    </header>
    <section class="modal-card-body">
      <b-message type="is-info" class="after-line-dot5">
        <p>{{ $t("dns.messages.0") }}</p>
        <p>{{ $t("dns.messages.1") }}</p>
      </b-message>
      <b-field :label="$t('dns.dnsPriorityList')">
        <b-input v-model="dnslist" type="textarea" />
      </b-field>
      <b-message type="is-danger" class="after-line-dot5">
        <p>{{ $t("dns.messages.2") }}</p>
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
  name: "ModalDnsSetting",
  data: () => ({
    dnslist: ""
  }),
  created() {
    this.$axios({
      url: apiRoot + "/dnsList"
    }).then(res => {
      handleResponse(res, this, () => {
        if (res.data.data.dnslist) {
          let dnslist = res.data.data.dnslist;
          dnslist.trim();
          let arr = dnslist.split("\n");
          if (arr.length > 0) {
            this.dnslist = dnslist;
          }
        }
      });
    });
  },
  methods: {
    handleClickSubmit() {
      this.$axios({
        url: apiRoot + "/dnsList",
        method: "put",
        data: {
          dnslist: this.dnslist
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
