<template>
  <div class="modal-card" style="width: 680px; max-width: 100%; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">{{ $t("tinytun.routeScript.title") }}</p>
    </header>
    <section class="modal-card-body">
      <!-- Warning -->
      <b-message type="is-warning" has-icon>
        {{ $t("tinytun.routeScript.warning") }}
      </b-message>

      <!-- Shell type selector -->
      <b-field :label="$t('tinytun.routeScript.shellType')" label-position="on-border">
        <b-select v-model="localShellType" expanded @input="onShellTypeChange">
          <template v-if="isWindows">
            <option value="windows_powershell">Windows PowerShell</option>
            <option value="pwsh">PowerShell Core (pwsh)</option>
            <option value="cmd">Command Prompt (cmd)</option>
            <option value="git_bash">Git Bash</option>
          </template>
          <template v-else>
            <option value="bash">bash</option>
            <option value="zsh">zsh</option>
            <option value="sh">POSIX sh</option>
            <option value="fish">fish</option>
          </template>
          <option value="custom">{{ $t("tinytun.routeScript.customShell") }}</option>
        </b-select>
      </b-field>

      <!-- Custom shell path (shown when shell type is "custom" or has a custom path) -->
      <b-field :label="$t('tinytun.routeScript.shellPath')" label-position="on-border"
        v-if="localShellType === 'custom'">
        <b-input v-model="localShellPath" :placeholder="$t('tinytun.routeScript.shellPathPlaceholder')" expanded />
      </b-field>

      <!-- Setup script -->
      <b-field :label="$t('tinytun.routeScript.setupScript')" label-position="on-border" style="margin-top: 1rem">
        <b-input v-model="localSetupScript" type="textarea" rows="6"
          :placeholder="$t('tinytun.routeScript.setupScriptPlaceholder')" custom-class="code-font horizon-scroll"
          autocomplete="off" autocorrect="off" autocapitalize="off" spellcheck="false" />
      </b-field>

      <!-- Teardown script -->
      <b-field :label="$t('tinytun.routeScript.teardownScript')" label-position="on-border" style="margin-top: 1rem">
        <b-input v-model="localTeardownScript" type="textarea" rows="6"
          :placeholder="$t('tinytun.routeScript.teardownScriptPlaceholder')" custom-class="code-font horizon-scroll"
          autocomplete="off" autocorrect="off" autocapitalize="off" spellcheck="false" />
      </b-field>
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
  name: "ModalTinyTunRouteScript",
  props: {
    os: { type: String, default: "" },
    shellType: { type: String, default: "" },
    shellPath: { type: String, default: "" },
    setupScript: { type: String, default: "" },
    teardownScript: { type: String, default: "" },
  },
  data() {
    const defaultShell = this.os === "windows" ? "windows_powershell" : "bash";
    return {
      localShellType: this.shellType || defaultShell,
      localShellPath: this.shellPath || "",
      localSetupScript: this.setupScript || "",
      localTeardownScript: this.teardownScript || "",
    };
  },
  computed: {
    isWindows() {
      return this.os === "windows";
    },
  },
  methods: {
    onShellTypeChange(val) {
      // Clear custom path when switching away from custom
      if (val !== "custom") {
        this.localShellPath = "";
      }
    },
    handleClickSave() {
      this.$emit("save", {
        shellType: this.localShellType,
        shellPath: this.localShellPath,
        setupScript: this.localSetupScript,
        teardownScript: this.localTeardownScript,
      });
      this.$parent.close();
    },
  },
};
</script>

<style lang="scss" scoped>
.code-font {
  font-family: monospace;
}
</style>
