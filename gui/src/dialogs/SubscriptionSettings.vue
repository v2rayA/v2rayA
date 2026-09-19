<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { errorText } from "@/api/errors";
import { useNotify } from "@/composables";
import { required } from "@/dialogs/Server/forms/parts/rules";
import { useSettings } from "@/views/settings/model";
import {
  subscriptionUpdateModes,
  updateProxyModes,
} from "@/views/settings/options";

defineOptions({ name: "SubscriptionSettingsDialog" });
const emit = defineEmits<{ close: [saved?: boolean] }>();
const { t } = useI18n();
const notify = useNotify();
const settings = useSettings();
const { form, ready } = settings;
const validation = ref<{ validate(): Promise<{ valid: boolean }> } | null>(
  null,
);
const saving = ref(false);
const loadError = ref("");
const modes = computed(() => subscriptionUpdateModes(t));
const proxyModes = computed(() =>
  updateProxyModes(t, form.transparent === "close"),
);

async function load() {
  loadError.value = "";
  try {
    await settings.load();
  } catch (err) {
    loadError.value = errorText(err);
    notify.warning(loadError.value);
  }
}
onMounted(load);
async function save() {
  if (!ready.value || saving.value) return;
  if (!(await validation.value?.validate())?.valid) return;
  saving.value = true;
  try {
    await settings.save();
    notify.success(t("setting.saved"));
    emit("close", true);
  } catch (err) {
    notify.warning(t("setting.saveFailed", { message: errorText(err) }));
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <v-card :loading="!ready && !loadError">
    <v-card-title class="md3-headline-small px-6 pt-6 pb-2 text-wrap">{{
      t("subscription.settingsTitle")
    }}</v-card-title>
    <v-card-text class="px-6">
      <v-alert v-if="loadError" type="error" variant="tonal" class="mb-4">
        {{ loadError }}
        <template #append
          ><v-btn variant="text" @click="load">{{
            t("operations.update")
          }}</v-btn></template
        >
      </v-alert>
      <v-form
        ref="validation"
        :disabled="!ready || saving"
        @submit.prevent="save"
      >
        <v-select
          v-model="form.subscriptionAutoUpdateMode"
          :label="t('setting.autoUpdateSub')"
          :items="modes"
        />
        <v-expand-transition>
          <v-text-field
            v-if="
              form.subscriptionAutoUpdateMode === 'auto_update_at_intervals'
            "
            v-model.number="form.subscriptionAutoUpdateIntervalHour"
            type="number"
            min="1"
            step="1"
            dir="ltr"
            :label="t('setting.options.updateSubAtIntervals')"
            :rules="[
              required,
              (v) =>
                (Number.isInteger(Number(v)) && Number(v) >= 1) ||
                t('configureServer.required'),
            ]"
          />
        </v-expand-transition>
        <v-select
          v-model="form.proxyModeWhenSubscribe"
          :label="t('setting.preferModeWhenUpdate')"
          :items="proxyModes"
        />
      </v-form>
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" :disabled="saving" @click="emit('close')">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn
        variant="flat"
        color="primary"
        :disabled="!ready"
        :loading="saving"
        @click="save"
        >{{ t("operations.save") }}</v-btn
      >
    </v-card-actions>
  </v-card>
</template>
