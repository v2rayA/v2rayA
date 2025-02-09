<template>
  <div class="modal-card" style="max-width: 450px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $t("domainsExcluded.title") }}
      </p>
    </header>
    <section class="modal-card-body">
      <b-message type="is-info" class="after-line-dot5">
        <p>{{ $t("domainsExcluded.messages.0") }}</p>
      </b-message>
      <b-field :label="$t('domainsExcluded.formName')">
        <b-input
          v-model="domains"
          type="textarea"
          :placeholder="$t('domainsExcluded.formPlaceholder')"
          custom-class="full-min-height horizon-scroll code-font"
        />
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
import { handleResponse } from "@/assets/js/utils";

export default {
  name: "modalDomainsExcluded",
  data: () => ({
    domains: "",
  }),
  created() {
    this.$axios({
      url: apiRoot + "/domainsExcluded",
    }).then((res) => {
      handleResponse(res, this, () => {
        if (res.data.data.domains) {
          this.domains = res.data.data.domains;
        }
      });
    });
  },
  methods: {
    handleClickSubmit() {
      this.$axios({
        url: apiRoot + "/domainsExcluded",
        method: "put",
        data: {
          domains: this.domains,
        },
      }).then((res) => {
        handleResponse(res, this, () => {
          this.$emit("close");
        });
      });
    },
  },
};
</script>
