<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { mdiDeleteOutline, mdiPlus } from "@mdi/js";
import {
  getDnsRules,
  getOutbounds,
  getSetting,
  postNodeDnsOptions,
} from "@/api";
import type { DnsMode, DnsRule, NodeDnsOption } from "@/api/types";
import { errorText } from "@/api/errors";
import NodeDnsChoice from "./NodeDnsChoice.vue";
import DocsLink from "@/components/DocsLink.vue";
import { useNotify } from "@/composables";
import {
  cloneDnsRules,
  defaultDnsRules,
  dnsModeHelp,
  dnsModeItems,
  dnsRulesPayload,
  normalizeDnsRules,
  resolveDnsMode,
  saveDnsSettings,
} from "./dnsModel";

defineOptions({ name: "DnsDialog" });
const emit = defineEmits<{ close: [changed?: boolean] }>();
const { t } = useI18n();
const notify = useNotify();
const mode = ref<DnsMode>("hijack");
/** what the service holds before this edit */
const storedMode = ref<DnsMode>("hijack");
const rules = ref<DnsRule[]>(cloneDnsRules(defaultDnsRules));
/** what the service holds, as the dialog folds it into the form */
const storedRules = ref<DnsRule[]>(cloneDnsRules(defaultDnsRules));
const nodeDns = ref("auto");
const storedNodeDns = ref("auto");
const nodeDnsOptions = ref<NodeDnsOption[]>([]);
const nodeDnsWarnings = ref<string[]>([]);
const nodeDnsError = ref<unknown>(null);
const nodeDnsLoading = ref(false);
const nodeDnsLoaded = ref(false);
const nodeDnsFallbackFrom = computed(() =>
  nodeDnsLoaded.value &&
  !nodeDnsError.value &&
  nodeDns.value !== "auto" &&
  !nodeDnsOptions.value.some((option) => option.value === nodeDns.value) &&
  nodeDnsOptions.value.some((option) => option.value === "auto")
    ? nodeDns.value
    : "",
);
// Keep the explicit edit until the user saves or chooses another source.
const displayedNodeDns = computed({
  get: () => (nodeDnsFallbackFrom.value ? "auto" : nodeDns.value),
  set: (value: string) => {
    nodeDns.value = value;
  },
});
let nodeDnsRequest = 0;
let previewTimer: ReturnType<typeof setTimeout> | undefined;
const outbounds = ref(["proxy"]);
const choices = computed(() => [...new Set(["direct", ...outbounds.value])]);
const modeItems = computed(() => dnsModeItems(t));
const modeTexts = computed(() => dnsModeHelp(t));
const loading = ref(true);
const loaded = ref(false);
const saving = ref(false);
const loadError = ref("");

const modeOff = computed(() => mode.value === "off");
const rulesChanged = computed(
  () => JSON.stringify(rules.value) !== JSON.stringify(storedRules.value),
);
// An off mode keeps the rules rather than writing them, so what it hides is
// never part of what this dialog owes the service.
const dirty = computed(
  () =>
    mode.value !== storedMode.value ||
    (!modeOff.value &&
      (rulesChanged.value || displayedNodeDns.value !== storedNodeDns.value)),
);

onMounted(async () => {
  try {
    const [dns, groups, setting] = await Promise.all([
      getDnsRules(),
      getOutbounds(),
      getSetting(),
    ]);
    storedRules.value = normalizeDnsRules(dns.rules);
    rules.value = cloneDnsRules(storedRules.value);
    mode.value = storedMode.value = resolveDnsMode(setting.setting);
    nodeDns.value = storedNodeDns.value = setting.setting.nodeDns || "auto";
    outbounds.value = groups.outbounds;
    loaded.value = true;
  } catch (err) {
    loadError.value = errorText(err);
    notify.warning(loadError.value);
  } finally {
    loading.value = false;
  }
});

function resetDefault() {
  rules.value = cloneDnsRules(defaultDnsRules);
}

async function loadNodeDnsOptions() {
  clearTimeout(previewTimer);
  const request = ++nodeDnsRequest;
  nodeDnsLoading.value = true;
  nodeDnsError.value = null;
  try {
    const result = await postNodeDnsOptions({
      rules: dnsRulesPayload(rules.value),
    });
    if (request !== nodeDnsRequest) return;
    nodeDnsOptions.value = result.options;
    nodeDnsWarnings.value = result.warnings ?? [];
    nodeDnsLoaded.value = true;
  } catch (err) {
    if (request === nodeDnsRequest) nodeDnsError.value = err;
  } finally {
    if (request === nodeDnsRequest) nodeDnsLoading.value = false;
  }
}

watch(
  () => [
    loaded.value,
    modeOff.value,
    JSON.stringify(dnsRulesPayload(rules.value)),
  ],
  () => {
    clearTimeout(previewTimer);
    ++nodeDnsRequest;
    if (!loaded.value || modeOff.value) return;
    nodeDnsLoading.value = true;
    // Avoid a request for every keystroke, while excluding stale responses
    // as soon as the edited rules change.
    if (!nodeDnsLoaded.value) void loadNodeDnsOptions();
    else previewTimer = setTimeout(() => void loadNodeDnsOptions(), 250);
  },
  { flush: "sync" },
);
onBeforeUnmount(() => {
  clearTimeout(previewTimer);
  ++nodeDnsRequest;
});

