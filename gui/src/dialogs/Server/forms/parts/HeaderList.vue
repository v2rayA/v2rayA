<script setup lang="ts">
// The xhttp custom headers: name–value rows, add and remove.
import { useI18n } from "vue-i18n";
import { mdiDeleteOutline, mdiPlus } from "@mdi/js";

defineProps<{ readonly?: boolean }>();
const headers = defineModel<{ key: string; value: string }[]>({
  required: true,
});
const { t } = useI18n();
</script>

<template>
  <div>
    <span class="md3-body-small headers__label">{{
      t("configureServer.customHeaders")
    }}</span>
    <div
      v-for="(header, i) in headers"
      :key="i"
      class="d-flex ga-2 align-center mb-2"
    >
      <v-text-field
        v-model="header.key"
        :label="t('configureServer.headerName')"
        :readonly="readonly"
        hide-details="auto"
      />
      <v-text-field
        v-model="header.value"
        :label="t('configureServer.headerValue')"
        :readonly="readonly"
        hide-details="auto"
      />
      <v-btn
        v-if="!readonly"
        :icon="mdiDeleteOutline"
        variant="text"
        size="small"
        :aria-label="t('operations.delete')"
        @click="headers.splice(i, 1)"
      />
    </div>
    <v-btn
      v-if="!readonly"
      variant="tonal"
      size="small"
      :prepend-icon="mdiPlus"
      @click="headers.push({ key: '', value: '' })"
    >
      {{ t("configureServer.addHeader") }}
    </v-btn>
  </div>
</template>

<style scoped>
.headers__label {
  display: block;
  color: rgb(var(--v-theme-on-surface-variant));
  margin-bottom: 4px;
}
</style>
