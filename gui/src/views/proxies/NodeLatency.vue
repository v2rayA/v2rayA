<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
defineOptions({ name: "NodeLatency" });
const props = defineProps<{ latency: string }>();
const { t } = useI18n();
const testing = computed(() => props.latency === t("latency.testing"));
const color = computed(() => {
  if (props.latency === "TIMEOUT") return "text-error";
  const value = parseFloat(props.latency);
  if (!Number.isFinite(value)) return "text-on-surface-variant";
  return value < 200
    ? "text-primary"
    : value < 500
      ? "text-tertiary"
      : "text-error";
});
</script>
<template>
  <span class="node-latency md3-label-large" :class="color" dir="ltr">
    <v-progress-circular
      v-if="testing"
      indeterminate
      size="16"
      width="2"
      :aria-label="t('latency.testing')"
    />
    <template v-else>{{ latency || "—" }}</template>
  </span>
</template>
<style scoped>
.node-latency {
  flex-shrink: 0;
  white-space: nowrap;
}
</style>
