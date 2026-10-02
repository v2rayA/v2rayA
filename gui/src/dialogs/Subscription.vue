<script setup lang="ts">
// Subscription refresh policy.
import { computed, reactive, ref } from "vue";
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
  updateMode: "disabled" as NonNullable<TouchSubscription["updateMode"]>,
  updateIntervalMinutes: 0,
  failureIntervalMinutes: 1,
  allowDirectRecovery: false,
  ...props.subscription,
  servers: [],
});
const updateModes = computed(() => [
  { value: "disabled", title: t("subscription.updateModes.disabled") },
  { value: "on_start", title: t("subscription.updateModes.onStart") },
  { value: "at_interval", title: t("subscription.updateModes.interval") },
  {
    value: "interval_failsafe",
    title: t("subscription.updateModes.intervalFailsafe"),
  },
]);
const needsRegularInterval = computed(
  () =>
    form.updateMode === "at_interval" ||
    form.updateMode === "interval_failsafe",
);
const needsFailureInterval = computed(
  () => form.updateMode === "interval_failsafe",
);
const updateModeHelp = computed(() => {
  switch (form.updateMode) {
    case "on_start":
      return t("subscription.updateModeHelp.onStart");
    case "at_interval":
      return t("subscription.updateModeHelp.interval");
    case "interval_failsafe":
      return t("subscription.updateModeHelp.intervalFailsafe");
    default:
      return t("subscription.updateModeHelp.disabled");
  }
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
        <v-select
          v-model="form.updateMode"
          :items="updateModes"
          :label="t('subscription.updateMode')"
        />
        <p class="md3-body-large mb-4">
          {{ updateModeHelp }}
        </p>
        <div v-if="needsRegularInterval" style="padding-bottom: 28px">
          <v-text-field
            v-model.number="form.updateIntervalMinutes"
            type="number"
            min="1"
            max="525600"
            step="1"
            :label="t('subscription.updateIntervalMinutes')"
            :rules="[intervalRule(1)]"
            :hint="t('subscription.regularHelp')"
            persistent-hint
          />
        </div>
        <div v-if="needsFailureInterval" style="padding-bottom: 16px">
          <v-text-field
            v-model.number="form.failureIntervalMinutes"
            type="number"
            min="1"
            max="525600"
            step="1"
            :label="t('subscription.failureIntervalMinutes')"
            :rules="[intervalRule(1)]"
            :hint="t('subscription.failureHelp')"
            persistent-hint
          />
        </div>
        <div v-if="needsFailureInterval">
          <v-switch
            v-model="form.allowDirectRecovery"
            :label="t('subscription.allowDirectRecovery')"
            hide-details
          />
          <p class="md3-body-small text-medium-emphasis mb-4">
            {{ t("subscription.recoveryRouteHelp") }}
          </p>
        </div>
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