async function save() {
  if (!loaded.value || saving.value || !dirty.value) return;
  if (
    !modeOff.value &&
    rulesChanged.value &&
    !dnsRulesPayload(rules.value).length
  ) {
    notify.warning(t("dns.errNoRules"));
    return;
  }
  saving.value = true;
  try {
    if (!modeOff.value) {
      await loadNodeDnsOptions();
      if (nodeDnsError.value) {
        notify.warning(errorText(nodeDnsError.value));
        return;
      }
      if (
        !nodeDnsOptions.value.some(
          (option) => option.value === displayedNodeDns.value,
        )
      ) {
        notify.warning(t("nodeDns.unavailable"));
        return;
      }
    }
    const result = await saveDnsSettings({
      mode: mode.value,
      storedMode: storedMode.value,
      rules: rules.value,
      rulesDirty: rulesChanged.value,
      nodeDns: displayedNodeDns.value,
      storedNodeDns: storedNodeDns.value,
    });
    if (result.status === "saved") {
      notify.success(t("dns.saved"));
      emit("close", true);
      return;
    }
    notify.warning(t("dns.saveFailed", { message: result.error }));
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <v-card tag="form" rounded="xl" @submit.prevent="save">
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0 d-flex align-center ga-1">
        {{ t("dns.title") }}
        <DocsLink section="transparent-proxy" new-tab />
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6">
      <v-skeleton-loader v-if="loading" type="article" />
      <v-alert v-else-if="loadError" type="warning" variant="tonal">
        {{ loadError }}
      </v-alert>
      <template v-else>
        <v-sheet color="surface-container-low" rounded="lg" class="pa-4 mb-3">
          <v-select
            v-model="mode"
            :items="modeItems"
            item-title="title"
            item-value="value"
            :label="t('dns.mode')"
            :disabled="saving"
            hide-details="auto"
          />
          <p class="md3-body-small text-on-surface-variant mt-2 mb-0">
            {{ modeTexts[mode] }}
          </p>
        </v-sheet>
        <v-expand-transition>
          <div v-if="modeOff">
            <v-sheet
              color="surface-container-low"
              rounded="lg"
              class="pa-4 mb-3 md3-body-small text-on-surface-variant"
            >
              {{ t("dns.offKeepsRules", { n: storedRules.length }) }}
            </v-sheet>
          </div>
        </v-expand-transition>
        <v-expand-transition>
          <div v-if="!modeOff">
            <v-sheet
              v-for="(rule, index) in rules"
              :key="index"
              color="surface-container-low"
              rounded="lg"
              class="pa-4 mb-3"
            >
              <div class="d-flex align-center mb-3">
                <span class="md3-title-small">
                  {{ t("dns.rule", { n: index + 1 }) }}
                </span>
                <v-spacer />
                <v-tooltip :text="t('operations.delete')">
                  <template #activator="{ props }">
                    <v-btn
                      v-bind="props"
                      icon
                      variant="text"
                      type="button"
                      size="40"
                      :aria-label="t('operations.delete')"
                      :disabled="saving"
                      @click="rules.splice(index, 1)"
                    >
                      <v-icon :icon="mdiDeleteOutline" size="20" />
                    </v-btn>
                  </template>
                </v-tooltip>
              </div>
              <v-row dense>
                <v-col cols="12" sm="6">
                  <v-text-field
                    v-model="rule.server"
                    :label="t('dns.colServer')"
                    :placeholder="t('dns.serverPlaceholder')"
                    :disabled="saving"
                    hide-details="auto"
                    dir="ltr"
                  />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-select
                    v-model="rule.outbound"
                    :label="t('dns.colOutbound')"
                    :items="choices"
                    :disabled="saving"
                    hide-details="auto"
                    dir="ltr"
                  />
                </v-col>
                <v-col cols="12">
                  <v-textarea
                    v-model="rule.domains"
                    :label="t('dns.colDomains')"
                    :placeholder="t('dns.domainsPlaceholder')"
                    :disabled="saving"
                    rows="2"
                    auto-grow
                    hide-details="auto"
                    dir="ltr"
                  />
                </v-col>
              </v-row>
            </v-sheet>
            <div class="d-flex flex-wrap ga-2">
              <v-btn
                variant="tonal"
                type="button"
                :prepend-icon="mdiPlus"
                :disabled="saving"
                @click="
                  rules.push({ server: '', domains: '', outbound: 'direct' })
                "
              >
                {{ t("dns.addRule") }}
              </v-btn>
              <v-btn
                variant="text"
                type="button"
                :disabled="saving"
                @click="resetDefault"
              >
                {{ t("dns.resetDefault") }}
              </v-btn>
            </div>
            <v-sheet color="surface-container-low" rounded="lg" class="mt-4">
              <NodeDnsChoice
                v-model="displayedNodeDns"
                :options="nodeDnsOptions"
                :warnings="nodeDnsWarnings"
                :error="nodeDnsError"
                :loading="nodeDnsLoading"
                :loaded="nodeDnsLoaded"
                :disabled="saving"
                :fallback-from="nodeDnsFallbackFrom"
                @refresh="loadNodeDnsOptions"
              />
            </v-sheet>
          </div>
        </v-expand-transition>
      </template>
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" type="button" @click="emit('close')">
        {{ t("operations.cancel") }}
      </v-btn>
      <v-btn
        variant="flat"
        color="primary"
        type="submit"
        :loading="saving"
        :disabled="!loaded || !dirty"
      >
        {{ t("operations.save") }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>
