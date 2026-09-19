<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { mdiDeleteOutline, mdiEye, mdiEyeOff, mdiLockOutline } from "@mdi/js";
import {
  deleteCustomInbound,
  getCustomInbound,
  getOutbounds,
  postCustomInbound,
} from "@/api";
import type { CustomInbound } from "@/api/types";
import { errorText } from "@/api/errors";
import { useConfirm, useNotify } from "@/composables";

defineOptions({ name: "CustomInboundDialog" });
const emit = defineEmits<{ close: [] }>();
const { t } = useI18n();
const confirm = useConfirm();
const notify = useNotify();
const inbounds = ref<CustomInbound[]>([]);
const outbounds = ref<string[]>([]);
const loading = ref(true);
const adding = ref(false);
const deleting = ref<string | null>(null);
const loadError = ref("");
const showPassword = ref(false);
const busy = computed(
  () => loading.value || adding.value || deleting.value !== null,
);
const form = ref(emptyForm());
const protocols = [
  { value: "socks", title: "SOCKS" },
  { value: "http", title: "HTTP" },
];
const modes = computed(() => [
  { value: "direct", title: t("customInbound.outboundTypeDirect") },
  { value: "routingA", title: t("customInbound.outboundTypeRoutingA") },
]);
function emptyForm() {
  return {
    tag: "",
    protocol: "socks" as CustomInbound["protocol"],
    port: "",
    outbound: outbounds.value[0] ?? "",
    outboundType: "direct",
    routingARules: "",
    username: "",
    password: "",
  };
}

onMounted(async () => {
  try {
    const [rows, groups] = await Promise.all([
      getCustomInbound(),
      getOutbounds(),
    ]);
    inbounds.value = rows.inbounds ?? [];
    outbounds.value = groups.outbounds ?? [];
    form.value.outbound = outbounds.value[0] ?? "";
  } catch (err) {
    loadError.value = t("customInbound.saveFailed", {
      message: errorText(err),
    });
    notify.warning(loadError.value);
  } finally {
    loading.value = false;
  }
});

async function add(event: Event) {
  if (busy.value) return;
  if (!form.value.tag.trim() || !form.value.port) {
    notify.warning(t("customInbound.fillAll"));
    return;
  }
  if (!form.value.outbound) {
    notify.warning(t("customInbound.outboundRequired"));
    return;
  }
  if (!(event.target as HTMLFormElement).reportValidity()) return;
  adding.value = true;
  try {
    const res = (await postCustomInbound({
      tag: form.value.tag.trim(),
      protocol: form.value.protocol,
      port: Number(form.value.port),
      outbound: form.value.outbound,
      outboundType: form.value.outboundType,
      routingARules:
        form.value.outboundType === "routingA" ? form.value.routingARules : "",
      username: form.value.username.trim(),
      password: form.value.password,
    })) as { inbounds?: CustomInbound[] };
    inbounds.value = res.inbounds ?? [];
    form.value = emptyForm();
    showPassword.value = false;
  } catch (err) {
    notify.warning(t("customInbound.saveFailed", { message: errorText(err) }));
  } finally {
    adding.value = false;
  }
}

async function remove(tag: string) {
  if (busy.value) return;
  deleting.value = tag;
  try {
    if (
      !(await confirm({
        message: t("customInbound.deleteConfirm", { tag }),
        confirmText: t("operations.delete"),
        cancelText: t("operations.cancel"),
        destructive: true,
      }))
    )
      return;
    const res = (await deleteCustomInbound({ tag })) as {
      inbounds?: CustomInbound[];
    };
    inbounds.value = res.inbounds ?? [];
  } catch (err) {
    notify.warning(
      t("customInbound.deleteFailed", { message: errorText(err) }),
    );
  } finally {
    deleting.value = null;
  }
}
</script>

