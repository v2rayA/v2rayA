<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { formatBytes, formatRate } from "@/lib/format";
import TrafficChart from "./TrafficChart.vue";

const props = defineProps<{
  up: number;
  down: number;
  upTotal: number;
  downTotal: number;
  upSeries: number[];
  downSeries: number[];
}>();
const { t } = useI18n();
</script>

<template>
  <v-card color="surface-container-high" rounded="xl" class="pa-5 traffic">
    <div class="traffic__rates">
      <div>
        <p class="md3-label-medium text-on-surface-variant ma-0">
          {{ t("traffic.download") }}
        </p>
        <p class="md3-headline-small traffic__value ma-0" dir="ltr">
          {{ formatRate(props.down) }}
        </p>
        <p class="md3-body-small text-on-surface-variant ma-0" dir="ltr">
          {{ t("traffic.total", { value: formatBytes(props.downTotal) }) }}
        </p>
      </div>
      <div>
        <p class="md3-label-medium text-on-surface-variant ma-0">
          {{ t("traffic.upload") }}
        </p>
        <p class="md3-headline-small traffic__value ma-0" dir="ltr">
          {{ formatRate(props.up) }}
        </p>
        <p class="md3-body-small text-on-surface-variant ma-0" dir="ltr">
          {{ t("traffic.total", { value: formatBytes(props.upTotal) }) }}
        </p>
      </div>
    </div>
    <div class="traffic__chart mt-4">
      <TrafficChart :down="props.downSeries" :up="props.upSeries" />
    </div>
  </v-card>
</template>

<style scoped>
.traffic__rates {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}
.traffic__value {
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}
.traffic__chart {
  height: 96px;
}
</style>
