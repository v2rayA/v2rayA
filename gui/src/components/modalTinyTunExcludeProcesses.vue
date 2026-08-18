<template>
  <div class="modal-card" style="width: 680px; max-width: 100%; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">{{ $t("tinytun.processExclude.title") }}</p>
    </header>
    <section class="modal-card-body">
      <b-message type="is-warning" has-icon>
        {{ $t("tinytun.processExclude.warning") }}
      </b-message>

      <b-field :label="$t('tinytun.processExclude.listLabel')" label-position="on-border">
        <b-input
          v-model="localExcludeProcessesText"
          type="textarea"
          rows="10"
          :placeholder="$t('tinytun.processExclude.placeholder')"
          custom-class="code-font horizon-scroll"
          autocomplete="off"
          autocorrect="off"
          autocapitalize="off"
          spellcheck="false"
        />
      </b-field>
      <p class="help is-size-7">{{ $t("tinytun.processExclude.hint") }}</p>
    </section>
    <footer class="modal-card-foot" style="justify-content: flex-end">
      <button class="button" type="button" @click="$parent.close()">
        {{ $t("operations.cancel") }}
      </button>
      <button class="button is-primary" @click="handleClickSave">
        {{ $t("operations.save") }}
      </button>
    </footer>
  </div>
</template>

<script>
export default {
  name: "ModalTinyTunExcludeProcesses",
  props: {
    excludeProcesses: { type: String, default: "" },
  },
  data() {
    return {
      localExcludeProcessesText: this.formatExcludeProcesses(this.excludeProcesses),
    };
  },
  methods: {
    parseExcludeProcesses(raw) {
      const parts = (raw || "")
        .split(/[\n,;\t]/g)
        .map((p) => p.trim())
        .filter((p) => p.length > 0);
      const seen = new Set();
      const values = [];
      for (const part of parts) {
        if (seen.has(part)) {
          continue;
        }
        seen.add(part);
        values.push(part);
      }
      return values;
    },
    formatExcludeProcesses(raw) {
      return this.parseExcludeProcesses(raw).join("\n");
    },
    handleClickSave() {
      const values = this.parseExcludeProcesses(this.localExcludeProcessesText);
      this.$emit("save", {
        excludeProcesses: values.join(","),
      });
      this.$parent.close();
    },
  },
};
</script>

<style scoped>
.code-font {
  font-family: monospace;
}
</style>