<template>
  <v-card tag="form" rounded="xl" novalidate @submit.prevent="add">
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0">
        {{ t("customInbound.title") }}
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6">
      <v-alert v-if="loadError" type="warning" variant="tonal" class="mb-4">
        {{ loadError }}
      </v-alert>
      <v-skeleton-loader v-if="loading" type="list-item-two-line@2" />
      <v-list
        v-else-if="inbounds.length"
        bg-color="surface-container-low"
        rounded="lg"
        class="mb-6 inbound-list"
      >
        <v-list-item
          v-for="item in inbounds"
          :key="item.tag"
          :title="item.tag"
          lines="two"
        >
          <template #subtitle>
            <span class="d-inline-flex flex-wrap align-center ga-2" dir="ltr">
              <v-chip size="small" variant="tonal">{{
                item.protocol.toUpperCase()
              }}</v-chip>
              <span class="md3-body-medium">{{ item.port }}</span>
              <v-icon
                v-if="item.username"
                :icon="mdiLockOutline"
                size="18"
                :aria-label="t('customInbound.authEnabled')"
                role="img"
              />
              <v-chip v-if="item.outbound" size="small" variant="tonal">{{
                item.outbound
              }}</v-chip>
              <v-chip
                v-if="item.outboundType === 'routingA'"
                size="small"
                variant="text"
                >RoutingA</v-chip
              >
            </span>
          </template>
          <template #append>
            <v-btn
              :icon="mdiDeleteOutline"
              variant="text"
              size="40"
              :aria-label="`${t('operations.delete')}: ${item.tag}`"
              :disabled="busy"
              :loading="deleting === item.tag"
              @click="remove(item.tag)"
            >
              <v-icon :icon="mdiDeleteOutline" size="20" />
              <v-tooltip activator="parent">{{
                t("operations.delete")
              }}</v-tooltip>
            </v-btn>
          </template>
        </v-list-item>
      </v-list>
      <v-empty-state
        v-else
        :text="t('customInbound.empty')"
        class="mb-4 inbound-list"
      />
      <v-card-subtitle class="md3-title-medium pa-0 mb-4">
        {{ t("customInbound.addNew") }}
      </v-card-subtitle>
      <v-row dense>
        <v-col cols="12" sm="6">
          <v-text-field
            v-model="form.tag"
            name="tag"
            :label="t('customInbound.tag')"
            :placeholder="t('customInbound.tagPlaceholder')"
            :disabled="busy"
            required
          />
        </v-col>
        <v-col cols="12" sm="6">
          <v-select
            v-model="form.protocol"
            :items="protocols"
            :label="t('customInbound.protocol')"
            :disabled="busy"
          />
        </v-col>
        <v-col cols="12" sm="6">
          <v-text-field
            v-model="form.port"
            name="port"
            :label="t('customInbound.port')"
            :placeholder="t('customInbound.portPlaceholder')"
            type="number"
            min="1"
            max="65535"
            :disabled="busy"
            required
            dir="ltr"
          />
        </v-col>
        <v-col cols="12" sm="6">
          <v-select
            v-model="form.outbound"
            :items="outbounds"
            :label="t('customInbound.outbound')"
            :placeholder="t('customInbound.outboundPlaceholder')"
            :disabled="busy"
          />
        </v-col>
        <v-col cols="12">
          <v-select
            v-model="form.outboundType"
            :items="modes"
            :label="t('customInbound.outboundType')"
            :disabled="busy"
          />
          <v-expand-transition>
            <v-textarea
              v-if="form.outboundType === 'routingA'"
              v-model="form.routingARules"
              name="routingARules"
              :label="t('customInbound.routingARules')"
              :placeholder="t('customInbound.routingARulesPlaceholder')"
              :disabled="busy"
              rows="6"
              dir="ltr"
            />
          </v-expand-transition>
        </v-col>
        <v-col cols="12" sm="6">
          <v-text-field
            v-model="form.username"
            name="username"
            :label="t('customInbound.username')"
            :placeholder="t('customInbound.authOptional')"
            :disabled="busy"
            autocomplete="off"
            dir="ltr"
          />
        </v-col>
        <v-col cols="12" sm="6">
          <v-text-field
            v-model="form.password"
            name="password"
            :label="t('customInbound.password')"
            :placeholder="t('customInbound.authOptional')"
            :type="showPassword ? 'text' : 'password'"
            :append-inner-icon="showPassword ? mdiEyeOff : mdiEye"
            :disabled="busy"
            autocomplete="off"
            dir="ltr"
            @click:append-inner="showPassword = !showPassword"
          />
        </v-col>
      </v-row>
      <v-alert
        type="info"
        variant="tonal"
        density="compact"
        class="md3-body-small"
      >
        {{ t("customInbound.hint") }}
      </v-alert>
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn
        variant="text"
        :disabled="adding || deleting !== null"
        @click="emit('close')"
      >
        {{ t("operations.close") }}
      </v-btn>
      <v-btn
        type="submit"
        color="primary"
        variant="flat"
        :loading="adding"
        :disabled="loading || deleting !== null"
      >
        {{ t("operations.add") }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>
