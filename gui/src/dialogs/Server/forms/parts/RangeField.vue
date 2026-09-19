<script setup lang="ts">
// A from–to pair of numbers under one label, the way the xhttp options
// are written ("scMaxEachPostBytes 1000-2000").
import { useI18n } from "vue-i18n";

defineProps<{ label: string; readonly?: boolean }>();
const from = defineModel<string>("from", { required: true });
const to = defineModel<string>("to", { required: true });
const { t } = useI18n();
</script>

<template>
  <div class="range">
    <span class="md3-body-small range__label">{{ label }}</span>
    <div class="d-flex ga-2">
      <v-text-field
        v-model="from"
        type="number"
        :label="t('configureServer.rangeFrom')"
        :readonly="readonly"
        hide-details="auto"
      />
      <v-text-field
        v-model="to"
        type="number"
        :label="t('configureServer.rangeTo')"
        :readonly="readonly"
        hide-details="auto"
      />
    </div>
  </div>
</template>

<style scoped>
.range__label {
  display: block;
  color: rgb(var(--v-theme-on-surface-variant));
  margin-bottom: 4px;
}
</style>
