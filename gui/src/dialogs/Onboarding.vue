<script lang="ts">
export function shouldShowOnboarding(): boolean {
  return localStorage.getItem("onboardingSeen") === null;
}
</script>

<script setup lang="ts">
import { onBeforeUnmount, ref } from "vue";
import { useI18n } from "vue-i18n";
import {
  mdiClose,
  mdiPower,
  mdiRoutes,
  mdiServerNetwork,
  mdiTrayArrowDown,
} from "@mdi/js";
import { useDialog } from "@/composables/useDialog";
import ImportDialog from "@/dialogs/Import.vue";
import ServerDialog from "@/dialogs/Server/index.vue";
import RoutingADialog from "@/dialogs/settings/RoutingA.vue";
import { useAppStore } from "@/stores/app";

defineOptions({ name: "OnboardingDialog" });
const emit = defineEmits<{ close: [finished?: boolean] }>();
const { t } = useI18n();
const { open } = useDialog();
const store = useAppStore();
const step = ref(0);
const steps = [
  { key: "import", icon: mdiTrayArrowDown },
  { key: "group", icon: mdiServerNetwork },
  { key: "rules", icon: mdiRoutes },
  { key: "start", icon: mdiPower },
] as const;

function markSeen() {
  localStorage.setItem("onboardingSeen", "1");
}

function close(finished = false) {
  if (finished) store.view = "dashboard";
  markSeen();
  emit("close", finished);
}

// Escape and scrim dismissal remove the component through DialogHost.
onBeforeUnmount(markSeen);
</script>

<template>
  <v-card rounded="xl">
    <v-card-item class="px-6 pt-6 pb-4">
      <v-card-title class="md3-headline-small text-wrap pa-0">
        {{ t("onboarding.title") }}
      </v-card-title>
      <template #append>
        <v-tooltip :text="t('operations.close')">
          <template #activator="{ props }">
            <v-btn
              v-bind="props"
              :icon="mdiClose"
              :aria-label="t('operations.close')"
              variant="text"
              size="48"
              @click="close()"
            />
          </template>
        </v-tooltip>
      </template>
    </v-card-item>
    <v-card-text class="px-6 pb-6">
      <v-window v-model="step" :touch="false">
        <v-window-item
          v-for="(item, index) in steps"
          :key="item.key"
          :value="index"
        >
          <v-avatar color="primary-container" size="48" class="mb-4">
            <v-icon :icon="item.icon" class="text-on-primary-container" />
          </v-avatar>
          <h2 class="md3-headline-small mb-4">
            {{ t(`onboarding.${item.key}Title`) }}
          </h2>
          <p class="md3-body-medium text-on-surface-variant ma-0">
            {{ t(`onboarding.${item.key}Body`) }}
          </p>
          <div v-if="index === 0" class="d-flex flex-wrap ga-2 mt-6">
            <v-btn
              variant="tonal"
              color="primary"
              @click="open(ImportDialog, {}, { width: 480 })"
            >
              {{ t("operations.import") }}
            </v-btn>
            <v-btn
              variant="text"
              @click="open(ServerDialog, { which: null }, { width: 560 })"
            >
              {{ t("onboarding.newNode") }}
            </v-btn>
          </div>
          <v-btn
            v-else-if="index === 1"
            variant="text"
            class="mt-6"
            @click="store.view = 'proxies'"
          >
            {{ t("onboarding.goToProxies") }}
          </v-btn>
          <v-btn
            v-else-if="index === 2"
            variant="tonal"
            class="mt-6"
            @click="open(RoutingADialog, {}, { width: 960 })"
          >
            {{ t("routingA.editor") }}
          </v-btn>
        </v-window-item>
      </v-window>
    </v-card-text>
    <div
      class="d-flex justify-center ga-2 px-6 pb-4"
      role="status"
      :aria-label="
        t('onboarding.progress', { current: step + 1, total: steps.length })
      "
    >
      <span
        v-for="(_, index) in steps"
        :key="index"
        class="onboarding-dot rounded-pill"
        :class="{ 'onboarding-dot--active': step === index }"
        :aria-current="step === index ? 'step' : undefined"
        aria-hidden="true"
      />
    </div>
    <v-card-actions class="px-6 pb-6 ga-2">
      <v-btn variant="text" :disabled="step === 0" @click="step--">
        {{ t("onboarding.back") }}
      </v-btn>
      <v-spacer />
      <v-btn
        variant="flat"
        color="primary"
        @click="step === steps.length - 1 ? close(true) : step++"
      >
        {{
          t(step === steps.length - 1 ? "onboarding.finish" : "onboarding.next")
        }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<style scoped>
.onboarding-dot {
  width: 8px;
  height: 8px;
  background-color: rgb(var(--v-theme-outline-variant));
}
.onboarding-dot--active {
  width: 24px;
  background-color: rgb(var(--v-theme-primary));
}
</style>
