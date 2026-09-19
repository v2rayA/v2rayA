<script setup lang="ts">
// The traffic plot: the download rate as a filled area and the upload
// rate as a line, both smoothed through the midpoints of the samples and
// sharing one scale that uses the top 70 % of the plot. Drawn as one SVG
// so the path can transition when the next sample shifts everything left.
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const props = defineProps<{ down: number[]; up: number[] }>();
const { t } = useI18n();
const width = 100;
const height = 40;
const plot = height * 0.7;

const max = computed(() => Math.max(1, ...props.down, ...props.up));

/** points maps the samples to plot coordinates, oldest at the left */
function points(series: number[]): [number, number][] {
  const step = width / Math.max(1, series.length - 1);
  return series.map((v, i) => [i * step, height - (v / max.value) * plot]);
}

/** curve joins the points with quadratic curves through their midpoints, as a speed graph reads best */
function curve(series: number[]): string {
  const p = points(series);
  if (p.length < 2) return "";
  let d = `M${p[0][0]},${p[0][1]}`;
  for (let i = 1; i < p.length - 1; i++) {
    const [cx, cy] = p[i];
    const mx = (cx + p[i + 1][0]) / 2;
    const my = (cy + p[i + 1][1]) / 2;
    d += ` Q${cx},${cy} ${mx},${my}`;
  }
  const [lx, ly] = p[p.length - 1];
  return `${d} L${lx},${ly}`;
}
const downLine = computed(() => curve(props.down));
const downArea = computed(() =>
  downLine.value ? `${downLine.value} L${width},${height} L0,${height} Z` : "",
);
const upLine = computed(() => curve(props.up));
</script>

<template>
  <svg
    :viewBox="`0 0 ${width} ${height}`"
    preserveAspectRatio="none"
    class="traffic-chart"
    role="img"
    :aria-label="`${t('traffic.download')} / ${t('traffic.upload')}`"
  >
    <defs>
      <linearGradient id="traffic-down-fill" x1="0" y1="0" x2="0" y2="1">
        <stop
          offset="0"
          stop-color="rgb(var(--v-theme-primary))"
          stop-opacity="0.38"
        />
        <stop
          offset="1"
          stop-color="rgb(var(--v-theme-primary))"
          stop-opacity="0.08"
        />
      </linearGradient>
    </defs>
    <path
      :d="downArea"
      fill="url(#traffic-down-fill)"
      class="traffic-chart__path"
    />
    <path
      :d="downLine"
      fill="none"
      stroke="rgb(var(--v-theme-primary))"
      stroke-width="2"
      vector-effect="non-scaling-stroke"
      stroke-linejoin="round"
      class="traffic-chart__path traffic-chart__down"
    />
    <path
      :d="upLine"
      fill="none"
      stroke="rgb(var(--v-theme-tertiary))"
      stroke-width="1.5"
      vector-effect="non-scaling-stroke"
      stroke-linejoin="round"
      class="traffic-chart__path traffic-chart__up"
    />
  </svg>
</template>

<style scoped>
.traffic-chart {
  display: block;
  width: 100%;
  height: 100%;
  overflow: visible;
}
/* the next sample slides the curves left where the browser can animate d */
.traffic-chart__path {
  transition: d 300ms cubic-bezier(0.2, 0, 0, 1);
}
</style>
