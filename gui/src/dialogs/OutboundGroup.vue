<script setup lang="ts">
// One outbound group's membership checks and connection strategy. Resolves
// true after a save.
import { computed, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { mdiHelpCircleOutline } from "@mdi/js";
import {
  getOutbound,
  postOutboundRefresh,
  putOutbound,
  type OutboundSetting,
} from "@/api";
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
  selected: "",
});
const wasAutomatic = ref(false);
const form = ref<{ validate(): Promise<{ valid: boolean }> } | null>(null);
const saving = ref(false);
const refreshing = ref(false);
const strategyHelpOpen = ref(false);
const strategyItems = computed(() => [
  { value: "fixed", title: t("outbound.strategies.fixed") },
  { value: "leastping", title: t("outbound.strategies.leastPing") },
  { value: "keepcurrent", title: t("outbound.strategies.keepCurrent") },
  { value: "roundrobin", title: t("outbound.strategies.roundRobin") },
  { value: "random", title: t("outbound.strategies.random") },
]);
const isFixed = computed(() => setting.type === "fixed");
const fixedReady = computed(() => !isFixed.value || !!setting.selected);
const strategyDetails = computed(() =>
  t(`outbound.strategyDetails.${setting.type}`),
);

function writableSetting(): OutboundSetting {
  return {
    autoAdd: setting.autoAdd,
    probeURL: setting.probeURL,
    probeInterval: setting.probeInterval,
    type: setting.type,
  };
}

onMounted(async () => {
  try {
    Object.assign(setting, (await getOutbound(props.outbound)).setting);
    wasAutomatic.value = !!setting.autoAdd;
  } catch (err) {
    notify.warning(errorText(err));
  }
});

async function save() {
  if (!fixedReady.value) {
    notify.warning(t("outbound.fixedRequiresServer"));
    return;
  }
  const check = await form.value?.validate();
  if (check && !check.valid) return;
  saving.value = true;
  try {
    await putOutbound({ outbound: props.outbound, setting: writableSetting() });
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

async function refreshMembers() {
  const check = await form.value?.validate();
  if (check && !check.valid) return;
  refreshing.value = true;
  try {
    await postOutboundRefresh(props.outbound);
    wasAutomatic.value = true;
    notify.success(t("outbound.membersUpdated"));
  } catch (err) {
    notify.warning(
      t("outbound.membersUpdateFailed", { message: errorText(err) }),
    );
  } finally {
    refreshing.value = false;
  }
}

function setAutomatic(enabled: boolean | null) {
  setting.autoAdd = !!enabled;
  if (enabled && !wasAutomatic.value && setting.probeInterval === "60s") {
    setting.probeInterval = "300s";
  }
}

function setStrategy(type: OutboundSetting["type"]) {
  setting.type = type;
  if (type === "fixed") setting.autoAdd = false;
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
        <div class="d-flex align-start ga-1">
          <v-select
            :model-value="setting.type"
            :items="strategyItems"
            :label="t('outbound.strategy')"
            class="flex-grow-1"
            @update:model-value="setStrategy"
          />
          <v-btn
            :icon="mdiHelpCircleOutline"
            variant="text"
            class="mt-1"
            :aria-label="t('outbound.strategyHelpAction')"
            :aria-expanded="strategyHelpOpen"
            aria-controls="strategy-help"
            @click="strategyHelpOpen = !strategyHelpOpen"
          />
        </div>
        <v-expand-transition>
          <div
            v-show="strategyHelpOpen"
            id="strategy-help"
            class="strategy-help text-on-surface-variant mb-4"
          >
            <p class="md3-body-small mb-2">{{ t("outbound.strategyHelp") }}</p>
            <p class="md3-body-small mb-0">{{ strategyDetails }}</p>
          </div>
        </v-expand-transition>
        <v-alert
          v-if="isFixed"
          type="info"
          variant="tonal"
          density="compact"
          class="mb-4"
        >
          {{
            fixedReady
              ? t("outbound.fixedServerSelected")
              : t("outbound.fixedRequiresServer")
          }}
        </v-alert>
        <v-text-field
          v-model="setting.probeURL"
          :label="t('outbound.probeUrl')"
          :rules="[required]"
          :disabled="isFixed"
          dir="ltr"
        />
        <v-text-field
          v-model="setting.probeInterval"
          :label="t('outbound.probeInterval')"
          :rules="[required]"
          :disabled="isFixed"
          dir="ltr"
        />
        <v-switch
          :model-value="setting.autoAdd"
          :label="t('outbound.autoAdd')"
          :disabled="isFixed"
          hide-details
          @update:model-value="setAutomatic"
        />
        <p class="md3-body-large mb-2" :class="{ 'text-disabled': isFixed }">
          {{ t("outbound.autoAddHelp") }}
        </p>
        <p
          class="md3-body-small text-on-surface-variant mb-4"
          :class="{ 'text-disabled': isFixed }"
        >
          {{ t("outbound.autoAddDetails") }}
        </p>
        <v-btn
          block
          variant="outlined"
          class="mb-2"
          :disabled="isFixed || !setting.autoAdd || !wasAutomatic"
          :loading="refreshing"
          @click="refreshMembers"
        >
          {{ t("outbound.updateMembers") }}
        </v-btn>
      </v-form>
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn
        variant="flat"
        color="primary"
        :loading="saving"
        :disabled="!fixedReady"
        @click="save"
        >{{ t("operations.save") }}</v-btn
      >
    </v-card-actions>
  </v-card>
</template>
