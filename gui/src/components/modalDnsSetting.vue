<template>
  <div class="modal-card" style="max-width: 450px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $t("dns.title") }}
      </p>
    </header>
    <section class="modal-card-body">
      <b-message type="is-info" class="after-line-dot5">
        <p>{{ $t("dns.messages.0") }}</p>
      </b-message>
      <b-field :label="$t('dns.internalQueryServers')">
        <b-input
          v-model="internal"
          type="textarea"
          custom-class="full-min-height horizon-scroll code-font"
        />
      </b-field>
      <b-field :label="$t('dns.externalQueryServers')">
        <b-input
          v-model="external"
          type="textarea"
          custom-class="full-min-height horizon-scroll code-font"
          autocomplete="off"
          autocorrect="off"
          autocapitalize="off"
          spellcheck="false"
        />
      </b-field>
      <b-message type="is-danger" class="after-line-dot5">
        <p>{{ $t("dns.messages.1") }}</p>
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
    internal: "",
    external: "",
  }),
  created() {
    this.$axios({
      url: apiRoot + "/dnsList",
    }).then((res) => {
      handleResponse(res, this, () => {
        if (res.data.data.internal) {
          let internal = res.data.data.internal;
          let external = res.data.data.external;
          console.log(res.data.data);
          if (internal.length) {
            this.internal = internal.join("\n");
          }
          if (external.length) {
            this.external = external.join("\n");
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
          internal: this.internal,
          external: this.external,
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
