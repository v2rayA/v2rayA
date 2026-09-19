<script setup lang="ts">
// The scripts that set up and tear down routing when auto route is off:
// the shell to run them with, and the two scripts. Returns the four
// values to the settings form.
import { computed, reactive } from "vue";
import { useI18n } from "vue-i18n";

defineOptions({ name: "TunRouteScriptDialog" });
export interface TunRouteScript {
  shellType: string;
  shellPath: string;
  setupScript: string;
  teardownScript: string;
}
const props = defineProps<{ os: string; value: TunRouteScript }>();
const emit = defineEmits<{ close: [value?: TunRouteScript] }>();
const { t } = useI18n();

const windows = props.os === "windows";
const form = reactive<TunRouteScript>({
  ...props.value,
  shellType: props.value.shellType || (windows ? "windows_powershell" : "bash"),
});
const shells = computed(() => [
  ...(windows
    ? [
        { value: "windows_powershell", title: "Windows PowerShell" },
        { value: "pwsh", title: "PowerShell Core (pwsh)" },
        { value: "cmd", title: "Command Prompt (cmd)" },
        { value: "git_bash", title: "Git Bash" },
      ]
    : [
        { value: "bash", title: "bash" },
        { value: "zsh", title: "zsh" },
        { value: "sh", title: "POSIX sh" },
        { value: "fish", title: "fish" },
      ]),
  { value: "custom", title: t("tun.routeScript.customShell") },
]);

function save() {
  if (form.shellType !== "custom") form.shellPath = "";
  emit("close", { ...form });
}
</script>

<template>
  <v-card>
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0">
        {{ t("tun.routeScript.title") }}
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6">
      <v-alert
        type="warning"
        variant="tonal"
        density="compact"
        class="md3-body-small mb-4"
      >
        {{ t("tun.routeScript.warning") }}
      </v-alert>
      <v-select
        v-model="form.shellType"
        :items="shells"
        :label="t('tun.routeScript.shellType')"
      />
      <v-text-field
        v-if="form.shellType === 'custom'"
        v-model="form.shellPath"
        :label="t('tun.routeScript.shellPath')"
        :placeholder="t('tun.routeScript.shellPathPlaceholder')"
        dir="ltr"
      />
      <v-textarea
        v-model="form.setupScript"
        :label="t('tun.routeScript.setupScript')"
        :placeholder="t('tun.routeScript.setupScriptPlaceholder')"
        rows="5"
        auto-grow
        dir="ltr"
        spellcheck="false"
        class="code"
      />
      <v-textarea
        v-model="form.teardownScript"
        :label="t('tun.routeScript.teardownScript')"
        :placeholder="t('tun.routeScript.teardownScriptPlaceholder')"
        rows="5"
        auto-grow
        dir="ltr"
        spellcheck="false"
        class="code"
      />
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn variant="flat" color="primary" @click="save">{{
        t("operations.save")
      }}</v-btn>
    </v-card-actions>
  </v-card>
</template>

<style scoped>
.code :deep(textarea) {
  font-family: ui-monospace, "Cascadia Mono", "Fira Mono", Menlo, monospace;
}
</style>
