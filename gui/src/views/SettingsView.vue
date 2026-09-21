<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useDisplay } from "vuetify";
import dayjs from "dayjs";
import { errorText } from "@/api/errors";
import { useDialog, useNotify, useUnsavedGuard } from "@/composables";
import DocsLink from "@/components/DocsLink.vue";
import CustomInboundDialog from "@/dialogs/settings/CustomInbound.vue";
import DnsDialog from "@/dialogs/settings/Dns.vue";
import DomainsExcludedDialog from "@/dialogs/settings/DomainsExcluded.vue";
import GfwListDialog from "@/dialogs/settings/GfwList.vue";
import PortsDialog from "@/dialogs/settings/Ports.vue";
import RoutingADialog from "@/dialogs/settings/RoutingA.vue";
import TproxyWhiteIpsDialog from "@/dialogs/settings/TproxyWhiteIps.vue";
import TunProcessesDialog from "@/dialogs/settings/TunProcesses.vue";
import TunRouteScriptDialog, {
  type TunRouteScript,
} from "@/dialogs/settings/TunRouteScript.vue";
import { useAppStore } from "@/stores/app";
import { useSettings } from "./settings/model";
import {
  pacModes as pacModeItems,
  transparentModes as transparentModeItems,
  transparentTypes as transparentTypeItems,
} from "./settings/options";
import AboutDialog from "./settings/AboutDialog.vue";
import SettingChoice from "./settings/SettingChoice.vue";
import SettingRow from "./settings/SettingRow.vue";

defineOptions({ name: "SettingsView" });
const { t } = useI18n();
const { width } = useDisplay();
const compact = computed(() => width.value < 600);
const notify = useNotify();
const { open } = useDialog();
const store = useAppStore();
const settings = useSettings();
useUnsavedGuard(() => settings.dirty.value);
const {
  form,
  ready,
  localGFWListVersion,
  localGeositeVersion,
  remoteGFWListVersion,
} = settings;
const saving = ref(false);
const formRef = ref<{ validate(): Promise<{ valid: boolean }> } | null>(null);

onMounted(() => {
  settings.load().catch((err) => notify.warning(errorText(err)));
  settings.loadRemoteVersion().catch(() => {});
});

// ---- the transparent proxy's implementation ---------------------------------------
// The mode itself, LAN sharing and IP forwarding are the dashboard's tiles.

const os = computed(() => store.version?.os ?? "");
const isRoot = computed(() => store.version?.isRoot ?? false);
const tunSupported = computed(() => store.version?.tunSupported ?? false);
const transparentModes = computed(() => transparentModeItems(t));
const pacModes = computed(() => pacModeItems(t));
const transparentTypes = computed(() =>
  transparentTypeItems(t, {
    lite: store.lite,
    os: os.value,
    isRoot: isRoot.value,
    tunSupported: tunSupported.value,
  }),
);
const transparentOn = computed(() => form.transparent !== "close");
const usesTproxy = computed(
  () =>
    transparentOn.value &&
    ["tproxy", "redirect"].includes(form.transparentType),
);
const usesTun = computed(
  () =>
    transparentOn.value && form.transparentType === "tun" && tunSupported.value,
);
async function openTunProcesses() {
  const value = await open<string>(
    TunProcessesDialog,
    { value: form.tunExcludeProcesses },
    { width: 520 },
  ).result;
  if (value !== undefined) form.tunExcludeProcesses = value;
}
async function openTunScript() {
  const value = await open<TunRouteScript>(
    TunRouteScriptDialog,
    {
      os: os.value,
      value: {
        shellType: form.tunRouteShellType,
        shellPath: form.tunRouteShellPath,
        setupScript: form.tunSetupScript,
        teardownScript: form.tunTeardownScript,
      },
    },
    { width: 640 },
  ).result;
  if (!value) return;
  form.tunRouteShellType = value.shellType;
  form.tunRouteShellPath = value.shellPath;
  form.tunSetupScript = value.setupScript;
  form.tunTeardownScript = value.teardownScript;
}
const openWhiteIps = () => open(TproxyWhiteIpsDialog, {}, { width: 520 });

