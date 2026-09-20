<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { mdiPlus } from "@mdi/js";
import { getPorts, putPorts } from "@/api";
import { probe } from "@/api/client";
import { errorText } from "@/api/errors";
import { useDialog, useNotify } from "@/composables";
import DocsLink from "@/components/DocsLink.vue";
import { resetSession } from "@/session";
import { useAppStore } from "@/stores/app";
import SharingDialog from "@/dialogs/Sharing.vue";
import CustomInboundDialog from "./CustomInbound.vue";

defineOptions({ name: "PortsDialog" });
const emit = defineEmits<{ close: [saved?: boolean] }>();
const { t } = useI18n();
const store = useAppStore();
const { open } = useDialog();
const notify = useNotify();
const address = ref(store.backendAddress);
const ports = reactive({
  socks5: "20170",
  http: "20171",
  socks5WithPac: "0",
  httpWithPac: "20172",
  vmess: "0",
  api: "0",
});
const services = ref<string[]>(["LoggerService"]);
const supportedServices = ["HandlerService", "LoggerService", "StatsService"];
const vmessLink = ref("");
const loading = ref(true);
const backendReady = ref(false);
const saving = ref(false);
const loadError = ref("");
const fields = [
  {
    key: "socks5",
    label: "customAddressPort.portSocks5",
    placeholder: "20170",
  },
  { key: "http", label: "customAddressPort.portHttp", placeholder: "20171" },
  {
    key: "socks5WithPac",
    label: "customAddressPort.portSocks5WithPac",
    placeholder: "0",
  },
  {
    key: "httpWithPac",
    label: "customAddressPort.portHttpWithPac",
    placeholder: "20172",
  },
  { key: "vmess", label: "customAddressPort.portVmess", placeholder: "0" },
  { key: "api", label: "customAddressPort.portApi", placeholder: "0" },
] as const;
const normalizedAddress = computed(() => address.value.replace(/\/$/, ""));
const addressChanged = computed(
  () => normalizedAddress.value !== store.backendAddress.replace(/\/$/, ""),
);
const messages = computed(() =>
  [
    ...(store.docker
      ? [t("customAddressPort.messages.1"), t("customAddressPort.messages.2")]
      : [t("customAddressPort.messages.0")]),
    t("customAddressPort.messages.3"),
  ].map((message) => message.replace(/<[^>]*>/g, "")),
);

onMounted(async () => {
  try {
    const res = await getPorts();
    for (const { key } of fields) {
      ports[key] = String(key === "api" ? res.api.port : res[key]);
    }
    services.value = res.api.services ?? [];
    vmessLink.value = res.vmessLink ?? "";
    backendReady.value = true;
  } catch (err) {
    loadError.value = t("customAddressPort.saveFailed", {
      message: errorText(err),
    });
    notify.warning(loadError.value);
  } finally {
    loading.value = false;
  }
});

function share(link: string) {
  if (!link) {
    notify.warning(t("customAddressPort.noVmessLink"));
    return;
  }
  open(
    SharingDialog,
    {
      title: t("customAddressPort.portVmessLink"),
      link,
      name: "VMess | v2rayA",
      type: "server",
    },
    { width: 420 },
  );
}

async function save(event: Event) {
  if (
    saving.value ||
    loading.value ||
    !(event.target as HTMLFormElement).reportValidity()
  )
    return;
  saving.value = true;
  const nextAddress = normalizedAddress.value;
  try {
    if (backendReady.value && !addressChanged.value) {
      const res = await putPorts({
        socks5: parseInt(ports.socks5),
        http: parseInt(ports.http),
        socks5WithPac: parseInt(ports.socks5WithPac),
        httpWithPac: parseInt(ports.httpWithPac),
        vmess: parseInt(ports.vmess),
        api: { port: parseInt(ports.api), services: services.value },
      });
      store.setBackendAddress(nextAddress);
      emit("close", true);
      if (
        res &&
        typeof res === "object" &&
        "vmessLink" in res &&
        typeof res.vmessLink === "string" &&
        res.vmessLink
      ) {
        share(res.vmessLink);
      }
    } else {
      await probe(nextAddress);
      store.setBackendAddress(nextAddress);
      emit("close", true);
      await resetSession({ backendAddress: nextAddress });
    }
  } catch (err) {
    notify.warning(
      t("customAddressPort.saveFailed", { message: errorText(err) }),
    );
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <v-card tag="form" rounded="xl" @submit.prevent="save">
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0 d-flex align-center ga-1">
        {{ t("customAddressPort.title") }}
        <DocsLink section="inbounds" new-tab />
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6">
      <v-text-field
        v-model="address"
        name="backendAddress"
        :label="t('customAddressPort.serviceAddress')"
        placeholder="http://localhost:2017"
        pattern="(https?://.+|/.+)"
        :disabled="saving"
        dir="ltr"
      />
      <v-skeleton-loader v-if="loading" type="list-item-two-line@3" />
      <v-alert
        v-else-if="loadError"
        type="warning"
        variant="tonal"
        class="mb-4"
      >
        {{ loadError }}
      </v-alert>
      <v-expand-transition>
        <v-sheet v-if="backendReady && !addressChanged" color="transparent">
          <v-row dense>
            <v-col v-for="field in fields" :key="field.key" cols="12" sm="6">
              <v-text-field
                v-model="ports[field.key]"
                :name="field.key"
                :label="t(field.label)"
                :placeholder="field.placeholder"
                type="number"
                min="0"
                required
                :disabled="saving"
                dir="ltr"
              />
            </v-col>
            <v-col cols="12">
              <v-select
                v-model="services"
                :items="supportedServices"
                :label="t('customAddressPort.apiServices')"
                multiple
                chips
                closable-chips
                :disabled="saving"
                dir="ltr"
              />
            </v-col>
          </v-row>
          <v-btn
            v-if="Number(ports.vmess) > 0 && vmessLink"
            variant="text"
            class="mb-4"
            :disabled="saving"
            @click="share(vmessLink)"
          >
            {{ t("customAddressPort.portVmessLink") }}
          </v-btn>
          <v-alert
            v-for="message in messages"
            :key="message"
            type="info"
            variant="tonal"
            density="compact"
            class="md3-body-small mb-2"
          >
            {{ message }}
          </v-alert>
        </v-sheet>
      </v-expand-transition>
    </v-card-text>
    <v-card-actions class="px-6 pb-4 flex-wrap ga-2">
      <v-btn
        variant="tonal"
        :prepend-icon="mdiPlus"
        :disabled="saving"
        @click="open(CustomInboundDialog, {}, { width: 720 })"
      >
        {{ t("customInbound.title") }}
      </v-btn>
      <v-spacer />
      <v-btn variant="text" :disabled="saving" @click="emit('close')">
        {{ t("operations.cancel") }}
      </v-btn>
      <v-btn
        type="submit"
        color="primary"
        variant="flat"
        :loading="saving"
        :disabled="loading"
      >
        {{ t("operations.confirm") }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>
