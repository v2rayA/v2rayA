<script setup lang="ts">
import { computed } from "vue";
import { mdiRefresh } from "@mdi/js";
import { useI18n } from "vue-i18n";
import type { NodeDnsOption } from "@/api/types";
import { errorText } from "@/api/errors";

const model = defineModel<string>({ required: true });
const props = defineProps<{
  options: NodeDnsOption[];
  warnings: string[];
  error: unknown;
  loading: boolean;
  loaded: boolean;
  disabled: boolean;
}>();
defineEmits<{ refresh: [] }>();
const { t } = useI18n();
const categories = computed(() => ({
  auto: t("nodeDns.categories.auto"),
  direct: t("nodeDns.categories.direct"),
  localhost: t("nodeDns.categories.localhost"),
  proxy: t("nodeDns.categories.proxy"),
  fallback: t("nodeDns.categories.fallback"),
}));
const items = computed(() =>
  props.options.map((option) => ({
    ...option,
    address: option.url,
    categoryText: categories.value[option.category],
    title: `${option.url} (${categories.value[option.category]})`,
  })),
);
const selection = computed(
  () =>
    items.value.find((item) => item.value === model.value) ?? {
      address: model.value,
      url: model.value,
      categoryText: "",
    },
);
const unavailable = computed(
  () =>
    !props.disabled &&
    props.loaded &&
    !props.error &&
    !props.loading &&
    !props.options.some((option) => option.value === model.value),
);
</script>

<template>
  <div class="node-dns px-4 py-3">
    <div class="d-flex align-center ga-2">
      <v-select
        v-model="model"
        class="node-dns__select"
        :label="t('nodeDns.title')"
        :items="items"
        :loading="loading"
        :disabled="disabled || loading || !!error"
        :error-messages="unavailable ? t('nodeDns.unavailable') : []"
        :rules="[() => !unavailable || t('nodeDns.unavailable')]"
        hide-details="auto"
        dir="ltr"
      >
        <template #selection>
          <v-tooltip :text="selection.url" max-width="360" open-on-click>
            <template #activator="{ props: tip }">
              <span v-bind="tip" class="node-dns__value">
                <bdi dir="ltr">{{ selection.address }}</bdi
                >{{ " " }}
                <span v-if="selection.categoryText"
                  >(<bdi dir="auto">{{ selection.categoryText }}</bdi
                  >)</span
                >
              </span>
            </template>
          </v-tooltip>
        </template>
        <template #item="{ item, props: entry }">
          <v-list-item v-bind="entry" :title="undefined">
            <span class="node-dns__value">
              <bdi dir="ltr">{{ item.address }}</bdi>
              (<bdi dir="auto">{{ item.categoryText }}</bdi
              >)
            </span>
          </v-list-item>
        </template>
      </v-select>
      <v-btn
        :icon="mdiRefresh"
        :aria-label="t('nodeDns.refresh')"
        :loading="loading"
        :disabled="disabled"
        variant="text"
        @click="$emit('refresh')"
      />
    </div>
    <v-alert v-if="error" type="error" variant="tonal" class="mt-2">
      {{ t("nodeDns.loadFailed", { message: errorText(error) }) }}
      <v-btn variant="text" :disabled="disabled" @click="$emit('refresh')">
        {{ t("nodeDns.refresh") }}
      </v-btn>
    </v-alert>
    <v-alert
      v-if="warnings.includes('NODE_DNS_SYSTEM_READ_FAILED')"
      type="warning"
      variant="tonal"
      class="mt-2"
    >
      {{ t("nodeDns.systemReadFailed") }}
    </v-alert>
  </div>
</template>

<style scoped>
.node-dns__select {
  min-width: 0;
}
.node-dns__value {
  white-space: normal;
  overflow-wrap: anywhere;
}
.node-dns :deep(.v-select__selection) {
  max-width: 100%;
}
</style>
