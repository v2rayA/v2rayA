<template>
  <div class="modal-card" style="max-width: 450px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $t("gfwList.title") }}
      </p>
      <button type="button" class="delete" aria-label="close" @click="$emit('close')"></button>
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
      <button
        :disabled="disableDeleteBtn"
        class="button is-danger"
        @click="handleClickDelete"
      >
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
import { openLoading } from "@/plugins/session";

export default {
  name: "modalUpdateGfwList",
  emits: ["close"],
  data: () => ({
    disableDeleteBtn: false,
    downloadLink: "",
  }),
  created() {
    this.$axios({
      url: apiRoot + "/setting",
    }).then((res) => {
      handleResponse(res, this, () => {
        this.disableDeleteBtn = res.data.data.localGFWListVersion == "";
      });
    });
  },
  methods: {
    handleClickDelete() {
      this.$axios({
        url: apiRoot + "/gfwList",
        method: "delete",
      }).then((res) => {
        handleResponse(res, this, () => {
          this.$emit("close");
        }, null, "delete.failed");
      });
    },
    handleClickSubmit() {
      if (!this.downloadLink.startsWith("http") && this.downloadLink != "") {
        this.$buefy.toast.open({
          message: this.$t("gfwList.wrongCustomLink"),
          type: "is-warning",
          position: "is-top",
          duration: 5000,
        });
        return;
      }
      let loading = openLoading(this);
      this.$axios({
        url: apiRoot + "/gfwList",
        method: "put",
        timeout: 0,
        data: {
          downloadLink: this.downloadLink,
        },
      }).then((res) => {
        loading.close();
        handleResponse(res, this, () => {
          this.$emit("close");
          // "Already the latest" is a result, not a failure: the backend used
          // to report it as an error and the dialog said "could not update".
          const upToDate = res.data.data && res.data.data.alreadyUpToDate;
          this.$buefy.toast.open({
            message: upToDate
              ? this.$t("gfwList.alreadyUpToDate", {
                  version: res.data.data.localGFWListVersion,
                })
              : this.$t("gfwList.updated"),
            type: upToDate ? "is-info" : "is-success",
            position: "is-top",
            duration: 5000,
          });
        }, null, "gfwList.saveFailed");
      }).catch(() => {
        loading.close();
      });
    },
  },
};
</script>