// ---- the choices -------------------------------------------------------------

const onOffDefault = computed(() => [
  { value: "default", title: t("setting.options.default") },
  { value: "yes", title: t("setting.options.on") },
  { value: "no", title: t("setting.options.off") },
]);
const logLevels = computed(() =>
  ["trace", "debug", "info", "warn", "error"].map((v) => ({
    value: v,
    title: t(`setting.options.${v}`),
  })),
);
const sniffing = computed(() => [
  { value: "disable", title: t("setting.options.off") },
  { value: "http,tls", title: "HTTP + TLS" },
  { value: "http,tls,quic", title: "HTTP + TLS + QUIC" },
]);
const gfwUpdateModes = computed(() => [
  { value: "none", title: t("setting.options.off") },
  { value: "auto_update", title: t("setting.options.updateGfwlistWhenStart") },
  {
    value: "auto_update_at_intervals",
    title: t("setting.options.updateGfwlistAtIntervals"),
  },
]);

const gfwlistInUse = computed(
  () => form.pacMode === "gfwlist" || form.transparent === "gfwlist",
);
const localVersionAhead = computed(
  () =>
    !!localGFWListVersion.value &&
    !!remoteGFWListVersion.value &&
    dayjs(localGFWListVersion.value).isAfter(dayjs(remoteGFWListVersion.value)),
);
const localVersionDisplay = computed(() => {
  if (localGFWListVersion.value) return localGFWListVersion.value;
  if (localGeositeVersion.value)
    return t("gfwList.geosite", { date: localGeositeVersion.value });
  return t("common.none");
});
const positive = (v: unknown) =>
  Number(v) >= 1 || t("configureServer.required");

// ---- the dialogs -----------------------------------------------------------------

async function openGfwList() {
  const changed = await open<boolean>(
    GfwListDialog,
    { localVersion: localGFWListVersion.value },
    { width: 520 },
  ).result;
  if (changed) settings.load().catch(() => {});
}
const openDomains = () => open(DomainsExcludedDialog, {}, { width: 520 });
const openRoutingA = () => open(RoutingADialog, {}, { width: 960 });
const openDns = () => open(DnsDialog, {}, { width: 640 });
const openPorts = () => open(PortsDialog, {}, { width: 520 });
const openInbounds = () => open(CustomInboundDialog, {}, { width: 640 });
const openAbout = () => open(AboutDialog, {}, { width: 840 });

async function save() {
  const check = await formRef.value?.validate();
  if (check && !check.valid) return;
  saving.value = true;
  try {
    await settings.save();
    notify.success(t("setting.saved"));
  } catch (err) {
    notify.warning(t("setting.saveFailed", { message: errorText(err) }));
  } finally {
    saving.value = false;
  }
}
defineExpose({ sync: () => settings.load() });
</script>

