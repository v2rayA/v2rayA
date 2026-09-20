<script setup lang="ts">
// The logo in a frame that changes shape on every tap: the shapes of the
// M3 expressive set, drawn as polygons so the browser morphs one into the
// next. At rest there is no frame; the seventh tap is back to none. Every
// sixth tap, the frame gone again, mirrors the layout, the way the RTL
// locales see it.
import { computed, ref } from "vue";
import { useLocale } from "vuetify";
import logo from "@/assets/img/v2raya-icon.svg";

defineProps<{ size: number }>();

const shapes: ((a: number) => number)[] = [
  () => 1,
  (a) => 1 - 0.08 * Math.cos(9 * a),
  (a) => 0.8 + 0.2 * Math.cos(4 * a),
  (a) => 0.86 + 0.14 * Math.cos(8 * a),
  (a) =>
    Math.cos(Math.PI / 5) /
    Math.cos(((a + Math.PI / 2) % (Math.PI / 2.5)) - Math.PI / 5),
  (a) => (Math.abs(Math.cos(a)) ** 4 + Math.abs(Math.sin(a)) ** 4) ** -0.25,
];
const polygon = (r: (a: number) => number) =>
  "polygon(" +
  Array.from({ length: 90 }, (_, i) => {
    const a = (i / 90) * 2 * Math.PI;
    return `${50 + 49 * r(a) * Math.cos(a)}% ${50 + 49 * r(a) * Math.sin(a)}%`;
  }).join(", ") +
  ")";
const index = ref(0);
const clip = computed(() => polygon(shapes[index.value % shapes.length]));
const { current, rtl, isRtl } = useLocale();
function tap() {
  index.value++;
  if (index.value % shapes.length) return;
  rtl.value = { ...rtl.value, [current.value]: !isRtl.value };
  document.documentElement.dir = isRtl.value ? "rtl" : "ltr";
}
</script>

<template>
  <button
    type="button"
    class="brand-shape"
    :class="{ 'brand-shape--on': index % shapes.length }"
    :style="{
      width: `${size}px`,
      height: `${size}px`,
      clipPath: clip,
      transform: `rotate(${index * 60}deg)`,
    }"
    tabindex="-1"
    aria-hidden="true"
    @click="tap"
  >
    <img
      :src="logo"
      alt=""
      :style="{ transform: `rotate(${-index * 60}deg)` }"
    />
  </button>
</template>

<style scoped>
.brand-shape {
  display: grid;
  place-items: center;
  border: none;
  padding: 0;
  background: transparent;
  cursor: default;
  transition:
    clip-path 0.5s cubic-bezier(0.2, 0, 0, 1),
    transform 0.5s cubic-bezier(0.2, 0, 0, 1),
    background-color 0.3s;
}
.brand-shape--on {
  background: rgb(var(--v-theme-primary-container));
}
.brand-shape img {
  width: 72%;
  height: 72%;
  transition: transform 0.5s cubic-bezier(0.2, 0, 0, 1);
}
</style>
