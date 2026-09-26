<script setup lang="ts">
// One outbound group's membership checks and connection strategy. Resolves
// true after a save.
import { computed, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { getOutbound, putOutbound, type OutboundSetting } from "@/api";
import { errorText } from "@/api/errors";
import { useNotify } from "@/composables";
import { required } from "@/dialogs/Server/forms/parts/rules";

defineOptions({ name: "OutboundGroupDialog" });
const props = defineProps<{ outbound: string }>();
const emit = defineEmits<{ close: [saved?: boolean] }>();
const { t } = useI18n();
const notify = useNotify();
const setting = reactive<OutboundSetting>({
  probeURL: "",
  probeInterval: "300s",
  autoAdd: false,
  type: "leastping",
});
const wasAutomatic = ref(false);
const form = ref<{ validate(): Promise<{ valid: boolean }> } | null>(null);
const saving = ref(false);
const strategyItems = computed(() => [
  { value: "leastping", title: t("outbound.strategies.leastPing") },
  { value: "keepcurrent", title: t("outbound.strategies.keepCurrent") },
  { value: "roundrobin", title: t("outbound.strategies.roundRobin") },
  { value: "random", title: t("outbound.strategies.random") },
]);
const strategyDetails = computed(() =>
  t(`outbound.strategyDetails.${setting.type}`),
);

onMounted(async () => {
  try {
    Object.assign(setting, (await getOutbound(props.outbound)).setting);
    wasAutomatic.value = !!setting.autoAdd;
  } catch (err) {
    notify.warning(errorText(err));
  }
});

async function save() {
  const check = await form.value?.validate();
  if (check && !check.valid) return;
  saving.value = true;
  try {
    await putOutbound({ outbound: props.outbound, setting: { ...setting } });
    notify.success(t("outbound.settingSaved"));
    emit("close", true);
  } catch (err) {
    notify.warning(
      t("outbound.settingSaveFailed", { message: errorText(err) }),
    );
  } finally {
    saving.value = false;
  }
}

function setAutomatic(enabled: boolean | null) {
  setting.autoAdd = !!enabled;
  if (enabled && !wasAutomatic.value && setting.probeInterval === "60s") {
    setting.probeInterval = "300s";
  }
}
</script>

<template>
  <v-card>
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0">
        {{ t("common.proxyGroups") }}
        <span class="text-on-surface-variant ms-2">{{
          outbound.toUpperCase()
        }}</span>
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6">
      <v-form ref="form" @submit.prevent="save">
        <v-switch
          :model-value="setting.autoAdd"
          :label="t('outbound.autoAdd')"
          hide-details
          @update:model-value="setAutomatic"
        />
        <p class="md3-body-large mb-2">
          {{ t("outbound.autoAddHelp") }}
        </p>
        <p class="md3-body-small text-on-surface-variant mb-4">
          {{ t("outbound.autoAddDetails") }}
        </p>
        <v-text-field
          v-model="setting.probeURL"
          :label="t('outbound.probeUrl')"
          :rules="[required]"
          dir="ltr"
        />
        <v-text-field
          v-model="setting.probeInterval"
          :label="t('outbound.probeInterval')"
          :rules="[required]"
          dir="ltr"
        />
        <v-select
          v-model="setting.type"
          :items="strategyItems"
          :label="t('outbound.strategy')"
        />
        <p class="md3-body-large mb-2">
          {{ t("outbound.strategyHelp") }}
        </p>
        <p class="md3-body-small text-on-surface-variant mb-0">
          {{ strategyDetails }}
        </p>
      </v-form>
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn variant="flat" color="primary" :loading="saving" @click="save">{{
        t("operations.save")
      }}</v-btn>
    </v-card-actions>
  </v-card>
</template>