<template>
  <v-form ref="formRef" class="settings" @submit.prevent="save">
    <v-skeleton-loader
      v-if="!ready"
      type="list-item-two-line@6"
      class="bg-transparent"
    />
    <template v-else>
      <v-list class="mb-4" bg-color="surface-container-low" rounded="xl">
        <v-list-subheader>
          {{ t("setting.sections.proxy") }}
          <DocsLink section="transparent-proxy" />
        </v-list-subheader>
        <SettingChoice
          v-model="form.transparent"
          :title="t('setting.transparentProxy')"
          :hint="t('setting.messages.transparentProxy')"
          :items="transparentModes"
        />
        <v-expand-transition>
          <SettingChoice
            v-if="transparentOn"
            v-model="form.transparentType"
            :title="t('setting.transparentType')"
            :items="transparentTypes"
          />
        </v-expand-transition>
        <v-expand-transition>
          <SettingRow
            v-if="usesTproxy"
            :title="t('setting.tproxyExcludedInterfaces')"
            :hint="t('setting.messages.tproxyExcludedInterfaces')"
          >
            <v-text-field
              v-model="form.tproxyExcludedInterfaces"
              class="settings__interfaces"
              :aria-label="t('setting.tproxyExcludedInterfaces')"
              :placeholder="t('setting.tproxyExcludedInterfacesPlaceholder')"
              hide-details="auto"
              density="compact"
              dir="ltr"
            />
          </SettingRow>
        </v-expand-transition>
        <v-expand-transition>
          <SettingRow
            v-if="transparentOn && form.transparentType === 'tproxy'"
            :title="t('operations.tproxyWhiteIpGroups')"
            :hint="t('tproxyWhiteIpGroups.messages.0')"
            action
            @click="openWhiteIps"
          />
        </v-expand-transition>
        <v-expand-transition>
          <div v-if="usesTun">
            <SettingRow
              :title="t('setting.tunAutoRoute')"
              :hint="t('setting.messages.tunAutoRoute')"
            >
              <v-switch
                v-model="form.tunAutoRoute"
                :aria-label="t('setting.tunAutoRoute')"
                hide-details
              />
            </SettingRow>
            <v-expand-transition>
              <SettingRow
                v-if="!form.tunAutoRoute"
                :title="t('operations.configureTunRouteScript')"
                :subtitle="form.tunRouteShellPath"
                action
                @click="openTunScript"
              />
            </v-expand-transition>
            <SettingRow
              :title="t('setting.tunExcludeProcesses')"
              :hint="t('setting.messages.tunExcludeProcesses')"
              action
              @click="openTunProcesses"
            >
              <v-badge
                v-if="form.tunExcludeProcesses"
                :content="
                  form.tunExcludeProcesses.split(',').filter((p) => p).length
                "
                inline
                color="primary"
                class="me-2"
              />
            </SettingRow>
          </div>
        </v-expand-transition>
        <SettingRow v-if="!store.lite" :title="t('setting.ipForwardOn')">
          <v-switch
            v-model="form.ipforward"
            :aria-label="t('setting.ipForwardOn')"
            hide-details
          />
        </SettingRow>
        <SettingRow :title="t('setting.portSharingOn')">
          <v-switch
            v-model="form.portSharing"
            :aria-label="t('setting.portSharingOn')"
            hide-details
          />
        </SettingRow>
      </v-list>

      <v-list class="mb-4" bg-color="surface-container-low" rounded="xl">
        <v-list-subheader>
          {{ t("setting.sections.traffic") }}
          <DocsLink section="routing" />
        </v-list-subheader>
        <SettingChoice
          v-model="form.pacMode"
          :title="t('setting.pacMode')"
          :hint="t('setting.messages.pacMode')"
          :items="pacModes"
        />
        <SettingRow
          :title="t('routingA.title')"
          :subtitle="t('operations.configure')"
          action
          @click="openRoutingA"
        />
        <SettingRow
          :title="t('gfwList.title')"
          :hint="localVersionAhead ? t('setting.messages.gfwlist') : undefined"
          :subtitle="`${t('common.latest')}: ${remoteGFWListVersion || t('common.checkRunning')}  ${t('common.local')}: ${localVersionDisplay}`"
          action
          @click="openGfwList"
        />
        <v-expand-transition>
          <div v-if="gfwlistInUse">
            <SettingChoice
              v-model="form.pacAutoUpdateMode"
              :title="t('setting.autoUpdateGfwlist')"
              :items="gfwUpdateModes"
            />
            <v-expand-transition>
              <SettingRow
                v-if="form.pacAutoUpdateMode === 'auto_update_at_intervals'"
                :title="t('setting.options.updateGfwlistAtIntervals')"
              >
                <v-text-field
                  v-model="form.pacAutoUpdateIntervalHour"
                  class="settings__number"
                  type="number"
                  min="1"
                  :rules="[positive]"
                  :aria-label="t('setting.options.updateGfwlistAtIntervals')"
                  hide-details="auto"
                />
              </SettingRow>
            </v-expand-transition>
          </div>
        </v-expand-transition>
      </v-list>

      <v-list class="mb-4" bg-color="surface-container-low" rounded="xl">
        <v-list-subheader>{{ t("setting.sections.core") }}</v-list-subheader>
        <SettingChoice
          v-model="form.logLevel"
          :title="t('setting.logLevel')"
          :items="logLevels"
        />
        <SettingChoice
          v-model="form.tcpFastOpen"
          :title="t('setting.tcpFastOpen')"
          :hint="t('setting.messages.tcpFastOpen')"
          :items="onOffDefault"
        />
        <SettingChoice
          v-model="form.inboundSniffing"
          :title="t('setting.inboundSniffing')"
          :hint="t('setting.messages.inboundSniffing')"
          :items="sniffing"
        />
        <v-expand-transition>
          <div v-if="form.inboundSniffing !== 'disable'">
            <SettingRow title="RouteOnly">
              <v-switch
                v-model="form.routeOnly"
                aria-label="RouteOnly"
                hide-details
              />
            </SettingRow>
            <SettingRow
              :title="t('operations.domainsExcluded')"
              :hint="t('domainsExcluded.messages.0')"
              action
              @click="openDomains"
            />
          </div>
        </v-expand-transition>
        <SettingRow :title="t('setting.mux')" :hint="t('setting.messages.mux')">
          <v-switch
            :model-value="form.muxOn === 'yes'"
            :aria-label="t('setting.mux')"
            hide-details
            @update:model-value="(v) => (form.muxOn = v ? 'yes' : 'no')"
          />
        </SettingRow>
        <v-expand-transition>
          <SettingRow
            v-if="form.muxOn === 'yes'"
            :title="t('setting.concurrency')"
          >
            <v-text-field
              v-model="form.mux"
              class="settings__number"
              type="number"
              min="1"
              max="1024"
              :rules="[positive]"
              :aria-label="t('setting.concurrency')"
              hide-details="auto"
            />
          </SettingRow>
        </v-expand-transition>
      </v-list>

      <v-list bg-color="surface-container-low" rounded="xl">
        <v-list-subheader>{{ t("setting.sections.more") }}</v-list-subheader>
        <SettingRow
          :title="t('customAddressPort.title')"
          :subtitle="t('customAddressPort.serviceAddress')"
          action
          @click="openPorts"
        />
        <SettingRow
          :title="t('customInbound.title')"
          :hint="t('customInbound.hint')"
          action
          @click="openInbounds"
        />
        <SettingRow
          :title="t('dns.title')"
          :subtitle="t('dns.colServer')"
          action
          @click="openDns"
        />
        <SettingRow
          :title="t('common.about')"
          :subtitle="t('about.intro')"
          action
          @click="openAbout"
        />
      </v-list>
    </template>
    <v-sheet
      color="surface-container"
      class="settings__save pa-4"
      :class="{ 'settings__save--compact': compact }"
    >
      <div class="settings__save-content d-flex justify-end">
        <v-btn
          color="primary"
          variant="flat"
          :loading="saving"
          :disabled="!ready"
          type="submit"
        >
          {{ t("operations.saveApply") }}
        </v-btn>
      </div>
    </v-sheet>
  </v-form>
</template>

<style scoped>
.settings {
  padding-bottom: 88px;
}
.settings__number {
  width: 96px;
}
.settings__interfaces {
  width: 200px;
}
.settings__save {
  position: fixed;
  left: var(--v-layout-left, 0px);
  right: var(--v-layout-right, 0px);
  bottom: var(--v-layout-bottom, 0px);
  z-index: 2;
}
.settings__save--compact {
  bottom: max(80px, var(--v-layout-bottom, 0px));
}
.settings__save-content {
  max-width: 1040px;
  margin-inline: auto;
  padding-inline: 8px;
}
@media (max-width: 599px) {
  .settings__interfaces {
    width: 112px;
  }
}
</style>
