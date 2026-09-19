<script setup lang="ts">
import { computed, nextTick, onMounted, ref, useId, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useDisplay } from "vuetify";
import { mdiClose, mdiCodeTags, mdiFormSelect, mdiOpenInNew } from "@mdi/js";
import { getRoutingA, putRoutingA } from "@/api";
import { errorText } from "@/api/errors";
import { useConfirm, useNotify } from "@/composables";
import RoutingEditor from "./routingA/RoutingEditor.vue";
import RoutingReference from "./routingA/RoutingReference.vue";
import RoutingForm from "./routingA/RoutingForm.vue";
import { template } from "./routingA/template";

defineOptions({ name: "RoutingADialog" });
const emit = defineEmits<{ close: [saved?: boolean] }>();
const { t } = useI18n();
const { width } = useDisplay();
const confirm = useConfirm();
const notify = useNotify();
const routingA = ref("");
const original = ref("");
const loading = ref(true);
const saving = ref(false);
const confirming = ref(false);
const backendError = ref("");
const warningDismissed = ref(false);
const editor = ref<InstanceType<typeof RoutingEditor>>();
const referenceId = useId();
const wide = computed(() => width.value >= 840);
const referencePreference = ref<string | null>(null);
const view = ref<"form" | "text">("form");
const picker = ref<HTMLInputElement>();
try {
  referencePreference.value = localStorage.getItem("routingA.reference");
  if (localStorage.getItem("routingA.view") === "text") view.value = "text";
} catch {
  // Storage can be unavailable in private browsing.
}
const showReference = computed(() =>
  referencePreference.value === null
    ? width.value >= 600
    : referencePreference.value === "true",
);
const hasInboundDef = computed(() =>
  routingA.value.split("\n").some((line) => /^\s*inbound\s*\(/.test(line)),
);
const busy = computed(() => loading.value || saving.value || confirming.value);
watch(routingA, () => {
  backendError.value = "";
  warningDismissed.value = false;
});
watch(view, (value) => {
  try {
    localStorage.setItem("routingA.view", value);
  } catch {
    // View switching does not require storage access.
  }
});
async function insertExample(example: string) {
  if (busy.value) return;
  view.value = "text";
  await nextTick();
  await editor.value?.insert(example, true);
}
function exportRules() {
  const url = URL.createObjectURL(
    new Blob([routingA.value], { type: "text/plain;charset=utf-8" }),
  );
  const link = document.createElement("a");
  link.href = url;
  link.download = "routingA.txt";
  link.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}
async function importRules(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = "";
  if (!file || busy.value) return;
  confirming.value = true;
  try {
    const text = await file.text();
    if (
      await confirm({
        message: t("routingA.import.confirm"),
        confirmText: t("routingA.import.title"),
        cancelText: t("operations.cancel"),
      })
    )
      routingA.value = text;
  } catch (err) {
    notify.warning(errorText(err));
  } finally {
    confirming.value = false;
  }
}
function toggleReference() {
  referencePreference.value = String(!showReference.value);
  try {
    localStorage.setItem("routingA.reference", referencePreference.value);
  } catch {
    // The panel still works without persisted preferences.
  }
}
onMounted(async () => {
  try {
    const res = await getRoutingA();
    routingA.value = original.value = res.routingA;
  } catch (err) {
    notify.warning(errorText(err));
    emit("close");
  } finally {
    loading.value = false;
  }
});
async function close() {
  if (saving.value || confirming.value) return;
  confirming.value = true;
  try {
    if (
      routingA.value !== original.value &&
      !(await confirm({
        message: t("routingA.discard"),
        confirmText: t("operations.confirm"),
        cancelText: t("operations.cancel"),
      }))
    )
      return;
    emit("close");
  } finally {
    confirming.value = false;
  }
}
async function reset() {
  if (busy.value) return;
  confirming.value = true;
  try {
    if (
      await confirm({
        message: t("routingA.resetConfirm"),
        confirmText: t("routingA.resetDefault"),
        cancelText: t("operations.cancel"),
      })
    )
      routingA.value = template;
  } finally {
    confirming.value = false;
  }
}
async function save() {
  if (busy.value) return;
  saving.value = true;
  try {
    if (
      hasInboundDef.value &&
      !(await confirm({
        message: t("routingA.inboundDeprecatedConfirm"),
        confirmText: t("operations.save"),
        cancelText: t("operations.cancel"),
      }))
    )
      return;
    const res = await putRoutingA({ routingA: routingA.value });
    if (res && typeof res === "object" && "warning" in res && res.warning) {
      notify.warning(t("routingA.savedWithWarning", { warning: res.warning }), {
        timeout: 8000,
      });
    }
    emit("close", true);
  } catch (err) {
    backendError.value = errorText(err);
    notify.warning(t("routingA.saveFailed", { message: backendError.value }));
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <v-card>
    <v-card-item class="routing-title px-6 pt-6 pb-4">
      <v-card-title class="md3-headline-small pa-0">RoutingA</v-card-title>
      <template #append>
        <v-btn-toggle
          v-model="view"
          mandatory
          variant="text"
          selected-class="bg-secondary-container text-on-secondary-container"
          class="me-2"
        >
          <v-btn
            value="form"
            :icon="mdiFormSelect"
            width="48"
            :aria-label="t('routingA.form.title')"
            :title="t('routingA.form.title')"
          />
          <v-btn
            value="text"
            :icon="mdiCodeTags"
            width="48"
            :aria-label="t('routingA.form.text')"
            :title="t('routingA.form.text')"
          />
        </v-btn-toggle>
        <v-btn
          variant="text"
          :aria-expanded="showReference"
          :aria-controls="referenceId"
          @click="toggleReference"
          >{{ t("routingA.reference.title") }}</v-btn
        >
        <v-tooltip :text="t('operations.close')">
          <template #activator="{ props: tip }">
            <v-btn
              v-bind="tip"
              :icon="mdiClose"
              variant="text"
              size="40"
              :aria-label="t('operations.close')"
              :disabled="saving || confirming"
              @click="close"
            />
          </template>
        </v-tooltip>
      </template>
    </v-card-item>
    <v-card-text class="px-6">
      <v-alert
        v-if="hasInboundDef && !warningDismissed"
        type="warning"
        variant="tonal"
        density="compact"
        closable
        class="md3-body-small mb-4"
        @click:close="warningDismissed = true"
        >{{ t("routingA.inboundDeprecated") }}</v-alert
      >
      <v-progress-linear
        v-if="loading"
        indeterminate
        color="primary"
        :aria-label="t('routingA.loading')"
        class="mb-4"
      />
      <div
        class="routing-layout"
        :class="{ 'routing-layout--wide': wide && showReference }"
      >
        <RoutingForm
          v-if="view === 'form'"
          v-model="routingA"
          :disabled="busy"
        />
        <RoutingEditor
          v-else
          ref="editor"
          v-model="routingA"
          :disabled="loading"
          :readonly="saving || confirming"
          @save="save"
        />
        <v-expand-transition>
          <RoutingReference
            v-show="showReference"
            :id="referenceId"
            :disabled="busy"
            @insert="insertExample"
          />
        </v-expand-transition>
      </div>
      <v-alert
        v-if="backendError"
        type="error"
        variant="tonal"
        class="mt-4"
        data-testid="routing-error"
      >
        <pre class="routing-error md3-body-small" dir="ltr">{{
          backendError
        }}</pre>
      </v-alert>
    </v-card-text>
    <v-card-actions class="px-6 pb-4 flex-wrap ga-2">
      <input
        ref="picker"
        type="file"
        accept=".txt,.routinga,text/plain"
        hidden
        @change="importRules"
      />
      <v-btn variant="text" :disabled="busy" @click="picker?.click()">{{
        t("routingA.import.title")
      }}</v-btn>
      <v-btn variant="text" :disabled="loading" @click="exportRules">{{
        t("routingA.export")
      }}</v-btn>
      <v-btn variant="text" :disabled="busy" @click="reset">{{
        t("routingA.resetDefault")
      }}</v-btn>
      <v-btn
        variant="text"
        :append-icon="mdiOpenInNew"
        href="https://github.com/v2rayA/v2rayA/wiki/RoutingA"
        target="_blank"
        rel="noopener noreferrer"
        >{{ t("operations.helpManual") }}</v-btn
      >
      <v-spacer />
      <v-btn variant="text" :disabled="saving || confirming" @click="close">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn
        color="primary"
        variant="flat"
        :loading="saving"
        :disabled="loading || confirming"
        @click="save"
        >{{ t("operations.save") }}</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<style scoped>
.routing-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 24px;
  align-items: start;
}
.routing-layout--wide {
  grid-template-columns: minmax(0, 1fr) 320px;
}
.routing-error {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font-family: ui-monospace, "Cascadia Mono", "Fira Mono", Menlo, monospace;
}
@media (max-width: 599px) {
  .routing-title {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .routing-title :deep(.v-card-item__append) {
    padding-inline-start: 0;
    margin-inline-start: auto;
  }
}
</style>
