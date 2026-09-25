<script setup lang="ts">
// Subscription refresh policy; group membership is configured separately.
import { reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { patchSubscription } from "@/api";
import { errorText } from "@/api/errors";
import type { TouchSubscription } from "@/api/types";
import { useNotify } from "@/composables/useNotify";

defineOptions({ name: "SubscriptionDialog" });
const props = defineProps<{ subscription: TouchSubscription }>();
const emit = defineEmits<{ close: [saved?: boolean] }>();
const { t } = useI18n();
const notify = useNotify();

// the old page sent the row with its servers emptied
const form = reactive({
  autoUpdate: false,
  updateIntervalMinutes: 0,
  failureIntervalMinutes: 1,
  ...props.subscription,
  servers: [],
});
const saving = ref(false);
const validation = ref<{ validate(): Promise<{ valid: boolean }> } | null>(
  null,
);
const intervalRule = (minimum: number) => (v: unknown) =>
  (v !== "" &&
    v !== null &&
    Number.isInteger(Number(v)) &&
    Number(v) >= minimum &&
    Number(v) <= 525600) ||
  t("subscription.intervalInvalid", { minimum });

async function save() {
  if (saving.value) return;
  if (!(await validation.value?.validate())?.valid) return;
  saving.value = true;
  try {
    await patchSubscription({ subscription: form });
    notify.success(t("subscription.saved"));
    emit("close", true);
  } catch (err) {
    notify.warning(t("subscription.saveFailed", { message: errorText(err) }));
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <v-card>
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0">
        {{ t("configureSubscription.title") }}
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6">
      <v-form ref="validation" @submit.prevent="save">
        <v-textarea
          v-model="form.address"
          :label="t('subscription.subscription')"
          rows="2"
          auto-grow
          dir="ltr"
          @keydown.enter="
            (e: KeyboardEvent) => (e.ctrlKey || e.metaKey) && save()
          "
        />
        <v-text-field
          v-model="form.remarks"
          :label="t('subscription.remarks')"
        />
        <v-switch
          v-model="form.autoUpdate"
          :label="t('subscription.autoUpdate')"
          hide-details
        />
        <p class="md3-body-large mb-4">
          {{ t("subscription.autoUpdateHelp") }}
        </p>
        <template v-if="form.autoUpdate">
          <v-text-field
            v-model.number="form.updateIntervalMinutes"
            type="number"
            min="0"
            max="525600"
            step="1"
            :label="t('subscription.updateIntervalMinutes')"
            :rules="[intervalRule(0)]"
            hide-details="auto"
          />
          <p class="md3-body-small text-on-surface-variant mt-2 mb-4">
            {{ t("subscription.regularHelp") }}
          </p>
          <v-text-field
            v-model.number="form.failureIntervalMinutes"
            type="number"
            min="1"
            max="525600"
            step="1"
            :label="t('subscription.failureIntervalMinutes')"
            :rules="[intervalRule(1)]"
            hide-details="auto"
          />
          <p class="md3-body-small text-on-surface-variant mt-2">
            {{ t("subscription.failureHelp") }}
          </p>
        </template>
      </v-form>
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn variant="flat" color="primary" :loading="saving" @click="save">
        {{ t("operations.saveApply") }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>
