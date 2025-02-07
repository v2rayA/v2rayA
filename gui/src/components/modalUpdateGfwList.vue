<template>
  <div class="modal-card" style="max-width: 450px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $t("gfwList.title") }}
      </p>
    </header>
    <section class="modal-card-body">
      <b-message type="is-info" class="after-line-dot5">
        <p>{{ $t("gfwList.messages.0") }}</p>
      </b-message>
      <b-message type="is-info" class="after-line-dot5">
        <p>{{ $t("gfwList.messages.1") }}</p>
      </b-message>
      <b-field :label="$t('gfwList.formName')">
        <b-input
          v-model="downloadLink"
          placeholder="https://example.com/LoyalsoldierSite.dat"
          custom-class="full-min-height horizon-scroll code-font"
        />
      </b-field>
      <b-message type="is-warning" class="after-line-dot5">
        <p>{{ $t("gfwList.messages.2") }}</p>
      </b-message>
    </section>
    <footer class="modal-card-foot flex-end">
      <button class="button" @click="$emit('close')">
        {{ $t("operations.cancel") }}
      </button>
      <button class="button is-danger" @click="handleClickDelete">
        {{ $t("operations.delete") }}
      </button>
      <button class="button is-primary" @click="handleClickSubmit">
        {{
          downloadLink == "" ? $t("operations.autoUpdate") : $t("operations.manualUpdate")
        }}
      </button>
    </footer>
  </div>
</template>
<script>
import { handleResponse } from "@/assets/js/utils";

export default {
  name: "modalUpdateGfwList",
  data: () => ({
    downloadLink: "",
  }),
  methods: {
    handleClickDelete() {
      this.$axios({
        url: apiRoot + "/gfwList",
        method: "delete",
        data: {
          domains: this.domains,
        },
      }).then((res) => {
        handleResponse(res, this, () => {
          this.$emit("close");
        });
      });
    },
    handleClickSubmit() {
      this.$axios({
        url: apiRoot + "/gfwList",
        method: "put",
        timeout: 0,
      }).then((res) => {
        handleResponse(res, this, () => {
          this.$emit("close");
          this.$buefy.toast.open({
            message: this.$t("common.success"),
            type: "is-warning",
            position: "is-top",
            duration: 5000,
            queue: false,
          });
        });
      });
    },
  },
};
</script>
